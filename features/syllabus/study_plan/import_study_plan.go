package study_plan // nolint

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userCreatesAValidBookContent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	bookResult, err := utils.GenerateBooksV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), 1, nil, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateBooksV2: %w", err)
	}
	stepState.BookID = bookResult.BookIDs[0]

	chapterResult, err := utils.GenerateChaptersV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.BookID, 1, nil, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateChaptersV2: %w", err)
	}
	stepState.ChapterID = chapterResult.ChapterIDs[0]

	topicResult, err := utils.GenerateTopicsV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.ChapterID, 1, nil, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateTopicsV2: %w", err)
	}
	stepState.TopicID = topicResult.TopicIDs[0]

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreatesACourseAndAddStudentsIntoTheCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateCourse: %w", err)
	}
	stepState.CourseID = courseID

	studentIDs, err := utils.InsertMultiUserIntoBob(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.BobDB, 1)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("InsertMultiUserIntoBob: %w", err)
	}
	stepState.StudentIDs = studentIDs

	_, err = utils.AValidCourseWithIDs(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, stepState.StudentIDs, stepState.CourseID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userAddsAMasterStudyPlanWithTheCreatedBook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studyPlanID, err := utils.GenerateStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateStudyPlan: %w", err)
	}
	stepState.StudyPlanID = studyPlanID

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateALearningMaterialInType(ctx context.Context, lmType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	for i := 0; i < 5; i++ {
		time.Sleep(10 * time.Millisecond)
		lo := utils.GenerateLearningObjective(stepState.TopicID)
		switch lmType {
		case "learning objective":
			stepState.LearningMaterialType = sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE
		case "flash card":
			lo.Type = cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_FLASH_CARD
			lo.Info.Name = fmt.Sprint("flashcard-name+%w", idutil.ULIDNow())
			stepState.LearningMaterialType = sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD
		}
		resp, err := epb.NewLearningObjectiveModifierServiceClient(s.EurekaConn).UpsertLOs(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertLOsRequest{
			LearningObjectives: []*cpb.LearningObjective{
				lo,
			},
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("NewLearningObjectiveModifierServiceClient.UpsertLOs: %w", err)
		}

		stepState.LoIDs = append(stepState.LoIDs, resp.LoIds...)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userBulkUploadCSV(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp, err := sspb.NewStudyPlanClient(s.EurekaConn).ImportStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ImportStudyPlanRequest{
		StudyPlanItems: []*sspb.StudyPlanItemImport{
			{
				StudyPlanId:        stepState.StudyPlanID,
				LearningMaterialId: stepState.LoIDs[0],
				AvailableFrom:      timestamppb.Now(),
			},
			{
				StudyPlanId:        stepState.StudyPlanID,
				LearningMaterialId: stepState.LoIDs[1],
				AvailableTo:        timestamppb.Now(),
			},
			{
				StudyPlanId:        stepState.StudyPlanID,
				LearningMaterialId: stepState.LoIDs[2],
				AvailableFrom:      timestamppb.New(time.Now().Add(-1 * time.Hour)),
				StartDate:          timestamppb.Now(),
				AvailableTo:        timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
			{
				StudyPlanId:        stepState.StudyPlanID,
				LearningMaterialId: stepState.LoIDs[3],
				EndDate:            timestamppb.Now(),
				AvailableFrom:      timestamppb.New(time.Now().Add(-1 * time.Hour)),
				AvailableTo:        timestamppb.New(time.Now().Add(1 * time.Hour)),
			},
		},
	})

	fmt.Printf("%+v %+v\n", resp, err)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) hasCreatedAStudyplanForStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	// Find master study plan Items
	time.Sleep(1 * time.Second)
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	for _, masterStudyPlanItem := range masterStudyPlanItems {
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range stepState.StudyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		err = item.ContentStructure.AssignTo(cs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
		}

		if len(cse.LoID) != 0 {
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
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

	return utils.StepStateToContext(ctx, stepState), nil
}
