package irs

import (
	"context"
	"sync"
	"time"

	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/errors"
)

type Driver interface {
	Connect(ctx context.Context) error
	Disconnect() error
	SetPhaseShifts(ctx context.Context, phaseShifts []float64) error
	GetStatus(ctx context.Context) (*model.IRSStatus, error)
	IsConnected() bool
}

type Controller struct {
	driver         Driver
	config         *model.IRSConfig
	status         *model.IRSStatus
	mu             sync.RWMutex
	onStatusChange func(status *model.IRSStatus)
}

func NewController(driver Driver) *Controller {
	return &Controller{
		driver: driver,
	}
}

func (c *Controller) Connect(ctx context.Context) error {
	return c.driver.Connect(ctx)
}

func (c *Controller) Disconnect() error {
	return c.driver.Disconnect()
}

func (c *Controller) Configure(ctx context.Context, config *model.IRSConfigRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := config.Validate(); err != nil {
		return errors.Wrap(errors.CodeInvalidIRSConfig, "invalid IRS configuration", err)
	}

	if !c.driver.IsConnected() {
		if err := c.driver.Connect(ctx); err != nil {
			return errors.Wrap(errors.CodeIRSDeviceError, "failed to connect IRS device", err)
		}
	}

	if err := c.driver.SetPhaseShifts(ctx, config.PhaseShifts); err != nil {
		return errors.Wrap(errors.CodeIRSConfigFailed, "failed to set phase shifts", err)
	}

	c.config = &model.IRSConfig{
		Name:          config.Name,
		ElementCount:  config.ElementCount,
		PhaseShifts:   config.PhaseShifts,
		FrequencyBand: config.FrequencyBand,
		Status:        model.ConfigStatusApplied,
	}

	status, err := c.driver.GetStatus(ctx)
	if err != nil {
		return errors.Wrap(errors.CodeIRSStatusError, "failed to get IRS status", err)
	}
	c.status = status

	return nil
}

func (c *Controller) GetStatus(ctx context.Context) (*model.IRSStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.driver.IsConnected() {
		return nil, errors.New(errors.CodeIRSDeviceError, "IRS device not connected")
	}

	status, err := c.driver.GetStatus(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.CodeIRSStatusError, "failed to get IRS status", err)
	}

	c.status = status
	return status, nil
}

func (c *Controller) GetCurrentConfig() *model.IRSConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

func (c *Controller) SetStatusChangeCallback(fn func(status *model.IRSStatus)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onStatusChange = fn
}

func (c *Controller) StartMonitoring(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := c.driver.GetStatus(ctx)
			if err != nil {
				continue
			}
			c.mu.Lock()
			c.status = status
			callback := c.onStatusChange
			c.mu.Unlock()

			if callback != nil {
				callback(status)
			}
		}
	}
}
