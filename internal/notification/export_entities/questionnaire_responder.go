package exportentities

import "time"

type QuestionnaireAnswer struct {
	QuestionnaireQuestionID string
	Answer                  string
}

type QuestionnaireCSVResponder struct {
	UserNotificationID   string
	UserID               string
	ResponderName        string
	IsParent             bool
	StudentID            string
	TargetID             string
	TargetName           string
	SubmittedAt          time.Time
	IsIndividual         bool
	StudentExternalID    string
	LocationNames        []string
	QuestionnaireAnswers []*QuestionnaireAnswer
}
