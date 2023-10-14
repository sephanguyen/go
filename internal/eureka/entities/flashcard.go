package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Flashcard struct {
	LearningMaterial
}

func (t *Flashcard) FieldMap() ([]string, []interface{}) {
	return t.LearningMaterial.FieldMap()
}

func (t *Flashcard) TableName() string {
	return "flash_card"
}

type Flashcards []*Flashcard

func (t *Flashcards) Add() database.Entity {
	e := &Flashcard{}
	*t = append(*t, e)

	return e
}

func (t Flashcards) Get() []*Flashcard {
	return []*Flashcard(t)
}

type FlashcardBase struct {
	Flashcard
	TotalQuestion pgtype.Int4
}
