package doa

import (
	"math/cmplx"
	"testing"

	"isac-cran-system/internal/model"
)

func TestEstimator_Estimate(t *testing.T) {
	estimator := NewEstimator(64, 3, 1024, "MUSIC")

	data := make([]complex128, 1024)
	for i := range data {
		t := float64(i) / 1024.0
		data[i] = complex(cmplx.Abs(complex(t, 0)), 0)
	}

	params := &model.DOAParams{
		ElementCount:   64,
		NumSources:     3,
		SnapshotLength: 1024,
		Method:         "MUSIC",
	}

	result, err := estimator.Estimate(data, params)
	if err != nil {
		t.Fatalf("Estimate failed: %v", err)
	}

	if len(result.Spectrum) == 0 {
		t.Error("Expected spectrum to be generated")
	}
}

func TestEstimator_MUSICAlgorithm(t *testing.T) {
	estimator := NewEstimator(64, 3, 1024, "MUSIC")

	data := make([]complex128, 1024)
	for i := range data {
		data[i] = complex(1, 0)
	}

	params := &model.DOAParams{
		ElementCount:   64,
		NumSources:     2,
		SnapshotLength: 1024,
		Method:         "MUSIC",
	}

	spectrum, angles := estimator.musicAlgorithm(data, params)

	if len(spectrum) == 0 {
		t.Error("Expected spectrum to be generated")
	}

	if len(angles) > params.NumSources {
		t.Errorf("Expected at most %d angles, got %d", params.NumSources, len(angles))
	}
}

func TestMUSIC_ComputeSpectrum(t *testing.T) {
	music := NewMUSIC(64, 3, 0.5)

	covMatrix := make([][]complex128, 64)
	for i := range covMatrix {
		covMatrix[i] = make([]complex128, 64)
		for j := range covMatrix[i] {
			if i == j {
				covMatrix[i][j] = complex(1, 0)
			}
		}
	}

	searchAngles := make([]float64, 360)
	for i := range searchAngles {
		searchAngles[i] = float64(i) * 3.14159265359 / 180.0
	}

	spectrum := music.ComputeSpectrum(covMatrix, searchAngles)

	if len(spectrum) != 360 {
		t.Errorf("Expected 360 spectrum points, got %d", len(spectrum))
	}
}

func TestMUSIC_EstimateDOA(t *testing.T) {
	music := NewMUSIC(64, 3, 0.5)

	covMatrix := make([][]complex128, 64)
	for i := range covMatrix {
		covMatrix[i] = make([]complex128, 64)
		for j := range covMatrix[i] {
			if i == j {
				covMatrix[i][j] = complex(1, 0)
			}
		}
	}

	angles := music.EstimateDOA(covMatrix)

	if len(angles) > 3 {
		t.Errorf("Expected at most 3 angles, got %d", len(angles))
	}
}

func BenchmarkEstimator_Estimate(b *testing.B) {
	estimator := NewEstimator(64, 3, 1024, "MUSIC")

	data := make([]complex128, 1024)
	for i := range data {
		data[i] = complex(1, 0)
	}

	params := &model.DOAParams{
		ElementCount:   64,
		NumSources:     3,
		SnapshotLength: 1024,
		Method:         "MUSIC",
	}

	for i := 0; i < b.N; i++ {
		_, _ = estimator.Estimate(data, params)
	}
}
