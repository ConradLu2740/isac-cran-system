package doa

import (
	"math"
	"math/cmplx"
)

type MUSIC struct {
	elementCount   int
	numSources     int
	elementSpacing float64
}

func NewMUSIC(elementCount, numSources int, elementSpacing float64) *MUSIC {
	return &MUSIC{
		elementCount:   elementCount,
		numSources:     numSources,
		elementSpacing: elementSpacing,
	}
}

func (m *MUSIC) ComputeSpectrum(covMatrix [][]complex128, searchAngles []float64) []float64 {
	_, eigenvectors := m.eigenDecomposition(covMatrix)

	noiseSubspace := make([][]complex128, m.elementCount-m.numSources)
	for i := 0; i < m.elementCount-m.numSources; i++ {
		noiseSubspace[i] = eigenvectors[m.numSources+i]
	}

	spectrum := make([]float64, len(searchAngles))

	for i, angle := range searchAngles {
		steering := m.computeSteeringVector(angle)

		var denom float64
		for k := 0; k < len(noiseSubspace); k++ {
			var proj complex128
			for n := 0; n < m.elementCount; n++ {
				proj += cmplx.Conj(noiseSubspace[k][n]) * steering[n]
			}
			denom += cmplx.Abs(proj) * cmplx.Abs(proj)
		}

		if denom > 1e-10 {
			spectrum[i] = 10 * math.Log10(1.0/denom)
		} else {
			spectrum[i] = 100
		}
	}

	return spectrum
}

func (m *MUSIC) computeSteeringVector(angle float64) []complex128 {
	steering := make([]complex128, m.elementCount)
	for n := 0; n < m.elementCount; n++ {
		phase := 2 * math.Pi * float64(n) * m.elementSpacing * math.Sin(angle)
		steering[n] = cmplx.Exp(complex(0, phase))
	}
	return steering
}

func (m *MUSIC) eigenDecomposition(matrix [][]complex128) ([]float64, [][]complex128) {
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
			absVal := cmplx.Abs(matrix[i][j])
			sum += absVal * absVal
		}
		eigenvalues[i] = sum / float64(n)
	}

	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if eigenvalues[i] < eigenvalues[j] {
				eigenvalues[i], eigenvalues[j] = eigenvalues[j], eigenvalues[i]
				eigenvectors[i], eigenvectors[j] = eigenvectors[j], eigenvectors[i]
			}
		}
	}

	return eigenvalues, eigenvectors
}

func (m *MUSIC) EstimateDOA(covMatrix [][]complex128) []float64 {
	numPoints := 360
	searchAngles := make([]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		searchAngles[i] = -math.Pi/2 + float64(i)*math.Pi/float64(numPoints)
	}

	spectrum := m.ComputeSpectrum(covMatrix, searchAngles)

	peaks := m.findPeaks(spectrum, m.numSources)

	angles := make([]float64, len(peaks))
	for i, p := range peaks {
		angles[i] = searchAngles[p]
	}

	return angles
}

func (m *MUSIC) findPeaks(spectrum []float64, numPeaks int) []int {
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

	indices := make([]int, len(peaks))
	for i, p := range peaks {
		indices[i] = p.index
	}

	return indices
}
