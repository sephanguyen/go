package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_UpsertQuestionnaireTemplate(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	questionnaireTemplateRepo := &mock_repositories.MockQuestionnaireTemplateRepo{}
	questionnaireTemplateQuestionRepo := &mock_repositories.MockQuestionnaireTemplateQuestionRepo{}

	svc := &NotificationModifierService{
		DB:                                mockDB,
		QuestionnaireTemplateRepo:         questionnaireTemplateRepo,
		QuestionnaireTemplateQuestionRepo: questionnaireTemplateQuestionRepo,
	}

	ctx := context.Background()

	questionnaireTemplatePb := utils.GetSampleQuestionnaireTemplate()

	questionnaireTemplateEnt, _ := mappers.PbToQuestionnaireTemplateEnt(questionnaireTemplatePb)
	questionnaireTemplateQuestionEnt, _ := mappers.PbToQuestionnaireTemplateQuestionEnts(questionnaireTemplatePb)

	t.Run("happy case", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		questionnaireTemplateRepo.
			On("Upsert", ctx, mockTx, questionnaireTemplateEnt).
			Once().
			Return(nil)
		questionnaireTemplateRepo.
			On("CheckIsExistNameAndType", ctx, mockDB, mock.Anything).
			Once().
			Return(false, nil)
		questionnaireTemplateQuestionRepo.On("SoftDelete", ctx, mockTx, mock.Anything).Once().Return(nil)
		questionnaireTemplateQuestionRepo.On("BulkForceUpsert", ctx, mockTx, questionnaireTemplateQuestionEnt).Once().Return(nil)

		res, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
			QuestionnaireTemplate: questionnaireTemplatePb,
		})

		assert.Equal(t, questionnaireTemplatePb.QuestionnaireTemplateId, res.QuestionnaireTemplateId)
		assert.Nil(t, err)
	})

	t.Run("upsert questionnaire template error", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)
		mockTx.On("Rollback", mock.Anything).Return(pgx.ErrNoRows)

		questionnaireTemplateRepo.
			On("Upsert", ctx, mockTx, questionnaireTemplateEnt).
			Once().
			Return(pgx.ErrNoRows)
		questionnaireTemplateRepo.
			On("CheckIsExistNameAndType", ctx, mockDB, mock.Anything).
			Once().
			Return(false, nil)
		questionnaireTemplateQuestionRepo.On("SoftDelete", ctx, mockTx, mock.Anything).Once().Return(nil)
		questionnaireTemplateQuestionRepo.On("BulkForceUpsert", ctx, mockTx, questionnaireTemplateQuestionEnt).Once().Return(nil)

		res, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
			QuestionnaireTemplate: questionnaireTemplatePb,
		})

		assert.Nil(t, res)
		assert.Equal(t, status.Error(codes.Internal, fmt.Sprintf("svc.QuestionnaireTemplateRepo.Upsert: %v", pgx.ErrNoRows)), err)
	})

	t.Run("questionnaire template name is exist", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		questionnaireTemplateRepo.
			On("CheckIsExistNameAndType", ctx, mockDB, mock.Anything).
			Once().
			Return(true, nil)
		questionnaireTemplateRepo.
			On("Upsert", ctx, mockTx, questionnaireTemplateEnt).
			Once().
			Return(nil)
		questionnaireTemplateQuestionRepo.On("SoftDelete", ctx, mockTx, mock.Anything).Once().Return(nil)
		questionnaireTemplateQuestionRepo.On("BulkForceUpsert", ctx, mockTx, questionnaireTemplateQuestionEnt).Once().Return(nil)

		res, err := svc.UpsertQuestionnaireTemplate(ctx, &npb.UpsertQuestionnaireTemplateRequest{
			QuestionnaireTemplate: questionnaireTemplatePb,
		})

		assert.Nil(t, res)
		assert.Equal(t, status.Error(codes.InvalidArgument, "questionnaire template name is exist"), err)
	})
}
