package entities

import (
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
)

// Question is record in question table
type Question struct {
	QuestionID                     pgtype.Text `sql:"question_id"`
	MasterQuestionID               pgtype.Text `sql:"master_question_id"`
	Country                        pgtype.Text
	Question                       pgtype.Text
	Answers                        pgtype.TextArray
	Explanation                    pgtype.Text
	DifficultyLevel                pgtype.Int2
	UpdatedAt                      pgtype.Timestamptz
	CreatedAt                      pgtype.Timestamptz
	QuestionRendered               pgtype.Text
	AnswersRendered                pgtype.TextArray
	ExplanationRendered            pgtype.Text
	IsWaitingForRender             pgtype.Bool
	ExplanationWrongAnswer         pgtype.TextArray
	ExplanationWrongAnswerRendered pgtype.TextArray
	QuestionURL                    pgtype.Text      `sql:"question_url"`
	AnswersURL                     pgtype.TextArray `sql:"answers_url"`
	ExplanationURL                 pgtype.Text      `sql:"explanation_url"`
	ExplanationWrongAnswerURL      pgtype.TextArray `sql:"explanation_wrong_answer_url"`
	RenderingQuestion              pgtype.Bool
	DeletedAt                      pgtype.Timestamptz
}

// FieldMap return a map of field name and pointer to field
func (e *Question) FieldMap() ([]string, []interface{}) {
	return []string{
			"question_id", "master_question_id", "country", "question", "answers", "explanation", "difficulty_level", "updated_at", "created_at", "question_rendered",
			"answers_rendered", "explanation_rendered", "is_waiting_for_render", "explanation_wrong_answer", "explanation_wrong_answer_rendered",
			"question_url", "explanation_url", "answers_url", "explanation_wrong_answer_url", "rendering_question", "deleted_at",
		}, []interface{}{
			&e.QuestionID, &e.MasterQuestionID, &e.Country, &e.Question, &e.Answers, &e.Explanation, &e.DifficultyLevel, &e.UpdatedAt, &e.CreatedAt, &e.QuestionRendered,
			&e.AnswersRendered, &e.ExplanationRendered, &e.IsWaitingForRender, &e.ExplanationWrongAnswer, &e.ExplanationWrongAnswerRendered,
			&e.QuestionURL, &e.ExplanationURL, &e.AnswersURL, &e.ExplanationWrongAnswerURL, &e.RenderingQuestion, &e.DeletedAt,
		}
}

// TableName returns "question"
func (e *Question) TableName() string {
	return "questions"
}

// Questions list of question
type Questions []*Question

// Add used for Select.ScanAll()
func (qs *Questions) Add() database.Entity {
	e := Question{}
	*qs = append(*qs, &e)
	return &e
}

// QuestionTagLo is record in question_tagged_learning_objectives
type QuestionTagLo struct {
	QuestionID   pgtype.Text `sql:"question_id"`
	LoID         pgtype.Text `sql:"lo_id"`
	DisplayOrder pgtype.Int4
}

// FieldMap return a map of field name and pointer to field
func (e *QuestionTagLo) FieldMap() ([]string, []interface{}) {
	return []string{
			"question_id", "lo_id", "display_order",
		}, []interface{}{
			&e.QuestionID, &e.LoID, &e.DisplayOrder,
		}
}

// TableName returns "question"
func (e *QuestionTagLo) TableName() string {
	return "questions_tagged_learning_objectives"
}
