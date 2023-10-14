package entities

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type (
	LessonStatus         string
	LessonType           string
	LessonTeachingMethod string
	LessonTeachingMedium string
	SchedulingStatus     string
)

const (
	LessonStatusNone       LessonStatus = "LESSON_STATUS_NONE"
	LessonStatusCompleted  LessonStatus = "LESSON_STATUS_COMPLETED"
	LessonStatusInProgress LessonStatus = "LESSON_STATUS_IN_PROGRESS"
	LessonStatusNotStarted LessonStatus = "LESSON_STATUS_NOT_STARTED"
	LessonStatusDraft      LessonStatus = "LESSON_STATUS_DRAFT"

	LessonTypeNone    LessonType = "LESSON_TYPE_NONE"
	LessonTypeOnline  LessonType = "LESSON_TYPE_ONLINE"
	LessonTypeOffline LessonType = "LESSON_TYPE_OFFLINE"
	LessonTypeHybrid  LessonType = "LESSON_TYPE_HYBRID"

	LessonTeachingMethodIndividual LessonTeachingMethod = "LESSON_TEACHING_METHOD_INDIVIDUAL"
	LessonTeachingMethodGroup      LessonTeachingMethod = "LESSON_TEACHING_METHOD_GROUP"

	LessonTeachingMediumOffline LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_OFFLINE"
	LessonTeachingMediumOnline  LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_ONLINE"
	LessonTeachingMediumZoom    LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_ZOOM"

	LessonSchedulingStatusPublished SchedulingStatus = "LESSON_SCHEDULING_STATUS_PUBLISHED"
	LessonSchedulingStatusDraft     SchedulingStatus = "LESSON_SCHEDULING_STATUS_DRAFT"
	LessonSchedulingStatusCompleted SchedulingStatus = "LESSON_SCHEDULING_STATUS_COMPLETED"
	LessonSchedulingStatusCanceled  SchedulingStatus = "LESSON_SCHEDULING_STATUS_CANCELED"
)

type TeacherIDs struct {
	TeacherIDs pgtype.TextArray
}

func (ids TeacherIDs) HaveID(id pgtype.Text) bool {
	for _, v := range ids.TeacherIDs.Elements {
		if v.String == id.String {
			return true
		}
	}

	return false
}

type CourseIDs struct {
	CourseIDs pgtype.TextArray
}

type LearnerIDs struct {
	LearnerIDs pgtype.TextArray
}

func (ids LearnerIDs) HaveID(id pgtype.Text) bool {
	for _, v := range ids.LearnerIDs.Elements {
		if v.String == id.String {
			return true
		}
	}

	return false
}

type Lesson struct {
	LessonID             pgtype.Text
	Name                 pgtype.Text
	TeacherID            pgtype.Text
	CourseID             pgtype.Text
	ControlSettings      pgtype.JSONB
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
	EndAt                pgtype.Timestamptz
	StartTime            pgtype.Timestamptz
	EndTime              pgtype.Timestamptz
	LessonGroupID        pgtype.Text
	RoomID               pgtype.Text
	LessonType           pgtype.Text
	Status               pgtype.Text
	StreamLearnerCounter pgtype.Int4
	LearnerIds           pgtype.TextArray
	RoomState            pgtype.JSONB
	TeachingModel        pgtype.Text
	ClassID              pgtype.Text
	CenterID             pgtype.Text
	TeachingMethod       pgtype.Text
	TeachingMedium       pgtype.Text
	SchedulingStatus     pgtype.Text
	IsLocked             pgtype.Bool
	ZoomLink             pgtype.Text
	TeacherIDs
	CourseIDs
	LearnerIDs
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
			"teaching_method",
			"teaching_medium",
			"scheduling_status",
			"is_locked",
			"zoom_link",
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
			&l.TeachingMethod,
			&l.TeachingMedium,
			&l.SchedulingStatus,
			&l.IsLocked,
			&l.ZoomLink,
		}
}

func (l *Lesson) TableName() string {
	return "lessons"
}

type LessonsTeachers struct {
	TeacherID pgtype.Text
	LessonID  pgtype.Text
}

func (t *LessonsTeachers) FieldMap() ([]string, []interface{}) {
	return []string{
			"teacher_id", "lesson_id",
		}, []interface{}{
			&t.TeacherID, &t.LessonID,
		}
}

func (t *LessonsTeachers) TableName() string {
	return "lessons_teachers"
}

