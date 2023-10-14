package entity

type StqQuestionData struct {
	*BaseQuestionData
	Validation StqValidation `json:"validation,omitempty"`
}

type StqValidation struct {
	ScoringType   string        `json:"scoring_type,omitempty"`
	ValidResponse StqResponse   `json:"valid_response,omitempty"`
	AltResponses  []StqResponse `json:"alt_responses,omitempty"`
}

type StqResponse struct {
	Score int    `json:"score,omitempty"`
	Value string `json:"value,omitempty"`
}

func (q *StqQuestionData) GetBasicData() *BaseQuestionData {
	return q.BaseQuestionData
}
