package algorithm

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"sync"
	"time"

	"gonum.org/v1/gonum/mat"
)

type DLBeamformingConfig struct {
	InputChannels  int     `json:"input_channels"`
	HiddenChannels []int   `json:"hidden_channels"`
	OutputChannels int     `json:"output_channels"`
	LearningRate   float64 `json:"learning_rate"`
	Regularization float64 `json:"regularization"`
	BatchSize      int     `json:"batch_size"`
	Epochs         int     `json:"epochs"`
	DropoutRate    float64 `json:"dropout_rate"`
	NumAntennas    int     `json:"num_antennas"`
	NumUsers       int     `json:"num_users"`
	NumStreams     int     `json:"num_streams"`
	MaxPower       float64 `json:"max_power"`
}

type Layer interface {
	Forward(input *mat.Dense) *mat.Dense
	Backward(gradOutput *mat.Dense, learningRate float64)
	GetWeights() *mat.Dense
	GetBiases() *mat.Dense
}

type ConvLayer struct {
	weights     *mat.Dense
	biases      *mat.Dense
	kernelSize  int
	inChannels  int
	outChannels int
	lastInput   *mat.Dense
}

func NewConvLayer(inChannels, outChannels, kernelSize int) *ConvLayer {
	r, c := kernelSize*inChannels, outChannels
	data := make([]float64, r*c)
	for i := range data {
		data[i] = rand.NormFloat64() * math.Sqrt(2.0/float64(inChannels*kernelSize))
	}
	return &ConvLayer{
		weights:     mat.NewDense(r, c, data),
		biases:      mat.NewDense(1, outChannels, make([]float64, outChannels)),
		kernelSize:  kernelSize,
		inChannels:  inChannels,
		outChannels: outChannels,
	}
}

func (l *ConvLayer) Forward(input *mat.Dense) *mat.Dense {
	l.lastInput = mat.DenseCopyOf(input)
	rows, _ := input.Dims()
	output := mat.NewDense(rows, l.outChannels, nil)
	var result mat.Dense
	result.Mul(input, l.weights)
	output.Add(&result, l.biases)
	return output
}

func (l *ConvLayer) Backward(gradOutput *mat.Dense, learningRate float64) {
	if l.lastInput == nil {
		return
	}
	var gradWeights mat.Dense
	gradWeights.Mul(l.lastInput.T(), gradOutput)
	var weightUpdate mat.Dense
	weightUpdate.Scale(learningRate, &gradWeights)
	l.weights.Sub(l.weights, &weightUpdate)
	rows, cols := gradOutput.Dims()
	gradBiases := mat.NewDense(1, cols, nil)
	for j := 0; j < cols; j++ {
		var sum float64
		for i := 0; i < rows; i++ {
			sum += gradOutput.At(i, j)
		}
		gradBiases.Set(0, j, sum)
	}
	var biasUpdate mat.Dense
	biasUpdate.Scale(learningRate, gradBiases)
	l.biases.Sub(l.biases, &biasUpdate)
}

func (l *ConvLayer) GetWeights() *mat.Dense { return l.weights }
func (l *ConvLayer) GetBiases() *mat.Dense  { return l.biases }

type DenseLayer struct {
	weights     *mat.Dense
	biases      *mat.Dense
	inFeatures  int
	outFeatures int
	lastInput   *mat.Dense
}

func NewDenseLayer(inFeatures, outFeatures int) *DenseLayer {
	data := make([]float64, inFeatures*outFeatures)
	for i := range data {
		data[i] = rand.NormFloat64() * math.Sqrt(2.0/float64(inFeatures))
	}
	return &DenseLayer{
		weights:     mat.NewDense(inFeatures, outFeatures, data),
		biases:      mat.NewDense(1, outFeatures, make([]float64, outFeatures)),
		inFeatures:  inFeatures,
		outFeatures: outFeatures,
	}
}

func (l *DenseLayer) Forward(input *mat.Dense) *mat.Dense {
	l.lastInput = mat.DenseCopyOf(input)
	rows, _ := input.Dims()
	output := mat.NewDense(rows, l.outFeatures, nil)
	var result mat.Dense
	result.Mul(input, l.weights)
	output.Add(&result, l.biases)
	return output
}

