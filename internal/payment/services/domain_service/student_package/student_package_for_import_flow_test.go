package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestStudentPackage_UpsertStudentPackage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Invalid student location", func(t *testing.T) {
		studentPackageService := StudentPackageService{}
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{},
			map[string]interface{}{},
			map[string]entities.StudentPackageAccessPath{},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &entities.StudentPackages{},
					StudentPackageAccessPath: studentPackageAccess,
				},
			})
		require.NotNil(t, errs)
		require.Nil(t, events)
		require.Equal(t, "student id 1 can't access location id 1", errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Invalid course location", func(t *testing.T) {
		studentPackageService := StudentPackageService{}
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]interface{}{},
			map[string]entities.StudentPackageAccessPath{},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &entities.StudentPackages{},
					StudentPackageAccessPath: studentPackageAccess,
				},
			})
		require.NotNil(t, errs)
		require.Nil(t, events)
		require.Equal(t, "course id 1 can't access location id 1", errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Error when insert student package", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo: studentPackageRepo,
		}
		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]entities.StudentPackageAccessPath{},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &entities.StudentPackages{},
					StudentPackageAccessPath: studentPackageAccess,
				},
			})
		require.NotNil(t, errs)
		require.Nil(t, events)
		require.Equal(t, fmt.Sprintf("insert student package by student id %s and course id %s have error %s", "1", "1", constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Error when insert student package access path", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentPackageRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageAccessPathRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
		}
		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]entities.StudentPackageAccessPath{},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &entities.StudentPackages{},
					StudentPackageAccessPath: studentPackageAccess,
				},
			})
		require.NotNil(t, errs)
		require.Nil(t, events)
		require.Equal(t, fmt.Sprintf("insert student package access path by student id %s and course id %s have error %s", "1", "1", constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Happy case for insert", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentCourseRepo := new(mockRepositories.MockStudentCourseRepo)
		studentPackageRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageAccessPathRepo.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentCourseRepo.On("UpsertStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageLogRepo := new(mockRepositories.MockStudentPackageLogRepo)
		studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			StudentCourseRepo:            studentCourseRepo,
			StudentPackageLogRepo:        studentPackageLogRepo,
		}
		studentPackage := entities.StudentPackages{}
		_ = multierr.Combine(
			studentPackage.ID.Set(nil),
			studentPackage.StudentID.Set(nil),
			studentPackage.PackageID.Set(nil),
			studentPackage.StartAt.Set(nil),
			studentPackage.EndAt.Set(nil),
			studentPackage.Properties.Set(nil),
			studentPackage.IsActive.Set(nil),
			studentPackage.LocationIDs.Set(nil),
			studentPackage.CreatedAt.Set(nil),
			studentPackage.UpdatedAt.Set(nil),
			studentPackage.DeletedAt.Set(nil),
		)

		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]entities.StudentPackageAccessPath{},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &studentPackage,
					StudentPackageAccessPath: studentPackageAccess,
					StudentPackageEvent:      &npb.EventStudentPackage{},
				},
			})
		require.Nil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, 0, len(errs))
		require.Equal(t, 1, len(events))
	})
	t.Run("Error when update student package", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo: studentPackageRepo,
		}
		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &entities.StudentPackages{},
					StudentPackageAccessPath: studentPackageAccess,
				},
			})
		require.NotNil(t, errs)
		require.Nil(t, events)
		require.Equal(t, fmt.Sprintf("update student package by student id %s and course id %s have error %s", "1", "1", constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Error when update student package access path", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageAccessPathRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
		}
		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &entities.StudentPackages{},
					StudentPackageAccessPath: studentPackageAccess,
					StudentPackageEvent: &npb.EventStudentPackage{
						StudentPackage: &npb.EventStudentPackage_StudentPackage{
							Package: &npb.EventStudentPackage_Package{
								StudentPackageId: "1",
							},
						},
					},
				},
			})
		require.NotNil(t, errs)
		require.Nil(t, events)
		require.Equal(t, fmt.Sprintf("update student package access path by student id %s and course id %s have error %s", "1", "1", constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Happy case for update", func(t *testing.T) {
		db := new(mockDb.Ext)
		studentPackage := entities.StudentPackages{}
		_ = multierr.Combine(
			studentPackage.ID.Set(nil),
			studentPackage.StudentID.Set(nil),
			studentPackage.PackageID.Set(nil),
			studentPackage.StartAt.Set(nil),
			studentPackage.EndAt.Set(nil),
			studentPackage.Properties.Set(nil),
			studentPackage.IsActive.Set(nil),
			studentPackage.LocationIDs.Set(nil),
			studentPackage.CreatedAt.Set(nil),
			studentPackage.UpdatedAt.Set(nil),
			studentPackage.DeletedAt.Set(nil),
		)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageAccessPathRepo := new(mockRepositories.MockStudentPackageAccessPathRepo)
		studentPackageRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageAccessPathRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageLogRepo := new(mockRepositories.MockStudentPackageLogRepo)
		studentPackageLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			StudentPackageLogRepo:        studentPackageLogRepo,
		}
		events, errs := studentPackageService.UpsertStudentPackage(
			ctx,
			db,
			"1",
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]interface{}{
				"1_1": mapVal,
			},
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			[]utils.ImportedStudentCourseRow{
				{
					Row:                      2,
					StudentPackage:           &studentPackage,
					StudentPackageAccessPath: studentPackageAccess,
					StudentPackageEvent: &npb.EventStudentPackage{
						StudentPackage: &npb.EventStudentPackage_StudentPackage{
							Package: &npb.EventStudentPackage_Package{
								StudentPackageId: "1",
							},
						},
					},
				},
			})
		require.Nil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, 0, len(errs))
		require.Equal(t, 1, len(events))
	})
}

