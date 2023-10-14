package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	masterMgmt "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"

	"github.com/jackc/pgx/v4"
)

type StudentJoinClass struct {
	ClassID          string
	StudentID        string
	OldClassID       string
	Ctx              context.Context
	DB               database.Ext
	ClassMemberRepo  masterMgmt.ClassMemberRepo
	LessonMemberRepo infrastructure.LessonMemberRepo
	LessonReportRepo infrastructure.LessonReportRepo
}

func (s *StudentJoinClass) GetQueryLesson() (*domain.QueryLesson, error) {
	mapClassMember, err := s.ClassMemberRepo.GetByClassIDAndUserIDs(s.Ctx, s.DB, s.ClassID, []string{s.StudentID})
	if err != nil {
		return nil, err
	}
	classMember := mapClassMember[s.StudentID]
	if classMember == nil {
		return nil, fmt.Errorf("Not found student %s in class %s", s.StudentID, s.ClassID)
	}

	now := time.Now()
	startJoin := now
	if s.OldClassID == "" || classMember.StartDate.After(now) {
		startJoin = classMember.StartDate
	}

	return &domain.QueryLesson{ClassID: s.ClassID, StartTime: &startJoin, EndTime: &classMember.EndDate}, nil
}

func (s *StudentJoinClass) GetLessonMember(lesson *domain.Lesson) *domain.LessonMember {
	now := time.Now()
	return &domain.LessonMember{LessonID: lesson.LessonID, StudentID: s.StudentID, CourseID: lesson.CourseID, UpdatedAt: now, CreatedAt: now}
}

func getBeginOfDate(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.In(loc).Date()
	beginningOfDate := time.Date(y, m, d, 0, 0, 0, 0, loc)
	return beginningOfDate
}

func (s *StudentJoinClass) UpdateLessonMember(db database.Ext, lessonMembers []*domain.LessonMember) error {
	return database.ExecInTx(s.Ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		// because locVN +7 is always less than locJP  +9, so we can only care date now at locVN +7
		locVN, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
		beginningOfDate := getBeginOfDate(time.Now(), locVN)
		lessonID, err := s.LessonMemberRepo.DeleteLessonMembersByStartDate(s.Ctx, tx, s.StudentID, s.ClassID, beginningOfDate)
		if err != nil {
			return fmt.Errorf("s.LessonMemberRepo.DeleteLessonMembersByStartDate: %w", err)
		}
		if len(lessonID) > 0 {
			if err := s.LessonReportRepo.DeleteLessonReportWithoutStudent(s.Ctx, tx, lessonID); err != nil {
				return fmt.Errorf("s.LessonReportRepo.DeleteLessonReportWithoutStudent: %w", err)
			}
		}
		if len(lessonMembers) > 0 {
			return s.LessonMemberRepo.InsertLessonMembers(s.Ctx, tx, lessonMembers)
		}
		return nil
	})
}
