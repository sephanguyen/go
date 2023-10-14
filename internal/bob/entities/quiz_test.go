package entities

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/stretchr/testify/assert"
)

func TestQuiz_GetOptionsWithAlternatives(t *testing.T) {
	t.Parallel()
	t.Run("happy case", func(t *testing.T) {
		quiz := generateQuiz()
		quiz.Options.Set(`[
		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},

		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"2", "key":"key-2", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"3", "key":"key-3", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"3", "key":"key-3", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
	]`)
		optionWithAlternatives, err := quiz.GetOptionsWithAlternatives()
		assert.Nil(t, err)
		options, err := quiz.GetOptions()
		assert.Nil(t, err)
		expect := []*QuizOptionWithAlternatives{
			{
				Key:                "key-1",
				AlternativeOptions: []*QuizOption{options[0], options[1], options[4]},
			},
			{
				Key:                "key-2",
				AlternativeOptions: []*QuizOption{options[2]},
			},
			{
				Key:                "key-3",
				AlternativeOptions: []*QuizOption{options[3], options[5]},
			},
		}
		assert.Equal(t, len(expect), len(optionWithAlternatives))
		for i := range expect {
			assert.Equal(t, len(expect[i].Key), len(optionWithAlternatives[i].Key))
			assert.Equal(t, len(expect[i].AlternativeOptions), len(optionWithAlternatives[i].AlternativeOptions))
			for j := range expect[i].AlternativeOptions {
				assert.Equal(t, expect[i].AlternativeOptions[j], optionWithAlternatives[i].AlternativeOptions[j])
			}
		}
	})
}

func TestQuizOptionWithAlternatives_GetText(t *testing.T) {
	t.Parallel()
	t.Run("happy case", func(t *testing.T) {
		quiz := generateQuiz()
		quiz.Options.Set(`[
		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},

		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"2", "key":"key-2", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"3", "key":"key-3", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"bonjour\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"greeting\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"3", "key":"key-3", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
		]`)
		optionWithAlternatives, err := quiz.GetOptionsWithAlternatives()
		assert.Nil(t, err)
		expected := []string{"qwewqeqweqwe", "goodbye", "bonjour"}
		assert.Equal(t, len(expected), len(optionWithAlternatives))
		for i, opt := range optionWithAlternatives {
			assert.Equal(t, expected[i], opt.GetText())
		}
	})
}

func TestQuizOptionWithAlternatives_IsCorrect(t *testing.T) {
	t.Parallel()
	t.Run("happy case", func(t *testing.T) {
		quiz := generateQuiz()
		quiz.Options.Set(`[
		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},

		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"2", "key":"key-2", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"3", "key":"key-3", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"bonjour\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"1", "key":"key-1", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"greeting\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"3", "key":"key-3", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
		]`)
		optionWithAlternatives, err := quiz.GetOptionsWithAlternatives()
		assert.Nil(t, err)
		answers := []string{"qwewqeqweqwe", "goodbye", "wrong answer"}
		expected := []bool{true, true, false}
		assert.Equal(t, len(expected), len(optionWithAlternatives))
		for i, opt := range optionWithAlternatives {
			assert.Equal(t, expected[i], opt.IsCorrect(answers[i]))
		}
	})
}

func generateQuiz() *Quiz {
	quiz := &Quiz{}

	quiz.ID.Set(idutil.ULIDNow())
	quiz.ExternalID.Set("external-id-1")
	quiz.Country.Set(cpb.Country_name[int32(cpb.Country_COUNTRY_VN)])
	quiz.SchoolID.Set(1)
	quiz.Kind.Set(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)])
	quiz.Question.Set(`{"raw":"{\"blocks\":[{\"key\":\"2bsgi\",\"text\":\"qeqweqewq\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/150cb1b73bc9d3bbe4011a55476a6913.html"}`)
	quiz.Explanation.Set(`{"raw":"{\"blocks\":[{\"key\":\"f5lms\",\"text\":\"\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/24061416a35eb51f403307148c5f4cef.html"}`)
	quiz.Options.Set(`[
		{"label":"1","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false}
	]`)
	quiz.TaggedLOs.Set([]string{"tagLO1", "tagLO2", "tagLO3"})
	quiz.CreatedBy.Set("adminid")
	quiz.ApprovedBy.Set("adminid")
	quiz.Status.Set("approved")
	quiz.DifficultLevel.Set(2)
	createdAt := database.Timestamptz(time.Now())
	updatedAt := database.Timestamptz(time.Now())
	quiz.CreatedAt.Set(createdAt)
	quiz.UpdatedAt.Set(updatedAt)

	return quiz
}

