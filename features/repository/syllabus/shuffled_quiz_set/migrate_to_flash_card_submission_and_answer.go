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
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

func (s *Suite) userCreateAStudyPlanOfFlashcardToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.aValidFlashcardInDatabase(ctx)
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

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userSubmittedWithNumberOfFlashcardAnswers(ctx context.Context, numOfAnswers int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	answersEnt := make([]*entities.QuizAnswer, 0)
	if numOfAnswers == 0 {
		return utils.StepStateToContext(ctx, stepState), nil
	}
	for _, quiz := range stepState.Quizzes {
		if len(answersEnt) < numOfAnswers {
			answer := &entities.QuizAnswer{
				QuizID:        quiz.ExternalID.String,
				QuizType:      quiz.Kind.String,
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
	}
	shuffledQuizSetRepo := &repositories.ShuffledQuizSetRepo{}
	if err := shuffledQuizSetRepo.UpdateSubmissionHistory(ctx, s.DB, database.Text(stepState.ShuffledQuizSetID), database.JSONB(answersEnt)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to UpdateSubmissionHistory shuffled quiz set: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) aValidFlashcardInDatabase(ctx context.Context) (context.Context, error) {
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

	flashcard := &entities.Flashcard{}
	database.AllNullEntity(flashcard)
	if err := multierr.Combine(
		flashcard.CreatedAt.Set(time.Now()),
		flashcard.UpdatedAt.Set(time.Now()),
		flashcard.ID.Set(stepState.LoID),
		flashcard.TopicID.Set(stepState.TopicID),
		flashcard.Name.Set(fmt.Sprintf("flashcard-%s", idutil.ULIDNow())),
		flashcard.Type.Set(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup flashcard: %w", err)
	}

	flashcardRepo := repositories.FlashcardRepo{}
	if err := flashcardRepo.Insert(ctx, s.DB, flashcard); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create flashcard: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateAShuffleQuizSetForFlashcardWithNumberOfQuizzes(ctx context.Context, numOfQuizzes int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studentID := idutil.ULIDNow()
	ctx, err := s.aValidUserInDB(ctx, s.DBTrace, studentID, constant.RoleStudent, constant.RoleStudent)
	if err != nil {
		return ctx, err
	}

	// quizzes
	stepState.Quizzes = entities.Quizzes{}
	extIDs := []string{}
	for i := 0; i < numOfQuizzes; i++ {
		extID := idutil.ULIDNow()

		quiz := &entities.Quiz{}
		database.AllNullEntity(quiz)
		if err := multierr.Combine(
			quiz.ExternalID.Set(extID),
			quiz.Country.Set(cpb.Country_name[int32(cpb.Country_COUNTRY_VN)]),
			quiz.SchoolID.Set(constants.ManabieSchool),
			quiz.Kind.Set(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)]),
			quiz.Question.Set(`{"raw":"{\"blocks\":[{\"key\":\"2bsgi\",\"text\":\"qeqweqewq\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/150cb1b73bc9d3bbe4011a55476a6913.html"}`),
			quiz.Explanation.Set(`{"raw":"{\"blocks\":[{\"key\":\"f5lms\",\"text\":\"\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/24061416a35eb51f403307148c5f4cef.html"}`),
			quiz.Options.Set(`[
				{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},
				{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},
				{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
			]`),
			quiz.CreatedBy.Set("repo testing"),
			quiz.ApprovedBy.Set("repo testing"),
			quiz.Status.Set(cpb.QuizStatus_name[int32(cpb.QuizStatus_QUIZ_STATUS_APPROVED)]),
			quiz.DifficultLevel.Set(rand.Intn(5)+1),
			quiz.Point.Set(rand.Intn(5)+1),
		); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup quiz: %w", err)
		}

		stepState.Quizzes = append(stepState.Quizzes, quiz)
		extIDs = append(extIDs, extID)
	}

	quizRepo := &repositories.QuizRepo{}
	for _, quiz := range stepState.Quizzes {
		if err := quizRepo.Create(ctx, s.DB, quiz); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create quiz: %w", err)
		}
	}

	// shuffled_quiz_sets
	stepState.ShuffledQuizSetID = idutil.ULIDNow()
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	database.AllNullEntity(shuffledQuizSet)
	if err := multierr.Combine(
		shuffledQuizSet.ID.Set(stepState.ShuffledQuizSetID),
		shuffledQuizSet.StudentID.Set(studentID),
		shuffledQuizSet.StudyPlanItemID.Set(stepState.StudyPlanItemID),
		shuffledQuizSet.TotalCorrectness.Set(1),
		shuffledQuizSet.SubmissionHistory.Set(database.JSONB([]*entities.QuizAnswer{})),
		shuffledQuizSet.QuizExternalIDs.Set(extIDs),
		shuffledQuizSet.OriginalQuizSetID.Set(idutil.ULIDNow()),
		shuffledQuizSet.Status.Set(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
		shuffledQuizSet.CreatedAt.Set(time.Now()),
		shuffledQuizSet.UpdatedAt.Set(time.Now()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup shuffled quiz set: %w", err)
	}

	shuffledQuizSetRepo := &repositories.ShuffledQuizSetRepo{}
	if _, err := shuffledQuizSetRepo.Create(ctx, s.DB, shuffledQuizSet); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create shuffled quiz set: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) databaseHasOfRecordInFlashcardSubmissionAndAnswerInFlashcardSubmissionAnswer(ctx context.Context, numOfSubmissions int, numOfRecordAnswers int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var total int

	if err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM flash_card_submission WHERE shuffled_quiz_set_id = $1;`,
		database.Text(stepState.ShuffledQuizSetID),
	).Scan(&total); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
	}
	if total != numOfSubmissions {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("number of submission not equal, expected: %d, get: %d", numOfSubmissions, total)
	}

	if err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM flash_card_submission_answer WHERE shuffled_quiz_set_id = $1;`,
		database.Text(stepState.ShuffledQuizSetID),
	).Scan(&total); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
	}
	if total != numOfRecordAnswers {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("number of submission answer not equal, expected: %d, get: %d", numOfRecordAnswers, total)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
