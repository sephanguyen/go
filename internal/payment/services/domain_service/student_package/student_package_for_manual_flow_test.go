package service

import (
	"context"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudentPackage_UpsertStudentPackageForManualFlow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Error when check access path", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentPackageAccessPathRepo.On("CheckExistStudentPackageAccessPath",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
		}
		events, err := studentPackageService.UpsertStudentPackageForManualFlow(
			ctx,
			db,
			"1",
			&pb.StudentCourseData{
				StudentPackageId: wrapperspb.String(constant.StudentPackageID),
			},
		)
		require.NotNil(t, err)
		require.Nil(t, events)
		require.Equal(t, constant.ErrDefault, err)
	})
	t.Run("Error when insert student package", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageAccessPathRepo.On("CheckExistStudentPackageAccessPath",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			StudentPackageRepo:           studentPackageRepo,
		}
		events, err := studentPackageService.UpsertStudentPackageForManualFlow(
			ctx,
			db,
			"1",
			&pb.StudentCourseData{
				StudentPackageId: wrapperspb.String(constant.StudentPackageID),
			},
		)
		require.NotNil(t, err)
		require.Nil(t, events)
		require.Equal(t, status.Errorf(codes.Internal, "upsert student package have error %v", constant.ErrDefault.Error()), err)
	})
	t.Run("Error when insert student package access path", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageLogRepo := new(mockRepositories.MockStudentPackageLogRepo)
		studentPackageAccessPathRepo.On("CheckExistStudentPackageAccessPath",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageAccessPathRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageService := StudentPackageService{
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageLogRepo:        studentPackageLogRepo,
		}
		events, err := studentPackageService.UpsertStudentPackageForManualFlow(
			ctx,
			db,
			"1",
			&pb.StudentCourseData{
				StudentPackageId: wrapperspb.String(constant.StudentPackageID),
			},
		)
		require.NotNil(t, err)
		require.Nil(t, events)
		require.Equal(t, status.Errorf(codes.Internal, "upsert student package access path have error %v", constant.ErrDefault.Error()), err)
	})
	t.Run("Happy case", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageLogRepo := new(mockRepositories.MockStudentPackageLogRepo)
		studentPackageAccessPathRepo.On("CheckExistStudentPackageAccessPath",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageAccessPathRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageService := StudentPackageService{
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageLogRepo:        studentPackageLogRepo,
		}
		events, err := studentPackageService.UpsertStudentPackageForManualFlow(
			ctx,
			db,
			"1",
			&pb.StudentCourseData{
				StudentPackageId: wrapperspb.String(constant.StudentPackageID),
			},
		)
		require.Nil(t, err)
		require.NotNil(t, events)
	})
}

func TestStudentPackage_UpdateTimeStudentPackageForManualFlow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Error when update student package", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo: studentPackageRepo,
		}
		events, err := studentPackageService.UpdateTimeStudentPackageForManualFlow(
			ctx,
			db,
			"1",
			&pb.StudentCourseData{
				StudentPackageId: wrapperspb.String(constant.StudentPackageID),
			},
		)
		require.NotNil(t, err)
		require.Nil(t, events)
		require.Equal(t, status.Errorf(codes.Internal, "upsert student package have error %v", constant.ErrDefault.Error()), err)
	})
	t.Run("Happy case", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageLogRepo := new(mockRepositories.MockStudentPackageLogRepo)
		studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:    studentPackageRepo,
			StudentPackageLogRepo: studentPackageLogRepo,
		}
		events, err := studentPackageService.UpdateTimeStudentPackageForManualFlow(
			ctx,
			db,
			"1",
			&pb.StudentCourseData{
				StudentPackageId: wrapperspb.String(constant.StudentPackageID),
				CourseId:         "1",
				LocationId:       "1",
				StartDate:        timestamppb.Now(),
				EndDate:          timestamppb.Now(),
			},
		)
		require.Nil(t, err)
		require.NotNil(t, events)
	})
}