func TestRichText_GetText(t *testing.T) {
	t.Parallel()
	t.Run("multiple line of raw", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{
			"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"Line 1\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"cge0o\",\"text\":\"Line 2\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"b6kk1\",\"text\":\"Line 3\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", 
			"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/42a25f86169ef2832f9709c78feb7cd6.html"
		}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "Line 1\nLine 2\nLine 3"
		assert.Equal(t, expect, text)
	})
	t.Run("one line of raw", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{
			"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"One Line With Camel Case Of Every Word\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", 
			"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/6b0f6276cb092c4e2741ad6fe1801211.html"
		}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "One Line With Camel Case Of Every Word"
		assert.Equal(t, expect, text)
	})
	t.Run("test case sensitive word", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{
			"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"TeSt CaSE SenSITIvE Of WoRD\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", 
			"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/2c0895233c34870fa264a75aecc2a935.html"
		}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "TeSt CaSE SenSITIvE Of WoRD"
		assert.Equal(t, expect, text)
	})
	t.Run("test case sensitive with multiple line word", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{
			"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"TeSt CaSE SenSITIvE Of WoRD WiTH MulTIPLe Line\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"as5u\",\"text\":\"LiNE 2\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"6t04h\",\"text\":\"LIne 3\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"c8gas\",\"text\":\"LINE 4\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"9pq4g\",\"text\":\"LinE 5\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", 
			"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/44419b701ded7625b2da7c21f9d63558.html"
		}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "TeSt CaSE SenSITIvE Of WoRD WiTH MulTIPLe Line\nLiNE 2\nLIne 3\nLINE 4\nLinE 5"
		assert.Equal(t, expect, text)
	})
	t.Run("test only lower case", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"this text is only contain lower case\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/5c993b3d2b5d4686dac333a361a8fcfc.html"}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "this text is only contain lower case"
		assert.Equal(t, expect, text)
	})
	t.Run("test only lower case with multiple line", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"this is only lower case with multiple line\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"dnfpp\",\"text\":\"this is line 1\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"ckser\",\"text\":\"this is line 2\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"8rbvd\",\"text\":\"this is line 3\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"5b2mi\",\"text\":\"this is line 4\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/7954a7bbb5724488225ad29fb8ebb069.html"}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "this is only lower case with multiple line\nthis is line 1\nthis is line 2\nthis is line 3\nthis is line 4"
		assert.Equal(t, expect, text)
	})
	t.Run("test only upper case", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"THIS TEXT ONLY CONTAINS UPPERCASE CHARACTER\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/320a8e1a32623f5985eda5f38fde3d15.html"}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "THIS TEXT ONLY CONTAINS UPPERCASE CHARACTER"
		assert.Equal(t, expect, text)
	})
	t.Run("test only upper case with multiple line", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{"raw": "{\"blocks\":[{\"key\":\"7pa1\",\"text\":\"THIS TEXT ONLY CONTAINS UPPERCASE CHARACTER WITH MULTIPLE LINE\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"bi96l\",\"text\":\"THIS IS LINE 1\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"1m74r\",\"text\":\"THIS IS LINE 2\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"3no16\",\"text\":\"THIS IS LINE 3\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"5mrfk\",\"text\":\"THIS IS LINE 4\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}},{\"key\":\"4ehvq\",\"text\":\"THIS IS LINE 5\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/04e836206eacc00550cdebdb396ad1a7.html"}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetText()
		expect := "THIS TEXT ONLY CONTAINS UPPERCASE CHARACTER WITH MULTIPLE LINE\nTHIS IS LINE 1\nTHIS IS LINE 2\nTHIS IS LINE 3\nTHIS IS LINE 4\nTHIS IS LINE 5"
		assert.Equal(t, expect, text)
	})
}
