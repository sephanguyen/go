package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Lesson struct {
	LessonID             pgtype.Text
	Name                 pgtype.Text
	TeacherID            pgtype.Text // Deprecated
	CourseID             pgtype.Text // Deprecated
	ControlSettings      pgtype.JSONB
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
	EndAt                pgtype.Timestamptz // Deprecated
	StartTime            pgtype.Timestamptz
	EndTime              pgtype.Timestamptz
	LessonGroupID        pgtype.Text
	RoomID               pgtype.Text
	LessonType           pgtype.Text // Deprecated
	Status               pgtype.Text // Deprecated
	StreamLearnerCounter pgtype.Int4
	LearnerIds           pgtype.TextArray // Deprecated
	RoomState            pgtype.JSONB
	TeachingModel        pgtype.Text // Deprecated
	ClassID              pgtype.Text
	CenterID             pgtype.Text
	TeachingMedium       pgtype.Text // old field is LessonType
	TeachingMethod       pgtype.Text // old field is TeachingModel
	SchedulingStatus     pgtype.Text // old field is Status
}

func (l *Lesson) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_id",
			"teacher_id",
			"course_id",
			"control_settings",
			"created_at",
			"updated_at",
			"deleted_at",
			"end_at",
			"lesson_group_id",
			"room_id",
			"lesson_type",
			"status",
			"stream_learner_counter",
			"learner_ids",
			"name",
			"start_time",
			"end_time",
			"room_state",
			"teaching_model",
			"class_id",
			"center_id",
			"teaching_medium",
			"teaching_method",
			"scheduling_status",
		}, []interface{}{
			&l.LessonID,
			&l.TeacherID,
			&l.CourseID,
			&l.ControlSettings,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
			&l.EndAt,
			&l.LessonGroupID,
			&l.RoomID,
			&l.LessonType,
			&l.Status,
			&l.StreamLearnerCounter,
			&l.LearnerIds,
			&l.Name,
			&l.StartTime,
			&l.EndTime,
			&l.RoomState,
			&l.TeachingModel,
			&l.ClassID,
			&l.CenterID,
			&l.TeachingMedium,
			&l.TeachingMethod,
			&l.SchedulingStatus,
		}
}

func (l *Lesson) TableName() string {
	return "lessons"
}

type Lessons []*Lesson

// Add append new Lesson
func (l *Lessons) Add() database.Entity {
	e := &Lesson{}
	*l = append(*l, e)

	return e
}
