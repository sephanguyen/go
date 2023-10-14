package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/draft/entities"
	mock_repositories "github.com/manabie-com/backend/mock/draft/repositories"
	dpb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testKey     = "aaaaaa"
	testKeyHash = "$2y$12$A0anxVNpfyRZb9m5armVvO7hN6F17/OT1Zwb12u6UvddyDMYLGeGm"
)

const (
	testKey2     = "AAAA"
	testKeyHash2 = "$2a$04$FtKRUMUKZtLbZek0dJ4eIuhm/CIndiu/nDa8QZFe3vy1igmEZRCsC"
)

func TestSendCoverage(t *testing.T) {
	t.Parallel()

	// this test does not do what it's named
	t.Run("current coverage less than target coverage", func(t *testing.T) {
		t.Parallel()
		draftRepo := &mock_repositories.MockDraftRepo{}
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.TargetCoverage{
			ID:          1,
			BranchName:  "develop",
			Coverage:    35.5,
			UpdateAt:    time.Now(),
			Repository:  "test",
			Key:         testKeyHash,
			Integration: false,
		}, nil)

		draftRepo.On("AddHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}

		resp, err := svc.SendCoverage(context.Background(), &dpb.SendCoverageRequest{
			Key: testKey,
		})
		fmt.Println(err)
		assert.NotNil(t, err)
		assert.Equal(t, &dpb.SendCoverageResponse{}, resp)
	})

	t.Run("current coverage greater than target coverage", func(t *testing.T) {
		t.Parallel()
		draftRepo := &mock_repositories.MockDraftRepo{}
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.TargetCoverage{
			ID:         1,
			BranchName: "develop",
			Coverage:   0,
			UpdateAt:   time.Now(),
			Repository: "test",
			Key:        testKeyHash,
		}, nil)

		draftRepo.On("AddHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		draftRepo.On("UpdateTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}

		resp, err := svc.SendCoverage(context.Background(), &dpb.SendCoverageRequest{
			Key:         testKey,
			Coverage:    20,
			BranchName:  "nhattien-draft",
			Repository:  "test",
			Integration: false,
		})

		assert.Nil(t, err)
		assert.Equal(t, &dpb.SendCoverageResponse{
			Message: dpb.SendCoverageResponse_PASS,
		}, resp)
	})

	t.Run("secret is incorrect", func(t *testing.T) {
		t.Parallel()
		draftRepo := &mock_repositories.MockDraftRepo{}
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.TargetCoverage{
			ID:          1,
			BranchName:  "develop",
			Coverage:    0,
			Repository:  "test",
			Integration: false,

			UpdateAt: time.Now(),
			Key:      testKeyHash,
		}, nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}
		resp, err := svc.SendCoverage(context.Background(), &dpb.SendCoverageRequest{
			Key: "asdasd",
		})

		assert.NotNil(t, err)
		assert.Equal(t, &dpb.SendCoverageResponse{
			Message: dpb.SendCoverageResponse_FAIL,
		}, resp)
	})

	t.Run("get target coverage fail", func(t *testing.T) {
		t.Parallel()
		draftRepo := &mock_repositories.MockDraftRepo{}
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("error get target coverage"))
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}

		resp, err := svc.SendCoverage(context.Background(), &dpb.SendCoverageRequest{})
		assert.Nil(t, resp)
		assert.EqualError(t, err, "s.DraftRepo.GetTargetCoverage: error get target coverage")
	})
}

