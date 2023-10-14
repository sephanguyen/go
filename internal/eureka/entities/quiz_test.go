package entities

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		{"label":"3", "key":"key-3", "configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

 	 	{"label": "4",
			"key": "01GEEMA1E8ZHCF7450J18X5PMV",
			"configs": [],
			"content": {
			"raw": "{\"blocks\":[{\"key\":\"e33fb\",\"text\":\"   \",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[{\"offset\":0,\"length\":1,\"key\":0},{\"offset\":1,\"length\":1,\"key\":1},{\"offset\":2,\"length\":1,\"key\":2}],\"data\":{}}],\"entityMap\":{\"0\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\overrightarrow{n_{3}}=(1 ; 2 ;-1)\"}},\"1\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\overrightarrow{n_{2}}=(2 ; 3 ;-1)\"}},\"2\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\vec{n}_{1}=(1 ; 3 ;-1)\"}}}}",
			"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/294fd4fa23ccca4b38ef72ba552ad49b.html"
			},
			"attribute": {
				"configs": ["MATH_CONFIG"],
				"img_link": "",
				"audio_link": ""
			},
			"correctness": true
		},

		{"label":"5", 
			"key":"key-5",
			"configs":[],
			"content": {
				"raw": "{\"blocks\":[{\"key\":\"bp8asdfmr\",\"text\":\" \",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[{\"offset\":0,\"length\":1,\"key\":0}],\"data\":{}}],\"entityMap\":{\"0\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\overrightarrow{n_{4}}=(1 ; 2 ; 3)\"}}}}",
				"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"
			},
			"attribute": {
				"configs": ["MATH_CONFIG"],
				"img_link": "",
				"audio_link": ""
			},
			"correctness": true
		}
		]`)
		optionWithAlternatives, err := quiz.GetOptionsWithAlternatives()
		assert.Nil(t, err)
		answers := []string{"qwewqeqweqwe", "goodbye", "wrong answer", "\\overrightarrow{n_{3}}=(1 ; 2 ;-1)\n\\overrightarrow{n_{2}}=(2 ; 3 ;-1)\n\\vec{n}_{1}=(1 ; 3 ;-1)", "\\overrightarrow{n_{4}}=(1 ; 2 ; 3)"}
		expected := []bool{true, true, false, true, true}
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

func TestRichText_GetEntityMapData(t *testing.T) {
	t.Parallel()
	t.Run("multiple line of raw", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{"raw": "{\"blocks\":[{\"key\":\"e33fb\",\"text\":\"   \",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[{\"offset\":0,\"length\":1,\"key\":0},{\"offset\":1,\"length\":1,\"key\":1},{\"offset\":2,\"length\":1,\"key\":2}],\"data\":{}}],\"entityMap\":{\"0\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\overrightarrow{n_{3}}=(1 ; 2 ;-1)\"}},\"1\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\overrightarrow{n_{2}}=(2 ; 3 ;-1)\"}},\"2\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\vec{n}_{1}=(1 ; 3 ;-1)\"}}}}",
		"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/294fd4fa23ccca4b38ef72ba552ad49b.html"}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetEntityMapData()
		expect := "\\overrightarrow{n_{3}}=(1 ; 2 ;-1)\n\\overrightarrow{n_{2}}=(2 ; 3 ;-1)\n\\vec{n}_{1}=(1 ; 3 ;-1)"
		assert.Equal(t, expect, text)
	})
	t.Run("one line of raw", func(t *testing.T) {
		t.Parallel()
		rc := &RichText{}
		raw := `{"raw": "{\"blocks\":[{\"key\":\"bp8mr\",\"text\":\" \",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[{\"offset\":0,\"length\":1,\"key\":0}],\"data\":{}}],\"entityMap\":{\"0\":{\"type\":\"INLINE_MATHJAX\",\"mutability\":\"IMMUTABLE\",\"data\":{\"data\":\"\\\\overrightarrow{n_{4}}=(1 ; 2 ; 3)\"}}}}",
		"rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/04e836206eacc00550cdebdb396ad1a7.html"}`
		err := json.Unmarshal([]byte(raw), &rc)
		assert.Nil(t, err)
		text := rc.GetEntityMapData()
		expect := "\\overrightarrow{n_{4}}=(1 ; 2 ; 3)"
		assert.Equal(t, expect, text)
	})
}

func TestGenerateRandomSeed(t *testing.T) {
	quizzes := getQuizzes(
		8,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_FIB.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
		cpb.QuizType_QUIZ_TYPE_POW.String(),
		cpb.QuizType_QUIZ_TYPE_TAD.String(),
		cpb.QuizType_QUIZ_TYPE_MIQ.String(),
		cpb.QuizType_QUIZ_TYPE_MAQ.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
	)
	onlyOrdering := getQuizzes(
		8,
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
	)
	quizzes2options := getQuizzes(
		3,
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
	)
	for i := range quizzes2options {
		quizzes2options[i].Options = database.JSONB(`[
				{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true},
				{"key":"key-2" , "label": "2", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so84\",\"text\":\"3213214\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true}
				]`)
	}

	tcs := []struct {
		name            string
		shuffledQuizSet *ShuffledQuizSet
		quizzes         Quizzes
		hasError        bool
	}{
		{
			name: "full quiz type",
			shuffledQuizSet: &ShuffledQuizSet{
				QuizExternalIDs: database.TextArray(quizzes.GetExternalIDs()),
			},
			quizzes:  quizzes,
			hasError: false,
		},
		{
			name: "all is ordering quiz type",
			shuffledQuizSet: &ShuffledQuizSet{
				QuizExternalIDs: database.TextArray(onlyOrdering.GetExternalIDs()),
			},
			quizzes:  onlyOrdering,
			hasError: false,
		},
		{
			name: "quizzes have only 2 options",
			shuffledQuizSet: &ShuffledQuizSet{
				QuizExternalIDs: database.TextArray(quizzes2options.GetExternalIDs()),
			},
			quizzes:  quizzes2options,
			hasError: false,
		},
		{
			name: "not match between quiz external ids and list quizzes",
			shuffledQuizSet: &ShuffledQuizSet{
				QuizExternalIDs: database.TextArray(quizzes2options.GetExternalIDs()),
			},
			quizzes:  onlyOrdering,
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			allReSorted, err := tc.shuffledQuizSet.GenerateRandomSeed(tc.quizzes)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if !allReSorted {
					fmt.Printf("case %s is not successfully \n", tc.name)
					return
				}
				for i, quiz := range tc.quizzes {
					if quiz.Kind.String != cpb.QuizType_QUIZ_TYPE_ORD.String() {
						continue
					}
					var correctOrderKeys []string
					opts, err := quiz.GetOptions()
					require.NoError(t, err)
					for _, opt := range opts {
						correctOrderKeys = append(correctOrderKeys, opt.Key)
					}
					// shuffle option by random seed of shuffle quiz set
					seed, err := strconv.ParseInt(tc.shuffledQuizSet.RandomSeed.String, 10, 64)
					require.NoError(t, err)
					err = quiz.shuffleOptions(seed + int64(i))
					require.NoError(t, err)

					// check options after shuffle options
					haveAnIncorrectOrder := false
					opts, err = quiz.GetOptions()
					require.NoError(t, err)
					for j, opt := range opts {
						if opt.Key != correctOrderKeys[j] {
							haveAnIncorrectOrder = true
							break
						}
					}
					assert.True(t, haveAnIncorrectOrder, "could not generate a random seed to shuffle quiz's option which have at least one answer incorrect order")
				}
			}
		})
	}
}

func getQuizzes(numOfQuizzes int, kind ...string) Quizzes {
	start := database.Timestamptz(timeutil.Now())
	quizzes := Quizzes{}
	r := ring.New(len(kind))
	for i := 0; i < len(kind); i++ {
		r.Value = kind[i]
		r = r.Next()
	}
	ra := rand.New(rand.NewSource(99))
	for i := 0; i < numOfQuizzes; i++ {
		quiz := &Quiz{
			ID:          database.Text(idutil.ULIDNow()),
			ExternalID:  database.Text(idutil.ULIDNow()),
			Country:     database.Text("COUNTRY_VN"),
			SchoolID:    database.Int4(-2147483648),
			Kind:        database.Text("QUIZ_TYPE_MCQ"),
			Question:    database.JSONB(`{"raw": "{\"blocks\":[{\"key\":\"eq20k\",\"text\":\"3213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html"}`),
			Explanation: database.JSONB(`{"raw": "{\"blocks\":[{\"key\":\"4rpf3\",\"text\":\"213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html"}`),
			Options: database.JSONB(`[
				{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true},
				{"key":"key-2" , "label": "2", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so84\",\"text\":\"3213214\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true},
				{"key":"key-3" , "label": "3", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so85\",\"text\":\"3213215\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": false}
				]`),
			TaggedLOs:      database.TextArray([]string{"VN10-CH-01-L-001.1"}),
			DifficultLevel: database.Int4(1),
			CreatedBy:      database.Text("QC6KC30TZWc97APf99ydhonPRct1"),
			ApprovedBy:     database.Text("QC6KC30TZWc97APf99ydhonPRct1"),
			Status:         database.Text("QUIZ_STATUS_APPROVED"),
			UpdatedAt:      start,
			CreatedAt:      start,
			DeletedAt:      pgtype.Timestamptz{},
			Point:          database.Int4(ra.Int31()),
		}
		if len(kind) != 0 {
			quiz.Kind = database.Text(fmt.Sprintf("%v", r.Value))
			if fmt.Sprintf("%v", r.Value) == cpb.QuizType_QUIZ_TYPE_ESQ.String() {
				quiz.Options = database.JSONB(`
				[
					{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true, "answer_config": {"essay": {"limit": 10, "limit_type": "ESSAY_LIMIT_TYPE_WORD", "limit_enabled": true}}}
				]
				`)
			}
			r = r.Next()
		}
		quizzes = append(quizzes, quiz)
	}
	return quizzes
}
