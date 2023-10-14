package exam_lo

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/strings/slices"
)

func (s *Suite) userListExamLoSubmissionScores(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.ListExamLOSubmissionScoreRequest{
		SubmissionId:      stepState.SubmissionID,
		ShuffledQuizSetId: stepState.ShuffledQuizSetID,
	}

	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).ListExamLOSubmissionScore(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	stepState.Request = req

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreExamLOSubmissionScoresExisted(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	now := time.Now()
	topicID := idutil.ULIDNow()

	// insert topic
	topic := &entities.Topic{}
	database.AllNullEntity(topic)

	err := multierr.Combine(
		topic.ID.Set(topicID),
		topic.Name.Set("topic-1"),
		topic.Country.Set(cpb.Country_COUNTRY_VN.String()),
		topic.Grade.Set(1),
		topic.Subject.Set(cpb.Subject_SUBJECT_BIOLOGY.String()),
		topic.TopicType.Set(cpb.TopicType_TOPIC_TYPE_NONE.String()),
		topic.TotalLOs.Set(0),
		topic.SchoolID.Set(constants.ManabieSchool),
		topic.CopiedTopicID.Set("copied-topic-id"),
		topic.EssayRequired.Set(false),
		topic.CreatedAt.Set(now),
		topic.UpdatedAt.Set(now))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}

	if _, err := database.Insert(ctx, topic, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert topic: %w", err)
	}

	examLOID := idutil.ULIDNow()
	examLO := &entities.ExamLO{}
	database.AllNullEntity(examLO)
	if err := multierr.Combine(
		examLO.ID.Set(examLOID),
		examLO.TopicID.Set(topicID),
		examLO.Name.Set("exam-lo-1"),
		examLO.CreatedAt.Set(now),
		examLO.UpdatedAt.Set(now),
		examLO.ApproveGrading.Set(false),
		examLO.GradeCapping.Set(true),
		examLO.ReviewOption.Set(sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()),
		examLO.SetDefaultVendorType(),
		examLO.IsPublished.Set(false),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}

	if _, err := database.Insert(ctx, examLO, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert exam_lo: %w", err)
	}

	shuffledQuizSetID := idutil.ULIDNow()
	studentID := idutil.ULIDNow()
	submissionID := idutil.ULIDNow()
	studyPlanID := idutil.ULIDNow()
	examLOSubmission := &entities.ExamLOSubmission{}
	database.AllNullEntity(examLOSubmission)
	if err := multierr.Combine(
		examLOSubmission.SubmissionID.Set(submissionID),
		examLOSubmission.LearningMaterialID.Set(examLOID),
		examLOSubmission.ShuffledQuizSetID.Set(shuffledQuizSetID),
		examLOSubmission.StudyPlanID.Set(studyPlanID),
		examLOSubmission.StudentID.Set(studentID),
		examLOSubmission.TeacherFeedback.Set("teacher-feedback"),
		examLOSubmission.Status.Set(sspb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()),
		examLOSubmission.Result.Set(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String()),
		examLOSubmission.CreatedAt.Set(now),
		examLOSubmission.UpdatedAt.Set(now),
		examLOSubmission.LastAction.Set(sspb.ApproveGradingAction_APPROVE_ACTION_NONE.String()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}

	if _, err := database.Insert(ctx, examLOSubmission, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert exam_lo_submission: %w", err)
	}

	bookID := idutil.ULIDNow()
	courseID := idutil.ULIDNow()
	studyPlan := &entities.StudyPlan{}
	database.AllNullEntity(studyPlan)
	if err := multierr.Combine(
		studyPlan.ID.Set(studyPlanID),
		studyPlan.Name.Set("study-plan-name"),
		studyPlan.BookID.Set(bookID),
		studyPlan.CourseID.Set(courseID),
		studyPlan.SchoolID.Set(constants.ManabieSchool),
		studyPlan.CreatedAt.Set(now),
		studyPlan.UpdatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}
	if _, err := database.Insert(ctx, studyPlan, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert study_plan: %w", err)
	}

	individualStudyPlan := &entities.IndividualStudyPlan{}
	database.AllNullEntity(individualStudyPlan)
	if err := multierr.Combine(
		individualStudyPlan.ID.Set(studyPlanID),
		individualStudyPlan.LearningMaterialID.Set(examLOID),
		individualStudyPlan.StudentID.Set(studentID),
		individualStudyPlan.Status.Set(sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE.String()),
		individualStudyPlan.StartDate.Set(now.Add(-2400*time.Hour)),
		individualStudyPlan.EndDate.Set(now.Add(2400*time.Hour)),
		individualStudyPlan.AvailableFrom.Set(now.Add(-2400*time.Hour)),
		individualStudyPlan.AvailableTo.Set(now.Add(2400*time.Hour)),
		individualStudyPlan.CreatedAt.Set(now),
		individualStudyPlan.UpdatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
	}
	if _, err := database.Insert(ctx, individualStudyPlan, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert individual_study_plan: %w", err)
	}

	stepState.SubmissionID = submissionID
	stepState.ShuffledQuizSetID = shuffledQuizSetID
	stepState.ExamLOSubmissionEnts = append(stepState.ExamLOSubmissionEnts, examLOSubmission)

	ctx, err = s.createSomeTags(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.createSomeTags: %w", err)
	}
	teacherID := idutil.ULIDNow()
	numOfExamLOSubmissionAnswers := 5 + rand.Int31n(5)
	externalQuizIDs := make([]string, 0, numOfExamLOSubmissionAnswers)
	for i := 0; i < int(numOfExamLOSubmissionAnswers); i++ {
		quizID := idutil.ULIDNow()
		externalQuizIDs = append(externalQuizIDs, "external-id-"+quizID)

		var quizOptions, kind string
		num := rand.Int31() % 3 //nolint:gosec
		switch num {
		case 0: // FIB
			kind = cpb.QuizType_QUIZ_TYPE_FIB.String()
			quizRawObj := raw{
				Blocks: []block{
					{
						Key:               "eq20k",
						Text:              "3213",
						Type:              "unstyled",
						Depth:             "0",
						InlineStyleRanges: []string{},
						EntityRanges:      []string{},
						Data:              nil,
					},
				},
			}
			quizRaw, _ := json.Marshal(quizRawObj)

			quizOptionObjs := []*entities.QuizOption{
				{
					Content: entities.RichText{
						Raw:         string(quizRaw),
						RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
					},
					Correctness: false,
					Configs:     []string{},
					Label:       "A",
					Key:         "key-A",
					Attribute:   entities.QuizItemAttribute{},
				},
				{
					Content: entities.RichText{
						Raw:         string(quizRaw),
						RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
					},
					Correctness: true,
					Configs:     []string{},
					Label:       "A",
					Key:         "key-A",
					Attribute:   entities.QuizItemAttribute{},
				},
				{
					Content: entities.RichText{
						Raw:         string(quizRaw),
						RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
					},
					Correctness: true,
					Configs:     []string{},
					Label:       "B",
					Key:         "key-B",
					Attribute:   entities.QuizItemAttribute{},
				},
				{
					Content: entities.RichText{
						Raw:         string(quizRaw),
						RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
					},
					Correctness: false,
					Configs:     []string{},
					Label:       "B",
					Key:         "key-B",
					Attribute:   entities.QuizItemAttribute{},
				},
				{
					Content: entities.RichText{
						Raw:         string(quizRaw),
						RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
					},
					Correctness: true,
					Configs:     []string{},
					Label:       "C",
					Key:         "key-C",
					Attribute:   entities.QuizItemAttribute{},
				},
				{
					Content: entities.RichText{
						Raw:         string(quizRaw),
						RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
					},
					Correctness: false,
					Configs:     []string{},
					Label:       "C",
					Key:         "key-C",
					Attribute:   entities.QuizItemAttribute{},
				},
				{
					Content: entities.RichText{
						Raw:         string(quizRaw),
						RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
					},
					Correctness: false,
					Configs:     []string{},
					Label:       "D",
					Key:         "key-D",
					Attribute:   entities.QuizItemAttribute{},
				},
			}
			quizOptionsBytes, _ := json.Marshal(quizOptionObjs)
			quizOptions = string(quizOptionsBytes)
		case 1: // MAQ
			kind = cpb.QuizType_QUIZ_TYPE_MAQ.String()
			quizOptions = `[
				{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},
				{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},
				{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
			]`
		case 2:
			kind = cpb.QuizType_QUIZ_TYPE_ORD.String()
			quizOptions = `[
				{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true},
				{"key":"key-2" , "label": "2", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so84\",\"text\":\"3213214\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true},
				{"key":"key-3" , "label": "3", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so85\",\"text\":\"3213215\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": false}
				]`
		case 3:
			kind = cpb.QuizType_QUIZ_TYPE_ESQ.String()
			quizOptions = `[
					{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": false, "answer_config": {"essay": {"limit": 10, "limit_type": "ESSAY_LIMIT_TYPE_WORD", "limit_enabled": true}}}
					]`
		}

		quiz := &entities.Quiz{}
		database.AllNullEntity(quiz)
		err := multierr.Combine(
			quiz.ID.Set(quizID),
			quiz.ExternalID.Set("external-id-"+quizID),
			quiz.Country.Set(cpb.Country_COUNTRY_VN.String()),
			quiz.SchoolID.Set(fmt.Sprintf("%d", constants.ManabieSchool)),
			quiz.LoIDs.Set([]string{"lo-id-1", "lo-id-2"}),
			quiz.Kind.Set(kind),
			quiz.Question.Set(`{"raw":"{\"blocks\":[{\"key\":\"2bsgi\",\"text\":\"qeqweqewq\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/150cb1b73bc9d3bbe4011a55476a6913.html"}`),
			quiz.Explanation.Set(`{"raw":"{\"blocks\":[{\"key\":\"f5lms\",\"text\":\"\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/24061416a35eb51f403307148c5f4cef.html"}`),
			quiz.Options.Set(quizOptions),
			quiz.CreatedBy.Set("repo testing"),
			quiz.ApprovedBy.Set("repo testing"),
			quiz.Status.Set(cpb.QuizStatus_name[int32(cpb.QuizStatus_QUIZ_STATUS_APPROVED)]),
			quiz.DifficultLevel.Set(rand.Intn(5)+1),
			quiz.Point.Set(rand.Intn(5)+1),
			quiz.CreatedAt.Set(now),
			quiz.UpdatedAt.Set(now),
			quiz.QuestionTagIds.Set(stepState.QuestionTagIDs),
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
		}

		if stepState.QuestionGroupID != "" {
			err = quiz.QuestionGroupID.Set(stepState.QuestionGroupID)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("questionGroupID.Set: %w", err)
			}
		}

		if _, err := database.Insert(ctx, quiz, s.EurekaDB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert quiz: %w", err)
		}

		if rand.Intn(2) == 1 {
			examLOSubmissionScore := &entities.ExamLOSubmissionScore{}
			database.AllNullEntity(examLOSubmissionScore)
			err := multierr.Combine(
				examLOSubmissionScore.SubmissionID.Set(submissionID),
				examLOSubmissionScore.ShuffledQuizSetID.Set(shuffledQuizSetID),
				examLOSubmissionScore.QuizID.Set("external-id-"+quizID),
				examLOSubmissionScore.TeacherID.Set(teacherID),
				examLOSubmissionScore.TeacherComment.Set(fmt.Sprintf("teacher comment %d", i)),
				examLOSubmissionScore.IsCorrect.Set(true),
				examLOSubmissionScore.IsAccepted.Set(true),
				examLOSubmissionScore.Point.Set(rand.Int31n(11)),
				examLOSubmissionScore.CreatedAt.Set(now),
				examLOSubmissionScore.UpdatedAt.Set(now))
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
			}

			if _, err := database.Insert(ctx, examLOSubmissionScore, s.EurekaDB.Exec); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert exam_lo_submission_score: %w", err)
			}

			stepState.ExamLOSubmissionScoreEnts = append(stepState.ExamLOSubmissionScoreEnts, examLOSubmissionScore)
		}

		examLOSubmissionAnswer := &entities.ExamLOSubmissionAnswer{}
		database.AllNullEntity(examLOSubmissionAnswer)
		err = multierr.Combine(
			examLOSubmissionAnswer.SubmissionID.Set(submissionID),
			examLOSubmissionAnswer.ShuffledQuizSetID.Set(shuffledQuizSetID),
			examLOSubmissionAnswer.QuizID.Set("external-id-"+quizID),
			examLOSubmissionAnswer.StudentID.Set(studentID),
			examLOSubmissionAnswer.LearningMaterialID.Set(examLOID),
			examLOSubmissionAnswer.StudyPlanID.Set(studyPlanID),
			examLOSubmissionAnswer.StudentID.Set(studentID),
			examLOSubmissionAnswer.StudentIndexAnswer.Set([]int32{1, 2, 3}),
			examLOSubmissionAnswer.StudentTextAnswer.Set([]string{"1", "2", "3"}),
			examLOSubmissionAnswer.CorrectKeysAnswer.Set([]string{"key-1", "key-2", "key-3"}),
			examLOSubmissionAnswer.SubmittedKeysAnswer.Set([]string{"key-1", "key-2", "key-3"}),
			examLOSubmissionAnswer.IsCorrect.Set(true),
			examLOSubmissionAnswer.IsAccepted.Set(false),
			examLOSubmissionAnswer.Point.Set(rand.Int31n(11)),
			examLOSubmissionAnswer.CreatedAt.Set(now),
			examLOSubmissionAnswer.UpdatedAt.Set(now),
		)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
		}

		if _, err := database.Insert(ctx, examLOSubmissionAnswer, s.EurekaDB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert exam_lo_submission_answer: %w", err)
		}

		stepState.ExamLOSubmissionAnswerEnts = append(stepState.ExamLOSubmissionAnswerEnts, examLOSubmissionAnswer)
	}

	shuffledQuizSet := &entities.ShuffledQuizSet{}
	database.AllNullEntity(shuffledQuizSet)
	if err := multierr.Combine(
		shuffledQuizSet.ID.Set(shuffledQuizSetID),
		shuffledQuizSet.LearningMaterialID.Set(examLOID),
		shuffledQuizSet.QuizExternalIDs.Set(externalQuizIDs),
		shuffledQuizSet.RandomSeed.Set(strconv.FormatInt(time.Now().UTC().UnixNano(), 10)),
		shuffledQuizSet.TotalCorrectness.Set(1),
		shuffledQuizSet.SubmissionHistory.Set("{}"),
		shuffledQuizSet.StudyPlanID.Set(studyPlanID),
		shuffledQuizSet.StudentID.Set(studentID),
		shuffledQuizSet.CreatedAt.Set(now),
		shuffledQuizSet.UpdatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if _, err := database.Insert(ctx, shuffledQuizSet, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert shuffled_quiz_set: %w", err)
	}

	stepState.StudyPlanItemIdentities = append(stepState.StudyPlanItemIdentities, &sspb.StudyPlanItemIdentity{
		StudyPlanId:        studyPlanID,
		LearningMaterialId: examLOID,
		StudentId:          wrapperspb.String(studentID),
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnsListExamLOSubmissionScoresCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.ListExamLOSubmissionScoreResponse)
	request := stepState.Request.(*sspb.ListExamLOSubmissionScoreRequest)

	if response.TeacherFeedback != stepState.ExamLOSubmissionEnts[0].TeacherFeedback.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected teacher_feedback %s, but got %s", stepState.ExamLOSubmissionEnts[0].TeacherFeedback.String, response.TeacherFeedback)
	}
	if response.SubmissionStatus.String() != stepState.ExamLOSubmissionEnts[0].Status.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected submission_status %s, but got %s", stepState.ExamLOSubmissionEnts[0].Status.String, response.SubmissionStatus.String())
	}
	if response.SubmissionResult.String() != stepState.ExamLOSubmissionEnts[0].Result.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected submission_result %s, but got %s", stepState.ExamLOSubmissionEnts[0].Result.String, response.SubmissionResult.String())
	}
	if response.TotalPoint != nil && response.TotalPoint.Value != uint32(stepState.ExamLOSubmissionEnts[0].TotalPoint.Int) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected total_point %d, but got %d", stepState.ExamLOSubmissionEnts[0].TotalPoint.Int, response.TotalPoint.Value)
	}

	scoresMap := make(map[string]*entities.ExamLOSubmissionScore)
	for _, score := range stepState.ExamLOSubmissionScoreEnts {
		key := fmt.Sprintf("%s-%s-%s", score.ShuffledQuizSetID.String, score.SubmissionID.String, score.QuizID.String)
		scoresMap[key] = score
	}
	answersMap := make(map[string]*entities.ExamLOSubmissionAnswer)
	for _, answer := range stepState.ExamLOSubmissionAnswerEnts {
		key := fmt.Sprintf("%s-%s-%s", answer.ShuffledQuizSetID.String, answer.SubmissionID.String, answer.QuizID.String)
		answersMap[key] = answer
	}

	for _, scoreResp := range response.SubmissionScores {
		key := fmt.Sprintf("%s-%s-%s", scoreResp.ShuffleQuizSetId, request.SubmissionId, scoreResp.Core.ExternalId)
		score := scoresMap[key]
		answer := answersMap[key]

		if scoreResp.ShuffleQuizSetId != answer.ShuffledQuizSetID.String {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected shuffled_quiz_set_id %s, but got %s", score.ShuffledQuizSetID.String, scoreResp.ShuffleQuizSetId)
		}
		if scoreResp.IsAccepted != answer.IsAccepted.Bool {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected is_accepted %t, but got %t", score.IsAccepted.Bool, scoreResp.IsAccepted)
		}

		teacherComment := ""
		gradedPoint := answer.Point.Int
		point := answer.Point.Int
		if score != nil {
			teacherComment = score.TeacherComment.String
			gradedPoint = score.Point.Int
		}

		if scoreResp.TeacherComment != teacherComment {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected teacher_comment %s, but got %s", teacherComment, scoreResp.TeacherComment)
		}
		if scoreResp.GradedPoint != nil && scoreResp.GradedPoint.Value != uint32(gradedPoint) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected graded_point %d, but got %d", gradedPoint, scoreResp.GradedPoint.Value)
		}
		if scoreResp.Point != nil && scoreResp.Point.Value != uint32(point) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected point %d, but got %d", point, scoreResp.Point.Value)
		}
		switch scoreResp.Core.Kind {
		case cpb.QuizType_QUIZ_TYPE_FIB:
			if len(scoreResp.CorrectText) == 0 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected correct_text not empty, but got %v", scoreResp.CorrectText)
			}

		case cpb.QuizType_QUIZ_TYPE_MAQ:
			if len(scoreResp.CorrectIndex) == 0 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected correct_index not empty, but got %v", scoreResp.CorrectIndex)
			}

			for _, correctIndex := range scoreResp.CorrectIndex {
				if correctIndex == 0 {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected correct_index value must greater than zero")
				}
			}
		case cpb.QuizType_QUIZ_TYPE_ORD:
			if scoreResp.Result != nil {
				result := scoreResp.Result.(*sspb.ExamLOSubmissionScore_OrderingResult)
				if !slices.Equal(result.OrderingResult.SubmittedKeys, []string{"key-1", "key-2", "key-3"}) {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected submitted_keys %v, but got %v", []string{"key-1", "key-2", "key-3"}, result.OrderingResult.SubmittedKeys)
				}

				if !slices.Equal(result.OrderingResult.CorrectKeys, []string{"key-1", "key-2", "key-3"}) {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected correct_keys %v, but got %v", []string{"key-1", "key-2", "key-3"}, result.OrderingResult.CorrectKeys)
				}
			} else {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("scoreResp.Result should not empty")
			}
		}
		if len(scoreResp.Core.QuestionTagIds) != len(scoreResp.Core.TagNames) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected length of tag names %d, but got %d", len(scoreResp.Core.QuestionTagIds), len(scoreResp.Core.TagNames))
		}
		for _, id := range scoreResp.Core.QuestionTagIds {
			tagName := fmt.Sprintf("tag-name-%s", id)
			if !slices.Contains(scoreResp.Core.TagNames, tagName) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected tag name: %s belong to tag names, but got %v", tagName, scoreResp.Core.TagNames)
			}
		}
	}

	if stepState.QuestionGroupID == "" {
		if len(response.QuestionGroups) != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected no question group, but got %v", len(response.QuestionGroups))
		}
	} else {
		if response.QuestionGroups[0].QuestionGroupId != stepState.QuestionGroupID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected question group id is %v, but got %v", response.QuestionGroups[0].QuestionGroupId, stepState.QuestionGroupID)
		}

		if len(response.QuestionGroups) != 1 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected 1 question group, but got %v", len(response.QuestionGroups))
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
