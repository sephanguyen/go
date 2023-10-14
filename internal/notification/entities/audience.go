package entities

import (
	"github.com/jackc/pgtype"
)

type Audience struct {
	StudentID pgtype.Text
	ParentID  pgtype.Text

	UserID       pgtype.Text
	Name         pgtype.Text
	Email        pgtype.Text
	CurrentGrade pgtype.Int2
	GradeID      pgtype.Text
	GradeName    pgtype.Text
	UserGroup    pgtype.Text
	ChildIDs     pgtype.TextArray
	ChildNames   pgtype.TextArray
	IsIndividual pgtype.Bool
}
