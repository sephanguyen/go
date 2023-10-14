package entities

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LessonGroup struct {
	LessonGroupID pgtype.Text
	CourseID      pgtype.Text
	MediaIDs      pgtype.TextArray
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func (l *LessonGroup) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_group_id",
			"course_id",
			"media_ids",
			"created_at",
			"updated_at",
		}, []interface{}{
			&l.LessonGroupID,
			&l.CourseID,
			&l.MediaIDs,
			&l.CreatedAt,
			&l.UpdatedAt,
		}
}

func (l *LessonGroup) TableName() string {
	return "lesson_groups"
}

func (l *LessonGroup) PreInsert() error {
	if l.LessonGroupID.Status != pgtype.Present {
		if err := l.LessonGroupID.Set(idutil.ULIDNow()); err != nil {
			return err
		}
	}

	now := time.Now()
	err := multierr.Combine(
		l.CreatedAt.Set(now),
		l.UpdatedAt.Set(now),
	)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonGroup) PreUpdate() error {
	return l.UpdatedAt.Set(time.Now())
}

type LessonGroups []*LessonGroup

func (u *LessonGroups) Add() database.Entity {
	e := &LessonGroup{}
	*u = append(*u, e)

	return e
}
