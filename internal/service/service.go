package service

import (
	"context"
	"encoding/json"
	"time"

	"isac-cran-system/internal/algorithm/beamforming"
	"isac-cran-system/internal/algorithm/doa"
	"isac-cran-system/internal/device/irs"
	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/errors"
)

type IRSService struct {
	controller *irs.Controller
}

func NewIRSService(controller *irs.Controller) *IRSService {
	return &IRSService{controller: controller}
}

func (s *IRSService) Configure(ctx context.Context, req *model.IRSConfigRequest) (*model.IRSConfig, error) {
	if err := s.controller.Configure(ctx, req); err != nil {
		return nil, err
	}

	config := s.controller.GetCurrentConfig()
	return config, nil
}

func (s *IRSService) GetStatus(ctx context.Context) (*model.IRSStatus, error) {
	return s.controller.GetStatus(ctx)
}

func (s *IRSService) GetCurrentConfig() *model.IRSConfig {
	return s.controller.GetCurrentConfig()
}

func (s *IRSService) ApplyOptimalPhaseShifts(ctx context.Context, targetAngle float64) (*model.IRSConfig, error) {
	config := s.controller.GetCurrentConfig()
	if config == nil {
		return nil, errors.New(errors.CodeIRSDeviceError, "no active IRS configuration")
	}

	optimizer := beamforming.NewWeightsCalculator(config.ElementCount, 0.5)
	weights := optimizer.ComputeConjugateBeamforming(targetAngle)
	phaseShifts := optimizer.ComputePhaseShifts(weights)

	req := &model.IRSConfigRequest{
		Name:          "optimal_" + time.Now().Format("20060102150405"),
		ElementCount:  config.ElementCount,
		PhaseShifts:   phaseShifts,
		FrequencyBand: config.FrequencyBand,
	}

	if err := s.controller.Configure(ctx, req); err != nil {
		return nil, err
	}

	return s.controller.GetCurrentConfig(), nil
}

type ChannelService struct {
	receiver  ChannelReceiver
	dataStore ChannelDataStore
}

type ChannelReceiver interface {
	CollectData(ctx context.Context, duration time.Duration) ([]model.ChannelDataPoint, error)
	GetConfig() (sampleRate, centerFreq float64)
}

type ChannelDataStore interface {
	Write(ctx context.Context, data *model.ChannelMeasurement) error
	Query(ctx context.Context, q *model.ChannelDataQuery) ([]*model.ChannelMeasurement, error)
}

func NewChannelService(receiver ChannelReceiver, store ChannelDataStore) *ChannelService {
	return &ChannelService{
		receiver:  receiver,
		dataStore: store,
	}
}

func (s *ChannelService) CollectData(ctx context.Context, req *model.ChannelCollectRequest) (*model.ChannelMeasurement, error) {
	duration := time.Duration(req.Duration * float64(time.Second))

	dataPoints, err := s.receiver.CollectData(ctx, duration)
	if err != nil {
		return nil, errors.Wrap(errors.CodeUSRPReceiveError, "failed to collect channel data", err)
	}

	amplitudes := make([]float64, len(dataPoints))
	phases := make([]float64, len(dataPoints))
	for i, dp := range dataPoints {
		amplitudes[i] = dp.Amplitude
		phases[i] = dp.Phase
	}

	snr := s.calculateSNR(amplitudes)

	measurement := &model.ChannelMeasurement{
		MeasurementID: generateMeasurementID(),
		ExperimentID:  req.ExperimentID,
		UserID:        req.UserID,
		FrequencyBand: req.FrequencyBand,
		Amplitude:     amplitudes,
		Phase:         phases,
		SNR:           snr,
		Timestamp:     time.Now(),
	}

	if s.dataStore != nil {
		if err := s.dataStore.Write(ctx, measurement); err != nil {
			return nil, err
		}
	}

	return measurement, nil
}

func (s *ChannelService) QueryData(ctx context.Context, q *model.ChannelDataQuery) ([]*model.ChannelMeasurement, int64, error) {
	if s.dataStore == nil {
		return []*model.ChannelMeasurement{}, 0, nil
	}

	data, err := s.dataStore.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	return data, int64(len(data)), nil
}

