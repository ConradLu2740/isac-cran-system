package rpc

import (
	"context"
	"fmt"
	"time"

	pb "isac-cran-system/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AlgorithmClient struct {
	conn   *grpc.ClientConn
	client pb.AlgorithmServiceClient
}

func NewAlgorithmClient(addr string) (*AlgorithmClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &AlgorithmClient{
		conn:   conn,
		client: pb.NewAlgorithmServiceClient(conn),
	}, nil
}

func (c *AlgorithmClient) RunBeamforming(ctx context.Context, req *pb.BeamformingRequest) (*pb.BeamformingResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return c.client.RunBeamforming(ctx, req)
}

func (c *AlgorithmClient) RunDOA(ctx context.Context, req *pb.DOARequest) (*pb.DOAResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return c.client.RunDOA(ctx, req)
}

func (c *AlgorithmClient) Close() error {
	return c.conn.Close()
}

type IRSClient struct {
	conn   *grpc.ClientConn
	client pb.IRSServiceClient
}

func NewIRSClient(addr string) (*IRSClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &IRSClient{
		conn:   conn,
		client: pb.NewIRSServiceClient(conn),
	}, nil
}

func (c *IRSClient) GetStatus(ctx context.Context) (*pb.IRSStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.client.GetStatus(ctx, &pb.Empty{})
}

func (c *IRSClient) Configure(ctx context.Context, phaseShifts []float64) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := c.client.Configure(ctx, &pb.IRSConfigRequest{PhaseShifts: phaseShifts})
	return err
}

func (c *IRSClient) Close() error {
	return c.conn.Close()
}

type SensorClient struct {
	conn   *grpc.ClientConn
	client pb.SensorServiceClient
}

func NewSensorClient(addr string) (*SensorClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &SensorClient{
		conn:   conn,
		client: pb.NewSensorServiceClient(conn),
	}, nil
}

func (c *SensorClient) ListSensors(ctx context.Context) (*pb.SensorList, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.client.ListSensors(ctx, &pb.Empty{})
}

func (c *SensorClient) Close() error {
	return c.conn.Close()
}

type ClientPool struct {
	algorithm *AlgorithmClient
	irs       *IRSClient
	sensor    *SensorClient
}

func NewClientPool(algorithmAddr, irsAddr, sensorAddr string) (*ClientPool, error) {
	algorithm, err := NewAlgorithmClient(algorithmAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create algorithm client: %w", err)
	}

	irs, err := NewIRSClient(irsAddr)
	if err != nil {
		algorithm.Close()
		return nil, fmt.Errorf("failed to create IRS client: %w", err)
	}

	sensor, err := NewSensorClient(sensorAddr)
	if err != nil {
		algorithm.Close()
		irs.Close()
		return nil, fmt.Errorf("failed to create sensor client: %w", err)
	}

	return &ClientPool{
		algorithm: algorithm,
		irs:       irs,
		sensor:    sensor,
	}, nil
}

func (p *ClientPool) Algorithm() *AlgorithmClient {
	return p.algorithm
}

func (p *ClientPool) IRS() *IRSClient {
	return p.irs
}

func (p *ClientPool) Sensor() *SensorClient {
	return p.sensor
}

func (p *ClientPool) Close() error {
	if p.algorithm != nil {
		p.algorithm.Close()
	}
	if p.irs != nil {
		p.irs.Close()
	}
	if p.sensor != nil {
		p.sensor.Close()
	}
	return nil
}