func TestStudentPackage_UpsertStudentClass(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Invalid course location", func(t *testing.T) {
		studentPackageService := StudentPackageService{}
		db := new(mockDb.Ext)
		events, errs := studentPackageService.UpsertStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{},
			map[string]entities.Class{},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, "this course 1 didn't register for student 1 so we can't register for this class", errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Invalid class", func(t *testing.T) {
		studentPackageService := StudentPackageService{}
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		events, errs := studentPackageService.UpsertStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, "this class 1 didn't exist in database", errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Location of class isn't same location with course", func(t *testing.T) {
		studentPackageService := StudentPackageService{}
		db := new(mockDb.Ext)
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		events, errs := studentPackageService.UpsertStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{
				"1": {},
			},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, "this class 1 difference location with course so we can't register for this class", errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Can't get student package", func(t *testing.T) {
		db := new(mockDb.Ext)
		class := entities.Class{}
		_ = class.ClassID.Set("1")
		_ = class.LocationID.Set("1")
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo: studentPackageRepo,
		}
		events, errs := studentPackageService.UpsertStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{
				"1": class,
			},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, fmt.Sprintf("can't get student package by id %v with err: %v", 1, constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Can't upsert student package class", func(t *testing.T) {
		db := new(mockDb.Ext)
		class := entities.Class{}
		_ = class.ClassID.Set("1")
		_ = class.LocationID.Set("1")
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageClassRepo := new(mockRepositories.MockStudentPackageClassRepo)
		studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
		studentPackageClassRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:      studentPackageRepo,
			StudentPackageClassRepo: studentPackageClassRepo,
		}
		events, errs := studentPackageService.UpsertStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{
				"1": class,
			},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, fmt.Sprintf("can't insert student package class with err: %v", constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Happy case", func(t *testing.T) {
		db := new(mockDb.Ext)
		class := entities.Class{}
		_ = class.ClassID.Set("1")
		_ = class.LocationID.Set("1")
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageClassRepo := new(mockRepositories.MockStudentPackageClassRepo)
		studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
		studentPackageClassRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:      studentPackageRepo,
			StudentPackageClassRepo: studentPackageClassRepo,
		}
		events, errs := studentPackageService.UpsertStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{
				"1": class,
			},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, 0, len(errs))
		require.Equal(t, 1, len(events))
	})
}

