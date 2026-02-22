package usrp

import (
	"context"
	"sync"
	"time"

	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/logger"

	"go.uber.org/zap"
)

type Driver interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Receive(ctx context.Context, duration time.Duration) ([]model.ChannelDataPoint, error)
	SetFrequency(freq float64) error
	SetSampleRate(rate float64) error
	IsConnected() bool
}

type Receiver struct {
	driver     Driver
	sampleRate float64
	centerFreq float64
	connected  bool
	mu         sync.RWMutex
	dataBuffer []model.ChannelDataPoint
	bufferSize int
}

func NewReceiver(driver Driver, sampleRate, centerFreq float64) *Receiver {
	return &Receiver{
		driver:     driver,
		sampleRate: sampleRate,
		centerFreq: centerFreq,
		bufferSize: 1024 * 1024,
	}
}

func (r *Receiver) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.driver.Connect(ctx); err != nil {
		return err
	}

	r.connected = true
	return nil
}

func (r *Receiver) Disconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connected = false
	return r.driver.Disconnect()
}

func (r *Receiver) CollectData(ctx context.Context, duration time.Duration) ([]model.ChannelDataPoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.connected {
		return nil, ErrReceiverNotConnected
	}

	data, err := r.driver.Receive(ctx, duration)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *Receiver) SetCenterFrequency(freq float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.connected {
		return ErrReceiverNotConnected
	}

	if err := r.driver.SetFrequency(freq); err != nil {
		return err
	}

	r.centerFreq = freq
	logger.Info("USRP center frequency updated", zap.Float64("freq_hz", freq))
	return nil
}

func (r *Receiver) SetSampleRate(rate float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.connected {
		return ErrReceiverNotConnected
	}

	if err := r.driver.SetSampleRate(rate); err != nil {
		return err
	}

	r.sampleRate = rate
	logger.Info("USRP sample rate updated", zap.Float64("rate_sps", rate))
	return nil
}

func (r *Receiver) GetConfig() (sampleRate, centerFreq float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sampleRate, r.centerFreq
}

var ErrReceiverNotConnected = &ReceiverError{Message: "receiver not connected"}

type ReceiverError struct {
	Message string
}

func (e *ReceiverError) Error() string {
	return e.Message
}