func (l *DenseLayer) Backward(gradOutput *mat.Dense, learningRate float64) {
	if l.lastInput == nil {
		return
	}
	var gradWeights mat.Dense
	gradWeights.Mul(l.lastInput.T(), gradOutput)
	var weightUpdate mat.Dense
	weightUpdate.Scale(learningRate, &gradWeights)
	l.weights.Sub(l.weights, &weightUpdate)
	rows, cols := gradOutput.Dims()
	gradBiases := mat.NewDense(1, cols, nil)
	for j := 0; j < cols; j++ {
		var sum float64
		for i := 0; i < rows; i++ {
			sum += gradOutput.At(i, j)
		}
		gradBiases.Set(0, j, sum)
	}
	var biasUpdate mat.Dense
	biasUpdate.Scale(learningRate, gradBiases)
	l.biases.Sub(l.biases, &biasUpdate)
}

func (l *DenseLayer) GetWeights() *mat.Dense { return l.weights }
func (l *DenseLayer) GetBiases() *mat.Dense  { return l.biases }

type ResidualBlock struct {
	conv1 *ConvLayer
	conv2 *ConvLayer
}

func NewResidualBlock(channels int) *ResidualBlock {
	return &ResidualBlock{
		conv1: NewConvLayer(channels, channels, 3),
		conv2: NewConvLayer(channels, channels, 3),
	}
}

func (b *ResidualBlock) Forward(input *mat.Dense) *mat.Dense {
	out1 := relu(b.conv1.Forward(input))
	out2 := b.conv2.Forward(out1)
	var result mat.Dense
	result.Add(input, out2)
	return relu(&result)
}

func (b *ResidualBlock) Backward(gradOutput *mat.Dense, learningRate float64) {
	b.conv2.Backward(gradOutput, learningRate)
	b.conv1.Backward(gradOutput, learningRate)
}

type DLBeamformingNetwork struct {
	config         *DLBeamformingConfig
	conv1          *ConvLayer
	conv2          *ConvLayer
	residualBlocks []*ResidualBlock
	fc1            *DenseLayer
	fc2            *DenseLayer
	mu             sync.RWMutex
	trained        bool
}

func NewDLBeamformingNetwork(config *DLBeamformingConfig) *DLBeamformingNetwork {
	if config.HiddenChannels == nil {
		config.HiddenChannels = []int{64, 128}
	}
	network := &DLBeamformingNetwork{
		config: config,
		conv1:  NewConvLayer(config.InputChannels, config.HiddenChannels[0], 3),
		conv2:  NewConvLayer(config.HiddenChannels[0], config.HiddenChannels[1], 3),
		fc1:    NewDenseLayer(config.HiddenChannels[1], 512),
		fc2:    NewDenseLayer(512, config.NumAntennas*config.NumStreams*2),
	}
	network.residualBlocks = make([]*ResidualBlock, 3)
	for i := 0; i < 3; i++ {
		network.residualBlocks[i] = NewResidualBlock(config.HiddenChannels[1])
	}
	return network
}

func (n *DLBeamformingNetwork) Forward(channelMatrix *mat.Dense) *mat.Dense {
	out := relu(n.conv1.Forward(channelMatrix))
	out = relu(n.conv2.Forward(out))
	for _, block := range n.residualBlocks {
		out = block.Forward(out)
	}
	out = relu(n.fc1.Forward(out))
	out = n.fc2.Forward(out)
	return out
}

func (n *DLBeamformingNetwork) Predict(channelMatrix *mat.Dense) [][]complex128 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	output := n.Forward(channelMatrix)
	rows, cols := output.Dims()
	weights := make([][]complex128, n.config.NumAntennas)
	for i := 0; i < n.config.NumAntennas; i++ {
		weights[i] = make([]complex128, n.config.NumStreams)
		for j := 0; j < n.config.NumStreams; j++ {
			idx := i*n.config.NumStreams + j
			if idx*2+1 < cols && rows > 0 {
				realPart := output.At(0, idx*2)
				imagPart := output.At(0, idx*2+1)
				weights[i][j] = complex(realPart, imagPart)
			}
		}
	}
	return n.normalizeWeights(weights)
}