func (s *ChannelService) GetRealtimeData(ctx context.Context) ([]model.ChannelDataPoint, error) {
	data, err := s.receiver.CollectData(ctx, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *ChannelService) calculateSNR(amplitudes []float64) float64 {
	if len(amplitudes) == 0 {
		return 0
	}

	var sum, sumSq float64
	for _, a := range amplitudes {
		sum += a
		sumSq += a * a
	}

	mean := sum / float64(len(amplitudes))
	variance := sumSq/float64(len(amplitudes)) - mean*mean

	if variance <= 0 {
		return 0
	}

	return 10 * (mean * mean / variance)
}

func generateMeasurementID() string {
	return "meas_" + time.Now().Format("20060102150405")
}

type AlgorithmService struct {
	beamformingOptimizer *beamforming.Optimizer
	doaEstimator         *doa.Estimator
	resultStore          AlgorithmResultStore
}

type AlgorithmResultStore interface {
	Create(ctx context.Context, result *model.ExperimentResult) error
	GetByExperimentID(ctx context.Context, experimentID string) (*model.ExperimentResult, error)
	UpdateStatus(ctx context.Context, id int64, status model.ExperimentStatus, resultData string) error
	List(ctx context.Context, algorithmType model.AlgorithmType, page, pageSize int) ([]model.ExperimentResult, int64, error)
}

func NewAlgorithmService(store AlgorithmResultStore) *AlgorithmService {
	return &AlgorithmService{
		beamformingOptimizer: beamforming.NewOptimizer(64, 100, 0.001),
		doaEstimator:         doa.NewEstimator(64, 3, 1024, "MUSIC"),
		resultStore:          store,
	}
}

func (s *AlgorithmService) RunBeamforming(ctx context.Context, experimentID string, params *model.BeamformingParams) (*model.BeamformingResult, error) {
	result := &model.ExperimentResult{
		ExperimentID:  experimentID,
		AlgorithmType: model.AlgorithmTypeBeamforming,
		Status:        model.ExperimentStatusRunning,
	}

	paramsJSON, _ := json.Marshal(params)
	result.Parameters = string(paramsJSON)

	if s.resultStore != nil {
		if err := s.resultStore.Create(ctx, result); err != nil {
			return nil, err
		}
	}

	bfResult, err := s.beamformingOptimizer.Optimize(params)
	if err != nil {
		if s.resultStore != nil {
			s.resultStore.UpdateStatus(ctx, result.ID, model.ExperimentStatusFailed, "")
		}
		return nil, errors.Wrap(errors.CodeAlgorithmRunError, "beamforming optimization failed", err)
	}

	resultJSON, _ := json.Marshal(bfResult)
	if s.resultStore != nil {
		s.resultStore.UpdateStatus(ctx, result.ID, model.ExperimentStatusCompleted, string(resultJSON))
	}

	return bfResult, nil
}

func (s *AlgorithmService) RunDOA(ctx context.Context, experimentID string, params *model.DOAParams) (*model.DOAResult, error) {
	result := &model.ExperimentResult{
		ExperimentID:  experimentID,
		AlgorithmType: model.AlgorithmTypeDOA,
		Status:        model.ExperimentStatusRunning,
	}

	paramsJSON, _ := json.Marshal(params)
	result.Parameters = string(paramsJSON)

	if s.resultStore != nil {
		if err := s.resultStore.Create(ctx, result); err != nil {
			return nil, err
		}
	}

	data := generateTestSignal(params.SnapshotLength)
	doaResult, err := s.doaEstimator.Estimate(data, params)
	if err != nil {
		if s.resultStore != nil {
			s.resultStore.UpdateStatus(ctx, result.ID, model.ExperimentStatusFailed, "")
		}
		return nil, errors.Wrap(errors.CodeAlgorithmRunError, "DOA estimation failed", err)
	}

	resultJSON, _ := json.Marshal(doaResult)
	if s.resultStore != nil {
		s.resultStore.UpdateStatus(ctx, result.ID, model.ExperimentStatusCompleted, string(resultJSON))
	}

	return doaResult, nil
}

func (s *AlgorithmService) GetResult(ctx context.Context, experimentID string) (*model.ExperimentResult, error) {
	if s.resultStore == nil {
		return nil, errors.New(errors.CodeNotFound, "result store not available")
	}

	return s.resultStore.GetByExperimentID(ctx, experimentID)
}

func (s *AlgorithmService) ListResults(ctx context.Context, algorithmType model.AlgorithmType, page, pageSize int) ([]model.ExperimentResult, int64, error) {
	if s.resultStore == nil {
		return []model.ExperimentResult{}, 0, nil
	}

	return s.resultStore.List(ctx, algorithmType, page, pageSize)
}

func generateTestSignal(length int) []complex128 {
	data := make([]complex128, length)
	for i := 0; i < length; i++ {
		t := float64(i) / float64(length)
		data[i] = complex(0.5+0.5*cos(2*3.14159265359*t), 0.5+0.5*sin(2*3.14159265359*t))
	}
	return data
}

func cos(x float64) float64 {
	return 1 - x*x/2
}

func sin(x float64) float64 {
	return x
}
