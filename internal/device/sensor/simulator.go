package sensor

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/logger"
)

type Simulator struct {
	sensors   map[string]*simulatedSensor
	connected bool
	mu        sync.RWMutex
	rand      *rand.Rand
}

type simulatedSensor struct {
	info      *model.SensorInfo
	baseValue float64
	variation float64
	trend     float64
	lastValue float64
}

func NewSimulator() *Simulator {
	return &Simulator{
		sensors: make(map[string]*simulatedSensor),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Simulator) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	time.Sleep(20 * time.Millisecond)

	s.connected = true

	s.initDefaultSensors()

	logger.Info("Sensor simulator connected")
	return nil
}

func (s *Simulator) initDefaultSensors() {
	defaultSensors := []struct {
		id         string
		sensorType model.SensorType
		location   string
		unit       string
		baseValue  float64
		variation  float64
	}{
		{"temp-001", model.SensorTypeTemperature, "Room-A", "°C", 25.0, 5.0},
		{"temp-002", model.SensorTypeTemperature, "Room-B", "°C", 26.0, 4.0},
		{"hum-001", model.SensorTypeHumidity, "Room-A", "%", 60.0, 10.0},
		{"hum-002", model.SensorTypeHumidity, "Room-B", "%", 55.0, 8.0},
		{"press-001", model.SensorTypePressure, "Room-A", "kPa", 101.3, 2.0},
		{"volt-001", model.SensorTypeVoltage, "Power-Unit-1", "V", 220.0, 5.0},
		{"curr-001", model.SensorTypeCurrent, "Power-Unit-1", "A", 10.0, 2.0},
		{"power-001", model.SensorTypePower, "Power-Unit-1", "kW", 2.2, 0.5},
	}

	for _, ds := range defaultSensors {
		info := &model.SensorInfo{
			SensorID:   ds.id,
			SensorType: ds.sensorType,
			Location:   ds.location,
			Unit:       ds.unit,
			MinValue:   ds.baseValue - ds.variation*2,
			MaxValue:   ds.baseValue + ds.variation*2,
			Status:     1,
		}

		s.sensors[ds.id] = &simulatedSensor{
			info:      info,
			baseValue: ds.baseValue,
			variation: ds.variation,
			trend:     (s.rand.Float64() - 0.5) * 0.1,
			lastValue: ds.baseValue,
		}
	}
}

func (s *Simulator) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = false
	logger.Info("Sensor simulator disconnected")
	return nil
}

func (s *Simulator) Read(ctx context.Context, sensorID string) (*model.SensorData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		return nil, ErrSimulatorNotConnected
	}

	sensor, ok := s.sensors[sensorID]
	if !ok {
		return nil, ErrSensorNotFound
	}

	value := s.generateValue(sensor)
	sensor.lastValue = value

	quality := 0.9 + s.rand.Float64()*0.1

	return &model.SensorData{
		SensorID:   sensorID,
		SensorType: string(sensor.info.SensorType),
		Location:   sensor.info.Location,
		Value:      value,
		Unit:       sensor.info.Unit,
		Quality:    quality,
		Timestamp:  time.Now(),
	}, nil
}

func (s *Simulator) ReadAll(ctx context.Context) ([]*model.SensorData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		return nil, ErrSimulatorNotConnected
	}

	data := make([]*model.SensorData, 0, len(s.sensors))
	for id, sensor := range s.sensors {
		value := s.generateValue(sensor)
		sensor.lastValue = value

		quality := 0.9 + s.rand.Float64()*0.1

		data = append(data, &model.SensorData{
			SensorID:   id,
			SensorType: string(sensor.info.SensorType),
			Location:   sensor.info.Location,
			Value:      value,
			Unit:       sensor.info.Unit,
			Quality:    quality,
			Timestamp:  time.Now(),
		})
	}

	return data, nil
}

func (s *Simulator) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

func (s *Simulator) generateValue(sensor *simulatedSensor) float64 {
	noise := (s.rand.Float64()*2 - 1) * sensor.variation * 0.3
	trendStep := sensor.trend * s.rand.Float64()

	value := sensor.lastValue + trendStep + noise

	if math.Abs(value-sensor.baseValue) > sensor.variation {
		sensor.trend = -sensor.trend
		value = sensor.baseValue + (s.rand.Float64()*2-1)*sensor.variation*0.5
	}

	value = math.Max(sensor.info.MinValue, math.Min(sensor.info.MaxValue, value))

	return math.Round(value*100) / 100
}

func (s *Simulator) AddSensor(info *model.SensorInfo, baseValue, variation float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sensors[info.SensorID] = &simulatedSensor{
		info:      info,
		baseValue: baseValue,
		variation: variation,
		trend:     (s.rand.Float64() - 0.5) * 0.1,
		lastValue: baseValue,
	}
}

func (s *Simulator) RemoveSensor(sensorID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sensors, sensorID)
}

func (s *Simulator) GetAllSensorInfo() []*model.SensorInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	infos := make([]*model.SensorInfo, 0, len(s.sensors))
	for _, sensor := range s.sensors {
		infos = append(infos, sensor.info)
	}
	return infos
}

var (
	ErrSimulatorNotConnected = &SimulatorError{Message: "simulator not connected"}
)

type SimulatorError struct {
	Message string
}

func (e *SimulatorError) Error() string {
	return e.Message
}
