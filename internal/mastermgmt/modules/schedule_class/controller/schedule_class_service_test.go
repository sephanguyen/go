package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mock_fatima "github.com/manabie-com/backend/mock/fatima/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_schedule_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/schedule_class/infrastructure/repo"
	mock_reserve_class_command "github.com/manabie-com/backend/mock/mastermgmt/modules/schedule_class/application/commands"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupMock() (*ScheduleClassService, *mock_database.Ext, *mock_fatima.SubscriptionModifierServiceClient, *mock_reserve_class_command.MockReserveClassCommandHandler) {
	mockDB := &mock_database.Ext{}
	mockSubscriptionModifierServiceClient := new(mock_fatima.SubscriptionModifierServiceClient)
	mockReserveClassCommand := new(mock_reserve_class_command.MockReserveClassCommandHandler)
	mockReserveClassRepo := new(mock_schedule_class_repo.MockReserveClassRepo)
	s := &ScheduleClassService{
		db:                         mockDB,
		fatimaService:              mockSubscriptionModifierServiceClient,
		ReserveClassRepo:           mockReserveClassRepo,
		ReserveClassCommandHandler: mockReserveClassCommand,
	}

	return s, mockDB, mockSubscriptionModifierServiceClient, mockReserveClassCommand
}

func TestScheduleClassService_ScheduleStudentClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	currentClassID := idutil.ULIDNow()

	s, _, mockSubscriptionModifierServiceClient, mockReserveClassCommand := setupMock()

	testCases := []struct {
		name        string
		req         *mpb.ScheduleStudentClassRequest
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "CheckWillReserveClass fail",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(30 * 24 * time.Hour)),
			},
			expectedErr: status.Error(codes.Internal, "ScheduleStudentClass fail: check will reserve class fail"),
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CheckWillReserveClass", ctx, mock.Anything).Once().Return(false, "", fmt.Errorf("check will reserve class fail"))
			},
		},
		{
			name: "register class success",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.Now(),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CheckWillReserveClass", ctx, mock.Anything).Once().Return(false, "", nil)
				mockSubscriptionModifierServiceClient.On("RegisterStudentClass", utils.SignCtx(ctx), mock.Anything).Once().Return(&fpb.RegisterStudentClassResponse{
					Successful: true,
				}, nil)
				mockReserveClassCommand.On("CancelReserveClass", ctx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "register class success but cancel old reserve class fail",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.Now(),
			},
			expectedErr: status.Error(codes.Internal, "ScheduleStudentClass.RegisterStudentClass cancel old reserve class fail: error"),
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CheckWillReserveClass", ctx, mock.Anything).Once().Return(false, "", nil)
				mockSubscriptionModifierServiceClient.On("RegisterStudentClass", utils.SignCtx(ctx), mock.Anything).Once().Return(&fpb.RegisterStudentClassResponse{
					Successful: true,
				}, nil)
				mockReserveClassCommand.On("CancelReserveClass", ctx, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name: "register class fail",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.Now(),
			},
			expectedErr: status.Error(codes.Internal, "ScheduleStudentClass.RegisterStudentClass fail: register class fail"),
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CheckWillReserveClass", ctx, mock.Anything).Once().Return(false, "", nil)
				mockSubscriptionModifierServiceClient.On("RegisterStudentClass", utils.SignCtx(ctx), mock.Anything).Once().Return(&fpb.RegisterStudentClassResponse{
					Successful: false,
				}, fmt.Errorf("register class fail"))
			},
		},
		{
			name: "reserve class success",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(30 * 24 * time.Hour)),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CheckWillReserveClass", ctx, mock.Anything).Once().Return(true, currentClassID, nil)
				mockReserveClassCommand.On("ReserveStudentClass", ctx, mock.Anything, currentClassID).Once().Return(nil)
			},
		},
		{
			name: "reserve class fail",
			req: &mpb.ScheduleStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
				StartTime:        timestamppb.New(now.Add(-10 * 24 * time.Hour)),
				EndTime:          timestamppb.New(now.Add(100 * 24 * time.Hour)),
				EffectiveDate:    timestamppb.New(now.Add(30 * 24 * time.Hour)),
			},
			expectedErr: status.Error(codes.Internal, "ScheduleStudentClass.ReserveStudentClass fail: insert error"),
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CheckWillReserveClass", ctx, mock.Anything).Once().Return(true, currentClassID, nil)
				mockReserveClassCommand.On("ReserveStudentClass", ctx, mock.Anything, currentClassID).Once().Return(fmt.Errorf("insert error"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := s.ScheduleStudentClass(ctx, tc.req)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				assert.False(t, resp.Successful)
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Successful)
			}
		})
	}
}

func TestScheduleClassService_CancelScheduledStudentClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s, _, _, mockReserveClassCommand := setupMock()

	testCases := []struct {
		name        string
		req         *mpb.CancelScheduledStudentClassRequest
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "success",
			req: &mpb.CancelScheduledStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CancelReserveClass", ctx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "fail",
			req: &mpb.CancelScheduledStudentClassRequest{
				StudentId:        "student_id_01",
				StudentPackageId: "student_package_id_01",
				CourseId:         "course_id_01",
				ClassId:          "class_id_01",
			},
			expectedErr: status.Error(codes.Internal, "CancelScheduledStudentClass fail: error"),
			setup: func(ctx context.Context) {
				mockReserveClassCommand.On("CancelReserveClass", ctx, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := s.CancelScheduledStudentClass(ctx, tc.req)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				assert.False(t, resp.Successful)
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Successful)
			}
		})
	}
}
