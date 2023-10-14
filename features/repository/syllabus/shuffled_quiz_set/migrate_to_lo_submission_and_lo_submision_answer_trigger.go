package shuffled_quiz_set

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

func (s *Suite) aValidLearningObjectiveInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
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
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topic: %w", err)
	}

	lo := &entities.LearningObjectiveV2{}
	database.AllNullEntity(lo)
	if err := multierr.Combine(
		lo.CreatedAt.Set(time.Now()),
		lo.UpdatedAt.Set(time.Now()),
		lo.ID.Set(stepState.LoID),
		lo.TopicID.Set(stepState.TopicID),
		lo.Name.Set(fmt.Sprintf("lo-name-%s", stepState.LoID)),
		lo.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup lo: %w", err)
	}

	loRepo := repositories.LearningObjectiveRepoV2{}
	if err := loRepo.Insert(ctx, s.DB, lo); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat lo: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aStudyPlanOfLOInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, err := s.aValidStudyPlanInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aValidStudyPlanItemInDatabase(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) studentCreateAValidShuffleQuizSetWithQuizzesForLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studentID := idutil.ULIDNow()
	ctx, err := s.aValidUserInDB(ctx, s.DBTrace, studentID, constant.RoleStudent, constant.RoleStudent)
	if err != nil {
		return ctx, err
	}
	// quizzes
	numOfQuizzes := rand.Intn(5) + 5
	stepState.Quizzes = entities.Quizzes{}
	extIDs := []string{}
	for i := 0; i < numOfQuizzes; i++ {
		extID := idutil.ULIDNow()
		now := timeutil.Now()
		quiz := &entities.Quiz{}
		database.AllNullEntity(quiz)
		if err := multierr.Combine(
			quiz.ID.Set(idutil.ULIDNow()),
			quiz.ExternalID.Set(extID),
			quiz.Country.Set(cpb.Country_COUNTRY_VN.String()),
			quiz.SchoolID.Set(constants.ManabieSchool),
			quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_MCQ.String()),
			quiz.Question.Set(`{"raw":"{\"blocks\":[{\"key\":\"2bsgi\",\"text\":\"qeqweqewq\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/150cb1b73bc9d3bbe4011a55476a6913.html"}`),
			quiz.Explanation.Set(`{"raw":"{\"blocks\":[{\"key\":\"f5lms\",\"text\":\"\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/24061416a35eb51f403307148c5f4cef.html"}`),
			quiz.Options.Set(`[
			{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},
			{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},
			{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
		]`),
			quiz.CreatedBy.Set("repo testing"),
			quiz.ApprovedBy.Set("repo testing"),
			quiz.Status.Set(cpb.QuizStatus_QUIZ_STATUS_APPROVED.String()),
			quiz.DifficultLevel.Set(rand.Intn(5)+1),
			quiz.Point.Set(rand.Intn(5)+1),
			quiz.CreatedAt.Set(now),
			quiz.UpdatedAt.Set(now),
		); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup quiz: %w", err)
		}

		stepState.Quizzes = append(stepState.Quizzes, quiz)
		extIDs = append(extIDs, extID)
	}

	quizRepo := &repositories.QuizRepo{}
	if _, err := quizRepo.Upsert(ctx, s.DB, stepState.Quizzes); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create quizzes: %w", err)
	}

	// shuffled_quiz_sets
	stepState.ShuffledQuizSetID = idutil.ULIDNow()
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	database.AllNullEntity(shuffledQuizSet)
	now := timeutil.Now()
	if err := multierr.Combine(
		shuffledQuizSet.ID.Set(stepState.ShuffledQuizSetID),
		shuffledQuizSet.StudentID.Set(studentID),
		shuffledQuizSet.StudyPlanItemID.Set(stepState.StudyPlanItemID),
		shuffledQuizSet.TotalCorrectness.Set(1),
		shuffledQuizSet.SubmissionHistory.Set(database.JSONB([]*entities.QuizAnswer{})),
		shuffledQuizSet.QuizExternalIDs.Set(extIDs),
		shuffledQuizSet.OriginalQuizSetID.Set(idutil.ULIDNow()),
		shuffledQuizSet.Status.Set(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
		shuffledQuizSet.CreatedAt.Set(now),
		shuffledQuizSet.UpdatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup shuffled quiz set: %w", err)
	}
	shuffledQuizSetRepo := &repositories.ShuffledQuizSetRepo{}
	if _, err := shuffledQuizSetRepo.Create(ctx, s.DB, shuffledQuizSet); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create shuffled quiz set: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) databaseMustHaveNumOfSubmissionRecordInLOSubmissionTableCorrectly(ctx context.Context, numOfLOSubmission int, numOfAnswer int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var (
		loSubmissionCount,
		loSubmissionAnswerCount int
	)
	if err := s.DB.QueryRow(ctx, `
		SELECT count(1) FROM lo_submission WHERE shuffled_quiz_set_id = $1;`,
		database.Text(stepState.ShuffledQuizSetID),
	).Scan(&loSubmissionCount); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
	}
	if loSubmissionCount != numOfLOSubmission {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected number of lo submission records %d, got %d", numOfLOSubmission, loSubmissionCount)
	}

	if err := s.DB.QueryRow(ctx, `
		SELECT count(1) FROM lo_submission_answer WHERE shuffled_quiz_set_id = $1;`,
		database.Text(stepState.ShuffledQuizSetID),
	).Scan(&loSubmissionAnswerCount); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
	}
	if loSubmissionAnswerCount != numOfAnswer {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected number of lo submission answer records  %d, got %d", numOfAnswer, loSubmissionAnswerCount)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentSubmitWithNumberOfAnswers(ctx context.Context, numOfAnswers int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	answersEnt := make([]*entities.QuizAnswer, 0)
	if numOfAnswers == 0 {
		return utils.StepStateToContext(ctx, stepState), nil
	}
	for i := 0; i < numOfAnswers; i++ {
		answer := &entities.QuizAnswer{
			QuizID:        stepState.Quizzes[i].ExternalID.String,
			QuizType:      stepState.Quizzes[i].Kind.String,
			FilledText:    []string{"text 1", "text 2"},
			CorrectText:   []string{"text 1", "text 3"},
			SelectedIndex: []uint32{1, 2},
			CorrectIndex:  []uint32{1, 3},
			Correctness:   []bool{true, false},
			IsAccepted:    true,
			IsAllCorrect:  false,
			SubmittedAt:   time.Now(),
		}
		answersEnt = append(answersEnt, answer)
	}
	shuffledQuizSetRepo := &repositories.ShuffledQuizSetRepo{}
	if err := shuffledQuizSetRepo.UpdateSubmissionHistory(ctx, s.DB, database.Text(stepState.ShuffledQuizSetID), database.JSONB(answersEnt)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to UpdateSubmissionHistory shuffled quiz set: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
