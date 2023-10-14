package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/multierr"
)

type ClassStudentService struct {
	DB database.Ext

	ClassStudentRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.ClassStudent) error
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, classIDs []string) error
	}
}

func (s *ClassStudentService) upsertClassStudent(ctx context.Context, msg *bobproto.EvtClassRoom_JoinClass) error {
	classStudent := &entities.ClassStudent{}
	classID := strconv.Itoa(int(msg.ClassId))
	database.AllNullEntity(classStudent)
	if err := multierr.Combine(
		classStudent.StudentID.Set(msg.GetUserId()),
		classStudent.ClassID.Set(classID),
	); err != nil {
		return err
	}

	if err := s.ClassStudentRepo.Upsert(ctx, s.DB, classStudent); err != nil {
		return err
	}
	return nil
}

func (s *ClassStudentService) softDeleteClassMember(ctx context.Context, msg *bobproto.EvtClassRoom_LeaveClass) error {
	if len(msg.GetUserIds()) == 0 && msg.GetClassId() == 0 {
		return nil
	}

	if err := s.ClassStudentRepo.SoftDelete(ctx, s.DB, msg.GetUserIds(), []string{strconv.Itoa(int(msg.GetClassId()))}); err != nil {
		return err
	}
	return nil
}

func (s *ClassStudentService) HandleClassEvent(ctx context.Context, req *bobproto.EvtClassRoom) error {
	switch req.Message.(type) {
	case *bobproto.EvtClassRoom_JoinClass_:
		msg := req.GetJoinClass()
		err := s.upsertClassStudent(ctx, msg)
		if err != nil {
			return fmt.Errorf("err s.upsertClassStudent: %v", err)
		}
	case *bobproto.EvtClassRoom_LeaveClass_:
		msg := req.GetLeaveClass()
		err := s.softDeleteClassMember(ctx, msg)
		if err != nil {
			return fmt.Errorf("err s.softDeleteClassMember: %v", err)
		}
	}
	return nil
}