func TestStudentPackage_DeleteStudentClass(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Invalid course location", func(t *testing.T) {
		studentPackageService := StudentPackageService{}
		db := new(mockDb.Ext)
		events, errs := studentPackageService.DeleteStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{},
			map[string]entities.Class{},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, "this course 1 didn't register for student 1 so we can't register for this class", errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Can't get student package", func(t *testing.T) {
		db := new(mockDb.Ext)
		class := entities.Class{}
		_ = class.ClassID.Set("1")
		_ = class.LocationID.Set("1")
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo: studentPackageRepo,
		}
		events, errs := studentPackageService.DeleteStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{
				"1": class,
			},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, fmt.Sprintf("can't get student package by id %v with err: %v", 1, constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Can't delete student package class", func(t *testing.T) {
		db := new(mockDb.Ext)
		class := entities.Class{}
		_ = class.ClassID.Set("1")
		_ = class.LocationID.Set("1")
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageClassRepo := new(mockRepositories.MockStudentPackageClassRepo)
		studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
		studentPackageClassRepo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:      studentPackageRepo,
			StudentPackageClassRepo: studentPackageClassRepo,
		}
		events, errs := studentPackageService.DeleteStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{
				"1": class,
			},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, fmt.Sprintf("can't insert student package class with err: %v", constant.ErrDefault.Error()), errs[0].Error)
		require.Equal(t, int32(2), errs[0].RowNumber)
	})
	t.Run("Happy case", func(t *testing.T) {
		db := new(mockDb.Ext)
		class := entities.Class{}
		_ = class.ClassID.Set("1")
		_ = class.LocationID.Set("1")
		studentPackageAccess := &entities.StudentPackageAccessPath{}
		_ = studentPackageAccess.StudentPackageID.Set("1")
		_ = studentPackageAccess.StudentID.Set("1")
		_ = studentPackageAccess.LocationID.Set("1")
		_ = studentPackageAccess.CourseID.Set("1")
		studentPackageRepo := new(mockRepositories.MockStudentPackageRepo)
		studentPackageClassRepo := new(mockRepositories.MockStudentPackageClassRepo)
		studentPackageRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentPackages{}, nil)
		studentPackageClassRepo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentPackageService := StudentPackageService{
			StudentPackageRepo:      studentPackageRepo,
			StudentPackageClassRepo: studentPackageClassRepo,
		}
		events, errs := studentPackageService.DeleteStudentClass(
			ctx,
			db,
			map[string]entities.StudentPackageAccessPath{
				"1_1": *studentPackageAccess,
			},
			map[string]entities.Class{
				"1": class,
			},
			[]utils.ImportedStudentClassRow{
				{
					Row:       2,
					StudentID: "1",
					ClassID:   "1",
					CourseID:  "1",
				},
			})
		require.NotNil(t, errs)
		require.NotNil(t, events)
		require.Equal(t, 0, len(errs))
		require.Equal(t, 1, len(events))
	})
}

func TestStudentPackage_GetMapStudentCourseWithStudentPackageIDByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Happy case", func(t *testing.T) {
		studentPackageAccessPathRepo := &mockRepositories.MockStudentPackageAccessPathRepo{}
		studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", ctx, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
		studentPackageService := StudentPackageService{StudentPackageAccessPathRepo: studentPackageAccessPathRepo}
		db := new(mockDb.Ext)
		_, err := studentPackageService.GetMapStudentCourseWithStudentPackageIDByIDs(ctx, db, []string{"1"})
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(t, db, studentPackageAccessPathRepo)
	})

	t.Run("Error when get from repo", func(t *testing.T) {
		studentPackageAccessPathRepo := &mockRepositories.MockStudentPackageAccessPathRepo{}
		studentPackageAccessPathRepo.On("GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs", ctx, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, constant.ErrDefault)
		studentPackageService := StudentPackageService{StudentPackageAccessPathRepo: studentPackageAccessPathRepo}
		db := new(mockDb.Ext)
		_, err := studentPackageService.GetMapStudentCourseWithStudentPackageIDByIDs(ctx, db, []string{"1"})
		require.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, db, studentPackageAccessPathRepo)
	})
}
