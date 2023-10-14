package repo

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func NewLessonGroupFromLessonEntity(l *domain.VirtualLesson, courseID pgtype.Text) *LessonGroupDTO {
	lg := &LessonGroupDTO{}
	database.AllNullEntity(lg)
	if l.Material != nil {
		lg.MediaIDs = database.TextArray(l.Material.MediaIDs)
	}
	lg.CourseID = courseID

	return lg
}

type LessonGroupDTO struct {
	LessonGroupID pgtype.Text
	CourseID      pgtype.Text
	MediaIDs      pgtype.TextArray
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func (l *LessonGroupDTO) FieldMap() ([]string, []interface{}) {
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

func (l *LessonGroupDTO) TableName() string {
	return "lesson_groups"
}

func (l *LessonGroupDTO) PreInsert() error {
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

func (l *LessonGroupDTO) PreUpdate() error {
	now := time.Now()
	err := multierr.Combine(
		l.UpdatedAt.Set(now),
	)
	if err != nil {
		return err
	}

	return nil
}
