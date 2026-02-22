package beamforming

import (
	"math"
	"math/cmplx"

	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/logger"

	"go.uber.org/zap"
)

type Optimizer struct {
	elementCount         int
	maxIterations        int
	convergenceThreshold float64
}

func NewOptimizer(elementCount int, maxIterations int, threshold float64) *Optimizer {
	return &Optimizer{
		elementCount:         elementCount,
		maxIterations:        maxIterations,
		convergenceThreshold: threshold,
	}
}

func (o *Optimizer) Optimize(params *model.BeamformingParams) (*model.BeamformingResult, error) {
	logger.Info("Starting beamforming optimization",
		zap.Int("element_count", params.ElementCount),
		zap.Float64("target_direction", params.TargetDirection),
	)

	weights := o.initializeWeights(params.ElementCount)

	targetSteering := o.computeSteeringVector(params.ElementCount, params.TargetDirection)

	interferenceSteerings := make([][]complex128, len(params.InterferenceAngles))
	for i, angle := range params.InterferenceAngles {
		interferenceSteerings[i] = o.computeSteeringVector(params.ElementCount, angle)
	}

	var converged bool
	var iterations int

	for iter := 0; iter < o.maxIterations; iter++ {
		iterations = iter + 1

		gradient := make([]complex128, params.ElementCount)

		for n := 0; n < params.ElementCount; n++ {
			grad := complex(0, 0)
			for m := 0; m < params.ElementCount; m++ {
				grad += weights[m] * cmplx.Conj(targetSteering[n]) * targetSteering[m]
			}
			gradient[n] = grad
		}

		stepSize := 0.1 / complex(float64(iterations+1), 0)
		for n := 0; n < params.ElementCount; n++ {
			weights[n] += stepSize * gradient[n]
		}

		o.normalizeWeights(weights)

		if o.checkConvergence(weights, params.TargetDirection, params.SNRThreshold) {
			converged = true
			break
		}
	}

	beamPattern := o.computeBeamPattern(weights, 360)

	mainLobeDir, mainLobeWidth, sll := o.analyzeBeamPattern(beamPattern)

	weightsSerializable := make([][]float64, len(weights))
	for i, w := range weights {
		weightsSerializable[i] = []float64{real(w), imag(w)}
	}

	result := &model.BeamformingResult{
		Weights:           weightsSerializable,
		BeamPattern:       beamPattern,
		MainLobeDirection: mainLobeDir,
		MainLobeWidth:     mainLobeWidth,
		SLL:               sll,
		Iterations:        iterations,
		Converged:         converged,
	}

	logger.Info("Beamforming optimization completed",
		zap.Int("iterations", iterations),
		zap.Bool("converged", converged),
		zap.Float64("main_lobe_dir", mainLobeDir),
		zap.Float64("sll_db", 20*math.Log10(sll)),
	)

	return result, nil
}

func (o *Optimizer) initializeWeights(elementCount int) []complex128 {
	weights := make([]complex128, elementCount)
	for i := range weights {
		phase := -2 * math.Pi * float64(i) * 0.5
		weights[i] = cmplx.Exp(complex(0, phase))
	}
	o.normalizeWeights(weights)
	return weights
}

func (o *Optimizer) computeSteeringVector(elementCount int, angle float64) []complex128 {
	steering := make([]complex128, elementCount)
	d := 0.5
	for n := 0; n < elementCount; n++ {
		phase := 2 * math.Pi * float64(n) * d * math.Sin(angle)
		steering[n] = cmplx.Exp(complex(0, phase))
	}
	return steering
}

func (o *Optimizer) normalizeWeights(weights []complex128) {
	var sum float64
	for _, w := range weights {
		sum += cmplx.Abs(w) * cmplx.Abs(w)
	}
	norm := math.Sqrt(sum)
	for i := range weights {
		weights[i] = weights[i] / complex(norm, 0)
	}
}

func (o *Optimizer) checkConvergence(weights []complex128, targetAngle, snrThreshold float64) bool {
	steering := o.computeSteeringVector(len(weights), targetAngle)

	var response complex128
	for n, w := range weights {
		response += w * cmplx.Conj(steering[n])
	}

	snr := cmplx.Abs(response)
	return snr >= snrThreshold
}

func (o *Optimizer) computeBeamPattern(weights []complex128, numPoints int) []float64 {
	pattern := make([]float64, numPoints)
	d := 0.5

	for i := 0; i < numPoints; i++ {
		angle := -math.Pi/2 + float64(i)*math.Pi/float64(numPoints)

		var response complex128
		for n, w := range weights {
			phase := 2 * math.Pi * float64(n) * d * math.Sin(angle)
			response += w * cmplx.Exp(complex(0, -phase))
		}

		pattern[i] = cmplx.Abs(response)
	}

	return pattern
}

func (o *Optimizer) analyzeBeamPattern(pattern []float64) (mainLobeDir, mainLobeWidth, sll float64) {
	maxIdx := 0
	maxVal := pattern[0]
	for i, v := range pattern {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	mainLobeDir = -math.Pi/2 + float64(maxIdx)*math.Pi/float64(len(pattern))

	halfPower := maxVal / math.Sqrt(2)
	leftIdx, rightIdx := maxIdx, maxIdx

	for i := maxIdx; i >= 0; i-- {
		if pattern[i] < halfPower {
			leftIdx = i
			break
		}
	}
	for i := maxIdx; i < len(pattern); i++ {
		if pattern[i] < halfPower {
			rightIdx = i
			break
		}
	}
	mainLobeWidth = float64(rightIdx-leftIdx) * math.Pi / float64(len(pattern))

	sll = 0
	for i := 0; i < len(pattern); i++ {
		if i < leftIdx || i > rightIdx {
			if pattern[i] > sll {
				sll = pattern[i]
			}
		}
	}

	return
}

func (o *Optimizer) ComputeArrayFactor(weights []complex128, angles []float64) []float64 {
	af := make([]float64, len(angles))
	d := 0.5

	for i, angle := range angles {
		var response complex128
		for n, w := range weights {
			phase := 2 * math.Pi * float64(n) * d * math.Sin(angle)
			response += w * cmplx.Exp(complex(0, -phase))
		}
		af[i] = cmplx.Abs(response)
	}

	return af
}
