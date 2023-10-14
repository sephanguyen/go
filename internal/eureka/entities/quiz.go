package entities

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/y-bash/go-gaga"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Quiz respresnts quizzes table
type Quiz struct {
	ID              pgtype.Text
	ExternalID      pgtype.Text
	Country         pgtype.Text
	SchoolID        pgtype.Int4
	LoIDs           pgtype.TextArray
	Kind            pgtype.Text
	Question        pgtype.JSONB // RichText
	Explanation     pgtype.JSONB // RichText
	Options         pgtype.JSONB // QuizOptions
	TaggedLOs       pgtype.TextArray
	DifficultLevel  pgtype.Int4
	Point           pgtype.Int4
	QuestionTagIds  pgtype.TextArray
	CreatedBy       pgtype.Text
	ApprovedBy      pgtype.Text
	Status          pgtype.Text
	UpdatedAt       pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
	QuestionGroupID pgtype.Text
	LabelType       pgtype.Text
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
	var r []*QuizOption
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
		"point",
		"question_tag_ids",
		"created_by",
		"approved_by",
		"status",
		"updated_at",
		"created_at",
		"deleted_at",
		"question_group_id",
		"label_type",
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
		&e.Point,
		&e.QuestionTagIds,
		&e.CreatedBy,
		&e.ApprovedBy,
		&e.Status,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
		&e.QuestionGroupID,
		&e.LabelType,
	}
}

// TableName returns "quizzes" table name
func (e *Quiz) TableName() string {
	return "quizzes"
}

func (e *Quiz) ShuffleOptions(seed int64) error {
	return e.shuffleOptions(seed)
}

