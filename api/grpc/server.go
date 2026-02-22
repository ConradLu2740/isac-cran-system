package grpc

import (
	"context"

	pb "isac-cran-system/api/proto"
	"isac-cran-system/internal/model"
	"isac-cran-system/internal/service"
)

type AlgorithmServer struct {
	pb.UnimplementedAlgorithmServiceServer
	service *service.AlgorithmService
}

func NewAlgorithmServer(service *service.AlgorithmService) *AlgorithmServer {
	return &AlgorithmServer{service: service}
}

func (s *AlgorithmServer) RunBeamforming(ctx context.Context, req *pb.BeamformingRequest) (*pb.BeamformingResponse, error) {
	params := &model.BeamformingParams{
		ElementCount:       int(req.ElementCount),
		TargetDirection:    req.TargetDirection,
		InterferenceAngles: req.InterferenceAngles,
		SNRThreshold:       req.SnrThreshold,
		MaxIterations:      int(req.MaxIterations),
	}

	result, err := s.service.RunBeamforming(ctx, req.ExperimentId, params)
	if err != nil {
		return nil, err
	}

	return &pb.BeamformingResponse{
		ExperimentId:      req.ExperimentId,
		BeamPattern:       result.BeamPattern,
		MainLobeDirection: result.MainLobeDirection,
		MainLobeWidth:     result.MainLobeWidth,
		SideLobeLevel:     result.SLL,
		Iterations:        int32(result.Iterations),
		Converged:         result.Converged,
	}, nil
}

func (s *AlgorithmServer) RunDOA(ctx context.Context, req *pb.DOARequest) (*pb.DOAResponse, error) {
	params := &model.DOAParams{
		ElementCount:   int(req.ElementCount),
		NumSources:     int(req.NumSources),
		SnapshotLength: int(req.SnapshotLength),
		Method:         req.Method,
	}

	result, err := s.service.RunDOA(ctx, req.ExperimentId, params)
	if err != nil {
		return nil, err
	}

	return &pb.DOAResponse{
		ExperimentId:    req.ExperimentId,
		EstimatedAngles: result.EstimatedAngles,
		Spectrum:        result.Spectrum,
		Rmse:            result.RMSE,
	}, nil
}

func (s *AlgorithmServer) StreamBeamforming(stream pb.AlgorithmService_StreamBeamformingServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		params := &model.BeamformingParams{
			ElementCount:       int(req.ElementCount),
			TargetDirection:    req.TargetDirection,
			InterferenceAngles: req.InterferenceAngles,
			SNRThreshold:       req.SnrThreshold,
			MaxIterations:      int(req.MaxIterations),
		}

		result, err := s.service.RunBeamforming(stream.Context(), req.ExperimentId, params)
		if err != nil {
			return err
		}

		if err := stream.Send(&pb.BeamformingResponse{
			ExperimentId:      req.ExperimentId,
			BeamPattern:       result.BeamPattern,
			MainLobeDirection: result.MainLobeDirection,
			Iterations:        int32(result.Iterations),
			Converged:         result.Converged,
		}); err != nil {
			return err
		}
	}
}

type IRSServer struct {
	pb.UnimplementedIRSServiceServer
	service *service.IRSService
}

func NewIRSServer(service *service.IRSService) *IRSServer {
	return &IRSServer{service: service}
}

func (s *IRSServer) GetStatus(ctx context.Context, _ *pb.Empty) (*pb.IRSStatus, error) {
	status, err := s.service.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.IRSStatus{
		ElementCount:  int32(status.ElementCount),
		FrequencyBand: status.FrequencyBand,
		Temperature:   status.Temperature,
		PowerStatus:   status.PowerStatus,
		PhaseShifts:   status.PhaseShifts,
	}, nil
}

func (s *IRSServer) Configure(ctx context.Context, req *pb.IRSConfigRequest) (*pb.IRSConfigResponse, error) {
	err := s.service.Configure(ctx, &model.IRSConfig{
		PhaseShifts: req.PhaseShifts,
	})
	if err != nil {
		return &pb.IRSConfigResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.IRSConfigResponse{Success: true, Message: "Configuration applied"}, nil
}
