package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/draft/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	dpb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SendCoverageServer struct {
	DB database.Ext
	dpb.UnimplementedSendCoverageServiceServer
	DraftRepo interface {
		GetTargetCoverage(ctx context.Context, db database.QueryExecer, repository, branchName string, integration bool) (*entities.TargetCoverage, error)
		AddHistory(ctx context.Context, db database.QueryExecer, coverage float64, branchName, targetbranch, status, repository string, integration bool) error
		UpdateTargetCoverage(ctx context.Context, db database.QueryExecer, coverage float64, id int64) error
		CreateTargetBranch(ctx context.Context, db database.QueryExecer, branchName, repository string, integration bool, key string, coverage float64) error
	}
}

func compareSecretKey(currentKey, targetKeyHashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(targetKeyHashed), []byte(currentKey))
	return err == nil
}

// SendCoverage return result
func (s *SendCoverageServer) SendCoverage(ctx context.Context, req *dpb.SendCoverageRequest) (*dpb.SendCoverageResponse, error) {
	targetCoverage, err := s.DraftRepo.GetTargetCoverage(ctx, s.DB, req.Repository, req.TargetBranch, req.Integration)
	if err != nil {
		return nil, fmt.Errorf("s.DraftRepo.GetTargetCoverage: %v", err)
	}

	if targetCoverage == nil {
		return nil, errors.New("target branch does not exist")
	}

	if !compareSecretKey(req.Key, targetCoverage.Key) {
		return &dpb.SendCoverageResponse{
			Message: dpb.SendCoverageResponse_FAIL,
		}, fmt.Errorf("secret key is incorrect")
	}

	if req.Coverage < targetCoverage.Coverage {
		err = s.DraftRepo.AddHistory(
			ctx,
			s.DB,
			float64(req.Coverage),
			req.BranchName,
			targetCoverage.BranchName,
			dpb.SendCoverageResponse_FAIL.String(),
			req.Repository,
			req.Integration,
		)
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrap(err, "error recording history").Error())
		}
		return &dpb.SendCoverageResponse{
			Message: dpb.SendCoverageResponse_FAIL,
		}, status.Error(codes.FailedPrecondition, fmt.Sprintf("code coverage of base branch is %v, your branch is %v", targetCoverage.Coverage, req.Coverage))
	}

	err = s.DraftRepo.AddHistory(
		ctx,
		s.DB,
		float64(req.Coverage),
		req.BranchName,
		targetCoverage.BranchName,
		dpb.SendCoverageResponse_PASS.String(),
		req.Repository,
		req.Integration,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "error recording history").Error())
	}

	return &dpb.SendCoverageResponse{
		Message: dpb.SendCoverageResponse_PASS,
	}, nil
}

// CreateTargetCoverage creates the coverage target for target branch.
// creates must be unique per (repository, branch, is_integration).
// a key is used to secure the update method.
func (s *SendCoverageServer) CreateTargetCoverage(ctx context.Context, req *dpb.CreateTargetCoverageRequest) (*dpb.CreateTargetCoverageResponse, error) {
	hashKey, err := bcrypt.GenerateFromPassword([]byte(req.Key), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt.GenerateFromPassword: %v", err)
	}

	err = s.DraftRepo.CreateTargetBranch(ctx, s.DB, req.BranchName, req.Repository, req.Integration, string(hashKey), float64(req.Coverage))
	if err != nil {
		return nil, fmt.Errorf("s.DraftRepo.CreateTargetBranch: %v", err)
	}

	return &dpb.CreateTargetCoverageResponse{}, nil
}

func (s *SendCoverageServer) UpdateTargetCoverage(ctx context.Context, req *dpb.UpdateTargetCoverageRequest) (*dpb.UpdateTargetCoverageResponse, error) {
	targetCoverage, err := s.DraftRepo.GetTargetCoverage(ctx, s.DB, req.Repository, req.BranchName, req.Integration)
	if err != nil {
		return nil, fmt.Errorf("s.DraftRepo.GetTargetCoverage: %v", err)
	}
	if targetCoverage == nil {
		return nil, errors.New("target branch does not exist")
	}

	if !compareSecretKey(req.Key, targetCoverage.Key) {
		return nil, fmt.Errorf("secret key is incorrect")
	}

	if req.Coverage < targetCoverage.Coverage {
		err = s.DraftRepo.AddHistory(ctx, s.DB, float64(req.Coverage), req.BranchName, targetCoverage.BranchName, dpb.SendCoverageResponse_FAIL.String(), req.Repository, req.Integration)
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrap(err, "error recording history").Error())
		}
	}

	err = s.DraftRepo.AddHistory(ctx, s.DB, float64(req.Coverage), req.BranchName, targetCoverage.BranchName, dpb.SendCoverageResponse_PASS.String(), req.Repository, req.Integration)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "error recording history").Error())
	}

	err = s.DraftRepo.UpdateTargetCoverage(ctx, s.DB, float64(req.Coverage), targetCoverage.ID)
	if err != nil {
		return nil, fmt.Errorf("s.DraftRepo.GetTargetCoverage: %v", err)
	}

	return &dpb.UpdateTargetCoverageResponse{}, nil
}
