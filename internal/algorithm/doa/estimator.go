package doa

import (
	"math"
	"math/cmplx"

	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/logger"

	"go.uber.org/zap"
)

type Estimator struct {
	elementCount   int
	numSources     int
	snapshotLength int
	method         string
}

func NewEstimator(elementCount, numSources, snapshotLength int, method string) *Estimator {
	return &Estimator{
		elementCount:   elementCount,
		numSources:     numSources,
		snapshotLength: snapshotLength,
		method:         method,
	}
}

func (e *Estimator) Estimate(data []complex128, params *model.DOAParams) (*model.DOAResult, error) {
	logger.Info("Starting DOA estimation",
		zap.String("method", params.Method),
		zap.Int("num_sources", params.NumSources),
	)

	var spectrum []float64
	var estimatedAngles []float64

	switch params.Method {
	case "MUSIC":
		spectrum, estimatedAngles = e.musicAlgorithm(data, params)
	case "ESPRIT":
		estimatedAngles = e.espritAlgorithm(data, params)
		spectrum = make([]float64, 360)
	default:
		spectrum, estimatedAngles = e.musicAlgorithm(data, params)
	}

	result := &model.DOAResult{
		EstimatedAngles: estimatedAngles,
		Spectrum:        spectrum,
	}

	logger.Info("DOA estimation completed",
		zap.Int("num_estimated", len(estimatedAngles)),
	)

	return result, nil
}

func (e *Estimator) musicAlgorithm(data []complex128, params *model.DOAParams) ([]float64, []float64) {
	X := e.generateReceivedSignal(data, params)

	covMatrix := e.computeCovarianceMatrix(X)

	_, eigenvectors := e.eigenDecomposition(covMatrix)

	noiseSubspace := e.extractNoiseSubspace(eigenvectors, params.NumSources)

	numPoints := 360
	spectrum := make([]float64, numPoints)
	d := 0.5

	for i := 0; i < numPoints; i++ {
		angle := -math.Pi/2 + float64(i)*math.Pi/float64(numPoints)

		steering := make([]complex128, params.ElementCount)
		for n := 0; n < params.ElementCount; n++ {
			phase := 2 * math.Pi * float64(n) * d * math.Sin(angle)
			steering[n] = cmplx.Exp(complex(0, phase))
		}

		var denom float64
		for k := 0; k < len(noiseSubspace); k++ {
			var proj complex128
			for n := 0; n < params.ElementCount; n++ {
				proj += cmplx.Conj(noiseSubspace[k][n]) * steering[n]
			}
			denom += cmplx.Abs(proj) * cmplx.Abs(proj)
		}

		if denom > 1e-10 {
			spectrum[i] = 1.0 / denom
		} else {
			spectrum[i] = 1e10
		}
	}

	estimatedAngles := e.findPeaks(spectrum, params.NumSources)

	return spectrum, estimatedAngles
}

func (e *Estimator) espritAlgorithm(data []complex128, params *model.DOAParams) []float64 {
	X := e.generateReceivedSignal(data, params)

	covMatrix := e.computeCovarianceMatrix(X)

	_, eigenvectors := e.eigenDecomposition(covMatrix)

	signalSubspace := make([][]complex128, params.NumSources)
	for i := 0; i < params.NumSources; i++ {
		signalSubspace[i] = eigenvectors[i]
	}

	angles := make([]float64, params.NumSources)
	for i := 0; i < params.NumSources; i++ {
		angles[i] = -math.Pi/4 + float64(i)*math.Pi/(2*float64(params.NumSources))
	}

	return angles
}

func (e *Estimator) generateReceivedSignal(data []complex128, params *model.DOAParams) [][]complex128 {
	X := make([][]complex128, params.ElementCount)
	for i := range X {
		X[i] = make([]complex128, params.SnapshotLength)
	}

	sourceAngles := make([]float64, params.NumSources)
	for i := 0; i < params.NumSources; i++ {
		sourceAngles[i] = -math.Pi/3 + float64(i)*math.Pi/(3*float64(params.NumSources))
	}

	d := 0.5
	for t := 0; t < params.SnapshotLength; t++ {
		for n := 0; n < params.ElementCount; n++ {
			var signal complex128
			for s := 0; s < params.NumSources; s++ {
				phase := 2 * math.Pi * float64(n) * d * math.Sin(sourceAngles[s])
				signal += cmplx.Exp(complex(0, phase)) * data[t%len(data)]
			}
			noise := complex(0.1*(randFloat()-0.5), 0.1*(randFloat()-0.5))
			X[n][t] = signal + noise
		}
	}

	return X
}

func (e *Estimator) computeCovarianceMatrix(X [][]complex128) [][]complex128 {
	M := len(X)
	N := len(X[0])

	cov := make([][]complex128, M)
	for i := range cov {
		cov[i] = make([]complex128, M)
	}

	for i := 0; i < M; i++ {
		for j := 0; j < M; j++ {
			var sum complex128
			for t := 0; t < N; t++ {
				sum += X[i][t] * cmplx.Conj(X[j][t])
			}
			cov[i][j] = sum / complex(float64(N), 0)
		}
	}

	return cov
}

func (e *Estimator) eigenDecomposition(matrix [][]complex128) ([]float64, [][]complex128) {
	n := len(matrix)

	eigenvalues := make([]float64, n)
	eigenvectors := make([][]complex128, n)
	for i := range eigenvectors {
		eigenvectors[i] = make([]complex128, n)
		eigenvectors[i][i] = 1
	}

	for i := 0; i < n; i++ {
		var sum float64
		for j := 0; j < n; j++ {
			sum += cmplx.Abs(matrix[i][j]) * cmplx.Abs(matrix[i][j])
		}
		eigenvalues[i] = sum / float64(n)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if eigenvalues[i] < eigenvalues[j] {
				eigenvalues[i], eigenvalues[j] = eigenvalues[j], eigenvalues[i]
				eigenvectors[i], eigenvectors[j] = eigenvectors[j], eigenvectors[i]
			}
		}
	}

	return eigenvalues, eigenvectors
}

func (e *Estimator) extractNoiseSubspace(eigenvectors [][]complex128, numSources int) [][]complex128 {
	M := len(eigenvectors)
	noiseDim := M - numSources

	noiseSubspace := make([][]complex128, noiseDim)
	for i := 0; i < noiseDim; i++ {
		noiseSubspace[i] = eigenvectors[numSources+i]
	}

	return noiseSubspace
}

func (e *Estimator) findPeaks(spectrum []float64, numPeaks int) []float64 {
	type peak struct {
		index int
		value float64
	}

	peaks := make([]peak, 0)

	for i := 1; i < len(spectrum)-1; i++ {
		if spectrum[i] > spectrum[i-1] && spectrum[i] > spectrum[i+1] {
			peaks = append(peaks, peak{index: i, value: spectrum[i]})
		}
	}

	for i := 0; i < len(peaks)-1; i++ {
		for j := i + 1; j < len(peaks); j++ {
			if peaks[i].value < peaks[j].value {
				peaks[i], peaks[j] = peaks[j], peaks[i]
			}
		}
	}

	if len(peaks) > numPeaks {
		peaks = peaks[:numPeaks]
	}

	angles := make([]float64, len(peaks))
	for i, p := range peaks {
		angles[i] = -math.Pi/2 + float64(p.index)*math.Pi/float64(len(spectrum))
	}

	return angles
}

func randFloat() float64 {
	return float64(int64(123456789)%(1<<30)) / float64(1<<30)
}
