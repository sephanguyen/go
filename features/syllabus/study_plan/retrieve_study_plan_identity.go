package study_plan

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) hasCreatedAStudyPlanForStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.StudentID = idutil.ULIDNow()

	studyPlanID, err := utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = studyPlanID

	studentStudyPlan := &entities.StudentStudyPlan{
		StudentID: database.Text(stepState.StudentID),
		BaseEntity: entities.BaseEntity{
			CreatedAt: database.Timestamptz(time.Now()),
			UpdatedAt: database.Timestamptz(time.Now()),
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		},
		StudyPlanID:       database.Text(stepState.StudyPlanID),
		MasterStudyPlanID: pgtype.Text{Status: pgtype.Null},
	}
	err = (&repositories.StudentStudyPlanRepo{}).BulkUpsert(ctx, s.EurekaDB, []*entities.StudentStudyPlan{studentStudyPlan})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert student study plan: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnStudyPlanIdentityCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp, ok := stepState.Response.(*sspb.RetrieveStudyPlanIdentityResponse)
	if !ok {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to cast response to RetrieveStudyPlanIdentityResponse")
	}

	if len(resp.StudyPlanIdentities) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected %d study plan items, got %d", len(stepState.LearningMaterialIDs), len(resp.StudyPlanIdentities))
	}

	for _, studyPlanIdentity := range resp.StudyPlanIdentities {
		if studyPlanIdentity.StudyPlanId != stepState.StudyPlanID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected study plan id %s, got %s", stepState.StudyPlanID, studyPlanIdentity.StudyPlanId)
		}

		if studyPlanIdentity.StudentId != stepState.StudentID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected student id %s, got %s", stepState.StudentID, studyPlanIdentity.StudentId)
		}

		if !golibs.InArrayString(studyPlanIdentity.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected learning material id %s, got %s", stepState.LearningMaterialIDs, studyPlanIdentity.LearningMaterialId)
		}

		if !golibs.InArrayString(studyPlanIdentity.StudyPlanItemId, stepState.StudyPlanItemIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected study plan item id %s, got %s", stepState.StudyPlanItemIDs, studyPlanIdentity.StudyPlanItemId)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) someStudyPlanItemsForTheStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan items: %w", err)
	}

	stepState.LoIDs = nil

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range studyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		_ = item.ContentStructure.AssignTo(cs)

		if len(cse.LoID) != 0 {
			stepState.LoIDs = append(stepState.LoIDs, cse.LoID)
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, cse.LoID)
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, cse.AssignmentID)
			cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}

	_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), upsertSpiReq)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}

	for _, studyPlanItem := range studyPlanItems {
		stepState.StudyPlanItemIDs = append(stepState.StudyPlanItemIDs, studyPlanItem.ID.String)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userRetrievesStudyPlanIdentity(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).RetrieveStudyPlanIdentity(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.RetrieveStudyPlanIdentityRequest{
		StudyPlanItemIds: stepState.StudyPlanItemIDs,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}
