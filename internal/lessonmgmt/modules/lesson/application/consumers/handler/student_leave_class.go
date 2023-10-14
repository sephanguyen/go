package handler

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
)

type StudentLeaveClass struct {
	ClassID          string
	StudentID        string
	Ctx              context.Context
	LessonMemberRepo infrastructure.LessonMemberRepo
}

func (s *StudentLeaveClass) GetQueryLesson() (*domain.QueryLesson, error) {
	startJoin := time.Now()
	return &domain.QueryLesson{ClassID: s.ClassID, StartTime: &startJoin}, nil
}

func (s *StudentLeaveClass) GetLessonMember(lesson *domain.Lesson) *domain.LessonMember {
	return &domain.LessonMember{LessonID: lesson.LessonID, StudentID: s.StudentID, CourseID: lesson.CourseID}
}

func (s *StudentLeaveClass) UpdateLessonMember(db database.Ext, lessonMembers []*domain.LessonMember) error {
	if len(lessonMembers) == 0 {
		return nil
	}
	return s.LessonMemberRepo.DeleteLessonMembers(s.Ctx, db, lessonMembers)
}
