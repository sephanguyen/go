package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Reallocation struct {
	StudentID        pgtype.Text
	CourseID         pgtype.Text
	OriginalLessonID pgtype.Text
	NewLessonID      pgtype.Text
	UpdatedAt        pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

type Reallocations []*Reallocation

func (r *Reallocation) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_id",
		"course_id",
		"original_lesson_id",
		"new_lesson_id",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	values = []interface{}{
		&r.StudentID,
		&r.CourseID,
		&r.OriginalLessonID,
		&r.NewLessonID,
		&r.UpdatedAt,
		&r.CreatedAt,
		&r.DeletedAt,
	}
	return
}

func (r *Reallocation) TableName() string {
	return "reallocation"
}

func (r *Reallocation) PreUpsert() error {
	now := time.Now()
	if r.CreatedAt.Status != pgtype.Present {
		if err := r.CreatedAt.Set(now); err != nil {
			return err
		}
	}
	if err := r.UpdatedAt.Set(now); err != nil {
		return err
	}

	return nil
}

func NewReallocateStudentFromLessonEntity(l *domain.Lesson) (Reallocations, error) {
	var reallocations Reallocations
	for _, learner := range l.Learners {
		if learner.AttendStatus == domain.StudentAttendStatusReallocate {
			studentReallocate := new(Reallocation)
			database.AllNullEntity(studentReallocate)
			if err := multierr.Combine(
				studentReallocate.StudentID.Set(learner.LearnerID),
				studentReallocate.OriginalLessonID.Set(l.LessonID),
				studentReallocate.CourseID.Set(learner.CourseID),
			); err != nil {
				return nil, fmt.Errorf("could not mapping from students of lesson entity to reallocate student dto: %w", err)
			}
			reallocations = append(reallocations, studentReallocate)
		}
		if learner.Reallocate != nil {
			studentReallocated := new(Reallocation)
			database.AllNullEntity(studentReallocated)
			if err := multierr.Combine(
				studentReallocated.StudentID.Set(learner.LearnerID),
				studentReallocated.OriginalLessonID.Set(learner.OriginalLessonID),
				studentReallocated.CourseID.Set(learner.CourseID),
				studentReallocated.NewLessonID.Set(l.LessonID),
			); err != nil {
				return nil, fmt.Errorf("could not mapping from students of lesson entity to reallocate student dto: %w", err)
			}
			reallocations = append(reallocations, studentReallocated)
		}
	}

	return reallocations, nil
}

func NewReallocateStudentFromEntity(reallocation []*domain.Reallocation) (Reallocations, error) {
	reallocations := make([]*Reallocation, 0, len(reallocation))
	for _, r := range reallocation {
		studentReallocate := new(Reallocation)
		database.AllNullEntity(studentReallocate)
		if err := multierr.Combine(
			studentReallocate.StudentID.Set(r.StudentID),
			studentReallocate.OriginalLessonID.Set(r.OriginalLessonID),
			studentReallocate.CourseID.Set(r.CourseID),
		); err != nil {
			return nil, fmt.Errorf("could not mapping from students of lesson entity to reallocate student dto: %w", err)
		}
		if len(r.NewLessonID) > 0 {
			if err := studentReallocate.NewLessonID.Set(r.NewLessonID); err != nil {
				return nil, err
			}
		}
		reallocations = append(reallocations, studentReallocate)
	}
	return reallocations, nil
}
