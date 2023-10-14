package entity

type OrdQuestionData struct {
	*BaseQuestionData
	List           []string      `json:"list,omitempty"`
	ShuffleOptions bool          `json:"shuffle_options,omitempty"`
	Validation     OrdValidation `json:"validation,omitempty"`
}

type OrdValidation struct {
	ScoringType   string      `json:"scoring_type,omitempty"`
	ValidResponse OrdResponse `json:"valid_response,omitempty"`
}
type OrdResponse struct {
	Score int   `json:"score,omitempty"`
	Value []int `json:"value,omitempty"`
}

func (q *OrdQuestionData) GetBasicData() *BaseQuestionData {
	return q.BaseQuestionData
}
