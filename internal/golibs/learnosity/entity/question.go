package entity

const (
	MultipleChoice = "mcq"
	Ordering       = "orderlist"
	FillInTheBlank = "clozetext"
	ShortText      = "shorttext"
)

const (
	ScoringTypeExactMatch = "exactMatch"
	UIStyleTypeHorizontal = "horizontal"
)

type Question struct {
	Reference string                `json:"reference,omitempty"`
	Type      string                `json:"type,omitempty"`
	Data      QuestionDataInterface `json:"data,omitempty"`
}

type QuestionDataInterface interface {
	GetBasicData() *BaseQuestionData
}

type Metadata struct {
	DistractorRationale string `json:"distractor_rationale,omitempty"`
	MIQHintExplanation  string `json:"miq_hint_explanation,omitempty"`
}

type BaseQuestionData struct {
	Type             string   `json:"type,omitempty"`
	Stimulus         string   `json:"stimulus,omitempty"`
	Metadata         Metadata `json:"metadata,omitempty"`
	FeedbackAttempts int      `json:"feedback_attempts,omitempty"`
	InstantFeedback  bool     `json:"instant_feedback,omitempty"`
	IsMath           bool     `json:"is_math,omitempty"`
}