func (l *Lesson) Normalize() error {
	// default CourseID by first element of CourseIDs
	if l.CourseID.Status != pgtype.Present &&
		l.CourseIDs.CourseIDs.Status == pgtype.Present &&
		len(l.CourseIDs.CourseIDs.Elements) != 0 {
		l.CourseID = l.CourseIDs.CourseIDs.Elements[0]
	} else if len(l.CourseIDs.CourseIDs.Elements) == 0 {
		err := l.CourseIDs.CourseIDs.Set([]pgtype.Text{l.CourseID})
		if err != nil {
			return err
		}
	}

	// default TeacherID by first element of TeacherIDs
	if l.TeacherID.Status != pgtype.Present &&
		l.TeacherIDs.TeacherIDs.Status == pgtype.Present &&
		len(l.TeacherIDs.TeacherIDs.Elements) != 0 {
		l.TeacherID = l.TeacherIDs.TeacherIDs.Elements[0]
	} else if len(l.TeacherIDs.TeacherIDs.Elements) == 0 {
		err := l.TeacherIDs.TeacherIDs.Set([]pgtype.Text{l.TeacherID})
		if err != nil {
			return err
		}
	}

	if l.LessonType.Status != pgtype.Present {
		err := l.LessonType.Set(string(LessonTypeNone))
		if err != nil {
			return err
		}
	}

	if l.Status.Status != pgtype.Present {
		err := l.Status.Set(string(LessonStatusNone))
		if err != nil {
			return err
		}
	}

	if l.StreamLearnerCounter.Status != pgtype.Present {
		err := l.StreamLearnerCounter.Set(0)
		if err != nil {
			return err
		}
	}

	if l.LearnerIds.Status != pgtype.Present {
		err := l.LearnerIds.Set([]string{})
		if err != nil {
			return err
		}
	}

	if l.IsLocked.Status != pgtype.Present {
		err := l.IsLocked.Set(false)
		if err != nil {
			return err
		}
	}

	teacherIDs := database.FromTextArray(l.TeacherIDs.TeacherIDs)
	teacherIDs = golibs.GetUniqueElementStringArray(teacherIDs)
	if err := l.TeacherIDs.TeacherIDs.Set(teacherIDs); err != nil {
		return err
	}

	courseIDs := database.FromTextArray(l.CourseIDs.CourseIDs)
	courseIDs = golibs.GetUniqueElementStringArray(courseIDs)
	if err := l.CourseIDs.CourseIDs.Set(courseIDs); err != nil {
		return err
	}

	learnerIDs := database.FromTextArray(l.LearnerIDs.LearnerIDs)
	learnerIDs = golibs.GetUniqueElementStringArray(learnerIDs)
	if err := l.LearnerIDs.LearnerIDs.Set(learnerIDs); err != nil {
		return err
	}

	return nil
}

// IsValid checks the Lesson before create/update.
// Spec: https://manabie.atlassian.net/browse/LT-1685
//   - Require lesson name, start time, end time
//   - Require at least 1 teacher
//   - Require at least 1 course
//   - Require at least 1 student
func (l *Lesson) IsValid() error {
	if l.Name.Status != pgtype.Present || len(l.Name.String) == 0 {
		return fmt.Errorf("Lesson.Name cannot be empty")
	}

	if l.TeacherID.Status != pgtype.Present || len(l.TeacherID.String) == 0 {
		return fmt.Errorf("Lesson.TeacherID cannot be empty")
	}

	if l.CourseID.Status != pgtype.Present || len(l.CourseID.String) == 0 {
		return fmt.Errorf("Lesson.CourseID cannot be empty")
	}

	if l.LessonType.Status != pgtype.Present || len(l.LessonType.String) == 0 {
		return fmt.Errorf("Lesson.LessonType cannot be empty")
	}

	if l.Status.Status != pgtype.Present || len(l.Status.String) == 0 {
		return fmt.Errorf("Lesson.Status cannot be empty")
	}

	if l.TeacherIDs.TeacherIDs.Status != pgtype.Present || len(l.TeacherIDs.TeacherIDs.Elements) == 0 {
		return fmt.Errorf("Lesson.TeacherIDs cannot be empty")
	}

	if l.LearnerIDs.LearnerIDs.Status != pgtype.Present || len(l.LearnerIDs.LearnerIDs.Elements) == 0 {
		return fmt.Errorf("Lesson.LearnerIDs cannot be empty")
	}

	if l.CourseIDs.CourseIDs.Status != pgtype.Present || len(l.CourseIDs.CourseIDs.Elements) == 0 {
		return fmt.Errorf("Lesson.CourseIDs cannot be empty")
	}

	if l.StartTime.Time.IsZero() || l.EndTime.Time.IsZero() {
		return fmt.Errorf("Lesson.StartTime and Lesson.EndTime cannot be empty")
	}

	if l.StartTime.Time.After(l.EndTime.Time) {
		return fmt.Errorf("Lesson.StartTime cannot be after Lesson.EndTime")
	}

	return nil
}

func (l *Lesson) PreInsert() error {
	if l.LessonID.Status != pgtype.Present {
		if err := l.LessonID.Set(idutil.ULIDNow()); err != nil {
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

func (l *Lesson) PreUpdate() error {
	return l.UpdatedAt.Set(time.Now())
}

func (l *Lesson) SetDefaultSchedulingStatus() error {
	return l.SchedulingStatus.Set(LessonSchedulingStatusPublished)
}

func (l *Lesson) Deletable() bool {
	return l.StartTime.Time.After(time.Now())
}

type Lessons []*Lesson

func (u *Lessons) Add() database.Entity {
	e := &Lesson{}
	*u = append(*u, e)

	return e
}
