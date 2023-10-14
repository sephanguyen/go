package services

type ItemsBankCsvRow struct {
	LoID                 string `csv:"lo_id"`
	ItemID               string `csv:"item_id"`
	ItemName             string `csv:"item_name"`
	ItemDescriptionText  string `csv:"item_description_text"`
	ItemDescriptionImage string `csv:"item_description_image"`
	QuestionType         string `csv:"question_type"`
	Point                string `csv:"point"`
	QuestionContentText  string `csv:"question_content_text"`
	QuestionContentImage string `csv:"question_content_image"`
	ExplanationText      string `csv:"explanation_text"`
	ExplanationImage     string `csv:"explanation_image"`
	OptionText           string `csv:"option_text"`
	OptionImage          string `csv:"option_image"`
	CorrectOption        string `csv:"correct_option"`
}

func (r *ItemsBankCsvRow) IsQuestionRow() bool {
	return r.QuestionType != "" ||
		r.ItemID != "" ||
		r.QuestionContentText != ""
}
