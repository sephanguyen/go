package eureka

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) insertAChapter(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	chapter := &entities.Chapter{}
	now := time.Now()
	database.AllNullEntity(chapter)
	stepState.ChapterID = s.newID()
	multierr.Combine(
		chapter.ID.Set(stepState.ChapterID),
		chapter.Country.Set(pb.COUNTRY_VN),
		chapter.Name.Set(fmt.Sprintf("name-%s", stepState.ChapterID)),
		chapter.Grade.Set(12),
		chapter.SchoolID.Set(stepState.SchoolIDInt),
		chapter.CurrentTopicDisplayOrder.Set(0),
		chapter.CreatedAt.Set(now),
		chapter.UpdatedAt.Set(now),
	)
	_, err := database.Insert(ctx, chapter, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfValidChaptersInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := 0; i < 10; i++ {
		c := new(entities.Chapter)
		database.AllNullEntity(c)
		now := time.Now()
		id := fmt.Sprintf("chapter_id_%d", i)
		name := fmt.Sprintf("chapter_name_%d", i)
		c.ID.Set(id)
		c.Name.Set(name)
		c.CreatedAt.Set(now)
		c.UpdatedAt.Set(now)
		c.Country.Set(pb.COUNTRY_VN.String())
		c.Grade.Set(1)
		c.Subject.Set(pb.SUBJECT_CHEMISTRY.String())
		c.DisplayOrder.Set(1)
		c.SchoolID.Set(stepState.SchoolIDInt)
		c.DeletedAt.Set(nil)

		_, err := database.Insert(ctx, c, s.DB.Exec)
		if e, ok := err.(*pgconn.PgError); ok && e.Code != "23505" {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.CurrentChapterIDs = append(stepState.CurrentChapterIDs, id)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLearningObjectiveBelongedToATopicWithType(ctx context.Context, topic string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.logins(ctx, schoolAdminRawText); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	topicList, err := pb.NewCourseClient(s.BobConn).ListTopic(
		helper.GRPCContext(s.signedCtx(ctx), "token", stepState.AuthToken), &pb.ListTopicRequest{
			TopicType: pb.TopicType(pb.TopicType_value[topic]),
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err list topic %s", err)
	}

	if len(topicList.Topics) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("can't find topic")
	}

	t1 := topicList.Topics[0]

	lo := s.generateValidLearningObjectiveEntity(t1.Id)
	if _, err = database.Insert(ctx, lo, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	topicLO := &entities.TopicsLearningObjectives{
		TopicID:      database.Text(t1.Id),
		LoID:         database.Text(lo.ID.String),
		DisplayOrder: database.Int2(lo.DisplayOrder.Int),
		CreatedAt:    database.Timestamptz(now),
		UpdatedAt:    database.Timestamptz(now),
		DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
	}
	if _, err = database.Insert(ctx, topicLO, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.LoID = lo.ID.String
	stepState.Request = lo.ID.String
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) learningObjectiveBelongedToATopicWithType(ctx context.Context, topicType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.logins(ctx, schoolAdminRawText)
	ctx, err2 := s.aListOfValidTopics(ctx)
	ctx, err3 := s.adminInsertsAListOfValidTopics(ctx)
	ctx, err4 := s.aLearningObjectiveBelongedToATopicWithType(ctx, topicType)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2, err3, err4)
}

func (s *suite) genPairOfWordQuiz(currentUserID, loID string) *entities.Quiz {
	quizRawObj := raw{
		Blocks: []block{
			{
				Key:               "1c0o5",
				Text:              "Banana",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
			},
		},
	}
	quizRaw, _ := json.Marshal(quizRawObj)

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/d41d8cd98f00b204e9800998ecf8427e.html",
		Attribute: entities.QuizItemAttribute{
			ImgLink:   "https://storage.googleapis.com/stag-manabie-backend/user-upload/ce196d896889ce984b8c36f6c8ed64b001FF6SDP89K003NKW655N9CY1J.jpg",
			AudioLink: "https://storage.googleapis.com/stag-manabie-backend/user-upload/Banana01FFF0KPK7RBFXTFWFBFMPB81C.mp3",
			Configs:   []string{"FLASHCARD_LANGUAGE_CONFIG_ENG"},
		},
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

	quizOptionObjs := []*entities.QuizOption{
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/d41d8cd98f00b204e9800998ecf8427e.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "A",
			Key:         "01FFF0KP9ZBMV0X0RKJ9NMB33V",
			Attribute: entities.QuizItemAttribute{
				ImgLink:   "https://storage.googleapis.com/stag-manabie-backend/user-upload/ce196d896889ce984b8c36f6c8ed64b001FF6SDP89K003NKW655N9CY1J.jpg",
				AudioLink: "https://storage.googleapis.com/stag-manabie-backend/user-upload/Banana%20term01FFF0KPX73PNPF4MYBWA8ZX1D.mp3",
				Configs:   []string{cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG.String()},
			},
		},
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(s.newID())
	quiz.ExternalID = database.Text(s.newID())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_POW)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(explanationQuestion)
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	quiz.Point = database.Int4(10)
	return quiz
}

type block struct {
	Key               string      `json:"key"`
	Text              string      `json:"text"`
	Type              string      `json:"type"`
	Depth             string      `json:"depth"`
	InlineStyleRanges []string    `json:"inlineStyleRanges"`
	EntityRanges      []string    `json:"entityRanges"`
	Data              interface{} `json:"data"`
	EntityMap         interface{} `json:"entityMap"`
}

type data struct {
	Data string `json:"data"`
}
type entityMapContent struct {
	Type       string `json:"type"`
	Mutability string `json:"mutability"`
	Data       data   `json:"data"`
}

type raw struct {
	Blocks    []block                     `json:"blocks"`
	EntityMap map[string]entityMapContent `json:"entityMap"`
}

func (s *suite) genFillInTheBlankQuiz(currentUserID, loID string, userFillInTheBlankOld bool) *entities.Quiz {
	tempEntityMap := make(map[string]entityMapContent)
	tempEntityMap["0"] = entityMapContent{
		Type:       "INLINE_MATHJAX",
		Mutability: "IMMUTABLE",
		Data: data{
			Data: "\\overrightarrow{n_{3}}=(1 ; 2 ;-1)",
		},
	}
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
		EntityMap: tempEntityMap,
	}
	quizRaw, _ := json.Marshal(quizRawObj)

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

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
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(s.newID())
	quiz.ExternalID = database.Text(s.newID())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_FIB)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	quiz.Point = database.Int4(10)
	if userFillInTheBlankOld {
		options, _ := quiz.GetOptions()
		for _, opt := range options {
			opt.Key = ""
		}
		quiz.Options.Set(options)
	}
	return quiz
}

func (s *suite) genMultipleChoicesQuiz(currentUserID, loID string) *entities.Quiz {
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

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

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
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(s.newID())
	quiz.ExternalID = database.Text(s.newID())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	quiz.Point = database.Int4(10)
	return quiz
}

func (s *suite) genManualInputQuiz(_ context.Context, currentUserID, loID string) *entities.Quiz {
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

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

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
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(s.newID())
	quiz.ExternalID = database.Text(s.newID())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MIQ)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	quiz.Point = database.Int4(10)
	return quiz
}

func (s *suite) genOrderingQuiz(currentUserID, loID string) *entities.Quiz {
	quizRaws := make([][]byte, 0, 3)
	for i := 0; i < 3; i++ {
		quizRawObj := raw{
			Blocks: []block{
				{
					Key:               "eq20k",
					Text:              fmt.Sprintf("Text-%d", i),
					Type:              "unstyled",
					Depth:             "0",
					InlineStyleRanges: []string{},
					EntityRanges:      []string{},
					Data:              nil,
				},
			},
		}
		quizRaw, _ := json.Marshal(quizRawObj)
		quizRaws = append(quizRaws, quizRaw)
	}

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaws[0]),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

	quizOptionObjs := []*entities.QuizOption{
		{
			Content: entities.RichText{
				Raw:         string(quizRaws[0]),
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
				Raw:         string(quizRaws[1]),
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
				Raw:         string(quizRaws[2]),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "C",
			Key:         "key-C",
			Attribute:   entities.QuizItemAttribute{},
		},
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(s.newID())
	quiz.ExternalID = database.Text(s.newID())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_ORD)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	quiz.Point = database.Int4(10)
	return quiz
}

func (s *suite) genEssayQuiz(currentUserID, loID string) *entities.Quiz {
	quizRaws := make([][]byte, 0, 1)
	for i := 0; i < 3; i++ {
		quizRawObj := raw{
			Blocks: []block{
				{
					Key:               "eq20k",
					Text:              fmt.Sprintf("Text-%d", i),
					Type:              "unstyled",
					Depth:             "0",
					InlineStyleRanges: []string{},
					EntityRanges:      []string{},
					Data:              nil,
				},
			},
		}
		quizRaw, _ := json.Marshal(quizRawObj)
		quizRaws = append(quizRaws, quizRaw)
	}

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaws[0]),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

	quizOptionObjs := []*entities.QuizOption{
		{
			Content: entities.RichText{
				Raw:         string(quizRaws[0]),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "A",
			Key:         "key-A",
			Attribute:   entities.QuizItemAttribute{},
			AnswerConfig: entities.AnswerConfig{
				Essay: entities.EssayConfig{
					LimitEnabled: true,
					LimitType:    entities.EssayLimitTypeCharacter,
					Limit:        5000,
				},
			},
		},
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(s.newID())
	quiz.ExternalID = database.Text(s.newID())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_ESQ)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	quiz.Point = database.Int4(10)
	return quiz
}

func (s *suite) genTermAndDefinitionQuiz(ctx context.Context) *entities.Quiz {
	stepState := StepStateFromContext(ctx)
	quiz := s.genPairOfWordQuiz(stepState.CurrentUserID, stepState.LoID)
	_ = quiz.Kind.Set(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_TAD)])
	return quiz
}

