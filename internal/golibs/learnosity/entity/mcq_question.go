package entity

type McqQuestionData struct {
	*BaseQuestionData
	MultipleResponses bool          `json:"multiple_responses,omitempty"`
	Options           []Options     `json:"options,omitempty"`
	ShuffleOptions    bool          `json:"shuffle_options,omitempty"`
	Validation        McqValidation `json:"validation,omitempty"`
}

type McqValidation struct {
	ScoringType   string      `json:"scoring_type,omitempty"`
	ValidResponse McqResponse `json:"valid_response,omitempty"`
}

type McqResponse struct {
	Value []string `json:"value,omitempty"`
	Score int      `json:"score,omitempty"`
}

type Options struct {
	Value string `json:"value,omitempty"`
	Label string `json:"label,omitempty"`
}

func (q *McqQuestionData) GetBasicData() *BaseQuestionData {
	return q.BaseQuestionData
}
