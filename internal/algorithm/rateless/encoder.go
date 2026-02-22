package rateless

import (
	"math"
	"math/rand"

	"isac-cran-system/pkg/logger"

	"go.uber.org/zap"
)

type Encoder struct {
	sourceLen  int
	degreeDist []float64
	rand       *rand.Rand
}

type EncodedSymbol struct {
	Index   int
	Degree  int
	Indices []int
	Value   float64
}

func NewEncoder(sourceLen int) *Encoder {
	e := &Encoder{
		sourceLen: sourceLen,
		rand:      rand.New(rand.NewSource(42)),
	}
	e.degreeDist = e.robustSolitonDistribution(sourceLen, 0.1, 0.01)
	return e
}

func (e *Encoder) Encode(source []float64, numSymbols int) []EncodedSymbol {
	symbols := make([]EncodedSymbol, numSymbols)

	for i := 0; i < numSymbols; i++ {
		degree := e.sampleDegree()
		indices := e.sampleIndices(degree)

		var value float64
		for _, idx := range indices {
			value += source[idx]
		}

		symbols[i] = EncodedSymbol{
			Index:   i,
			Degree:  degree,
			Indices: indices,
			Value:   value,
		}
	}

	logger.Debug("Rateless encoding completed",
		zap.Int("source_len", len(source)),
		zap.Int("num_symbols", numSymbols),
	)

	return symbols
}

func (e *Encoder) robustSolitonDistribution(K int, c, delta float64) []float64 {
	R := c * math.Sqrt(float64(K)) * math.Log(float64(K)/delta)

	ideal := make([]float64, K+1)
	ideal[1] = 1.0 / float64(K)
	for d := 2; d <= K; d++ {
		ideal[d] = 1.0 / (float64(d) * float64(d-1))
	}

	robust := make([]float64, K+1)
	for d := 1; d <= K; d++ {
		robust[d] = ideal[d]
	}

	Kf := float64(K)
	Rf := R

	for d := 1; d < int(Kf/Rf); d++ {
		robust[d] += 1.0 / (Rf * float64(d))
	}
	d := int(Kf / Rf)
	if d <= K {
		robust[d] += math.Log(Rf/delta) / Rf
	}

	var sum float64
	for d := 1; d <= K; d++ {
		sum += robust[d]
	}

	for d := 1; d <= K; d++ {
		robust[d] /= sum
	}

	return robust
}

func (e *Encoder) sampleDegree() int {
	r := e.rand.Float64()
	cumulative := 0.0

	for d := 1; d <= e.sourceLen; d++ {
		cumulative += e.degreeDist[d]
		if r <= cumulative {
			return d
		}
	}

	return 1
}

func (e *Encoder) sampleIndices(degree int) []int {
	indices := make(map[int]bool)
	for len(indices) < degree {
		idx := e.rand.Intn(e.sourceLen)
		indices[idx] = true
	}

	result := make([]int, 0, degree)
	for idx := range indices {
		result = append(result, idx)
	}
	return result
}

type Decoder struct {
	sourceLen int
}

func NewDecoder(sourceLen int) *Decoder {
	return &Decoder{
		sourceLen: sourceLen,
	}
}

func (d *Decoder) Decode(symbols []EncodedSymbol) ([]float64, bool) {
	decoded := make([]float64, d.sourceLen)
	decodedFlags := make([]bool, d.sourceLen)

	for {
		progress := false

		for _, sym := range symbols {
			undecodedCount := 0
			var undecodedIdx int
			var value float64

			for _, idx := range sym.Indices {
				if !decodedFlags[idx] {
					undecodedCount++
					undecodedIdx = idx
				} else {
					value -= decoded[idx]
				}
			}

			if undecodedCount == 1 {
				decoded[undecodedIdx] = sym.Value + value
				decodedFlags[undecodedIdx] = true
				progress = true
			}
		}

		if !progress {
			break
		}
	}

	allDecoded := true
	for _, flag := range decodedFlags {
		if !flag {
			allDecoded = false
			break
		}
	}

	logger.Debug("Rateless decoding completed",
		zap.Bool("success", allDecoded),
		zap.Int("decoded_count", countTrue(decodedFlags)),
	)

	return decoded, allDecoded
}

func countTrue(flags []bool) int {
	count := 0
	for _, f := range flags {
		if f {
			count++
		}
	}
	return count
}

func (d *Decoder) EstimateOverhead(sourceLen int) float64 {
	return 1.1 + 0.1*math.Sqrt(float64(sourceLen))
}
