package service

import (
	"context"
	"sync"

	"isac-cran-system/internal/device/sensor"
	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/errors"
)

type SensorService struct {
	collector *sensor.Collector
	dataStore SensorDataStore
	mu        sync.RWMutex
	running   bool
}

type SensorDataStore interface {
	Write(ctx context.Context, data *model.SensorData) error
	WriteBatch(ctx context.Context, dataPoints []*model.SensorData) error
	Query(ctx context.Context, q *model.SensorDataQuery) ([]*model.SensorData, error)
}

func NewSensorService(collector *sensor.Collector, store SensorDataStore) *SensorService {
	return &SensorService{
		collector: collector,
		dataStore: store,
	}
}

func (s *SensorService) ListSensors(ctx context.Context, sensorType model.SensorType) ([]*model.SensorInfo, error) {
	sensors := s.collector.GetAllSensors()

	if sensorType != "" {
		filtered := make([]*model.SensorInfo, 0)
		for _, info := range sensors {
			if info.SensorType == sensorType {
				filtered = append(filtered, info)
			}
		}
		return filtered, nil
	}

	return sensors, nil
}

func (s *SensorService) GetSensorData(ctx context.Context, q *model.SensorDataQuery) ([]*model.SensorData, error) {
	if s.dataStore == nil {
		return []*model.SensorData{}, nil
	}

	return s.dataStore.Query(ctx, q)
}

func (s *SensorService) ReadSensor(ctx context.Context, sensorID string) (*model.SensorData, error) {
	data, err := s.collector.ReadSensor(ctx, sensorID)
	if err != nil {
		return nil, errors.Wrap(errors.CodeSensorDataError, "failed to read sensor", err)
	}

	if s.dataStore != nil {
		if err := s.dataStore.Write(ctx, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (s *SensorService) StartCollection(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New(errors.CodeExperimentRunning, "collection already running")
	}

	s.collector.SetDataCallback(func(data *model.SensorData) {
		if s.dataStore != nil {
			s.dataStore.Write(context.Background(), data)
		}
	})

	go func() {
		s.collector.StartCollection(ctx)
	}()

	s.running = true
	return nil
}

func (s *SensorService) StopCollection() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.collector.StopCollection()
	s.running = false
}

func (s *SensorService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

type MATLABService struct {
	dataDir string
}

func NewMATLABService(dataDir string) *MATLABService {
	return &MATLABService{dataDir: dataDir}
}

func (s *MATLABService) ExportToJSON(data interface{}, filename string) error {
	return nil
}

func (s *MATLABService) ExportToCSV(data interface{}, filename string) error {
	return nil
}

func (s *MATLABService) ImportFromJSON(filename string, target interface{}) error {
	return nil
}