// nolint
func (s *suite) aListOfQuizzes(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOfQuizzes, _ := strconv.Atoi(arg1)
	stepState.Quizzes = entities.Quizzes{}
	for i := 0; i < numOfQuizzes; i++ {
		var quiz *entities.Quiz
		if i == 0 {
			quiz = s.genPairOfWordQuiz(stepState.CurrentUserID, stepState.LoID)
			stepState.Quizzes = append(stepState.Quizzes, quiz)
			continue
		}

		// currently, not gen Essay Quiz yet
		kind := cpb.QuizType(rand.Intn(len(cpb.QuizType_name)))
		switch kind {
		case cpb.QuizType_QUIZ_TYPE_MCQ:
			quiz = s.genMultipleChoicesQuiz(stepState.CurrentUserID, stepState.LoID)
		case cpb.QuizType_QUIZ_TYPE_FIB:
			quiz = s.genFillInTheBlankQuiz(stepState.CurrentUserID, stepState.LoID, stepState.UserFillInTheBlankOld)
		case cpb.QuizType_QUIZ_TYPE_POW:
			quiz = s.genPairOfWordQuiz(stepState.CurrentUserID, stepState.LoID)
		case cpb.QuizType_QUIZ_TYPE_TAD:
			quiz = s.genTermAndDefinitionQuiz(ctx)
		case cpb.QuizType_QUIZ_TYPE_MIQ:
			quiz = s.genManualInputQuiz(ctx, stepState.CurrentUserID, stepState.LoID)
		case cpb.QuizType_QUIZ_TYPE_MAQ:
			quiz = s.genMultipleChoicesQuiz(stepState.CurrentUserID, stepState.LoID)
			var k pgtype.Text
			k.Set(cpb.QuizType_QUIZ_TYPE_MAQ.String())
			quiz.Kind = k
		case cpb.QuizType_QUIZ_TYPE_ORD:
			quiz = s.genOrderingQuiz(stepState.CurrentUserID, stepState.LoID)
		case cpb.QuizType_QUIZ_TYPE_ESQ:
			quiz = s.genEssayQuiz(stepState.CurrentUserID, stepState.LoID)
		}
		stepState.Quizzes = append(stepState.Quizzes, quiz)
		stepState.ExistingQuestionHierarchy.AddQuestionID(quiz.ExternalID.String)
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

func (s *suite) aQuizset(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	quizExternalIDs := []string{}
	for _, quiz := range stepState.Quizzes {
		quizExternalIDs = append(quizExternalIDs, quiz.ExternalID.String)
	}

	questionHierarchy := make([]interface{}, 0)
	for _, extID := range quizExternalIDs {
		questionHierarchy = append(questionHierarchy, &entities.QuestionHierarchyObj{
			ID:   extID,
			Type: entities.QuestionHierarchyQuestion,
		})
	}

	quizSet := entities.QuizSet{}
	database.AllNullEntity(&quizSet)

	quizSet.ID = database.Text(s.newID())
	quizSet.LoID = database.Text(stepState.LoID)
	quizSet.QuizExternalIDs = database.TextArray(quizExternalIDs)
	quizSet.Status = database.Text("QUIZSET_STATUS_APPROVED")
	quizSet.QuestionHierarchy = database.JSONBArray(questionHierarchy)

	stepState.QuizSet = quizSet

	quizSetRepo := repositories.QuizSetRepo{}
	err := quizSetRepo.Create(ctx, s.DB, &stepState.QuizSet)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) executeCreateFlashcardStudyTestService(ctx context.Context, request *epb.CreateFlashCardStudyRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := epb.NewQuizModifierServiceClient(s.Conn).CreateFlashCardStudy(s.signedCtx(ctx), request)
	if err == nil {
		stepState.StudySetID = resp.StudySetId
	}
	stepState.Response = resp
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateFlashcardStudyTestWithValidRequestAndLimitTheFirstTime(ctx context.Context, limit string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studyPlan := &entities.StudyPlan{
		BaseEntity: entities.BaseEntity{
			CreatedAt: database.Timestamptz(time.Now()),
			UpdatedAt: database.Timestamptz(time.Now()),
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		},
		ID:              database.Text(s.newID()),
		MasterStudyPlan: pgtype.Text{Status: pgtype.Null},
		Name:            database.Text("name"),
		StudyPlanType:   database.Text("studyPlanType"),
		SchoolID:        database.Int4(stepState.SchoolIDInt),
		CourseID:        database.Text("courseID"),
	}
	err := multierr.Combine(
		studyPlan.Grades.Set(nil),
		studyPlan.Status.Set("STUDY_PLAN_STATUS_ACTIVE"),
		studyPlan.TrackSchoolProgress.Set(false),
		studyPlan.BookID.Set(nil))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot set studyplan : %w", err)
	}
	if _, err := database.Insert(ctx, studyPlan, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert studyPlan: %w", err)
	}

	studyPlanItem := &entities.StudyPlanItem{
		BaseEntity: entities.BaseEntity{
			CreatedAt: database.Timestamptz(time.Now()),
			UpdatedAt: database.Timestamptz(time.Now()),
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		},
		ID:                      database.Text(s.newID()),
		StudyPlanID:             studyPlan.ID,
		AvailableFrom:           database.Timestamptz(time.Now().Add(-30 * 24 * time.Hour)),
		AvailableTo:             database.Timestamptz(time.Now().Add(30 * 24 * time.Hour)),
		StartDate:               database.Timestamptz(time.Now().Add(-10 * 24 * time.Hour)),
		EndDate:                 database.Timestamptz(time.Now().Add(10 * 24 * time.Hour)),
		CompletedAt:             pgtype.Timestamptz{Status: pgtype.Null},
		ContentStructure:        pgtype.JSONB{Status: pgtype.Null},
		ContentStructureFlatten: pgtype.Text{Status: pgtype.Null},
		DisplayOrder:            database.Int4(rand.Int31n(100)),
		CopyStudyPlanItemID:     pgtype.Text{Status: pgtype.Null},
		SchoolDate:              pgtype.Timestamptz{Status: pgtype.Null},
		Status:                  pgtype.Text{Status: pgtype.Null},
	}
	if _, err := database.Insert(ctx, studyPlanItem, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert: %w", err)
	}
	stepState.StudyPlanItemID = studyPlanItem.ID.String

	stepState.Limit, _ = strconv.Atoi(limit)
	stepState.Offset = 1
	stepState.NextPage = nil
	stepState.SetID = ""

	request := &epb.CreateFlashCardStudyRequest{
		StudyPlanItemId: stepState.StudyPlanItemID,
		LoId:            stepState.LoID,
		StudentId:       stepState.StudentID,
		Paging: &cpb.Paging{
			Limit: uint32(stepState.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		KeepOrder: false,
	}

	return s.executeCreateFlashcardStudyTestService(ctx, request)
}

func (s *suite) userFinishFlashcardStudyWithoutRestart(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := "UPDATE flashcard_progressions SET remembered_question_ids = quiz_external_ids WHERE study_set_id = $1"
	if _, err := s.DB.Exec(ctx, query, stepState.StudySetID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	request := &epb.FinishFlashCardStudyProgressRequest{
		StudySetId:      stepState.StudySetID,
		StudentId:       stepState.StudentID,
		LoId:            stepState.LoID,
		StudyPlanItemId: stepState.StudyPlanItemID,
	}
	stepState.Request = request
	stepState.Response, stepState.ResponseErr = epb.NewCourseModifierServiceClient(s.Conn).
		FinishFlashCardStudyProgress(s.signedCtx(ctx), request)

	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *suite) userFinishFlashcardStudyWithoutRestartAndRememberedQuestions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request := &epb.FinishFlashCardStudyProgressRequest{
		StudySetId:      stepState.StudySetID,
		StudentId:       stepState.StudentID,
		LoId:            stepState.LoID,
		StudyPlanItemId: stepState.StudyPlanItemID,
	}
	stepState.Request = request
	stepState.Response, stepState.ResponseErr = epb.NewCourseModifierServiceClient(s.Conn).
		FinishFlashCardStudyProgress(s.signedCtx(ctx), request)

	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *suite) userFinishFlashcardStudyWithRestart(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := &epb.FinishFlashCardStudyProgressRequest{
		StudySetId:      stepState.StudySetID,
		StudentId:       stepState.StudentID,
		LoId:            stepState.LoID,
		StudyPlanItemId: stepState.StudyPlanItemID,
		IsRestart:       true,
	}
	stepState.Request = request
	stepState.Response, stepState.ResponseErr = epb.NewCourseModifierServiceClient(s.Conn).
		FinishFlashCardStudyProgress(s.signedCtx(ctx), request)

	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *suite) verifyDataAfterFinishFlashcardWithoutRestart(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var completedAt pgtype.Timestamptz

	query := "SELECT completed_at FROM flashcard_progressions WHERE study_set_id = $1"
	if err := s.DB.QueryRow(ctx, query, stepState.StudySetID).Scan(&completedAt); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if completedAt.Time.IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected: completed_at of flashcard_progressions is not null, but got null")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) verifyDataAfterFinishFlashcardWithRestart(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var deletedAt, completedAt pgtype.Timestamptz

	query := "SELECT deleted_at FROM flashcard_progressions WHERE study_set_id = $1"
	if err := s.DB.QueryRow(ctx, query, stepState.StudySetID).Scan(&deletedAt); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if deletedAt.Time.IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected: deleted_at of flashcard_progressions is not null, but got null")
	}

	query = "SELECT completed_at FROM study_plan_items WHERE study_plan_item_id = $1"
	if err := s.DB.QueryRow(ctx, query, stepState.StudyPlanItemID).Scan(&completedAt); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if deletedAt.Time.IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected: completed_at of study_plan_items is not null, but got null")
	}

	return StepStateToContext(ctx, stepState), nil
}