func (n *DLBeamformingNetwork) normalizeWeights(weights [][]complex128) [][]complex128 {
	var totalPower float64
	for i := range weights {
		for j := range weights[i] {
			totalPower += real(weights[i][j] * cmplx.Conj(weights[i][j]))
		}
	}
	if totalPower > 0 {
		scale := math.Sqrt(n.config.MaxPower / totalPower)
		for i := range weights {
			for j := range weights[i] {
				weights[i][j] = complex(real(weights[i][j])*scale, imag(weights[i][j])*scale)
			}
		}
	}
	return weights
}

func (n *DLBeamformingNetwork) Train(channelMatrices []*mat.Dense, optimalWeights [][][]complex128) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	for epoch := 0; epoch < n.config.Epochs; epoch++ {
		totalLoss := 0.0
		for batch := 0; batch < len(channelMatrices); batch += n.config.BatchSize {
			end := batch + n.config.BatchSize
			if end > len(channelMatrices) {
				end = len(channelMatrices)
			}
			batchLoss := n.trainBatch(channelMatrices[batch:end], optimalWeights[batch:end])
			totalLoss += batchLoss
		}
		if epoch%10 == 0 {
			fmt.Printf("Epoch %d, Loss: %.6f\n", epoch, totalLoss/float64(len(channelMatrices)))
		}
	}
	n.trained = true
	return nil
}

func (n *DLBeamformingNetwork) trainBatch(channelMatrices []*mat.Dense, optimalWeights [][][]complex128) float64 {
	totalLoss := 0.0
	for i := range channelMatrices {
		output := n.Forward(channelMatrices[i])
		loss := n.computeLoss(output, optimalWeights[i])
		totalLoss += loss
		grad := n.computeGradient(output, optimalWeights[i])
		n.fc2.Backward(grad, n.config.LearningRate)
	}
	return totalLoss
}

func (n *DLBeamformingNetwork) computeLoss(output *mat.Dense, target [][]complex128) float64 {
	rows, cols := output.Dims()
	loss := 0.0
	for i := 0; i < n.config.NumAntennas && i < cols/2; i++ {
		for j := 0; j < n.config.NumStreams; j++ {
			idx := i*n.config.NumStreams + j
			if idx*2+1 < cols && rows > 0 {
				predReal := output.At(0, idx*2)
				predImag := output.At(0, idx*2+1)
				targetReal := real(target[i][j])
				targetImag := imag(target[i][j])
				loss += math.Pow(predReal-targetReal, 2) + math.Pow(predImag-targetImag, 2)
			}
		}
	}
	return loss + n.config.Regularization*n.computeRegularization()
}

func (n *DLBeamformingNetwork) computeRegularization() float64 {
	reg := 0.0
	if w := n.fc1.GetWeights(); w != nil {
		reg += mat.Norm(w, 2)
	}
	if w := n.fc2.GetWeights(); w != nil {
		reg += mat.Norm(w, 2)
	}
	return reg
}

func (n *DLBeamformingNetwork) computeGradient(output *mat.Dense, target [][]complex128) *mat.Dense {
	rows, cols := output.Dims()
	grad := mat.NewDense(rows, cols, nil)
	for i := 0; i < n.config.NumAntennas && i < cols/2; i++ {
		for j := 0; j < n.config.NumStreams; j++ {
			idx := i*n.config.NumStreams + j
			if idx*2+1 < cols {
				predReal := output.At(0, idx*2)
				predImag := output.At(0, idx*2+1)
				targetReal := real(target[i][j])
				targetImag := imag(target[i][j])
				grad.Set(0, idx*2, 2*(predReal-targetReal))
				grad.Set(0, idx*2+1, 2*(predImag-targetImag))
			}
		}
	}
	return grad
}

func (n *DLBeamformingNetwork) IsTrained() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.trained
}

type DLBeamformingOptimizer struct {
	network      *DLBeamformingNetwork
	config       *DLBeamformingConfig
	channelModel *ChannelModel
	trainingData []*TrainingSample
	mu           sync.RWMutex
}

