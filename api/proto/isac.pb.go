package proto

import (
	"context"

	"google.golang.org/grpc"
)

type Empty struct{}

type BeamformingRequest struct {
	ExperimentId       string
	ElementCount       int32
	TargetDirection    float64
	InterferenceAngles []float64
	SnrThreshold       float64
	MaxIterations      int32
}

type BeamformingResponse struct {
	ExperimentId      string
	BeamPattern       []float64
	MainLobeDirection float64
	MainLobeWidth     float64
	SideLobeLevel     float64
	Iterations        int32
	Converged         bool
}

type DOARequest struct {
	ExperimentId   string
	ElementCount   int32
	NumSources     int32
	SnapshotLength int32
	Method         string
}

type DOAResponse struct {
	ExperimentId    string
	EstimatedAngles []float64
	Spectrum        []float64
	Rmse            float64
}

type IRSStatus struct {
	ElementCount  int32
	FrequencyBand string
	Temperature   float64
	PowerStatus   string
	PhaseShifts   []float64
}

type IRSConfigRequest struct {
	PhaseShifts []float64
}

type IRSConfigResponse struct {
	Success bool
	Message string
}

type SensorData struct {
	SensorId  string
	Value     float64
	Unit      string
	Timestamp int64
}

type SensorRequest struct {
	SensorId string
}

type SensorList struct {
	Sensors []*SensorInfo
}

type SensorInfo struct {
	Id          string
	Type        string
	Location    string
	Description string
}

type AlgorithmServiceServer interface {
	RunBeamforming(ctx context.Context, req *BeamformingRequest) (*BeamformingResponse, error)
	RunDOA(ctx context.Context, req *DOARequest) (*DOAResponse, error)
	StreamBeamforming(stream grpc.ServerStream) error
}

type UnimplementedAlgorithmServiceServer struct{}

func (UnimplementedAlgorithmServiceServer) RunBeamforming(ctx context.Context, req *BeamformingRequest) (*BeamformingResponse, error) {
	return nil, nil
}

func (UnimplementedAlgorithmServiceServer) RunDOA(ctx context.Context, req *DOARequest) (*DOAResponse, error) {
	return nil, nil
}

func (UnimplementedAlgorithmServiceServer) StreamBeamforming(stream grpc.ServerStream) error {
	return nil
}

type IRSServiceServer interface {
	GetStatus(ctx context.Context, req *Empty) (*IRSStatus, error)
	Configure(ctx context.Context, req *IRSConfigRequest) (*IRSConfigResponse, error)
}

type UnimplementedIRSServiceServer struct{}

func (UnimplementedIRSServiceServer) GetStatus(ctx context.Context, req *Empty) (*IRSStatus, error) {
	return nil, nil
}

func (UnimplementedIRSServiceServer) Configure(ctx context.Context, req *IRSConfigRequest) (*IRSConfigResponse, error) {
	return nil, nil
}

type SensorServiceServer interface {
	GetData(ctx context.Context, req *SensorRequest) (*SensorData, error)
	StreamData(stream grpc.ServerStream) error
	ListSensors(ctx context.Context, req *Empty) (*SensorList, error)
}

type UnimplementedSensorServiceServer struct{}

func (UnimplementedSensorServiceServer) GetData(ctx context.Context, req *SensorRequest) (*SensorData, error) {
	return nil, nil
}

func (UnimplementedSensorServiceServer) StreamData(stream grpc.ServerStream) error {
	return nil
}

func (UnimplementedSensorServiceServer) ListSensors(ctx context.Context, req *Empty) (*SensorList, error) {
	return nil, nil
}

func RegisterAlgorithmServiceServer(s *grpc.Server, srv AlgorithmServiceServer) {
}

func RegisterIRSServiceServer(s *grpc.Server, srv IRSServiceServer) {
}

func RegisterSensorServiceServer(s *grpc.Server, srv SensorServiceServer) {
}

type AlgorithmServiceClient interface {
	RunBeamforming(ctx context.Context, req *BeamformingRequest, opts ...grpc.CallOption) (*BeamformingResponse, error)
	RunDOA(ctx context.Context, req *DOARequest, opts ...grpc.CallOption) (*DOAResponse, error)
}

type algorithmServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAlgorithmServiceClient(cc grpc.ClientConnInterface) AlgorithmServiceClient {
	return &algorithmServiceClient{cc}
}

func (c *algorithmServiceClient) RunBeamforming(ctx context.Context, req *BeamformingRequest, opts ...grpc.CallOption) (*BeamformingResponse, error) {
	return &BeamformingResponse{}, nil
}

func (c *algorithmServiceClient) RunDOA(ctx context.Context, req *DOARequest, opts ...grpc.CallOption) (*DOAResponse, error) {
	return &DOAResponse{}, nil
}

type IRSServiceClient interface {
	GetStatus(ctx context.Context, req *Empty, opts ...grpc.CallOption) (*IRSStatus, error)
	Configure(ctx context.Context, req *IRSConfigRequest, opts ...grpc.CallOption) (*IRSConfigResponse, error)
}

type irsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIRSServiceClient(cc grpc.ClientConnInterface) IRSServiceClient {
	return &irsServiceClient{cc}
}

func (c *irsServiceClient) GetStatus(ctx context.Context, req *Empty, opts ...grpc.CallOption) (*IRSStatus, error) {
	return &IRSStatus{}, nil
}

func (c *irsServiceClient) Configure(ctx context.Context, req *IRSConfigRequest, opts ...grpc.CallOption) (*IRSConfigResponse, error) {
	return &IRSConfigResponse{}, nil
}

type SensorServiceClient interface {
	GetData(ctx context.Context, req *SensorRequest, opts ...grpc.CallOption) (*SensorData, error)
	ListSensors(ctx context.Context, req *Empty, opts ...grpc.CallOption) (*SensorList, error)
}

type sensorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSensorServiceClient(cc grpc.ClientConnInterface) SensorServiceClient {
	return &sensorServiceClient{cc}
}

func (c *sensorServiceClient) GetData(ctx context.Context, req *SensorRequest, opts ...grpc.CallOption) (*SensorData, error) {
	return &SensorData{}, nil
}

func (c *sensorServiceClient) ListSensors(ctx context.Context, req *Empty, opts ...grpc.CallOption) (*SensorList, error) {
	return &SensorList{}, nil
}
