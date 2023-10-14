package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) studentChooseOption(ctx context.Context, options string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	stepState.CurrentStudentID = id
	var err error
	if ctx, err := s.aValidUser(ctx, id, consta.RoleStudent); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidStudentInDB id: %s has error: %w", id, err)
	}
	ctx = StepStateToContext(ctx, stepState)
	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken id: %s has error: %w", id, err)
	}
	ctx = StepStateToContext(ctx, stepState)

	rand.Seed(time.Now().UnixNano())
	stepState.Limit = 1
	stepState.Offset = 1
	paging := &cpb.Paging{
		Limit: uint32(stepState.Limit),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: int64(stepState.Offset),
		},
	}
	studyPlanItemID := s.newID()

	resp, err := epb.NewQuizModifierServiceClient(s.Conn).CreateQuizTest(s.signedCtx(ctx), &epb.CreateQuizTestRequest{
		LoId:            stepState.LoID,
		StudentId:       stepState.CurrentStudentID,
		StudyPlanItemId: studyPlanItemID,
		Paging:          paging,
		KeepOrder:       true,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.NextPage = paging
	stepState.SetID = resp.QuizzesId
	stepState.Response = resp
	ctx, err = s.returnListOfQuizItems(ctx, fmt.Sprintf("%d", len(resp.Items)))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	order := []int{0, 1, 2, 3, 4}

	shuffleQuizRepo := &repositories.ShuffledQuizSetRepo{}
	shuffleQuizSetRepo := &repositories.ShuffledQuizSetRepo{}

	seedStr, err := shuffleQuizRepo.GetSeed(ctx, s.DB, database.Text(stepState.SetID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	seed, err := strconv.ParseInt(seedStr.String, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	idx, err := shuffleQuizSetRepo.GetQuizIdx(ctx, s.DB, database.Text(stepState.SetID), database.Text(resp.Items[0].Core.ExternalId))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	r := rand.New(rand.NewSource(seed + int64(idx.Int)))
	r.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

	orderMapToNewOrder := make(map[int]int)
	for newOrder, oldOrder := range order {
		orderMapToNewOrder[oldOrder] = newOrder
	}

	for quizIndex := 0; quizIndex < len(stepState.QuizItems); quizIndex++ {
		answers := make([]*epb.Answer, 0)
		switch stepState.QuizItems[quizIndex].Core.Kind {
		case cpb.QuizType(cpb.QuizType_QUIZ_TYPE_MCQ), cpb.QuizType(cpb.QuizType_QUIZ_TYPE_MAQ):
			for _, idx := range strings.Split(options, ",") {
				i, _ := strconv.Atoi(strings.TrimSpace(idx))
				answers = append(answers, &epb.Answer{Format: &epb.Answer_SelectedIndex{SelectedIndex: uint32(orderMapToNewOrder[i-1]) + 1}})
			}
		}

		stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(s.signedCtx(ctx), &epb.CheckQuizCorrectnessRequest{
			SetId:  stepState.SetID,
			QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
			Answer: answers,
		})
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsIsCorrectAll(ctx context.Context, expectRes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.CheckQuizCorrectnessResponse)
	bERes, _ := strconv.ParseBool(expectRes)

	if resp.IsCorrectAll != bERes {
		return StepStateToContext(ctx, stepState), fmt.Errorf("return data wrong")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aFIBQuizTestWithCaseSensitiveConfigAndCorrectAnswers(ctx context.Context, options string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	kinds := []cpb.QuizType{
		cpb.QuizType_QUIZ_TYPE_FIB,
		cpb.QuizType_QUIZ_TYPE_POW,
		cpb.QuizType_QUIZ_TYPE_TAD,
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	reOptions := make([]*cpb.QuizOption, 0)
	for _, text := range strings.Split(options, ",") {
		reOptions = append(reOptions, &cpb.QuizOption{
			Content: &cpb.RichText{
				Raw: fmt.Sprintf(`
				{
					"blocks": [
						{
							"key": "2lnf5",
							"text": "%s",
							"type": "unstyled",
							"depth": 0,
							"inlineStyleRanges": [],
							"entityRanges": [],
							"data": {}
						}
					],
					"entityMap": {}
				}
			`, text),
				Rendered: "rendered " + s.newID(),
			},
			Correctness: true,
			Label:       "(1)",
			Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
			Key:         s.newID(),
		})
	}

	kind := kinds[rand.Intn(len(kinds))]
	req := []*epb.UpsertQuizRequest{
		{
			Quiz: &epb.QuizCore{
				ExternalId: s.newID(),
				Kind:       kind,
				SchoolId:   stepState.CurrentSchoolID,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				TaggedLos:       []string{"tag1", "tag2"},
				DifficultyLevel: 2,
				Options:         reOptions,
			},
			LoId: stepState.LoID,
		},
	}

	quizExIds := make([]string, 0)
	srv := epb.NewQuizModifierServiceClient(s.Conn)
	for _, r := range req {
		res, err := srv.UpsertQuiz(s.signedCtx(ctx), r)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExIds = append(quizExIds, res.Id)
	}
	stepState.Response = quizExIds
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aQuizTestWithPartialConfigOnAndTestCasesDataWithCorrectNotCorrect(ctx context.Context, correct, notCorrect string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	kinds := []cpb.QuizType{
		cpb.QuizType_QUIZ_TYPE_MAQ,
		cpb.QuizType_QUIZ_TYPE_MCQ,
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	kind := kinds[rand.Intn(len(kinds))]
	req := []*epb.UpsertQuizRequest{
		{
			Quiz: &epb.QuizCore{
				ExternalId: s.newID(),
				Kind:       kind,
				SchoolId:   stepState.CurrentSchoolID,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				TaggedLos:       []string{"tag1", "tag2"},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: false,
						Label:       "(3)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: false,
						Label:       "(4)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: false,
						Label:       "(5)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT},
						Key:         s.newID(),
					},
				},
			},
			LoId: stepState.LoID,
		},
	}

	quizExIds := make([]string, 0)
	srv := epb.NewQuizModifierServiceClient(s.Conn)
	for _, r := range req {
		res, err := srv.UpsertQuiz(s.signedCtx(ctx), r)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExIds = append(quizExIds, res.Id)
	}
	stepState.Response = quizExIds
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aQuizTestWithPartialConfigOffAndTestCasesDataWithCorrectNotCorrect(ctx context.Context, correct, notCorrect string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	kinds := []cpb.QuizType{
		cpb.QuizType_QUIZ_TYPE_MAQ,
		cpb.QuizType_QUIZ_TYPE_MCQ,
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	kind := kinds[rand.Intn(len(kinds))]
	req := []*epb.UpsertQuizRequest{
		{
			Quiz: &epb.QuizCore{
				ExternalId: s.newID(),
				Kind:       kind,
				SchoolId:   stepState.CurrentSchoolID,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				TaggedLos:       []string{"tag1", "tag2"},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: true,
						Label:       "(2)",
						Configs:     []cpb.QuizOptionConfig{},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: false,
						Label:       "(3)",
						Configs:     []cpb.QuizOptionConfig{},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: false,
						Label:       "(4)",
						Configs:     []cpb.QuizOptionConfig{},
						Key:         s.newID(),
					},
					{
						Content: &cpb.RichText{
							Raw:      "raw",
							Rendered: "rendered " + s.newID(),
						},
						Correctness: false,
						Label:       "(5)",
						Configs:     []cpb.QuizOptionConfig{},
						Key:         s.newID(),
					},
				},
			},
			LoId: stepState.LoID,
		},
	}

	quizExIds := make([]string, 0)
	srv := epb.NewQuizModifierServiceClient(s.Conn)
	for _, r := range req {
		res, err := srv.UpsertQuiz(s.signedCtx(ctx), r)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExIds = append(quizExIds, res.Id)
	}
	stepState.Response = quizExIds
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aFIBQuizTestWithPartialConfigOnAndCorrectAnwsers(ctx context.Context, options string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	kinds := []cpb.QuizType{
		cpb.QuizType_QUIZ_TYPE_FIB,
		cpb.QuizType_QUIZ_TYPE_POW,
		cpb.QuizType_QUIZ_TYPE_TAD,
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	reOptions := make([]*cpb.QuizOption, 0)
	for _, text := range strings.Split(options, ",") {
		reOptions = append(reOptions, &cpb.QuizOption{
			Content: &cpb.RichText{
				Raw: fmt.Sprintf(`
				{
					"blocks": [
						{
							"key": "2lnf5",
							"text": "%s",
							"type": "unstyled",
							"depth": 0,
							"inlineStyleRanges": [],
							"entityRanges": [],
							"data": {}
						}
					],
					"entityMap": {}
				}
			`, text),
				Rendered: "rendered " + s.newID(),
			},
			Correctness: true,
			Label:       "(1)",
			Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT},
			Key:         s.newID(),
		})
	}

	kind := kinds[rand.Intn(len(kinds))]
	req := []*epb.UpsertQuizRequest{
		{
			Quiz: &epb.QuizCore{
				ExternalId: s.newID(),
				Kind:       kind,
				SchoolId:   stepState.CurrentSchoolID,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				TaggedLos:       []string{"tag1", "tag2"},
				DifficultyLevel: 2,
				Options:         reOptions,
			},
			LoId: stepState.LoID,
		},
	}

	quizExIds := make([]string, 0)
	srv := epb.NewQuizModifierServiceClient(s.Conn)
	for _, r := range req {
		res, err := srv.UpsertQuiz(s.signedCtx(ctx), r)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExIds = append(quizExIds, res.Id)
	}
	stepState.Response = quizExIds
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aFIBQuizTestWithNoConfigAndCorrectAnswers(ctx context.Context, options string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	kinds := []cpb.QuizType{
		cpb.QuizType_QUIZ_TYPE_FIB,
		cpb.QuizType_QUIZ_TYPE_POW,
		cpb.QuizType_QUIZ_TYPE_TAD,
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	reOptions := make([]*cpb.QuizOption, 0)
	for _, text := range strings.Split(options, ",") {
		reOptions = append(reOptions, &cpb.QuizOption{
			Content: &cpb.RichText{
				Raw: fmt.Sprintf(`
				{
					"blocks": [
						{
							"key": "2lnf5",
							"text": "%s",
							"type": "unstyled",
							"depth": 0,
							"inlineStyleRanges": [],
							"entityRanges": [],
							"data": {}
						}
					],
					"entityMap": {}
				}
			`, text),
				Rendered: "rendered " + s.newID(),
			},
			Correctness: true,
			Label:       "(1)",
			Configs:     []cpb.QuizOptionConfig{},
			Key:         s.newID(),
		})
	}

	kind := kinds[rand.Intn(len(kinds))]
	req := []*epb.UpsertQuizRequest{
		{
			Quiz: &epb.QuizCore{
				ExternalId: s.newID(),
				Kind:       kind,
				SchoolId:   stepState.CurrentSchoolID,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				TaggedLos:       []string{"tag1", "tag2"},
				DifficultyLevel: 2,
				Options:         reOptions,
			},
			LoId: stepState.LoID,
		},
	}

	quizExIds := make([]string, 0)
	srv := epb.NewQuizModifierServiceClient(s.Conn)
	for _, r := range req {
		res, err := srv.UpsertQuiz(s.signedCtx(ctx), r)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExIds = append(quizExIds, res.Id)
	}
	stepState.Response = quizExIds
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aFIBQuizTestWithCaseSensitiveAndPartialConfigAndCorrectAnswers(ctx context.Context, options string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	kinds := []cpb.QuizType{
		cpb.QuizType_QUIZ_TYPE_FIB,
		cpb.QuizType_QUIZ_TYPE_POW,
		cpb.QuizType_QUIZ_TYPE_TAD,
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	reOptions := make([]*cpb.QuizOption, 0)
	for _, text := range strings.Split(options, ",") {
		reOptions = append(reOptions, &cpb.QuizOption{
			Content: &cpb.RichText{
				Raw: fmt.Sprintf(`
				{
					"blocks": [
						{
							"key": "2lnf5",
							"text": "%s",
							"type": "unstyled",
							"depth": 0,
							"inlineStyleRanges": [],
							"entityRanges": [],
							"data": {}
						}
					],
					"entityMap": {}
				}
			`, text),
				Rendered: "rendered " + s.newID(),
			},
			Correctness: true,
			Label:       "(1)",
			Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT, cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
			Key:         s.newID(),
		})
	}

	kind := kinds[rand.Intn(len(kinds))]
	req := []*epb.UpsertQuizRequest{
		{
			Quiz: &epb.QuizCore{
				ExternalId: s.newID(),
				Kind:       kind,
				SchoolId:   stepState.CurrentSchoolID,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + s.newID(),
				},
				TaggedLos:       []string{"tag1", "tag2"},
				DifficultyLevel: 2,
				Options:         reOptions,
			},
			LoId: stepState.LoID,
		},
	}

	quizExIds := make([]string, 0)
	srv := epb.NewQuizModifierServiceClient(s.Conn)
	for _, r := range req {
		res, err := srv.UpsertQuiz(s.signedCtx(ctx), r)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		quizExIds = append(quizExIds, res.Id)
	}
	stepState.Response = quizExIds
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentFillInText(ctx context.Context, uAnswers string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	stepState.CurrentStudentID = id
	var err error
	ctx, err = s.aValidUser(ctx, id, consta.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	rand.Seed(time.Now().UnixNano())
	stepState.Limit = 1
	stepState.Offset = 1
	paging := &cpb.Paging{
		Limit: uint32(stepState.Limit),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: int64(stepState.Offset),
		},
	}
	studyPlanItemID := s.newID()

	resp, err := epb.NewQuizModifierServiceClient(s.Conn).CreateQuizTest(s.signedCtx(ctx), &epb.CreateQuizTestRequest{
		LoId:            stepState.LoID,
		StudentId:       stepState.CurrentStudentID,
		StudyPlanItemId: studyPlanItemID,
		Paging:          paging,
		KeepOrder:       true,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.NextPage = paging
	stepState.SetID = resp.QuizzesId
	stepState.Response = resp
	ctx, err = s.returnListOfQuizItems(ctx, fmt.Sprintf("%d", len(resp.Items)))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for quizIndex := 0; quizIndex < len(stepState.QuizItems); quizIndex++ {
		answers := make([]*epb.Answer, 0)
		switch stepState.QuizItems[quizIndex].Core.Kind {
		case cpb.QuizType(cpb.QuizType_QUIZ_TYPE_FIB), cpb.QuizType(cpb.QuizType_QUIZ_TYPE_TAD), cpb.QuizType(cpb.QuizType_QUIZ_TYPE_POW):
			for _, anwser := range strings.Split(uAnswers, ",") {
				answers = append(answers, &epb.Answer{Format: &epb.Answer_FilledText{FilledText: anwser}})
			}
		}

		stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(s.signedCtx(ctx), &epb.CheckQuizCorrectnessRequest{
			SetId:  stepState.SetID,
			QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
			Answer: answers,
		})
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) thisIsAbsolutelyAnAnswerWithIsCorrectAll(ctx context.Context, status, expectRes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.CheckQuizCorrectnessResponse)
	bERes, _ := strconv.ParseBool(expectRes)

	if resp.IsCorrectAll != bERes {
		return StepStateToContext(ctx, stepState), fmt.Errorf("return data wrong")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) useUpsertATopic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ChapterID == "" {
		if ctx, err := s.insertAChapter(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve ")
		}
	}
	req := &epb.UpsertTopicsRequest{
		Topics: []*epb.Topic{
			{
				Name:         "Topic",
				Country:      epb.Country_COUNTRY_VN,
				Grade:        "G11",
				Subject:      epb.Subject_SUBJECT_MATHS,
				Type:         epb.TopicType_TOPIC_TYPE_LEARNING,
				CreatedAt:    timestamppb.Now(),
				UpdatedAt:    timestamppb.Now(),
				DisplayOrder: int32(1),
				TotalLos:     1,
				ChapterId:    stepState.ChapterID,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	stepState.TopicID = stepState.Response.(*epb.UpsertTopicsResponse).TopicIds[0]

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userCreateALearningObjective(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	stepState.LoID = id
	loData, err := s.setLoDataLocal(stepState.Random, id, "lo 1 name", "COUNTRY_VN", 12, "SUBJECT_BIOLOGY", false, stepState.TopicID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx = StepStateToContext(s.signedCtx(ctx), stepState)

	loRepo := &repositories.LearningObjectiveRepo{}
	if err := loRepo.BulkImport(ctx, s.DB, []*entities.LearningObjective{
		loData,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setLoDataLocal(ranStr, id, name, country string, grade int, subject string, isDeleted bool, tID string) (*entities.LearningObjective, error) {
	c := &entities.LearningObjective{}
	database.AllNullEntity(c)
	if err := c.ID.Set(id); err != nil {
		return nil, err
	}
	if err := c.Name.Set(name + ranStr); err != nil {
		return nil, err
	}
	if err := c.Country.Set(country); err != nil {
		return nil, err
	}
	if err := c.Subject.Set(subject); err != nil {
		return nil, err
	}
	if err := c.DisplayOrder.Set(1); err != nil {
		return nil, err
	}
	if err := c.Grade.Set(grade); err != nil {
		return nil, err
	}
	if err := c.SchoolID.Set(constants.ManabieSchool); err != nil {
		return nil, err
	}
	if err := c.CreatedAt.Set(time.Now()); err != nil {
		return nil, err
	}
	if err := c.UpdatedAt.Set(time.Now()); err != nil {
		return nil, err
	}
	if err := c.TopicID.Set(tID); err != nil {
		return nil, err
	}
	if err := c.ApproveGrading.Set(false); err != nil {
		return nil, err
	}
	if err := c.GradeCapping.Set(false); err != nil {
		return nil, err
	}
	if err := c.ReviewOption.Set(cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()); err != nil {
		return nil, err
	}
	if err := c.VendorType.Set(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String()); err != nil {
		return nil, err
	}
	if isDeleted {
		if err := c.DeletedAt.Set(time.Now()); err != nil {
			return nil, err
		}
	} else {
		if err := c.DeletedAt.Set(nil); err != nil {
			return nil, err
		}
	}

	return c, nil
}