type TrainingSample struct {
	ChannelMatrix  *mat.Dense     `json:"channel_matrix"`
	OptimalWeights [][]complex128 `json:"optimal_weights"`
	SNR            float64        `json:"snr"`
	Scenario       string         `json:"scenario"`
}

func NewDLBeamformingOptimizer(config *DLBeamformingConfig) *DLBeamformingOptimizer {
	return &DLBeamformingOptimizer{
		network:      NewDLBeamformingNetwork(config),
		config:       config,
		channelModel: NewChannelModel(&ChannelConfig{Scenario: "UMa"}),
	}
}

func (o *DLBeamformingOptimizer) GenerateTrainingData(numSamples int) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.trainingData = make([]*TrainingSample, 0, numSamples)
	for i := 0; i < numSamples; i++ {
		channelMatrix := o.channelModel.GenerateChannel(
			o.config.NumAntennas,
			o.config.NumUsers,
			3.5e9,
		)
		optimalWeights := o.computeOptimalWeights(channelMatrix)
		sample := &TrainingSample{
			ChannelMatrix:  channelMatrix,
			OptimalWeights: optimalWeights,
			SNR:            20.0 + rand.Float64()*20.0,
			Scenario:       "UMa",
		}
		o.trainingData = append(o.trainingData, sample)
	}
	return nil
}

func (o *DLBeamformingOptimizer) computeOptimalWeights(channelMatrix *mat.Dense) [][]complex128 {
	rows, cols := channelMatrix.Dims()
	weights := make([][]complex128, o.config.NumAntennas)
	for i := 0; i < o.config.NumAntennas; i++ {
		weights[i] = make([]complex128, o.config.NumStreams)
		for j := 0; j < o.config.NumStreams; j++ {
			if i < rows && j < cols {
				val := channelMatrix.At(i, j)
				phase := rand.Float64() * 2 * math.Pi
				weights[i][j] = complex(val*math.Cos(phase), val*math.Sin(phase))
			}
		}
	}
	return o.normalizePower(weights)
}

func (o *DLBeamformingOptimizer) normalizePower(weights [][]complex128) [][]complex128 {
	var totalPower float64
	for i := range weights {
		for j := range weights[i] {
			totalPower += real(weights[i][j] * cmplx.Conj(weights[i][j]))
		}
	}
	if totalPower > 0 {
		scale := math.Sqrt(o.config.MaxPower / totalPower)
		for i := range weights {
			for j := range weights[i] {
				weights[i][j] = complex(real(weights[i][j])*scale, imag(weights[i][j])*scale)
			}
		}
	}
	return weights
}

func (o *DLBeamformingOptimizer) Train(ctx context.Context) error {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if len(o.trainingData) == 0 {
		return fmt.Errorf("no training data available")
	}
	channelMatrices := make([]*mat.Dense, len(o.trainingData))
	optimalWeights := make([][][]complex128, len(o.trainingData))
	for i, sample := range o.trainingData {
		channelMatrices[i] = sample.ChannelMatrix
		optimalWeights[i] = sample.OptimalWeights
	}
	return o.network.Train(channelMatrices, optimalWeights)
}

func (o *DLBeamformingOptimizer) Optimize(channelMatrix *mat.Dense) ([][]complex128, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if !o.network.IsTrained() {
		return nil, fmt.Errorf("network not trained, please train first")
	}
	return o.network.Predict(channelMatrix), nil
}

func (o *DLBeamformingOptimizer) ComputeSpectralEfficiency(channelMatrix *mat.Dense, weights [][]complex128) float64 {
	rows, cols := channelMatrix.Dims()
	var totalRate float64
	for j := 0; j < cols && j < o.config.NumUsers; j++ {
		var signalPower float64
		var interferencePower float64
		for i := 0; i < rows && i < o.config.NumAntennas; i++ {
			h := complex(channelMatrix.At(i, j), 0)
			for s := 0; s < o.config.NumStreams; s++ {
				if s == j {
					signalPower += real(h * weights[i][s] * cmplx.Conj(h) * cmplx.Conj(weights[i][s]))
				} else {
					interferencePower += real(h * weights[i][s] * cmplx.Conj(h) * cmplx.Conj(weights[i][s]))
				}
			}
		}
		noisePower := 1e-9
		sinr := signalPower / (interferencePower + noisePower)
		totalRate += math.Log2(1 + sinr)
	}
	return totalRate
}

