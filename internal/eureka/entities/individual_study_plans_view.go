package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type IndividualStudyPlansView struct {
	StudentID           pgtype.Text
	StudyPlanID         pgtype.Text
	BookID              pgtype.Text
	ChapterID           pgtype.Text
	ChapterDisplayOrder pgtype.Int2
	TopicID             pgtype.Text
	TopicDisplayOrder   pgtype.Int2
	LearningMaterialID  pgtype.Text
	LmDisplayOrder      pgtype.Int2
}

func (rcv *IndividualStudyPlansView) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_id",
		"study_plan_id",
		"book_id",
		"chapter_id",
		"chapter_display_order",
		"topic_id",
		"topic_display_order",
		"learning_material_id",
		"lm_display_order",
	}
	values = []interface{}{
		&rcv.StudentID,
		&rcv.StudyPlanID,
		&rcv.BookID,
		&rcv.ChapterID,
		&rcv.ChapterDisplayOrder,
		&rcv.TopicID,
		&rcv.TopicDisplayOrder,
		&rcv.LearningMaterialID,
		&rcv.LmDisplayOrder,
	}
	return
}

func (rcv *IndividualStudyPlansView) TableName() string {
	return "individual_study_plans_view"
}

type IndividualStudyPlansViews []*IndividualStudyPlansView

func (ss *IndividualStudyPlansViews) Add() database.Entity {
	e := &IndividualStudyPlansView{}
	*ss = append(*ss, e)

	return e
}
