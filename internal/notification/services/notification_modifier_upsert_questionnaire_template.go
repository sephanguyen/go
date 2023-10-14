package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) UpsertQuestionnaireTemplate(ctx context.Context, req *npb.UpsertQuestionnaireTemplateRequest) (*npb.UpsertQuestionnaireTemplateResponse, error) {
	req.QuestionnaireTemplate.Name = strings.TrimSpace(req.QuestionnaireTemplate.Name)

	if req.QuestionnaireTemplate.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "questionnaire template name is empty")
	}

	filter := repositories.NewCheckTemplateNameFilter()
	err := multierr.Combine(
		filter.TemplateID.Set(req.QuestionnaireTemplate.QuestionnaireTemplateId),
		filter.Name.Set(req.QuestionnaireTemplate.Name),
		filter.Type.Set(npb.QuestionnaireTemplateType_QUESTION_TEMPLATE_TYPE_DEFAULT.String()),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	isExist, err := svc.QuestionnaireTemplateRepo.CheckIsExistNameAndType(ctx, svc.DB, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if isExist {
		return nil, status.Error(codes.InvalidArgument, "questionnaire template name is exist")
	}

	if req.QuestionnaireTemplate.QuestionnaireTemplateId == "" {
		req.QuestionnaireTemplate.QuestionnaireTemplateId = idutil.ULIDNow()
	}

	questionnaireTemplate, err := mappers.PbToQuestionnaireTemplateEnt(req.QuestionnaireTemplate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot convert PbToQuestionnaireTemplateEnt")
	}

	questionnaireTemplateQuestions, err := mappers.PbToQuestionnaireTemplateQuestionEnts(req.QuestionnaireTemplate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot convert PbToQuestionnaireTemplateQuestionEnts")
	}

	err = database.ExecInTxWithRetry(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = svc.QuestionnaireTemplateRepo.Upsert(ctx, tx, questionnaireTemplate)
		if err != nil {
			return fmt.Errorf("svc.QuestionnaireTemplateRepo.Upsert: %v", err)
		}

		err = svc.QuestionnaireTemplateQuestionRepo.SoftDelete(ctx, tx, []string{req.QuestionnaireTemplate.QuestionnaireTemplateId})
		if err != nil {
			return fmt.Errorf("svc.QuestionnaireTemplateQuestionRepo.SoftDelete: %v", err)
		}

		for _, question := range questionnaireTemplateQuestions {
			_ = question.QuestionnaireTemplateID.Set(req.QuestionnaireTemplate.QuestionnaireTemplateId)
		}

		err = svc.QuestionnaireTemplateQuestionRepo.BulkForceUpsert(ctx, tx, questionnaireTemplateQuestions)
		if err != nil {
			return fmt.Errorf("svc.QuestionnaireTemplateQuestionRepo.BulkForceUpsert: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &npb.UpsertQuestionnaireTemplateResponse{
		QuestionnaireTemplateId: req.QuestionnaireTemplate.QuestionnaireTemplateId,
	}, nil
}
