package sensor

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
	Read(ctx context.Context, sensorID string) (*model.SensorData, error)
	ReadAll(ctx context.Context) ([]*model.SensorData, error)
	IsConnected() bool
}

type Collector struct {
	driver             Driver
	sensors            map[string]*model.SensorInfo
	mu                 sync.RWMutex
	collectionInterval time.Duration
	stopChan           chan struct{}
	running            bool
	onDataReceived     func(data *model.SensorData)
}

func NewCollector(driver Driver, interval time.Duration) *Collector {
	return &Collector{
		driver:             driver,
		sensors:            make(map[string]*model.SensorInfo),
		collectionInterval: interval,
		stopChan:           make(chan struct{}),
	}
}

func (c *Collector) Connect(ctx context.Context) error {
	if err := c.driver.Connect(ctx); err != nil {
		return err
	}

	if simulator, ok := c.driver.(*Simulator); ok {
		for _, info := range simulator.GetAllSensorInfo() {
			c.sensors[info.SensorID] = info
		}
	}

	return nil
}

func (c *Collector) Disconnect() error {
	c.StopCollection()
	return c.driver.Disconnect()
}

func (c *Collector) RegisterSensor(info *model.SensorInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sensors[info.SensorID] = info
	logger.Info("Sensor registered", zap.String("sensor_id", info.SensorID))
}

func (c *Collector) UnregisterSensor(sensorID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sensors, sensorID)
	logger.Info("Sensor unregistered", zap.String("sensor_id", sensorID))
}

func (c *Collector) GetSensor(sensorID string) (*model.SensorInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	info, ok := c.sensors[sensorID]
	return info, ok
}

func (c *Collector) GetAllSensors() []*model.SensorInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sensors := make([]*model.SensorInfo, 0, len(c.sensors))
	for _, info := range c.sensors {
		sensors = append(sensors, info)
	}
	return sensors
}

func (c *Collector) ReadSensor(ctx context.Context, sensorID string) (*model.SensorData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if _, ok := c.sensors[sensorID]; !ok {
		return nil, ErrSensorNotFound
	}

	data, err := c.driver.Read(ctx, sensorID)
	if err != nil {
		return nil, err
	}

	if c.onDataReceived != nil {
		c.onDataReceived(data)
	}

	return data, nil
}

func (c *Collector) ReadAllSensors(ctx context.Context) ([]*model.SensorData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.sensors) == 0 {
		return []*model.SensorData{}, nil
	}

	data, err := c.driver.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	if c.onDataReceived != nil {
		for _, d := range data {
			c.onDataReceived(d)
		}
	}

	return data, nil
}

func (c *Collector) StartCollection(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return ErrAlreadyRunning
	}
	c.running = true
	c.mu.Unlock()

	ticker := time.NewTicker(c.collectionInterval)
	defer ticker.Stop()

	logger.Info("Sensor data collection started",
		zap.Duration("interval", c.collectionInterval),
	)

	for {
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.running = false
			c.mu.Unlock()
			return ctx.Err()
		case <-c.stopChan:
			c.mu.Lock()
			c.running = false
			c.mu.Unlock()
			return nil
		case <-ticker.C:
			_, err := c.ReadAllSensors(ctx)
			if err != nil {
				logger.Warn("Failed to collect sensor data", zap.Error(err))
			}
		}
	}
}

func (c *Collector) StopCollection() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.running {
		close(c.stopChan)
		c.stopChan = make(chan struct{})
	}
}

func (c *Collector) SetDataCallback(fn func(data *model.SensorData)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onDataReceived = fn
}

func (c *Collector) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

var (
	ErrSensorNotFound = &CollectorError{Message: "sensor not found"}
	ErrAlreadyRunning = &CollectorError{Message: "collection already running"}
)

type CollectorError struct {
	Message string
}

func (e *CollectorError) Error() string {
	return e.Message
}
