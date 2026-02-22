package beamforming

import (
	"math"
	"math/cmplx"
)

type WeightsCalculator struct {
	elementCount   int
	elementSpacing float64
}

func NewWeightsCalculator(elementCount int, elementSpacing float64) *WeightsCalculator {
	return &WeightsCalculator{
		elementCount:   elementCount,
		elementSpacing: elementSpacing,
	}
}

func (w *WeightsCalculator) ComputeConjugateBeamforming(targetAngle float64) []complex128 {
	weights := make([]complex128, w.elementCount)
	for n := 0; n < w.elementCount; n++ {
		phase := 2 * math.Pi * float64(n) * w.elementSpacing * math.Sin(targetAngle)
		weights[n] = cmplx.Exp(complex(0, -phase))
	}
	w.normalize(weights)
	return weights
}

func (w *WeightsCalculator) ComputeZOFBeamforming(targetAngles []float64) []complex128 {
	weights := make([]complex128, w.elementCount)
	for n := 0; n < w.elementCount; n++ {
		var phase complex128
		for _, angle := range targetAngles {
			phase += cmplx.Exp(complex(0, 2*math.Pi*float64(n)*w.elementSpacing*math.Sin(angle)))
		}
		weights[n] = cmplx.Conj(phase)
	}
	w.normalize(weights)
	return weights
}

func (w *WeightsCalculator) ComputeMVDRWeights(covMatrix [][]complex128, steeringVector []complex128) []complex128 {
	n := len(steeringVector)
	invCov := w.matrixInverse(covMatrix)

	weights := make([]complex128, n)
	for i := 0; i < n; i++ {
		var sum complex128
		for j := 0; j < n; j++ {
			sum += invCov[i][j] * steeringVector[j]
		}
		weights[i] = sum
	}

	var denom complex128
	for i := 0; i < n; i++ {
		denom += cmplx.Conj(steeringVector[i]) * weights[i]
	}

	for i := range weights {
		weights[i] /= denom
	}

	return weights
}

func (w *WeightsCalculator) normalize(weights []complex128) {
	var sum float64
	for _, weight := range weights {
		sum += cmplx.Abs(weight) * cmplx.Abs(weight)
	}
	norm := math.Sqrt(sum)
	for i := range weights {
		weights[i] /= complex(norm, 0)
	}
}

func (w *WeightsCalculator) matrixInverse(matrix [][]complex128) [][]complex128 {
	n := len(matrix)
	aug := make([][]complex128, n)
	for i := range aug {
		aug[i] = make([]complex128, 2*n)
		for j := 0; j < n; j++ {
			aug[i][j] = matrix[i][j]
		}
		aug[i][n+i] = 1
	}

	for i := 0; i < n; i++ {
		pivot := aug[i][i]
		for j := 0; j < 2*n; j++ {
			aug[i][j] /= pivot
		}
		for k := 0; k < n; k++ {
			if k != i {
				factor := aug[k][i]
				for j := 0; j < 2*n; j++ {
					aug[k][j] -= factor * aug[i][j]
				}
			}
		}
	}

	inv := make([][]complex128, n)
	for i := range inv {
		inv[i] = make([]complex128, n)
		for j := 0; j < n; j++ {
			inv[i][j] = aug[i][n+j]
		}
	}

	return inv
}

func (w *WeightsCalculator) ComputePhaseShifts(weights []complex128) []float64 {
	phases := make([]float64, len(weights))
	for i, weight := range weights {
		phases[i] = cmplx.Phase(weight)
		if phases[i] < 0 {
			phases[i] += 2 * math.Pi
		}
	}
	return phases
}

func (w *WeightsCalculator) ApplyPhaseQuantization(phases []float64, bits int) []float64 {
	levels := math.Pow(2, float64(bits))
	step := 2 * math.Pi / levels
	quantized := make([]float64, len(phases))
	for i, phase := range phases {
		level := math.Round(phase / step)
		quantized[i] = math.Mod(level*step, 2*math.Pi)
	}
	return quantized
}
