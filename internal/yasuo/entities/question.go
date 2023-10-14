package entities

import (
	"github.com/manabie-com/backend/internal/bob/entities"
)

//QuestionTagLo is record in question_tagged_learning_objectives
type Question struct {
	// TableName struct{} `pg:",discard_unknown_columns"`
	entities.Question
}
