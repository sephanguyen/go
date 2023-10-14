package eureka

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
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

func (s *suite) studentChooseOptionOfTheQuizForSubmitQuizAnswers(ctx context.Context, selectedIdxsStr, quizIdxStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.SelectedIndex == nil {
		stepState.SelectedIndex = make(map[string]map[string][]*epb.Answer)
	}

	if stepState.SelectedIndex[stepState.SetID] == nil {
		stepState.SelectedIndex[stepState.SetID] = make(map[string][]*epb.Answer)
	}
	selectedIdxs := make([]*epb.Answer, 0)

	for _, idx := range strings.Split(selectedIdxsStr, ",") {
		i, _ := strconv.Atoi(strings.TrimSpace(idx))
		selectedIdxs = append(selectedIdxs, &epb.Answer{Format: &epb.Answer_SelectedIndex{SelectedIndex: uint32(i)}})
	}

	quizIndex, _ := strconv.Atoi(quizIdxStr)
	quizIndex--
	stepState.SelectedIndex[stepState.SetID][stepState.QuizItems[quizIndex].Core.ExternalId] = selectedIdxs
	stepState.SelectedQuiz = append(stepState.SelectedQuiz, quizIndex)

	stepState.QuizAnswers = append(stepState.QuizAnswers, &epb.QuizAnswer{
		QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
		Answer: selectedIdxs,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewCourseModifierServiceClient(s.Conn).SubmitQuizAnswers(s.signedCtx(ctx), &epb.SubmitQuizAnswersRequest{
		SetId:      stepState.SetID,
		QuizAnswer: stepState.QuizAnswers,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsExpectedResultMultipleChoiceTypeForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.SubmitQuizAnswersResponse)
	for _, log := range resp.Logs {
		idx := -1
		for i, quiz := range stepState.QuizItems {
			if quiz.Core.ExternalId == log.QuizId {
				idx = i
				break
			}
		}

		ctx, err := s.checkResultMultipleChoiceTypeWithArgs(ctx, idx, log.Correctness)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateAStudyPlanOfExamLoToDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.aValidExamLOInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aValidStudyPlanInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aValidStudyPlanItemInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidExamLOInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.LoID = idutil.ULIDNow()
	stepState.TopicID = idutil.ULIDNow()

	topic := &entities.Topic{}
	database.AllNullEntity(topic)
	if err := multierr.Combine(
		topic.SchoolID.Set(constants.ManabieSchool),
		topic.ID.Set(stepState.TopicID),
		topic.ChapterID.Set(idutil.ULIDNow()),
		topic.Name.Set(fmt.Sprintf("topic-%s", idutil.ULIDNow())),
		topic.Grade.Set(rand.Intn(5)+1),
		topic.Subject.Set(epb.Subject_SUBJECT_NONE),
		topic.Status.Set(epb.TopicStatus_TOPIC_STATUS_NONE),
		topic.CreatedAt.Set(time.Now()),
		topic.UpdatedAt.Set(time.Now()),
		topic.TotalLOs.Set(1),
		topic.TopicType.Set(epb.TopicType_TOPIC_TYPE_EXAM),
		topic.EssayRequired.Set(true),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup topic: %w", err)
	}

	topicRepo := repositories.TopicRepo{}
	if err := topicRepo.BulkUpsertWithoutDisplayOrder(ctx, s.DB, []*entities.Topic{topic}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat topic: %w", err)
	}

	examLO := &entities.ExamLO{}
	database.AllNullEntity(examLO)
	if err := multierr.Combine(
		examLO.CreatedAt.Set(time.Now()),
		examLO.UpdatedAt.Set(time.Now()),
		examLO.ID.Set(stepState.LoID),
		examLO.TopicID.Set(stepState.TopicID),
		examLO.Name.Set(fmt.Sprintf("exam lo-%s", idutil.ULIDNow())),
		examLO.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
		examLO.ApproveGrading.Set(false),
		examLO.GradeCapping.Set(true),
		examLO.ReviewOption.Set(sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()),
		examLO.SetDefaultVendorType(),
		examLO.IsPublished.Set(false),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup exam lo: %w", err)
	}

	examLORepo := repositories.ExamLORepo{}
	if err := examLORepo.BulkInsert(ctx, s.DB, []*entities.ExamLO{examLO}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat exam lo: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudyPlanInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudyPlanID = idutil.ULIDNow()
	studyplan := &entities.StudyPlan{}
	database.AllNullEntity(studyplan)
	now := timeutil.Now()
	if err := multierr.Combine(studyplan.ID.Set(stepState.StudyPlanID),
		studyplan.Name.Set(fmt.Sprintf("StudyPlan_name+%s", stepState.StudyPlanID)),
		studyplan.StudyPlanType.Set(fmt.Sprintf("%d", 2)),
		studyplan.SchoolID.Set(int32(1)),
		studyplan.CourseID.Set(idutil.ULIDNow()),
		studyplan.BookID.Set(idutil.ULIDNow()),
		studyplan.Status.Set(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE),
		studyplan.TrackSchoolProgress.Set(true),
		studyplan.Grades.Set(int32(1)),
		studyplan.UpdatedAt.Set(now),
		studyplan.CreatedAt.Set(now),
		studyplan.MasterStudyPlan.Set(stepState.StudyPlanID)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup StudyPlan: %w", err)
	}
	studyplanRepo := repositories.StudyPlanRepo{}
	if err := studyplanRepo.BulkUpsert(ctx, s.DB, []*entities.StudyPlan{studyplan}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat StudyPlan: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudyPlanItemInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudyPlanItemID = idutil.ULIDNow()
	studyplanItem := &entities.StudyPlanItem{}
	database.AllNullEntity(studyplanItem)
	now := timeutil.Now()
	if err := multierr.Combine(
		studyplanItem.ID.Set(stepState.StudyPlanItemID),
		studyplanItem.StudyPlanID.Set(stepState.StudyPlanID),
		studyplanItem.AvailableFrom.Set(now),
		studyplanItem.AvailableTo.Set(now),
		studyplanItem.StartDate.Set(now),
		studyplanItem.EndDate.Set(now),
		studyplanItem.CompletedAt.Set(now),
		studyplanItem.ContentStructure.Set(entities.ContentStructure{
			CourseID:  idutil.ULIDNow(),
			BookID:    idutil.ULIDNow(),
			ChapterID: idutil.ULIDNow(),
			TopicID:   stepState.TopicID,
			LoID:      stepState.LoID,
		}),
		studyplanItem.ContentStructureFlatten.Set("a"),
		studyplanItem.DisplayOrder.Set(0),
		studyplanItem.CopyStudyPlanItemID.Set(stepState.StudyPlanItemID),
		studyplanItem.Status.Set(epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE),
		studyplanItem.SchoolDate.Set(now),

		studyplanItem.UpdatedAt.Set(now),
		studyplanItem.CreatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup StudyPlanItem: %w", err)
	}
	studyplanItemRepo := repositories.StudyPlanItemRepo{}
	if err := studyplanItemRepo.BulkInsert(ctx, s.DB, []*entities.StudyPlanItem{studyplanItem}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat StudyPlanItem: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
