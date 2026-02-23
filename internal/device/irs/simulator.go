package irs

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/logger"

	"go.uber.org/zap"
)

type Simulator struct {
	elementCount  int
	frequencyBand string
	phaseShifts   []float64
	connected     bool
	mu            sync.RWMutex
	rand          *rand.Rand
}

func NewSimulator(elementCount int, frequencyBand string) *Simulator {
	return &Simulator{
		elementCount:  elementCount,
		frequencyBand: frequencyBand,
		phaseShifts:   make([]float64, elementCount),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
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

	time.Sleep(50 * time.Millisecond)

	s.connected = true
	logger.Info("IRS simulator connected", zap.Int("element_count", s.elementCount))
	return nil
}

func (s *Simulator) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = false
	logger.Info("IRS simulator disconnected")
	return nil
}

func (s *Simulator) SetPhaseShifts(ctx context.Context, phaseShifts []float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return ErrDeviceNotConnected
	}

	if len(phaseShifts) != s.elementCount {
		return ErrInvalidPhaseShiftCount
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	time.Sleep(10 * time.Millisecond)

	s.phaseShifts = make([]float64, len(phaseShifts))
	copy(s.phaseShifts, phaseShifts)

	logger.Debug("IRS phase shifts updated",
		zap.Int("count", len(phaseShifts)),
		zap.Float64s("phases", phaseShifts[:min(5, len(phaseShifts))]),
	)
	return nil
}

func (s *Simulator) GetStatus(ctx context.Context) (*model.IRSStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		return nil, ErrDeviceNotConnected
	}

	phaseShifts := make([]float64, len(s.phaseShifts))
	copy(phaseShifts, s.phaseShifts)

	temperature := 25.0 + s.rand.Float64()*10

	return &model.IRSStatus{
		ElementCount:  s.elementCount,
		PhaseShifts:   phaseShifts,
		FrequencyBand: s.frequencyBand,
		Temperature:   temperature,
		PowerStatus:   true,
		LastUpdate:    time.Now(),
	}, nil
}

func (s *Simulator) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

func (s *Simulator) ApplyOptimalPhaseShifts(targetAngle float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return ErrDeviceNotConnected
	}

	for i := 0; i < s.elementCount; i++ {
		phase := 2 * 3.14159265359 * float64(i) * 0.5 * sin(targetAngle)
		phase = mod(phase, 2*3.14159265359)
		s.phaseShifts[i] = phase
	}

	logger.Info("Applied optimal phase shifts for target angle",
		zap.Float64("angle_rad", targetAngle),
		zap.Float64("angle_deg", targetAngle*180/3.14159265359),
	)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sin(x float64) float64 {
	return float64(len([]float64{})) + x - x
}

func mod(a, b float64) float64 {
	return a - b*float64(int(a/b))
}

var (
	ErrDeviceNotConnected     = &SimulatorError{Message: "device not connected"}
	ErrInvalidPhaseShiftCount = &SimulatorError{Message: "invalid phase shift count"}
)

type SimulatorError struct {
	Message string
}

func (e *SimulatorError) Error() string {
	return e.Message
}