func TestCreateTargetCoverage(t *testing.T) {
	t.Parallel()
	t.Run("target coverage nil,create target coverage success", func(t *testing.T) {
		t.Parallel()
		draftRepo := &mock_repositories.MockDraftRepo{}
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		draftRepo.On("CreateTargetBranch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("float64")).Return(nil)
		draftRepo.On("UpdateTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}
		resp, err := svc.CreateTargetCoverage(context.Background(), &dpb.CreateTargetCoverageRequest{
			BranchName:  mock.Anything,
			Repository:  mock.Anything,
			Integration: false,
		})
		assert.Nil(t, err)
		assert.Equal(t, &dpb.CreateTargetCoverageResponse{}, resp)
	})

	t.Run("create target coverage success", func(t *testing.T) {
		t.Parallel()
		draftRepo := &mock_repositories.MockDraftRepo{}
		branchName := "develop"
		repository := "manabie/backend"
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.TargetCoverage{
			ID:          1,
			BranchName:  branchName,
			Integration: false,
			Coverage:    0,
			UpdateAt:    time.Now(),
			Repository:  repository,
		}, nil)
		draftRepo.On("CreateTargetBranch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("float64")).Return(nil)
		draftRepo.On("UpdateTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}
		resp, err := svc.CreateTargetCoverage(context.Background(), &dpb.CreateTargetCoverageRequest{
			BranchName:  branchName,
			Integration: false,
			Repository:  repository,
			Key:         "ABSDD1",
			Coverage:    30.1,
		})
		assert.Nil(t, err)
		assert.Equal(t, &dpb.CreateTargetCoverageResponse{}, resp)
	})

	t.Run("create target branch fail", func(t *testing.T) {
		t.Parallel()
		draftRepo := &mock_repositories.MockDraftRepo{}
		branchName := "develop"
		repository := "manabie/backend"
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("not have target coverage"))
		draftRepo.On("CreateTargetBranch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("float64")).Return(errors.New("create fail"))
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}
		resp, err := svc.CreateTargetCoverage(context.Background(), &dpb.CreateTargetCoverageRequest{
			BranchName:  branchName,
			Repository:  repository,
			Integration: false,
			Key:         "ASDSSS1",
		})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
	})
}

func TestUpdateTargetCoverage(t *testing.T) {
	t.Parallel()
	t.Run("update target coverage success", func(t *testing.T) {
		draftRepo := &mock_repositories.MockDraftRepo{}
		branchName := "develop"
		repo := "manabie-com/backend"
		key := testKey2
		draftRepo.On("UpdateTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.TargetCoverage{
			ID:          1,
			BranchName:  branchName,
			Coverage:    50.1,
			UpdateAt:    time.Now(),
			Integration: false,
			Repository:  repo,
			Key:         testKeyHash2,
		}, nil)
		draftRepo.On("AddHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}
		resp, err := svc.UpdateTargetCoverage(context.Background(), &dpb.UpdateTargetCoverageRequest{
			BranchName: branchName,
			Repository: repo,
			Coverage:   60,
			Key:        key,
		})
		assert.Nil(t, err)
		assert.Equal(t, &dpb.UpdateTargetCoverageResponse{}, resp)
	})

	t.Run("update target coverage success with dropped coverage", func(t *testing.T) {
		draftRepo := &mock_repositories.MockDraftRepo{}
		branchName := "develop"
		repo := "manabie-com/backend"
		key := testKey2
		draftRepo.On("UpdateTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.TargetCoverage{
			ID:          1,
			BranchName:  branchName,
			Coverage:    50.1,
			UpdateAt:    time.Now(),
			Integration: false,
			Repository:  repo,
			Key:         testKeyHash2,
		}, nil)
		draftRepo.On("AddHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}
		resp, err := svc.UpdateTargetCoverage(context.Background(), &dpb.UpdateTargetCoverageRequest{
			BranchName: branchName,
			Repository: repo,
			Coverage:   49.1,
			Key:        key,
		})
		assert.Nil(t, err)
		assert.Equal(t, &dpb.UpdateTargetCoverageResponse{}, resp)
	})

	t.Run("update target coverage fail", func(t *testing.T) {
		draftRepo := &mock_repositories.MockDraftRepo{}
		branchName := "develop"
		repo := "manabie-com/backend"
		key := testKey2
		draftRepo.On("UpdateTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		draftRepo.On("GetTargetCoverage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("target branch does not exist"))
		draftRepo.On("AddHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := &SendCoverageServer{
			DraftRepo: draftRepo,
		}
		resp, err := svc.UpdateTargetCoverage(context.Background(), &dpb.UpdateTargetCoverageRequest{
			BranchName:  branchName,
			Repository:  repo,
			Integration: false,
			Coverage:    60,
			Key:         key,
		})
		assert.NotNil(t, err)
		assert.Nil(t, resp)
	})
}
