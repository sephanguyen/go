package handler

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type IChangeClassHandler interface {
	GetQueryLesson() (*domain.QueryLesson, error)
	GetLessonMember(lesson *domain.Lesson) *domain.LessonMember
	UpdateLessonMember(db database.Ext, lessonMembers []*domain.LessonMember) error
}