func (e *Quiz) shuffleOptions(seed int64) error {
	if e.Kind.String != cpb.QuizType_QUIZ_TYPE_MCQ.String() &&
		e.Kind.String != cpb.QuizType_QUIZ_TYPE_MAQ.String() &&
		e.Kind.String != cpb.QuizType_QUIZ_TYPE_ORD.String() {
		return nil
	}

	var options []*QuizOption
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
// TODO: refactor later to make sure option always be shuffled
func (u Quizzes) ShuffleOptions(seed int64, offset int64, idMap map[string]int64) error {
	for _, quiz := range u {
		err := quiz.shuffleOptions(seed + offset + idMap[quiz.ExternalID.String])
		if err != nil {
			return err
		}
	}
	return nil
}

func (u Quizzes) GetExternalIDs() []string {
	ids := make([]string, 0, len(u))
	for _, q := range u {
		ids = append(ids, q.ExternalID.String)
	}
	return ids
}

// GetQuestionGroupIDs will return all question groups of Quizzes not duplicated
func (u Quizzes) GetQuestionGroupIDs() []string {
	ids := make([]string, 0, len(u))
	had := make(map[string]bool)
	for _, q := range u {
		if _, ok := had[q.QuestionGroupID.String]; ok {
			continue
		}
		had[q.QuestionGroupID.String] = true
		ids = append(ids, q.QuestionGroupID.String)
	}
	return ids
}

// QuizSet represent quiz_sets table
type QuizSet struct {
	ID                pgtype.Text
	LoID              pgtype.Text
	QuizExternalIDs   pgtype.TextArray
	Status            pgtype.Text
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	QuestionHierarchy pgtype.JSONBArray
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
		"question_hierarchy",
	}

	return names, []interface{}{
		&e.ID,
		&e.LoID,
		&e.QuizExternalIDs,
		&e.Status,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
		&e.QuestionHierarchy,
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
	StudyPlanID              pgtype.Text
	LearningMaterialID       pgtype.Text
	QuestionHierarchy        pgtype.JSONBArray
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
		"study_plan_id",
		"learning_material_id",
		"question_hierarchy",
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
		&e.StudyPlanID,
		&e.LearningMaterialID,
		&e.QuestionHierarchy,
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

func (e *ShuffledQuizSet) GetSeedOfQuestionByExternalID(quesExternalID string) (int64, error) {
	seed, err := strconv.ParseInt(e.RandomSeed.String, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse seed: %v", err)
	}

	if idx := slices.Index(database.FromTextArray(e.QuizExternalIDs), quesExternalID); idx > -1 {
		return seed + int64(idx), nil
	}

	return -1, fmt.Errorf("could not found %s question external id", quesExternalID)
}

const maximumTriesNumber = 10

// GenerateRandomSeed will generate a random seed by try use it to shuffle
// all quiz's options of ShuffledQuizSet to make sure there are no any
// options still correct, maximum <maximumTriesNumber> times. Currently, just care ordering question type.
func (e *ShuffledQuizSet) GenerateRandomSeed(quizzes Quizzes) (allReSorted bool, err error) {
	quizzesMap := make(map[string]*Quiz)
	for _, quiz := range quizzes {
		q := *quiz
		quizzesMap[quiz.ExternalID.String] = &q
	}

Label:
	for i := 0; i < maximumTriesNumber; i++ {
		_ = e.RandomSeed.Set(strconv.FormatInt(time.Now().UTC().UnixNano(), 10))

		for _, id := range e.QuizExternalIDs.Elements {
			q, ok := quizzesMap[id.String]
			if !ok {
				return false, fmt.Errorf("could not found quiz id %s", id.String)
			}
			if q.Kind.String != cpb.QuizType_QUIZ_TYPE_ORD.String() {
				continue
			}
			var options []*QuizOption
			if err = q.Options.AssignTo(&options); err != nil {
				return false, err
			}
			seed, _ := e.GetSeedOfQuestionByExternalID(id.String)
			if err = q.ShuffleOptions(seed); err != nil {
				return false, fmt.Errorf("q.ShuffleOptions: %w", err)
			}
			var lastOpts []*QuizOption
			if err = q.Options.AssignTo(&lastOpts); err != nil {
				return false, err
			}
			// check order answer after shuffle
			isSuccess := false
			for j := range options {
				if options[j].Key != lastOpts[j].Key {
					isSuccess = true
					break
				}
			}
			// if there is a quiz not success, retry generate random seed again
			if !isSuccess {
				continue Label
			}
		}
		allReSorted = true
		break
	}

	return allReSorted, nil
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

func (rc *RichText) GetEntityMapData() string {
	type data struct {
		Data string `json:"data"`
	}

	type entityMap struct {
		Type       string `json:"type"`
		Mutability string `json:"mutability"`
		DataField  data   `json:"data"`
	}
	type raw struct {
		EntityMap map[string]entityMap `json:"entityMap"`
	}

	r := raw{}
	err := json.Unmarshal([]byte(rc.Raw), &r)
	if err != nil {
		return ""
	}

	// sort key of map for result right
	keys := make([]string, 0, len(r.EntityMap))
	for k := range r.EntityMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	resData := []string{}
	for _, key := range keys {
		// every block is one line
		resData = append(resData, r.EntityMap[key].DataField.Data)
	}
	return strings.TrimSpace(strings.Join(resData, "\n"))
}

// QuizOptionWithAlternatives option with multiple alternatives
type QuizOptionWithAlternatives struct {
	Key                string
	AlternativeOptions []*QuizOption
}

func (opt *QuizOptionWithAlternatives) IsCorrect(userAnswer string) bool {
	norm, _ := gaga.Norm(gaga.Fold)
	tabSpaceRegex := regexp.MustCompile(`\s+`)

	answerText := norm.String(userAnswer)
	answerText = strings.TrimSpace(answerText)
	answerText = tabSpaceRegex.ReplaceAllString(answerText, " ")

	for _, alter := range opt.AlternativeOptions {
		inCasedAnswerText := answerText

		correctText := norm.String(alter.GetText())
		correctText = strings.TrimSpace(correctText)
		correctText = tabSpaceRegex.ReplaceAllString(correctText, " ")

		correctData := strings.TrimSpace(alter.GetEntityMapData())
		correctData = tabSpaceRegex.ReplaceAllString(correctData, " ")

		if correctText != "" && !utils.IsContain(alter.Configs, QuizOptionConfigCaseSensitive) {
			inCasedAnswerText = strings.ToLower(inCasedAnswerText)
			correctText = strings.ToLower(correctText)
		}
		switch utils.IsContain(alter.Attribute.Configs, cpb.QuizItemAttributeConfig_MATH_CONFIG.String()) {
		case true: // Math
			if inCasedAnswerText == correctData {
				return true
			}
		case false: // English, Japanese
			if inCasedAnswerText == correctText {
				return true
			}
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

func (opt *QuizOptionWithAlternatives) GetEntityMapData() string {
	if len(opt.AlternativeOptions) == 0 {
		return ""
	}
	data := opt.AlternativeOptions[0].GetEntityMapData()
	return data
}

// QuizOption option of quiz
type QuizOption struct {
	Content      RichText          `json:"content"`
	Correctness  bool              `json:"correctness"`
	Configs      []string          `json:"configs"`
	Label        string            `json:"label"`
	Key          string            `json:"key"`
	Attribute    QuizItemAttribute `json:"attribute"`
	AnswerConfig AnswerConfig      `json:"answer_config"`
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

// GetEntityMapData return raw entityMap data content of a RichText content
func (qo *QuizOption) GetEntityMapData() string {
	return qo.Content.GetEntityMapData()
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
	Point         uint32    `json:"point"`
	SubmittedKeys []string  `json:"submitted_keys"`
	CorrectKeys   []string  `json:"correct_keys"`
}

func (q *QuizAnswer) ToAnswerLogProtoMessage() *cpb.AnswerLog {
	res := &cpb.AnswerLog{
		QuizId:        q.QuizID,
		QuizType:      cpb.QuizType(cpb.QuizType_value[q.QuizType]),
		SelectedIndex: q.SelectedIndex,
		CorrectIndex:  q.CorrectIndex,
		IsAccepted:    q.IsAccepted,
		CorrectText:   q.CorrectText,
		Correctness:   q.Correctness,
		SubmittedAt:   timestamppb.New(q.SubmittedAt),
		FilledText:    q.FilledText,
	}
	if len(q.CorrectKeys) != 0 || len(q.SubmittedKeys) != 0 {
		res.Result = &cpb.AnswerLog_OrderingResult{
			OrderingResult: &cpb.OrderingResult{
				SubmittedKeys: q.SubmittedKeys,
				CorrectKeys:   q.CorrectKeys,
			},
		}
	}

	return res
}

type AnswerConfig struct {
	Essay EssayConfig `json:"essay"`
}

type EssayLimitType string

const (
	EssayLimitTypeCharacter EssayLimitType = "ESSAY_LIMIT_TYPE_CHAR"
	EssayLimitTypeWord      EssayLimitType = "ESSAY_LIMIT_TYPE_WORD"
)

type EssayConfig struct {
	LimitEnabled bool           `json:"limit_enabled"`
	LimitType    EssayLimitType `json:"limit_type"`
	Limit        uint32         `json:"limit"`
}

type CorrectnessInfo struct {
	RandomSeed               pgtype.Text
	QuizIndex                pgtype.Int4
	TotalSubmissionHistory   pgtype.Int4
	TotalQuizExternalIDs     pgtype.Int4
	OriginalShuffleQuizSetID pgtype.Text
	StudentID                pgtype.Text
	TotalCorrectness         pgtype.Int4
	LoID                     pgtype.Text
}

func (e *CorrectnessInfo) FieldMap() ([]string, []interface{}) {
	names := []string{
		"random_seed",
		"quiz_index",
		"total_submission_history",
		"total_quiz_external_ids",
		"original_shuffle_quiz_set_id",
		"student_id",
		"total_correctness",
		"lo_id",
	}
	return names, []interface{}{
		&e.RandomSeed,
		&e.QuizIndex,
		&e.TotalSubmissionHistory,
		&e.TotalQuizExternalIDs,
		&e.OriginalShuffleQuizSetID,
		&e.StudentID,
		&e.TotalCorrectness,
		&e.LoID,
	}
}

func (e *CorrectnessInfo) TableName() string {
	return "shuffled_quiz_sets"
}
