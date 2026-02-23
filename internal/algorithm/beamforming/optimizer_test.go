package beamforming

import (
	"math"
	"testing"

	"isac-cran-system/internal/model"
)

func TestOptimizer_Optimize(t *testing.T) {
	optimizer := NewOptimizer(64, 100, 0.001)

	params := &model.BeamformingParams{
		ElementCount:       64,
		TargetDirection:    0.5,
		InterferenceAngles: []float64{0.2, 0.8},
		SNRThreshold:       0.9,
		MaxIterations:      100,
	}

	result, err := optimizer.Optimize(params)
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	if len(result.Weights) != 64 {
		t.Errorf("Expected 64 weights, got %d", len(result.Weights))
	}

	if len(result.BeamPattern) != 360 {
		t.Errorf("Expected 360 beam pattern points, got %d", len(result.BeamPattern))
	}

	if result.Iterations < 1 {
		t.Error("Expected at least 1 iteration")
	}
}

func TestOptimizer_ComputeArrayFactor(t *testing.T) {
	optimizer := NewOptimizer(64, 100, 0.001)

	weights := make([]complex128, 64)
	for i := range weights {
		weights[i] = complex(1.0/8.0, 0)
	}

	angles := []float64{0, 0.5, 1.0}
	af := optimizer.ComputeArrayFactor(weights, angles)

	if len(af) != len(angles) {
		t.Errorf("Expected %d array factor values, got %d", len(angles), len(af))
	}
}

func TestWeightsCalculator_ComputeConjugateBeamforming(t *testing.T) {
	calc := NewWeightsCalculator(64, 0.5)

	targetAngle := 0.5
	weights := calc.ComputeConjugateBeamforming(targetAngle)

	if len(weights) != 64 {
		t.Errorf("Expected 64 weights, got %d", len(weights))
	}

	var sum float64
	for _, w := range weights {
		sum += real(w)*real(w) + imag(w)*imag(w)
	}
	norm := math.Sqrt(sum)

	if math.Abs(norm-1.0) > 0.01 {
		t.Errorf("Expected normalized weights, got norm %f", norm)
	}
}

func TestWeightsCalculator_ComputePhaseShifts(t *testing.T) {
	calc := NewWeightsCalculator(64, 0.5)

	weights := make([]complex128, 64)
	for i := range weights {
		phase := float64(i) * math.Pi / 32
		weights[i] = complex(math.Cos(phase), math.Sin(phase))
	}

	phases := calc.ComputePhaseShifts(weights)

	if len(phases) != 64 {
		t.Errorf("Expected 64 phases, got %d", len(phases))
	}

	for i, phase := range phases {
		if phase < 0 || phase > 2*math.Pi {
			t.Errorf("Phase %d out of range [0, 2π]: %f", i, phase)
		}
	}
}

func TestWeightsCalculator_ApplyPhaseQuantization(t *testing.T) {
	calc := NewWeightsCalculator(64, 0.5)

	phases := make([]float64, 64)
	for i := range phases {
		phases[i] = float64(i) * 0.1
	}

	quantized := calc.ApplyPhaseQuantization(phases, 3)

	if len(quantized) != 64 {
		t.Errorf("Expected 64 quantized phases, got %d", len(quantized))
	}
}

func BenchmarkOptimizer_Optimize(b *testing.B) {
	optimizer := NewOptimizer(64, 100, 0.001)

	params := &model.BeamformingParams{
		ElementCount:       64,
		TargetDirection:    0.5,
		InterferenceAngles: []float64{0.2, 0.8},
		SNRThreshold:       0.9,
		MaxIterations:      100,
	}

	for i := 0; i < b.N; i++ {
		_, _ = optimizer.Optimize(params)
	}
}
