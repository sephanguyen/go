package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRetrieveMapByStudyPlanID(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}

	s := &StudyPlanItemReaderService{
		DB:                mockDB,
		StudyPlanItemRepo: studyPlanItemRepo,
	}

	spItems := []*repositories.FindLearningMaterialByStudyPlanID{
		{
			LearningMaterialID: database.Text("lm_1"),
			StudyPlanItemID:    database.Text("sp_id_1"),
		},
		{
			LearningMaterialID: database.Text("lm_2"),
			StudyPlanItemID:    database.Text("sp_id_2"),
		},
	}

	lmIDMapSPItemID := make(map[string]string)
	for _, item := range spItems {
		lmIDMapSPItemID[item.LearningMaterialID.String] = item.StudyPlanItemID.String
	}

	testCases := []TestCase{
		{
			name: "Happy case",
			req: &epb.RetrieveMappingLmIDToStudyPlanItemIDRequest{
				StudyPlanId: "_SP_ID",
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindLearningMaterialByStudyPlanID", mock.Anything, mock.Anything, mock.Anything).Once().Return(spItems, nil)
			},
			expectedResp: &epb.RetrieveMappingLmIDToStudyPlanItemIDResponse{
				Pairs: lmIDMapSPItemID,
			},
		},
		{
			name: "Missing study plan id",
			req: &epb.RetrieveMappingLmIDToStudyPlanItemIDRequest{
				StudyPlanId: "",
			},
			setup:       func(ctx context.Context) {},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("validateRetrieveMappingLmIDToStudyPlanItemIDRequest: StudyPlanId cannot be empty").Error()),
		},
		{
			name: "FindLearningMaterialByStudyPlanID err",
			req: &epb.RetrieveMappingLmIDToStudyPlanItemIDRequest{
				StudyPlanId: "_SP_ID",
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FindLearningMaterialByStudyPlanID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoDatabaseName)
			},
			expectedErr: pgx.ErrNoDatabaseName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)

			res, err := s.RetrieveMappingLmIDToStudyPlanItemID(ctx, tc.req.(*epb.RetrieveMappingLmIDToStudyPlanItemIDRequest))

			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.Equal(t, tc.expectedResp, res)
				assert.NoError(t, err)
			}
		})
	}
}
