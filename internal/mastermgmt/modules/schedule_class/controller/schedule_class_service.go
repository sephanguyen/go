package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	classInfra "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScheduleClassService struct {
	db                         database.Ext
	fatimaService              fpb.SubscriptionModifierServiceClient
	ReserveClassRepo           infrastructure.ReserveClassRepo
	ReserveClassCommandHandler application.ReserveClassCommandHandlerPort
	ReserveClassQueryHandler   application.ReserveClassQueryHandlerPort
}

func NewScheduleClassService(db database.Ext, reserveClassRepo infrastructure.ReserveClassRepo, courseRepo infrastructure.CourseRepo, classRepo infrastructure.ClassRepo, studentPackageClassRepo infrastructure.StudentPackageClassRepo, internalFatimaServiceClient fpb.SubscriptionModifierServiceClient, classMemberRepo classInfra.ClassMemberRepo, jsm nats.JetStreamManagement) *ScheduleClassService {
	return &ScheduleClassService{
		db:               db,
		fatimaService:    internalFatimaServiceClient,
		ReserveClassRepo: reserveClassRepo,
		ReserveClassCommandHandler: &commands.ReserveClassCommandHandler{
			DB:               db,
			ReserveClassRepo: reserveClassRepo,
			JSM:              jsm,
			ClassMemberRepo:  classMemberRepo,
		},
		ReserveClassQueryHandler: &queries.ReserveClassQueryHandler{
			DB:                      db,
			ReserveClassRepo:        reserveClassRepo,
			CourseRepo:              courseRepo,
			ClassRepo:               classRepo,
			StudentPackageClassRepo: studentPackageClassRepo,
		},
	}
}

func (s *ScheduleClassService) redirectToRegisterStudentClass(ctx context.Context, req *mpb.ScheduleStudentClassRequest) error {
	classesInformation := []*fpb.RegisterStudentClassRequest_ClassInformation{
		{
			StudentId:        req.StudentId,
			StudentPackageId: req.StudentPackageId,
			ClassId:          req.ClassId,
			CourseId:         req.CourseId,
			StartTime:        req.EffectiveDate,
			EndTime:          req.EndTime,
		},
	}

	registerStudentClassRequest := &fpb.RegisterStudentClassRequest{
		ClassesInformation: classesInformation,
	}
	_, err := s.fatimaService.RegisterStudentClass(utils.SignCtx(ctx), registerStudentClassRequest)

	return err
}

func (s *ScheduleClassService) ScheduleStudentClass(ctx context.Context, req *mpb.ScheduleStudentClassRequest) (*mpb.ScheduleStudentClassResponse, error) {
	if isReserveClass, currentClassID, err := s.ReserveClassCommandHandler.CheckWillReserveClass(ctx, req); err != nil {
		return &mpb.ScheduleStudentClassResponse{Successful: false}, status.Error(codes.Internal, fmt.Errorf("ScheduleStudentClass fail: %w", err).Error())
	} else if isReserveClass {
		err = s.ReserveClassCommandHandler.ReserveStudentClass(ctx, req, currentClassID)
		if err != nil {
			return &mpb.ScheduleStudentClassResponse{Successful: false}, status.Error(codes.Internal, fmt.Errorf("ScheduleStudentClass.ReserveStudentClass fail: %w", err).Error())
		}
		return &mpb.ScheduleStudentClassResponse{Successful: true}, nil
	}

	err := s.redirectToRegisterStudentClass(ctx, req)
	if err != nil {
		return &mpb.ScheduleStudentClassResponse{Successful: false}, status.Error(codes.Internal, fmt.Errorf("ScheduleStudentClass.RegisterStudentClass fail: %w", err).Error())
	}

	cancelReserveClassPayload := &commands.CancelReserveClassCommandPayload{
		StudentPackageID: req.StudentPackageId,
		StudentID:        req.StudentId,
		CourseID:         req.CourseId,
		ActiveClassID:    req.ClassId,
	}

	err = s.ReserveClassCommandHandler.CancelReserveClass(ctx, *cancelReserveClassPayload)
	if err != nil {
		return &mpb.ScheduleStudentClassResponse{Successful: false}, status.Error(codes.Internal, fmt.Errorf("ScheduleStudentClass.RegisterStudentClass cancel old reserve class fail: %w", err).Error())
	}

	return &mpb.ScheduleStudentClassResponse{Successful: true}, nil
}

func (s *ScheduleClassService) CancelScheduledStudentClass(ctx context.Context, req *mpb.CancelScheduledStudentClassRequest) (*mpb.CancelScheduledStudentClassResponse, error) {
	cancelReserveClassPayload := &commands.CancelReserveClassCommandPayload{
		StudentPackageID: req.StudentPackageId,
		StudentID:        req.StudentId,
		CourseID:         req.CourseId,
		ActiveClassID:    req.ClassId,
	}

	err := s.ReserveClassCommandHandler.CancelReserveClass(ctx, *cancelReserveClassPayload)
	if err != nil {
		return &mpb.CancelScheduledStudentClassResponse{Successful: false}, status.Error(codes.Internal, fmt.Errorf("CancelScheduledStudentClass fail: %w", err).Error())
	}

	return &mpb.CancelScheduledStudentClassResponse{Successful: true}, nil
}

func (s *ScheduleClassService) RetrieveScheduledStudentClass(ctx context.Context, req *mpb.RetrieveScheduledStudentClassRequest) (*mpb.RetrieveScheduledStudentClassResponse, error) {
	result, err := s.ReserveClassQueryHandler.RetrieveScheduledClass(ctx, req.StudentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("RetrieveScheduledStudentClass fail: %w", err).Error())
	}

	return result, nil
}

func (s *ScheduleClassService) BulkAssignStudentsToClass(_ context.Context, _ *mpb.BulkAssignStudentsToClassRequest) (*mpb.BulkAssignStudentsToClassResponse, error) {
	return &mpb.BulkAssignStudentsToClassResponse{
		Successful: true,
	}, nil
}