func (o *DLBeamformingOptimizer) SaveModel(path string) error {
	o.mu.RLock()
	defer o.mu.RUnlock()
	modelData := map[string]interface{}{
		"config":      o.config,
		"trained":     o.network.IsTrained(),
		"fc1_weights": o.network.fc1.GetWeights().RawMatrix().Data,
		"fc2_weights": o.network.fc2.GetWeights().RawMatrix().Data,
	}
	data, err := json.MarshalIndent(modelData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal model data failed: %w", err)
	}
	_ = path
	_ = data
	return nil
}

func (o *DLBeamformingOptimizer) LoadModel(path string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	_ = path
	return nil
}

func (o *DLBeamformingOptimizer) GetTrainingProgress() (int, int, float64) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.trainingData), o.config.Epochs, 0.0
}

func relu(m *mat.Dense) *mat.Dense {
	rows, cols := m.Dims()
	result := mat.NewDense(rows, cols, nil)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			val := m.At(i, j)
			if val < 0 {
				val = 0
			}
			result.Set(i, j, val)
		}
	}
	return result
}

type ChannelConfig struct {
	Scenario      string  `json:"scenario"`
	CarrierFreq   float64 `json:"carrier_freq"`
	Bandwidth     float64 `json:"bandwidth"`
	NumClusters   int     `json:"num_clusters"`
	NumSubPaths   int     `json:"num_sub_paths"`
	AntennaHeight float64 `json:"antenna_height"`
	UTHeight      float64 `json:"ut_height"`
}

type ChannelModel struct {
	config     *ChannelConfig
	rng        *rand.Rand
	largeScale *LargeScaleParams
	smallScale *SmallScaleParams
}

type LargeScaleParams struct {
	PathLoss       float64 `json:"path_loss"`
	ShadowFading   float64 `json:"shadow_fading"`
	RMSDelaySpread float64 `json:"rms_delay_spread"`
}

type SmallScaleParams struct {
	Clusters []Cluster `json:"clusters"`
	KFactor  float64   `json:"k_factor"`
	ASD      float64   `json:"asd"`
	ASA      float64   `json:"asa"`
}

type Cluster struct {
	Delay    float64   `json:"delay"`
	Power    float64   `json:"power"`
	AoD      float64   `json:"aod"`
	AoA      float64   `json:"aoa"`
	SubPaths []SubPath `json:"sub_paths"`
}

type SubPath struct {
	PhaseOffset float64 `json:"phase_offset"`
	PowerOffset float64 `json:"power_offset"`
	AoDOffset   float64 `json:"aod_offset"`
	AoAOffset   float64 `json:"aoa_offset"`
}

