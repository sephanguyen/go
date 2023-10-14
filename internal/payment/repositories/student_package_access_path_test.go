package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StudentPackageAccessPathRepoWithSqlMock() (*StudentPackageAccessPathRepo, *testutil.MockDB) {
	studentPackageAccessPathRepo := &StudentPackageAccessPathRepo{}
	return studentPackageAccessPathRepo, testutil.NewMockDB()
}

func TestStudentPackageAccessPathRepo_Insert(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.StudentPackageAccessPath{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := studentPackageAccessPathRepoWithSqlMock.Insert(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update student package fail", func(t *testing.T) {
		studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := studentPackageAccessPathRepoWithSqlMock.Insert(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error when insert student package access path: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageAccessPathRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.StudentPackageAccessPath{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := studentPackageAccessPathRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update student package fail", func(t *testing.T) {
		studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := studentPackageAccessPathRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error when update student package access path: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentPackageAccessPathRepo_GetMapStudentCourseWithStudentPackageAccessPathByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		studentPackageAccessPathRepoWithSqlMock *StudentPackageAccessPathRepo
		mockDB                                  *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, constant.ErrDefault,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
			},
		},
		{
			Name:        constant.FailCaseErrorQuery,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			studentPackageAccessPathRepoWithSqlMock, mockDB = StudentPackageAccessPathRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			_, err := studentPackageAccessPathRepoWithSqlMock.GetMapStudentCourseKeyWithStudentPackageAccessPathByStudentIDs(testCase.Ctx, mockDB.DB, []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestStudentPackageAccessPathRepo_DeleteMulti(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	studentPackageAccessPathRepo := &StudentPackageAccessPathRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []entities.StudentPackageAccessPath{
				{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "happy case: upsert multiple parents",
			Req: []entities.StudentPackageAccessPath{
				{},
				{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []entities.StudentPackageAccessPath{
				{},
			},
			ExpectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := studentPackageAccessPathRepo.DeleteMulti(ctx, db, testCase.Req.([]entities.StudentPackageAccessPath))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}

func TestStudentPackageAccessPathRepo_InsertMulti(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	studentPackageAccessPathRepo := &StudentPackageAccessPathRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []entities.StudentPackageAccessPath{
				{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "happy case: upsert multiple parents",
			Req: []entities.StudentPackageAccessPath{
				{},
				{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []entities.StudentPackageAccessPath{
				{},
			},
			ExpectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := studentPackageAccessPathRepo.InsertMulti(ctx, db, testCase.Req.([]entities.StudentPackageAccessPath))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}

func TestStudentPackageAccessPathRepo_SoftDeleteByStudentPackageIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()
	db := mockDB.DB

	now := time.Now()
	studentPackageIDs := []string{"student_package_id_1"}
	mockEntity := &entities.StudentPackageAccessPath{}
	stmt := fmt.Sprintf(`UPDATE %s SET deleted_at = $1, updated_at = now() 
                         WHERE student_package_id = ANY($2) 
                           AND deleted_at IS NULL`, mockEntity.TableName())
	args := []interface{}{
		mock.Anything,
		stmt,
		now,
		database.TextArray(studentPackageIDs),
	}

	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  nil,
			Req: []interface{}{
				now,
				studentPackageIDs,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name: "Failed case: Error when exec",
			Req: []interface{}{
				now,
				studentPackageIDs,
			},
			ExpectedErr: fmt.Errorf("err db.Exec StudentPackageAccessPathRepo.SoftDeleteByStudentPackageIDs: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			deletedAtReq := (testCase.Req.([]interface{})[0]).(time.Time)
			studentPackageIDsReq := (testCase.Req.([]interface{})[1]).([]string)
			err := studentPackageAccessPathRepoWithSqlMock.SoftDeleteByStudentPackageIDs(ctx, db, studentPackageIDsReq, deletedAtReq)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentPackageAccessPathRepo_CheckExistStudentPackageAccessPath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t,
			mock.Anything,
			mock.Anything,
			constant.StudentID,
			constant.CourseID,
		)
		entity := &entities.StudentPackageAccessPath{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		err := studentPackageAccessPathRepoWithSqlMock.CheckExistStudentPackageAccessPath(ctx, mockDB.DB, constant.StudentID, constant.CourseID)
		assert.Nil(t, err)
	})
	t.Run("error when student package is exist", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t,
			mock.Anything,
			mock.Anything,
			constant.StudentID,
			constant.CourseID,
		)
		entity := &entities.StudentPackageAccessPath{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		err := studentPackageAccessPathRepoWithSqlMock.CheckExistStudentPackageAccessPath(ctx, mockDB.DB, constant.StudentID, constant.CourseID)
		require.NotNil(t, err)
		require.Equal(t, status.Errorf(codes.FailedPrecondition, "duplicate student course id"), err)
	})
	t.Run("error when db have problem", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t,
			mock.Anything,
			mock.Anything,
			constant.StudentID,
			constant.CourseID,
		)
		entity := &entities.StudentPackageAccessPath{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(constant.ErrDefault, fields, values)
		err := studentPackageAccessPathRepoWithSqlMock.CheckExistStudentPackageAccessPath(ctx, mockDB.DB, constant.StudentID, constant.CourseID)
		require.NotNil(t, err)
		require.Equal(t, status.Errorf(codes.Internal, "get student package access path have error %v", constant.ErrDefault.Error()), err)
	})
}

func TestStudentPackageAccessPathRepo_RevertByStudentIDAndCourseID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()
	db := mockDB.DB

	mockEntity := &entities.StudentPackageAccessPath{}
	stmt := fmt.Sprintf(`UPDATE %s SET deleted_at = NULL, updated_at = now() 
                         WHERE student_id = $1 AND course_id = $2`, mockEntity.TableName())
	args := []interface{}{
		mock.Anything,
		stmt,
		constant.StudentID,
		constant.CourseID,
	}

	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  nil,
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name: "Failed case: Error when exec",
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
			},
			ExpectedErr: fmt.Errorf("err db.Exec StudentPackageAccessPathRepo.RevertByStudentIDAndCourseID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)

			studentID := (testCase.Req.([]interface{})[0]).(string)
			courseID := (testCase.Req.([]interface{})[1]).(string)
			err := studentPackageAccessPathRepoWithSqlMock.RevertByStudentIDAndCourseID(ctx, db, studentID, courseID)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentPackageAccessPathRepo_GetByStudentIDAndCourseID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPackageAccessPathRepoWithSqlMock, mockDB := StudentPackageAccessPathRepoWithSqlMock()
	db := mockDB.DB

	mockEntity := &entities.StudentPackageAccessPath{}
	studentFieldNames, studentFieldValues := mockEntity.FieldMap()
	stmt := fmt.Sprintf(
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1 AND course_id = $2 AND deleted_at is NULL
		FOR NO KEY UPDATE`,
		strings.Join(studentFieldNames, ","),
		mockEntity.TableName(),
	)

	args := []interface{}{
		mock.Anything,
		stmt,
		constant.StudentID,
		constant.CourseID,
	}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when scan",
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
			},
			ExpectedErr: fmt.Errorf("row.Scan StudentPackageAccessPathRepo.GetByStudentIDAndCourseID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", args...).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", studentFieldValues...).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", args...).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", studentFieldValues...).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)

			studentID := (testCase.Req.([]interface{})[0]).(string)
			courseID := (testCase.Req.([]interface{})[1]).(string)
			_, err := studentPackageAccessPathRepoWithSqlMock.GetByStudentIDAndCourseID(ctx, db, studentID, courseID)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}
