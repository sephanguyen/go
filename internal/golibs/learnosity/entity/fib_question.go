package entity

type FibQuestionData struct {
	*BaseQuestionData
	Template   string        `json:"template,omitempty"`
	Validation FibValidation `json:"validation,omitempty"`
}

type FibValidation struct {
	ScoringType   string        `json:"scoring_type,omitempty"`
	ValidResponse FibResponse   `json:"valid_response,omitempty"`
	AltResponses  []FibResponse `json:"alt_responses,omitempty"`
}

type FibResponse struct {
	Value []string `json:"value,omitempty"`
	Score int      `json:"score,omitempty"`
}

func (q *FibQuestionData) GetBasicData() *BaseQuestionData {
	return q.BaseQuestionData
}
