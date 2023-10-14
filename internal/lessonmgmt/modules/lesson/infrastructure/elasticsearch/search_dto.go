package elasticsearch

import (
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type LessonSearch struct {
	LessonID       string            `json:"lesson_id"`
	LocationID     string            `json:"location_id"`
	TeachingMedium string            `json:"teaching_medium"`
	TeachingMethod string            `json:"teaching_method"`
	ClassID        string            `json:"class_id"`
	CourseID       string            `json:"course_id"`
	DeletedAt      *time.Time        `json:"deleted_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	CreatedAt      time.Time         `json:"created_at"`
	LessonMember   []LessonMemberDoc `json:"lesson_members"`
	LessonTeacher  []string          `json:"lesson_teachers"`
	StartTime      time.Time         `json:"start_time"`
	EndTime        time.Time         `json:"end_time"`
}

func (l *LessonSearch) AddLessonMembers(lm []LessonMemberDoc) {
	l.LessonMember = lm
}

type LessonSearchs []*LessonSearch

type LessonMemberDoc struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CurrentGrade int    `json:"current_grade"`
	CourseID     string `json:"course_id"`
}

func (l LessonSearch) GetFields() []interface{} {
	return []interface{}{
		l.LessonID,
		l.LocationID,
		l.TeachingMedium,
		l.TeachingMethod,
		l.ClassID,
		l.CourseID,
		l.UpdatedAt,
		l.CreatedAt,
		l.LessonMember,
		l.LessonTeacher,
		l.StartTime,
		l.EndTime,
	}
}

func (ls LessonSearchs) ToLessonSearchEntities() ([]*domain.Lesson, error) {
	len := len(ls)
	eLessons := make([]*domain.Lesson, 0, len)
	for i := 0; i < len; i++ {
		s := ls[i]

		eb := domain.NewLesson().
			WithID(s.LessonID).
			WithLocationID(s.LocationID).
			WithTeacherIDs(s.LessonTeacher).
			WithTeachingMethod(domain.LessonTeachingMethod(s.TeachingMethod)).
			WithTeachingMedium(domain.LessonTeachingMedium(s.TeachingMedium)).
			WithClassID(s.ClassID).
			WithCourseID(s.CourseID).
			WithModificationTime(s.CreatedAt, s.UpdatedAt).
			WithTimeRange(s.StartTime, s.EndTime).
			WithDeletedTime(s.DeletedAt)
		e := eb.BuildDraft()

		eLessons = append(eLessons, e)
	}
	return eLessons, nil
}

func NewLessonSearchsFromEntities(sLessons domain.LessonSearchs) *LessonSearchs {
	len := len(sLessons)
	ls := make(LessonSearchs, 0, len)

	for i := 0; i < len; i++ {
		s := sLessons[i]
		lm := []LessonMemberDoc{}
		for _, v := range s.LessonMember {
			m := LessonMemberDoc{ID: v.ID, Name: v.Name, CurrentGrade: v.CurrentGrade, CourseID: v.CourseID}
			lm = append(lm, m)
		}

		es := &LessonSearch{
			LessonID:       s.LessonID,
			LocationID:     s.LocationID,
			TeachingMedium: s.TeachingMedium,
			TeachingMethod: s.TeachingMethod,
			ClassID:        s.ClassID,
			CourseID:       s.CourseID,
			DeletedAt:      s.DeletedAt,
			UpdatedAt:      s.UpdatedAt,
			CreatedAt:      s.CreatedAt,
			LessonMember:   lm,
			LessonTeacher:  s.LessonTeacher,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
		}

		ls = append(ls, es)
	}
	return &ls
}
