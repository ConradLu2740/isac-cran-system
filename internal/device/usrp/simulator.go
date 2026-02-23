package usrp

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/logger"

	"go.uber.org/zap"
)

type Simulator struct {
	sampleRate float64
	centerFreq float64
	gain       float64
	connected  bool
	mu         sync.RWMutex
	rand       *rand.Rand
	noiseLevel float64
}

func NewSimulator(sampleRate, centerFreq float64) *Simulator {
	return &Simulator{
		sampleRate: sampleRate,
		centerFreq: centerFreq,
		gain:       30.0,
		noiseLevel: 0.1,
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
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

	time.Sleep(100 * time.Millisecond)

	s.connected = true
	logger.Info("USRP simulator connected",
		zap.Float64("sample_rate_mhz", s.sampleRate/1e6),
		zap.Float64("center_freq_ghz", s.centerFreq/1e9),
	)
	return nil
}

func (s *Simulator) Disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = false
	logger.Info("USRP simulator disconnected")
	return nil
}

func (s *Simulator) Receive(ctx context.Context, duration time.Duration) ([]model.ChannelDataPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		return nil, ErrSimulatorNotConnected
	}

	numSamples := int(float64(duration.Seconds()) * s.sampleRate)
	if numSamples > 100000 {
		numSamples = 100000
	}

	data := make([]model.ChannelDataPoint, numSamples)

	numSignals := 3
	signalFreqs := []float64{0.1, 0.3, 0.5}
	signalAmps := []float64{1.0, 0.7, 0.5}
	signalPhases := []float64{0, math.Pi / 4, math.Pi / 2}

	for i := 0; i < numSamples; i++ {
		t := float64(i) / s.sampleRate

		iVal := 0.0
		qVal := 0.0

		for j := 0; j < numSignals; j++ {
			phase := 2*math.Pi*signalFreqs[j]*t*s.sampleRate + signalPhases[j]
			fading := 0.5 + 0.5*math.Cos(2*math.Pi*0.01*t)
			iVal += signalAmps[j] * fading * math.Cos(phase)
			qVal += signalAmps[j] * fading * math.Sin(phase)
		}

		iVal += s.noiseLevel * (s.rand.Float64()*2 - 1)
		qVal += s.noiseLevel * (s.rand.Float64()*2 - 1)

		amplitude := math.Sqrt(iVal*iVal + qVal*qVal)
		phase := math.Atan2(qVal, iVal)

		data[i] = model.ChannelDataPoint{
			Index:     i,
			Amplitude: amplitude,
			Phase:     phase,
			I:         iVal,
			Q:         qVal,
		}
	}

	logger.Debug("USRP data received",
		zap.Int("samples", numSamples),
		zap.Duration("duration", duration),
	)

	return data, nil
}

func (s *Simulator) SetFrequency(freq float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return ErrSimulatorNotConnected
	}

	s.centerFreq = freq
	return nil
}

func (s *Simulator) SetSampleRate(rate float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return ErrSimulatorNotConnected
	}

	s.sampleRate = rate
	return nil
}

func (s *Simulator) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

func (s *Simulator) SetNoiseLevel(level float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.noiseLevel = level
}

func (s *Simulator) SetGain(gain float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gain = gain
}

var ErrSimulatorNotConnected = &SimulatorError{Message: "simulator not connected"}

type SimulatorError struct {
	Message string
}

func (e *SimulatorError) Error() string {
	return e.Message
}
