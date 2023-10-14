package entities

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
)

// Quiz respresnts quizzes table
type Quiz struct {
	ID             pgtype.Text
	ExternalID     pgtype.Text
	Country        pgtype.Text
	SchoolID       pgtype.Int4
	LoIDs          pgtype.TextArray
	Kind           pgtype.Text
	Question       pgtype.JSONB // RichText
	Explanation    pgtype.JSONB // RichText
	Options        pgtype.JSONB // QuizOptions
	TaggedLOs      pgtype.TextArray
	DifficultLevel pgtype.Int4
	CreatedBy      pgtype.Text
	ApprovedBy     pgtype.Text
	Status         pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

type QuizConfig struct {
	PartialCredit bool
	CaseSensitive bool
	PlanList      bool
}

func (e *Quiz) GetQuestionV2() (*QuizQuestion, error) {
	r := &QuizQuestion{}
	err := e.Question.AssignTo(r)
	return r, err
}

func (e *Quiz) GetQuestion() (*RichText, error) {
	r := &RichText{}
	err := e.Question.AssignTo(r)
	return r, err
}

func (e *Quiz) GetExplaination() (*RichText, error) {
	r := &RichText{}
	err := e.Explanation.AssignTo(r)
	return r, err
}

func (e *Quiz) GetOptions() ([]*QuizOption, error) {
	r := []*QuizOption{}
	err := e.Options.AssignTo(&r)
	return r, err
}

func (e *Quiz) GetOptionsWithAlternatives() ([]*QuizOptionWithAlternatives, error) {
	// Example:
	// input options with keys : ['1', '1', '2', '3', '1', '1', '2', '3', '1', '3']
	// output : map
	// '1': [option of key 1] with len = 5
	// '2': [option of key 2] with len = 2
	// '3': [option of key 3] with len = 3
	options, err := e.GetOptions()
	if err != nil {
		return nil, err
	}
	optionWithAlternatives := make([]*QuizOptionWithAlternatives, 0, len(options))

	orderdKeys := make([]string, 0, len(options))
	usedKey := make(map[string]struct{})

	for _, opt := range options {
		k := opt.Key
		_, ok := usedKey[k]
		if !ok {
			orderdKeys = append(orderdKeys, k)
			usedKey[k] = struct{}{}
		}
	}

	mp := make(map[string]*QuizOptionWithAlternatives)
	for _, opt := range options {
		k := opt.Key
		if _, ok := mp[k]; !ok {
			mp[k] = &QuizOptionWithAlternatives{
				Key:                k,
				AlternativeOptions: make([]*QuizOption, 0, len(options)),
			}
		}
		mp[k].AlternativeOptions = append(mp[k].AlternativeOptions, opt)
	}

	for _, key := range orderdKeys {
		if key != "" {
			optionWithAlternatives = append(optionWithAlternatives, mp[key])
		}
	}

	if _, ok := mp[""]; ok {
		// every option with empty key will be counted as separate options
		for _, opt := range mp[""].AlternativeOptions {
			optionWithAlternatives = append(optionWithAlternatives, &QuizOptionWithAlternatives{
				Key:                "",
				AlternativeOptions: []*QuizOption{opt},
			})
		}
	}

	return optionWithAlternatives, nil
}

// FieldMap return a map of field name and pointer to field
func (e *Quiz) FieldMap() ([]string, []interface{}) {
	names := []string{
		"quiz_id",
		"country",
		"school_id",
		"lo_ids",
		"external_id",
		"kind",
		"question",
		"explanation",
		"options",
		"tagged_los",
		"difficulty_level",
		"created_by",
		"approved_by",
		"status",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	return names, []interface{}{
		&e.ID,
		&e.Country,
		&e.SchoolID,
		&e.LoIDs,
		&e.ExternalID,
		&e.Kind,
		&e.Question,
		&e.Explanation,
		&e.Options,
		&e.TaggedLOs,
		&e.DifficultLevel,
		&e.CreatedBy,
		&e.ApprovedBy,
		&e.Status,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
	}
}

// TableName returns "quizzes" table name
func (e *Quiz) TableName() string {
	return "quizzes"
}

func (e *Quiz) shuffleOptions(seed int64) error {
	if e.Kind.String != cpb.QuizType_QUIZ_TYPE_MCQ.String() && e.Kind.String != cpb.QuizType_QUIZ_TYPE_MAQ.String() {
		return nil
	}

	options := []*QuizOption{}
	err := e.Options.AssignTo(&options)
	if err != nil {
		return err
	}
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	optionsJSONB, _ := json.Marshal(options)
	err = e.Options.Set(optionsJSONB)
	if err != nil {
		return err
	}
	return nil
}

// Quizzes implement Entities interface
type Quizzes []*Quiz

// Add append new QuizSet
func (u *Quizzes) Add() database.Entity {
	e := &Quiz{}
	*u = append(*u, e)

	return e
}

// Get Quiz by index
func (u Quizzes) Get(index int) *Quiz {
	q := ([]*Quiz)(u)[index]
	return q
}

// ShuffleOptions will shuffle options of all quiz
func (u Quizzes) ShuffleOptions(seed int64, offset int64, idMap map[string]int64) error {
	for _, quiz := range u {
		err := quiz.shuffleOptions(seed + offset + idMap[quiz.ExternalID.String])
		if err != nil {
			return err
		}
	}
	return nil
}

// QuizSet represent quiz_sets table
type QuizSet struct {
	ID              pgtype.Text
	LoID            pgtype.Text
	QuizExternalIDs pgtype.TextArray
	Status          pgtype.Text
	UpdatedAt       pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
}

// FieldMap return a map of field name and pointer to field
func (e *QuizSet) FieldMap() ([]string, []interface{}) {
	names := []string{
		"quiz_set_id",
		"lo_id",
		"quiz_external_ids",
		"status",
		"updated_at",
		"created_at",
		"deleted_at",
	}

	return names, []interface{}{
		&e.ID,
		&e.LoID,
		&e.QuizExternalIDs,
		&e.Status,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
	}
}

// TableName returns "quiz_sets" table name
func (e *QuizSet) TableName() string {
	return "quiz_sets"
}

// ShuffledQuizSet suffled quizzes list from original quiz set
type ShuffledQuizSet struct {
	ID                       pgtype.Text
	OriginalQuizSetID        pgtype.Text
	QuizExternalIDs          pgtype.TextArray
	Status                   pgtype.Text
	RandomSeed               pgtype.Text
	UpdatedAt                pgtype.Timestamptz
	CreatedAt                pgtype.Timestamptz
	DeletedAt                pgtype.Timestamptz
	StudentID                pgtype.Text
	StudyPlanItemID          pgtype.Text
	TotalCorrectness         pgtype.Int4
	SubmissionHistory        pgtype.JSONB
	SessionID                pgtype.Text
	OriginalShuffleQuizSetID pgtype.Text
}

// FieldMap return a map of field name and pointer to field
func (e *ShuffledQuizSet) FieldMap() ([]string, []interface{}) {
	names := []string{
		"shuffled_quiz_set_id",
		"original_quiz_set_id",
		"quiz_external_ids",
		"status",
		"random_seed",
		"updated_at",
		"created_at",
		"deleted_at",
		"student_id",
		"study_plan_item_id",
		"total_correctness",
		"submission_history",
		"session_id",
		"original_shuffle_quiz_set_id",
	}

	return names, []interface{}{
		&e.ID,
		&e.OriginalQuizSetID,
		&e.QuizExternalIDs,
		&e.Status,
		&e.RandomSeed,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
		&e.StudentID,
		&e.StudyPlanItemID,
		&e.TotalCorrectness,
		&e.SubmissionHistory,
		&e.SessionID,
		&e.OriginalShuffleQuizSetID,
	}
}

// TableName returns "shuffled_quiz_sets" table name
func (e *ShuffledQuizSet) TableName() string {
	return "shuffled_quiz_sets"
}

// ShuffleQuiz will shuffle the quiz external ids
func (e *ShuffledQuizSet) ShuffleQuiz() {
	seed, _ := strconv.ParseInt(e.RandomSeed.String, 10, 64)
	r := rand.New(rand.NewSource(seed))
	eqIDs := e.QuizExternalIDs.Elements
	r.Shuffle(len(eqIDs), func(i, j int) { eqIDs[i], eqIDs[j] = eqIDs[j], eqIDs[i] })
}

// QuizSets implement Entities interface
type QuizSets []*QuizSet

// Add append new QuizSet
func (u *QuizSets) Add() database.Entity {
	e := &QuizSet{}
	*u = append(*u, e)

	return e
}

// ShuffledQuizSets implement Entities interface
type ShuffledQuizSets []*ShuffledQuizSet

// Add append new ShuffledQuizSet
func (u *ShuffledQuizSets) Add() database.Entity {
	e := &ShuffledQuizSet{}
	*u = append(*u, e)

	return e
}

// RichText custom data type
type RichText struct {
	Raw         string `json:"raw"`
	RenderedURL string `json:"rendered_url"`
}

type QuizQuestion struct {
	Raw         string            `json:"raw"`
	RenderedURL string            `json:"rendered_url"`
	Configs     []string          `json:"configs"`
	Attribute   QuizItemAttribute `json:"attribute"`
}

type QuizItemAttribute struct {
	AudioLink string   `json:"audio_link"`
	ImgLink   string   `json:"img_link"`
	Configs   []string `json:"configs"`
}

func (qq *QuizQuestion) GetText() string {
	type block struct {
		Text string `json:"text"`
	}

	type raw struct {
		Blocks []block `json:"blocks"`
	}

	r := raw{}
	err := json.Unmarshal([]byte(qq.Raw), &r)
	if err != nil {
		return ""
	}

	content := []string{}
	for _, block := range r.Blocks {
		// every block is one line
		content = append(content, block.Text)
	}

	return strings.TrimSpace(strings.Join(content, "\n"))
}

func (rc *RichText) GetText() string {
	type block struct {
		Text string `json:"text"`
	}

	type raw struct {
		Blocks []block `json:"blocks"`
	}

	r := raw{}
	err := json.Unmarshal([]byte(rc.Raw), &r)

	if err != nil {
		return ""
	}

	content := []string{}
	for _, block := range r.Blocks {
		// every block is one line
		content = append(content, block.Text)
	}

	return strings.TrimSpace(strings.Join(content, "\n"))
}

// QuizOptionWithAlternatives option with multiple alternatives
type QuizOptionWithAlternatives struct {
	Key                string
	AlternativeOptions []*QuizOption
}

func (opt *QuizOptionWithAlternatives) IsCorrect(userAnswer string) bool {
	for _, alter := range opt.AlternativeOptions {
		userText := strings.TrimSpace(userAnswer)
		correctText := strings.TrimSpace(alter.GetText())

		rCompile := regexp.MustCompile(`([\t\x{3000}\x{0020}]{1,}|[\x{0020}]{2,})`)
		userText = rCompile.ReplaceAllString(userText, " ")
		correctText = rCompile.ReplaceAllString(correctText, " ")

		if !utils.IsContain(alter.Configs, QuizOptionConfigCaseSensitive) {
			userText = strings.ToLower(userText)
			correctText = strings.ToLower(correctText)
		}

		if userText == correctText {
			return true
		}
	}
	return false
}

func (opt *QuizOptionWithAlternatives) GetText() string {
	if len(opt.AlternativeOptions) == 0 {
		return ""
	}
	text := opt.AlternativeOptions[0].GetText()
	return text
}

// QuizOption option of quiz
type QuizOption struct {
	Content     RichText          `json:"content"`
	Correctness bool              `json:"correctness"`
	Configs     []string          `json:"configs"`
	Label       string            `json:"label"`
	Key         string            `json:"key"`
	Attribute   QuizItemAttribute `json:"attribute"`
}

var (
	QuizOptionConfigCaseSensitive = cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE.String()
	QuizOptionConfigPartialCredit = cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PARTIAL_CREDIT.String()
	QuizOptionConfigPlanList      = cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_PLAN_LIST.String()
)

// GetText return raw text content of a RichText Content
func (qo *QuizOption) GetText() string {
	return qo.Content.GetText()
}

// QuizAnswer store student answer
type QuizAnswer struct {
	QuizID        string    `json:"quiz_id"`
	QuizType      string    `json:"quiz_type"`
	SelectedIndex []uint32  `json:"selected_index"`
	CorrectIndex  []uint32  `json:"correct_index"`
	FilledText    []string  `json:"filled_text"`
	CorrectText   []string  `json:"correct_text"`
	Correctness   []bool    `json:"correctness"`
	IsAccepted    bool      `json:"is_accepted"`
	SubmittedAt   time.Time `json:"submitted_at"`
	IsAllCorrect  bool      `json:"is_all_correct"`
}
