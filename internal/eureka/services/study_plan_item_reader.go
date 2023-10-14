package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudyPlanItemReaderService struct {
	DB database.Ext

	StudyPlanItemRepo interface {
		FindLearningMaterialByStudyPlanID(ctx context.Context, db database.QueryExecer, id pgtype.Text) ([]*repositories.FindLearningMaterialByStudyPlanID, error)
	}
}

func NewStudyPlanItemReaderService(db database.Ext) epb.StudyPlanItemReaderServiceServer {
	return &StudyPlanItemReaderService{
		DB:                db,
		StudyPlanItemRepo: new(repositories.StudyPlanItemRepo), //nolint
	}
}

func validateRetrieveMappingLmIDToStudyPlanItemIDRequest(req *epb.RetrieveMappingLmIDToStudyPlanItemIDRequest) error {
	if req.StudyPlanId == "" {
		return errors.New("StudyPlanId cannot be empty")
	}

	return nil
}

func (s *StudyPlanItemReaderService) RetrieveMappingLmIDToStudyPlanItemID(ctx context.Context, req *epb.RetrieveMappingLmIDToStudyPlanItemIDRequest) (*epb.RetrieveMappingLmIDToStudyPlanItemIDResponse, error) {
	if err := validateRetrieveMappingLmIDToStudyPlanItemIDRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateRetrieveMappingLmIDToStudyPlanItemIDRequest: %w", err).Error())
	}

	spItems, err := s.StudyPlanItemRepo.FindLearningMaterialByStudyPlanID(ctx, s.DB, database.Text(req.StudyPlanId))
	if err != nil {
		return nil, fmt.Errorf("s.StudyPlanItemRepo.FindLearningMaterialByStudyPlanID: %w", err)
	}

	lmIDMapSPItemID := make(map[string]string)
	for _, item := range spItems {
		lmIDMapSPItemID[item.LearningMaterialID.String] = item.StudyPlanItemID.String
	}

	return &epb.RetrieveMappingLmIDToStudyPlanItemIDResponse{
		Pairs: lmIDMapSPItemID,
	}, nil
}
