package service

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type EchoService struct {
	pb.UnimplementedEchoServiceServer
}

func (s *EchoService) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{
		Message: req.Message,
	}, nil
}
