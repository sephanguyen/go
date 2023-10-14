package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	mock_study_plan_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/study_plan/repository/postgres"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
)

func TestStudyPlanItemUseCase_UpsertStudyPlanItems(t *testing.T) {
	t.Parallel()

	// mockLearningMaterialRepo := &mock_item_bank_learnosity_repo.MockItemBankRepo{}
	mockStudyPlanItemRepo := &mock_study_plan_repo.MockStudyPlanItemRepo{}
	mockLmListRepo := &mock_study_plan_repo.MockLmListRepo{}
	mockDB := new(mock_database.Ext)
	mockTxer := &mock_database.Tx{}

	now := time.Now()

	testCases := []struct {
		name           string
		studyPlanItems []domain.StudyPlanItem
		expectedErr    error
		setup          func(ctx context.Context)
	}{
		{
			name: "success",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				mockStudyPlanItemRepo.On("UpsertStudyPlanItems", mock.Anything, mock.Anything).Once().Return(nil)
				mockLmListRepo.On("UpsertLearningMaterialsIDList", mock.Anything, mock.Anything).Once().Return(nil)
			},
			studyPlanItems: []domain.StudyPlanItem{
				{
					Name:      "study plan item 1",
					LmList:    []string{"1", "2"},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedErr: nil,
		},
		{
			name: "error create learning material",
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				mockLmListRepo.On("UpsertLearningMaterialsIDList", mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error create learning material"))
				mockStudyPlanItemRepo.On("UpsertStudyPlanItems", mock.Anything, mock.Anything).Once().Return(nil)
			},
			studyPlanItems: []domain.StudyPlanItem{
				{
					Name:      "study plan item 1",
					LmList:    []string{"1", "2"},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			expectedErr: fmt.Errorf("ExecInTx: StudyPlanItemRepo.UpsertStudyPlanItems: error create learning material"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			uc := &StudyPlanItemUseCase{
				DB:                   mockDB,
				StudyPlanItemRepo:    mockStudyPlanItemRepo,
				LearningMaterialRepo: mockLmListRepo,
			}

			err := uc.UpsertStudyPlanItems(ctx, tc.studyPlanItems)
			if tc.expectedErr != nil {
				assert.Error(t, err, tc.expectedErr.Error())
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
		})
	}
}