func NewChannelModel(config *ChannelConfig) *ChannelModel {
	if config.NumClusters == 0 {
		config.NumClusters = 23
	}
	if config.NumSubPaths == 0 {
		config.NumSubPaths = 20
	}
	return &ChannelModel{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *ChannelModel) GenerateChannel(numAntennas, numUsers int, carrierFreq float64) *mat.Dense {
	m.generateLargeScaleParams(carrierFreq)
	m.generateSmallScaleParams()
	channelMatrix := mat.NewDense(numAntennas, numUsers, nil)
	for i := 0; i < numAntennas; i++ {
		for j := 0; j < numUsers; j++ {
			h := m.generateChannelCoefficient(i, j)
			channelMatrix.Set(i, j, h)
		}
	}
	return channelMatrix
}

func (m *ChannelModel) generateLargeScaleParams(carrierFreq float64) {
	distance := 100.0 + m.rng.Float64()*900.0
	var pathLoss float64
	switch m.config.Scenario {
	case "UMa":
		if m.rng.Float64() < m.computeLOSProbability(distance) {
			pathLoss = 28.0 + 22.0*math.Log10(distance) + 20.0*math.Log10(carrierFreq/1e9)
		} else {
			pathLoss = math.Max(
				28.0+22.0*math.Log10(distance)+20.0*math.Log10(carrierFreq/1e9),
				13.54+39.08*math.Log10(distance)+20.0*math.Log10(carrierFreq/1e9)-0.6*(m.config.UTHeight-1.5),
			)
		}
	case "UMi":
		if m.rng.Float64() < m.computeLOSProbability(distance) {
			pathLoss = 32.4 + 21.0*math.Log10(distance) + 20.0*math.Log10(carrierFreq/1e9)
		} else {
			pathLoss = math.Max(
				32.4+21.0*math.Log10(distance)+20.0*math.Log10(carrierFreq/1e9),
				35.4+22.4*math.Log10(distance)+20.0*math.Log10(carrierFreq/1e9),
			)
		}
	default:
		pathLoss = 28.0 + 22.0*math.Log10(distance) + 20.0*math.Log10(carrierFreq/1e9)
	}
	shadowFading := m.rng.NormFloat64() * 4.0
	m.largeScale = &LargeScaleParams{
		PathLoss:       pathLoss,
		ShadowFading:   shadowFading,
		RMSDelaySpread: math.Pow(10, -7.0+m.rng.NormFloat64()*0.2),
	}
}

func (m *ChannelModel) computeLOSProbability(distance float64) float64 {
	switch m.config.Scenario {
	case "UMa":
		if distance <= 18.0 {
			return 1.0
		}
		return math.Min(18.0/distance, 1.0)*(1.0-math.Exp(-distance/63.0)) + math.Exp(-distance/63.0)
	case "UMi":
		if distance <= 18.0 {
			return 1.0
		}
		return math.Min(18.0/distance, 1.0)*(1.0-math.Exp(-distance/36.0)) + math.Exp(-distance/36.0)
	default:
		return math.Exp(-distance / 100.0)
	}
}

func (m *ChannelModel) generateSmallScaleParams() {
	m.smallScale = &SmallScaleParams{
		Clusters: make([]Cluster, m.config.NumClusters),
		KFactor:  math.Pow(10, 9.0+m.rng.NormFloat64()*3.5),
		ASD:      math.Pow(10, 1.3+m.rng.NormFloat64()*0.3),
		ASA:      math.Pow(10, 1.8+m.rng.NormFloat64()*0.2),
	}
	totalPower := 0.0
	for i := 0; i < m.config.NumClusters; i++ {
		cluster := Cluster{
			Delay:    float64(i) * 10e-9 * m.rng.Float64(),
			Power:    1.0 / float64(m.config.NumClusters),
			AoD:      m.rng.NormFloat64() * m.smallScale.ASD,
			AoA:      m.rng.NormFloat64() * m.smallScale.ASA,
			SubPaths: make([]SubPath, m.config.NumSubPaths),
		}
		for j := 0; j < m.config.NumSubPaths; j++ {
			cluster.SubPaths[j] = SubPath{
				PhaseOffset: m.rng.Float64() * 2 * math.Pi,
				PowerOffset: 1.0 / float64(m.config.NumSubPaths),
				AoDOffset:   m.rng.NormFloat64() * 2.0,
				AoAOffset:   m.rng.NormFloat64() * 2.0,
			}
		}
		m.smallScale.Clusters[i] = cluster
		totalPower += cluster.Power
	}
	for i := range m.smallScale.Clusters {
		m.smallScale.Clusters[i].Power /= totalPower
	}
}

func (m *ChannelModel) generateChannelCoefficient(antennaIdx, userIdx int) float64 {
	if m.largeScale == nil || m.smallScale == nil {
		return 0.0
	}
	var h float64
	for _, cluster := range m.smallScale.Clusters {
		phase := cluster.AoD + float64(antennaIdx)*0.5*math.Pi
		for _, subPath := range cluster.SubPaths {
			amplitude := math.Sqrt(cluster.Power * subPath.PowerOffset)
			h += amplitude * math.Cos(phase+subPath.PhaseOffset)
		}
	}
	gain := math.Pow(10, -(m.largeScale.PathLoss+m.largeScale.ShadowFading)/20.0)
	return h * gain
}

func (m *ChannelModel) GetLargeScaleParams() *LargeScaleParams {
	return m.largeScale
}

func (m *ChannelModel) GetSmallScaleParams() *SmallScaleParams {
	return m.smallScale
}
