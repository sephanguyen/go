package quiz

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

	courseStudents, err := utils.AValidCourseWithIDs(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, stepState.StudentIDs, stepState.CourseID)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}
	stepState.CourseStudents = courseStudents

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
	stepState.LearningMaterialID = resp.LoIds[0]
	stepState.LoIDs = append(stepState.LoIDs, resp.LoIds...)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreatesAQuizInType(ctx context.Context, quizType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var vQuizType cpb.QuizType
	switch quizType {
	case "multiple choice":
		vQuizType = cpb.QuizType_QUIZ_TYPE_MCQ
	case "multiple answer":
		vQuizType = cpb.QuizType_QUIZ_TYPE_MAQ
	case "manual input":
		vQuizType = cpb.QuizType_QUIZ_TYPE_MIQ
	case "fill in blank":
		vQuizType = cpb.QuizType_QUIZ_TYPE_FIB
	case "pair of word":
		quizIDs, err := utils.GenerateUpsertFlashcardContent(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.LearningMaterialID, 1, s.EurekaConn)
		if err != nil {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateUpsertQuizV2: %w", err)
		}
		stepState.ExternalIDs = quizIDs
		return utils.StepStateToContext(ctx, stepState), nil
	case "term and definition":
		vQuizType = cpb.QuizType_QUIZ_TYPE_TAD
	case "order":
		vQuizType = cpb.QuizType_QUIZ_TYPE_ORD
	}

	resp, err := utils.GenerateUpsertSingleQuiz(s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.LearningMaterialID, vQuizType, 1, 1, s.EurekaConn)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("GenerateUpsertSingleQuiz: %w", err)
	}

	stepState.ExternalIDs = resp.ExternalIDs
	stepState.TotalPoint = resp.TotalPoint

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdatesStudyPlanForTheLearningMaterial(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	// Find master study plan Items
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	// Find child study plan item
	for _, masterStudyPlanItem := range masterStudyPlanItems {
		childStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).RetrieveChildStudyPlanItem(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaDB, masterStudyPlanItem.ID, database.TextArray(stepState.StudentIDs))
		if err != nil {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error retrieve child study plan item")
		}
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
		for _, childStudyPlanItem := range childStudyPlanItems {
			stepState.StudyPlanItems = append(stepState.StudyPlanItems, childStudyPlanItem)
			stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, childStudyPlanItem.ID.String)
		}
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for i := 0; i < len(stepState.LoIDs); i++ {
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
				stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
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
	}
	resp, err := epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), upsertSpiReq)
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("NewStudyPlanModifierServiceClient.UpsertStudyPlanItemV2: %w", err)
	}

	stepState.StudyPlanItemID = resp.StudyPlanItemIds[0]

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userStartsAndSubmitsAAnswerInKind(ctx context.Context, content string, kind string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	rand.Seed(time.Now().UnixNano())
	stepState.SessionID = idutil.ULIDNow()
	resp, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&epb.CreateQuizTestRequest{
			LoId:            stepState.LearningMaterialID,
			StudentId:       stepState.StudentIDs[0],
			StudyPlanItemId: stepState.StudyPlanItemID,
			SessionId:       stepState.SessionID,
			Paging: &cpb.Paging{
				Limit: uint32(5),
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 1,
				},
			},
		})
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}
	stepState.ShuffledQuizSetID = resp.QuizzesId

	answer := make([]*sspb.Answer, 0)
	switch kind {
	case "select":
		for _, idx := range strings.Split(content, ",") {
			i, _ := strconv.Atoi(strings.TrimSpace(idx))
			answer = append(answer, &sspb.Answer{Format: &sspb.Answer_SelectedIndex{SelectedIndex: uint32(i)}})
		}
	case "text":
		for _, text := range strings.Split(content, ",") {
			text = strings.TrimSpace(text)
			answer = append(answer, &sspb.Answer{Format: &sspb.Answer_FilledText{FilledText: text}})
		}
	case "order":
		for _, order := range strings.Split(content, ",") {
			order = strings.TrimSpace(order)
			answer = append(answer, &sspb.Answer{Format: &sspb.Answer_SubmittedKey{SubmittedKey: order}})
		}
	}

	checkResp, err := sspb.NewQuizClient(s.EurekaConn).CheckQuizCorrectness(s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&sspb.CheckQuizCorrectnessRequest{
			ShuffledQuizSetId: stepState.ShuffledQuizSetID,
			QuizId:            stepState.ExternalIDs[0],
			Answer:            answer,
			LmType:            stepState.LearningMaterialType,
		})
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}
	stepState.Response = checkResp

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnsCorrectnessAndIscorrectallCorrectly(ctx context.Context, correctness string, isCorrectAll string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if correctness != "" && isCorrectAll != "" {
		resp := stepState.Response.(*sspb.CheckQuizCorrectnessResponse)

		for i, item := range strings.Split(correctness, ",") {
			value, _ := strconv.ParseBool(item)
			if value != resp.Correctness[i] {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong Correctness, expected: %v, got: %v", value, resp.Correctness[i-1])
			}
		}

		vIsCorrectAll, _ := strconv.ParseBool(isCorrectAll)
		if vIsCorrectAll != resp.IsCorrectAll {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong IsCorrectAll, expected: %v, got: %v", vIsCorrectAll, resp.IsCorrectAll)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userRetryAndSubmitsAAnswerInKind(ctx context.Context, content string, kind string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.SessionID = idutil.ULIDNow()
	resp, err := sspb.NewQuizClient(s.EurekaConn).CreateRetryQuizTestV2(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.CreateRetryQuizTestV2Request{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        stepState.StudyPlanID,
			LearningMaterialId: stepState.LearningMaterialID,
			StudentId:          wrapperspb.String(stepState.StudentIDs[0]),
		},
		ShuffleQuizSetId: wrapperspb.String(stepState.ShuffledQuizSetID),
		SessionId:        stepState.SessionID,
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	})
	if err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}
	stepState.ShuffledQuizSetID = resp.ShuffleQuizSetId

	answer := make([]*sspb.Answer, 0)
	switch kind {
	case "select":
		for _, idx := range strings.Split(content, ",") {
			i, _ := strconv.Atoi(strings.TrimSpace(idx))
			answer = append(answer, &sspb.Answer{Format: &sspb.Answer_SelectedIndex{SelectedIndex: uint32(i)}})
		}
	case "text":
		for _, text := range strings.Split(content, ",") {
			text = strings.TrimSpace(text)
			answer = append(answer, &sspb.Answer{Format: &sspb.Answer_FilledText{FilledText: text}})
		}
	}

	stepState.Response, stepState.ResponseErr = sspb.NewQuizClient(s.EurekaConn).CheckQuizCorrectness(s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&sspb.CheckQuizCorrectnessRequest{
			ShuffledQuizSetId: stepState.ShuffledQuizSetID,
			QuizId:            stepState.ExternalIDs[0],
			Answer:            answer,
			LmType:            stepState.LearningMaterialType,
		})

	return utils.StepStateToContext(ctx, stepState), nil
}
