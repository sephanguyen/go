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

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

// nolint

func (s *Suite) userCreateAStudyPlanWithLearningMaterialToDatabase(ctx context.Context, lmType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var err error
	switch lmType {
	case "LO":
		ctx, err = s.validLOInDatabase(ctx)

	case "ExamLO":
		ctx, err = s.validExamLOInDatabase(ctx)

	case "FlashCard":
		ctx, err = s.validFlashcardInDatabase(ctx)

	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unknown lm type %s", lmType)
	}
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

func (s *Suite) validLOInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.LoID = idutil.ULIDNow()

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

func (s *Suite) validExamLOInDatabase(ctx context.Context) (context.Context, error) {
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
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup exam lo: %w", err)
	}

	examLORepo := repositories.ExamLORepo{}
	if err := examLORepo.BulkInsert(ctx, s.DB, []*entities.ExamLO{examLO}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to creat exam lo: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validFlashcardInDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.LoID = idutil.ULIDNow()

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

func (s *Suite) userCreateAShuffleQuizSetForLOsWithNumberOfQuizzes(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studentID := idutil.ULIDNow()
	stepState.StudentID = studentID
	numOfQuizzes := 5
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
			quiz.Point.Set(i),
		); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to setup quiz: %w", err)
		}
		stepState.TotalPoint += i
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

func (s *Suite) userSubmittedWithSomeAnswers(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	answersEnt := make([]*entities.QuizAnswer, 0)
	numOfAnswers := 5
	var flag bool
	for i, quiz := range stepState.Quizzes {
		if len(answersEnt) < numOfAnswers {
			if i%2 == 0 {
				flag = true
				stepState.TotalAnswerTrue += i
			} else {
				flag = false
			}
			answer := &entities.QuizAnswer{
				QuizID:        quiz.ExternalID.String,
				QuizType:      quiz.Kind.String,
				FilledText:    []string{"text 1", "text 2"},
				CorrectText:   []string{"text 1", "text 3"},
				SelectedIndex: []uint32{1, 2},
				CorrectIndex:  []uint32{1, 3},
				Correctness:   []bool{true, false},
				IsAccepted:    flag,
				IsAllCorrect:  false,
				SubmittedAt:   time.Now(),
			}

			answersEnt = append(answersEnt, answer)
			numOfAnswers++
		}
	}
	shuffledQuizSetRepo := &repositories.ShuffledQuizSetRepo{}
	if err := shuffledQuizSetRepo.UpdateSubmissionHistory(ctx, s.DB, database.Text(stepState.ShuffledQuizSetID), database.JSONB(answersEnt)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to UpdateSubmissionHistory shuffled quiz set: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

type CalculateHighestScoreResponse struct {
	StudyPlanItemID pgtype.Text
	Percentage      pgtype.Float4
}

func (s *Suite) userCalculateHighestSubmissionScore(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	query := `
	SELECT sqs.study_plan_item_id,
	max(coalesce(elss.graded_point, (SELECT sum(point) FROM get_submission_history() gsh where gsh.shuffled_quiz_set_id = sqs.shuffled_quiz_set_id  ))::float4 / 
	(SELECT sum(point) FROM public.quizzes q where q.deleted_at IS NULL AND q.external_id = any(sqs.quiz_external_ids))) * 100 as percentage
	FROM shuffled_quiz_sets sqs
	left join exam_lo_submission els on sqs.shuffled_quiz_set_id = els.shuffled_quiz_set_id 
	left join get_exam_lo_scores() elss on els.submission_id = elss.submission_id
	WHERE sqs.study_plan_item_id = ANY($1::_TEXT)
		AND sqs.deleted_at IS NULL
	GROUP BY sqs.study_plan_item_id
	`
	studyPlanItemIDs := make([]string, 0, 1)
	studyPlanItemIDs = append(studyPlanItemIDs, stepState.StudyPlanItemID)
	rows, err := s.DB.Query(ctx, query, studyPlanItemIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Query: %w", err)
	}
	defer rows.Close()
	var res []*CalculateHighestScoreResponse
	for rows.Next() {
		var studyPlanItemID pgtype.Text
		var percentage float32
		if err := rows.Scan(&studyPlanItemID, &percentage); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Scan: %w", err)
		}
		res = append(res, &CalculateHighestScoreResponse{
			StudyPlanItemID: studyPlanItemID,
			Percentage:      database.Float4(percentage),
		})
	}
	percentage := float32(stepState.TotalAnswerTrue) / float32(stepState.TotalPoint) * 100
	if int(res[0].Percentage.Float) != int(percentage) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected percentage want %.2f, got %.2f", percentage, res[0].Percentage.Float)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateAStudyPlanWithExamLOToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.validExamLOInDatabase(ctx)
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

func (s *Suite) userCalculateExamLOHighestSubmissionScore(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	updateQuery := `
	update exam_lo_submission 
	set 
		status = 'SUBMISSION_STATUS_NOT_MARKED'
	where
		shuffled_quiz_set_id = $1::TEXT`
	query := `
	SELECT sqs.study_plan_item_id,
	max(coalesce(elss.graded_point, (SELECT sum(point) FROM get_submission_history() gsh where gsh.shuffled_quiz_set_id = sqs.shuffled_quiz_set_id  ))::float4 /
	(SELECT sum(point) FROM public.quizzes q where q.deleted_at IS NULL AND q.external_id = any(sqs.quiz_external_ids))) * 100 as percentage
	FROM shuffled_quiz_sets sqs
	left join exam_lo_submission els on sqs.shuffled_quiz_set_id = els.shuffled_quiz_set_id
	left join get_exam_lo_returned_scores() elss on els.submission_id = elss.submission_id
	WHERE sqs.study_plan_item_id = ANY($1::_TEXT)
			AND sqs.deleted_at IS NULL
	GROUP BY sqs.study_plan_item_id
	`
	_, err := s.DB.Exec(ctx, updateQuery, stepState.ShuffledQuizSetID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't update status %w", err)
	}

	studyPlanItemIDs := make([]string, 0, 1)
	studyPlanItemIDs = append(studyPlanItemIDs, stepState.StudyPlanItemID)
	rows, err := s.DB.Query(ctx, query, studyPlanItemIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Query: %w", err)
	}
	defer rows.Close()
	var res []*CalculateHighestScoreResponse
	for rows.Next() {
		var studyPlanItemID pgtype.Text
		var percentage float32
		if err := rows.Scan(&studyPlanItemID, &percentage); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Scan: %w", err)
		}
		res = append(res, &CalculateHighestScoreResponse{
			StudyPlanItemID: studyPlanItemID,
			Percentage:      database.Float4(percentage),
		})
	}

	if int(res[0].Percentage.Float) != 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected percentage %.2f", res[0].Percentage.Float)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validStudentsLearningObjectivesCompleteness(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	now := time.Now()
	studentsLearningObjectivesCompleteness := &entities.StudentsLearningObjectivesCompleteness{
		StudentID:               database.Text(stepState.StudentID),
		LoID:                    database.Text(stepState.LoID),
		PresetStudyPlanWeeklyID: database.Text("text"),
		FirstAttemptScore:       database.Int2(1),
		IsFinishedQuiz:          database.Bool(true),
		IsFinishedVideo:         database.Bool(true),
		IsFinishedStudyGuide:    database.Bool(true),
		FirstQuizCorrectness:    database.Float4(1),
		FinishedQuizAt:          database.Timestamptz(now),
		HighestQuizScore:        database.Float4(1),
		CreatedAt:               database.Timestamptz(now),
		UpdatedAt:               database.Timestamptz(now),
	}

	if _, err := database.Insert(ctx, studentsLearningObjectivesCompleteness, s.DBTrace.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert : %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetExamLoScore(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lo := &entities.StudentsLearningObjectivesCompleteness{}
	loE := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	queryFindLos := fmt.Sprintf(`
	SELECT DISTINCT
	sloc.student_id,
	sloc.lo_id, 
	sloc.preset_study_plan_weekly_id,
	sloc.first_attempt_score, 
	sloc.is_finished_quiz,
	sloc.is_finished_video,
	sloc.is_finished_study_guide,
	sloc.first_quiz_correctness,
	sloc.finished_quiz_at,
	coalesce((select max(gelrs.graded_point::float4 / gelrs.total_point) *100 from get_exam_lo_returned_scores() gelrs where sloc.student_id = gelrs.student_id AND sloc.lo_id = gelrs.learning_material_id) , sloc.highest_quiz_score)::float4 as highest_quiz_score,
	sloc.updated_at,
	sloc.created_at
	FROM %s sloc 
		LEFT JOIN %s lo ON sloc.lo_id = lo.lo_id
		LEFT JOIN get_exam_lo_returned_scores() gelrs ON sloc.student_id = gelrs.student_id AND sloc.lo_id = gelrs.learning_material_id
	WHERE sloc.student_id = $1
		AND sloc.lo_id = ANY($2)
		AND lo.deleted_at IS NULL`, lo.TableName(), loE.TableName())

	rows, err := s.DB.Query(ctx, queryFindLos, stepState.StudentID, []string{stepState.LoID})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Query: %w", err)
	}
	defer rows.Close()

	pp := []*entities.StudentsLearningObjectivesCompleteness{}
	for rows.Next() {
		p := new(entities.StudentsLearningObjectivesCompleteness)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Scan: %w", err)
		}
		pp = append(pp, p)
	}

	percentage := float32(stepState.TotalAnswerTrue) / float32(stepState.TotalPoint) * 100
	if int(pp[0].HighestQuizScore.Float) != int(percentage) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected percentage want %.2f, got %.2f", percentage, pp[0].HighestQuizScore.Float)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
