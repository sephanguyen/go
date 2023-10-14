package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) aQuizsetWithQuizzesInLearningObjectiveBelongedToATopic(ctx context.Context, numberOfQuizzes, topicType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.learningObjectiveBelongedToATopic(ctx, topicType)
	ctx, err2 := s.aListOfQuizzes(ctx, numberOfQuizzes)
	ctx, err3 := s.aQuizset(ctx)

	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2, err3)
}

func (s *suite) learningObjectiveBelongedToATopic(ctx context.Context, topicType string) (context.Context, error) {
	ctx, err1 := s.aSignedIn(ctx, "school admin")
	ctx, err2 := s.aListOfValidTopics(ctx)
	ctx, err3 := s.adminInsertsAListOfValidTopics(ctx)
	ctx, err4 := s.aLearningObjectiveBelongedToATopic(ctx, topicType)
	return ctx, multierr.Combine(err1, err2, err3, err4)
}

func (s *suite) aStudyPlanItemId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// cause study plan item is from service enigma. So it's assumed that we have the study plan id
	stepState.StudyPlanItemID = s.newID()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateQuizTestWithValidRequestAndLimitTheFirstTime(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Limit, _ = strconv.Atoi(arg1)
	stepState.Offset = 1
	stepState.NextPage = nil
	stepState.SetID = ""
	stepState.SessionID = strconv.Itoa(rand.Int())
	request := &epb.CreateQuizTestRequest{
		LoId:            stepState.LoID,
		StudentId:       stepState.CurrentStudentID,
		StudyPlanItemId: stepState.StudyPlanItemID,
		SessionId:       stepState.SessionID,
		Paging: &cpb.Paging{
			Limit: uint32(stepState.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	}
	ctx, err := s.executeCreateQuizTestService(ctx, request)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) executeCreateQuizTestService(ctx context.Context, request *epb.CreateQuizTestRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).CreateQuizTest(s.signedCtx(ctx), request)
	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *suite) returnListOfQuizItems(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.returnsStatusCode(ctx, "OK")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	numOfQuizzes, _ := strconv.Atoi(arg1)

	resp, ok := stepState.Response.(*epb.CreateQuizTestResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not receive create quiz test response")
	}
	stepState.ShuffledQuizSetID = resp.QuizzesId
	var limit int
	var offset int
	if stepState.NextPage == nil {
		limit = stepState.Limit
		offset = stepState.Offset
	} else {
		limit = int(stepState.NextPage.Limit)
		offset = int(stepState.NextPage.GetOffsetInteger())
	}
	stepState.NextPage = resp.NextPage
	stepState.SetID = resp.QuizzesId
	stepState.QuizItems = resp.Items
	if len(resp.Items) != numOfQuizzes {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect number of quizzes are %v but got %v", numOfQuizzes, len(resp.Items))
	}

	quizExternalIDs := make([]string, 0)
	query := `SELECT quiz_external_id FROM shuffled_quiz_sets sqs INNER JOIN UNNEST(sqs.quiz_external_ids) AS quiz_external_id ON shuffled_quiz_set_id = $1 LIMIT $2 OFFSET $3;
	`

	rows, err := s.DB.Query(ctx, query, stepState.SetID, limit, offset-1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	for rows.Next() {
		var id pgtype.Text
		err := rows.Scan(&id)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExternalIDs = append(quizExternalIDs, id.String)
	}

	if len(resp.Items) != len(quizExternalIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect number of quizzes are %v but got %v", quizExternalIDs, len(resp.Items))
	}

	for i := range resp.Items {
		if resp.Items[i].Core.ExternalId != quizExternalIDs[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz id %v but got %v", len(quizExternalIDs), resp.Items[i].Core.ExternalId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetNextPageOfQuizTest(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	request := &epb.CreateQuizTestRequest{
		LoId:            stepState.LoID,
		StudentId:       stepState.CurrentStudentID,
		StudyPlanItemId: stepState.StudyPlanItemID,
		Paging:          stepState.NextPage,
		SetId:           wrapperspb.String(stepState.SetID),
	}

	s.executeCreateQuizTestService(ctx, request)
	return StepStateToContext(ctx, stepState)
}

func (s *suite) returnQuizItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stt, ok := status.FromError(stepState.ResponseErr)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	if stt.Code() != codes.OK {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s, got %s status code, message: %s", codes.OK.String(), stt.Code().String(), stt.Message())
	}

	resp, ok := stepState.Response.(*epb.CreateQuizTestResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not receive create quiz test response")
	}
	stepState.NextPage = resp.NextPage
	stepState.SetID = resp.QuizzesId
	stepState.QuizItems = resp.Items
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) quizSetWithAll(ctx context.Context, numberOfQuizzes, topicType string, quizType cpb.QuizType) (context.Context, error) {
	ctx, err1 := s.learningObjectiveBelongedToATopic(ctx, topicType)
	ctx, err2 := s.ListOfAllQuiz(ctx, numberOfQuizzes, quizType)
	ctx, err3 := s.aQuizset(ctx)
	return ctx, multierr.Combine(err1, err2, err3)
}

func (s *suite) ListOfAllQuiz(ctx context.Context, numStr string, quizType cpb.QuizType) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOfQuizzes, _ := strconv.Atoi(numStr)
	stepState.Quizzes = entities.Quizzes{}
	for i := 0; i < numOfQuizzes; i++ {
		quiz := s.genQuiz(ctx, quizType)
		stepState.Quizzes = append(stepState.Quizzes, quiz)
	}

	quizRepo := repositories.QuizRepo{}
	for _, quiz := range stepState.Quizzes {
		err := quizRepo.Create(ctx, s.DB, quiz)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) genQuiz(ctx context.Context, quizType cpb.QuizType) *entities.Quiz {
	stepState := StepStateFromContext(ctx)

	switch quizType {
	case cpb.QuizType_QUIZ_TYPE_MCQ:
		return s.genMultipleChoicesQuiz(stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_FIB:
		return s.genFillInTheBlankQuiz(stepState.CurrentUserID, stepState.LoID, stepState.UserFillInTheBlankOld)
	case cpb.QuizType_QUIZ_TYPE_TAD:
		return s.genTermAndDefinitionQuiz(ctx)
	case cpb.QuizType_QUIZ_TYPE_POW:
		return s.genPairOfWordQuiz(stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_MIQ:
		return s.genManualInputQuiz(ctx, stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_MAQ:
		return s.genMultipleChoicesQuiz(stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_ORD:
		return s.genOrderingQuiz(stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_ESQ:
		return s.genEssayQuiz(stepState.CurrentUserID, stepState.LoID)
	}
	return nil
}

func (s *suite) createAQuizV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	quizID := idutil.ULIDNow()
	stepState.QuizID = quizID
	externalID := idutil.ULIDNow()
	quizLO := &epb.QuizLO{
		Quiz: &cpb.QuizCore{
			ExternalId: externalID,
			Kind:       cpb.QuizType_QUIZ_TYPE_POW,
			Info: &cpb.ContentBasicInfo{
				SchoolId: constants.ManabieSchool,
				Country:  cpb.Country_COUNTRY_VN,
			},
			Question: &cpb.RichText{
				Raw:      "raw",
				Rendered: "rendered " + idutil.ULIDNow(),
			},
			Explanation: &cpb.RichText{
				Raw:      "raw",
				Rendered: "rendered " + idutil.ULIDNow(),
			},
			TaggedLos:       []string{"123", "abc"},
			DifficultyLevel: 2,
			Options: []*cpb.QuizOption{
				{
					Content:     &cpb.RichText{Raw: "raw", Rendered: "rendered " + idutil.ULIDNow()},
					Correctness: false,
					Configs: []cpb.QuizOptionConfig{
						cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE,
					},
					Attribute: &cpb.QuizItemAttribute{
						ImgLink:   "img.link",
						AudioLink: "audio.link",
						Configs: []cpb.QuizItemAttributeConfig{
							1,
						},
					},
					Label: "label",
					Key:   "key",
				},
			},
			Attribute: &cpb.QuizItemAttribute{
				ImgLink:   "img.link",
				AudioLink: "audio.link",
				Configs: []cpb.QuizItemAttributeConfig{
					1,
				},
			},
			Point: wrapperspb.Int32(10),
		},
		LoId: stepState.LoID,
	}
	stepState.QuizLOList = append(stepState.QuizLOList, quizLO)

	if _, err := sspb.NewQuizClient(s.Conn).UpsertFlashcardContent(contextWithToken(s, ctx), &sspb.UpsertFlashcardContentRequest{
		Quizzes:     []*cpb.QuizCore{quizLO.Quiz},
		FlashcardId: stepState.LoID,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createALO(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.LoID = idutil.ULIDNow()
	req := &epb.UpsertLOsRequest{
		LearningObjectives: []*cpb.LearningObjective{
			{
				Info: &cpb.ContentBasicInfo{
					Id:        stepState.LoID,
					Name:      "name",
					Country:   cpb.Country_COUNTRY_VN,
					Subject:   cpb.Subject_SUBJECT_BIOLOGY,
					SchoolId:  constants.ManabieSchool,
					Grade:     1,
					CreatedAt: timestamppb.Now(),
					UpdatedAt: timestamppb.Now(),
				},
				TopicId: stepState.TopicID,
			},
		},
	}
	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(contextWithToken(s, ctx), req); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateAQuizV2UsingV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.createALO(ctx)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}
	ctx, err = s.createAQuizV2(ctx)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), nil
}
