package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pbv1 "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterMgmtClassService struct {
	DB database.Ext

	MasterMgmtClassStudentRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entities.ClassStudent) error
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, classIDs []string) error
	}
	CourseClassRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseClass) error
		DeleteClass(ctx context.Context, db database.QueryExecer, classID string) error
	}
}

func (s *MasterMgmtClassService) upsertClassStudent(ctx context.Context, msg *pbv1.EvtClass_JoinClass) error {
	classStudent := &entities.ClassStudent{}
	database.AllNullEntity(classStudent)
	if err := multierr.Combine(
		classStudent.StudentID.Set(msg.GetUserId()),
		classStudent.ClassID.Set(msg.ClassId),
	); err != nil {
		return err
	}

	if err := s.MasterMgmtClassStudentRepo.Upsert(ctx, s.DB, classStudent); err != nil {
		return err
	}
	return nil
}

func (s *MasterMgmtClassService) upsertCourseClass(ctx context.Context, msg *pbv1.EvtClass_CreateClass) error {
	courseClass := &entities.CourseClass{}
	database.AllNullEntity(courseClass)
	if err := multierr.Combine(
		courseClass.CourseID.Set(msg.GetCourseId()),
		courseClass.ClassID.Set(msg.GetClassId()),
		courseClass.ID.Set(idutil.ULIDNow()),
	); err != nil {
		return err
	}
	err := s.CourseClassRepo.BulkUpsert(ctx, s.DB, []*entities.CourseClass{courseClass})
	if err != nil {
		return fmt.Errorf("err s.CourseClassRepo.BulkUpsert: %w", err)
	}
	return nil
}

func (s *MasterMgmtClassService) softDeleteClassMember(ctx context.Context, msg *pbv1.EvtClass_LeaveClass) error {
	if len(msg.GetUserId()) == 0 && len(msg.GetClassId()) == 0 {
		return nil
	}

	if err := s.MasterMgmtClassStudentRepo.SoftDelete(ctx, s.DB, []string{msg.GetUserId()}, []string{msg.GetClassId()}); err != nil {
		return err
	}
	return nil
}

func (s *MasterMgmtClassService) deleteClass(ctx context.Context, msg *pbv1.EvtClass_DeleteClass) error {
	classID := msg.GetClassId()
	if classID == "" {
		return status.Error(codes.InvalidArgument, fmt.Errorf("cannot empty class_id").Error())
	}

	if err := s.CourseClassRepo.DeleteClass(ctx, s.DB, classID); err != nil {
		return status.Errorf(codes.Internal, fmt.Errorf("s.CourseClassRepo.DeleteClass: %w", err).Error())
	}

	return nil
}

func (s *MasterMgmtClassService) HandleMasterMgmtClassEvent(ctx context.Context, req *pbv1.EvtClass) error {
	switch req.Message.(type) {
	case *pbv1.EvtClass_JoinClass_:
		msg := req.GetJoinClass()
		err := s.upsertClassStudent(ctx, msg)
		if err != nil {
			return fmt.Errorf("err s.upsertClassStudent: %v", err)
		}
	case *pbv1.EvtClass_LeaveClass_:
		msg := req.GetLeaveClass()
		err := s.softDeleteClassMember(ctx, msg)
		if err != nil {
			return fmt.Errorf("err s.softDeleteClassMember: %v", err)
		}
	case *pbv1.EvtClass_CreateClass_:
		msg := req.GetCreateClass()
		err := s.upsertCourseClass(ctx, msg)
		if err != nil {
			return fmt.Errorf("err s.softCreateClassMember: %v", err)
		}
	case *pbv1.EvtClass_DeleteClass_:
		msg := req.GetDeleteClass()

		err := s.deleteClass(ctx, msg)

		if err != nil {
			return fmt.Errorf("err s.deleteClass: %v", err)
		}
	}
	return nil
}
