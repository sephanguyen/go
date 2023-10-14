package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentCourseRepoWithSqlMock() (*StudentCourseRepo, *testutil.MockDB) {
	repo := &StudentCourseRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentCourseRepo_GetStudentCoursesByStudentPackageIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCourseRepoWithSqlMock, mockDB := StudentCourseRepoWithSqlMock()
	db := mockDB.DB
	rows := mockDB.Rows

	studentPackageID := "student_package_id_1"
	mockEntity := &entities.StudentCourse{}
	fields, _ := mockEntity.FieldMap()
	scanFields := database.GetScanFields(mockEntity, fields)
	stmt := fmt.Sprintf(`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_id = $1
		FOR NO KEY UPDATE
		`,
		strings.Join(fields, ","),
		mockEntity.TableName(),
	)
	args := []interface{}{
		mock.Anything,
		stmt,
		studentPackageID,
	}

	expectedStudentCourses := []*entities.StudentCourse{
		{
			StudentPackageID: pgtype.Text{
				String: studentPackageID,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
		},
	}
	testCases := []utils.TestCase{
		{
			Name:         "Happy case",
			Ctx:          nil,
			Req:          mock.Anything,
			ExpectedResp: &entities.BillItem{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedStudentCourses)).Run(func(args mock.Arguments) {
					rows.On("Scan", scanFields...).Once().Return(nil)
				}).Return(true)

				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:         "Failed case: Empty case",
			Req:          studentPackageID,
			ExpectedResp: []*entities.BillItem{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:         "Failed case: Error when query",
			Req:          studentPackageID,
			ExpectedResp: []*entities.BillItem{},
			ExpectedErr:  errors.New("error query"),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, errors.New("error query"))
			},
		},
		{
			Name:         "scan failed case",
			Req:          studentPackageID,
			ExpectedResp: []*entities.BillItem{},
			ExpectedErr:  fmt.Errorf("row.Scan: %w", errors.New("error scan")),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", scanFields...).Once().Return(errors.New("error scan"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			req := (testCase.Req).(string)
			_, err := studentCourseRepoWithSqlMock.GetStudentCoursesByStudentPackageIDForUpdate(ctx, db, req)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentCourseRepo_SoftDeleteByStudentPackageIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCourseRepoWithSqlMock, mockDB := StudentCourseRepoWithSqlMock()
	db := mockDB.DB

	now := time.Now()
	studentPackageIDs := []string{"student_package_id_1"}
	mockEntity := &entities.StudentCourse{}
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
			ExpectedErr: fmt.Errorf("err db.Exec StudentCourseRepo.SoftDeleteByStudentPackageIDs: %w", constant.ErrDefault),
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
			err := studentCourseRepoWithSqlMock.SoftDeleteByStudentPackageIDs(ctx, db, studentPackageIDsReq, deletedAtReq)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentCourseRepo_VoidStudentCoursesByStudentPackageID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCourseRepoWithSqlMock, mockDB := StudentCourseRepoWithSqlMock()
	db := mockDB.DB

	studentPackageID := "student_package_id_1"
	studentEndDate := time.Now().Add(24 * time.Hour)
	mockEntity := &entities.StudentCourse{}
	stmt := fmt.Sprintf(`
	UPDATE %s SET student_end_date = $1, updated_at = now() 
	WHERE student_package_id = $2 AND deleted_at IS NOT NULL`, mockEntity.TableName())
	args := []interface{}{
		mock.Anything,
		stmt,
		studentEndDate,
		studentPackageID,
	}

	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  nil,
			Req: []interface {
			}{
				studentEndDate,
				studentPackageID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name: "Failed case: Error when exec",
			Req: []interface {
			}{
				studentEndDate,
				studentPackageID,
			},
			ExpectedErr: fmt.Errorf("error when void student course: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			studentEndDateReq := (testCase.Req).([]interface{})[0].(time.Time)
			studentPackageIDReq := (testCase.Req).([]interface{})[1].(string)
			err := studentCourseRepoWithSqlMock.VoidStudentCoursesByStudentPackageID(ctx, db, studentEndDateReq, studentPackageIDReq)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentCourseRepo_GetStudentCoursesByStudentPackageIDsForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCourseRepoWithSqlMock, mockDB := StudentCourseRepoWithSqlMock()
	db := mockDB.DB
	rows := mockDB.Rows

	mockEntity := &entities.StudentCourse{}
	fields, _ := mockEntity.FieldMap()
	scanFields := database.GetScanFields(mockEntity, fields)
	stmt := fmt.Sprintf(`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_id = ANY($1::text[]) AND deleted_at is null
		FOR NO KEY UPDATE
		`,
		strings.Join(fields, ","),
		mockEntity.TableName(),
	)
	args := []interface{}{
		ctx,
		stmt,
		[]string{
			constant.StudentPackageID,
		},
	}

	expectedStudentCourses := []entities.StudentCourse{
		{
			StudentPackageID: pgtype.Text{
				String: constant.StudentPackageID,
			},
			StudentID: pgtype.Text{
				String: "student_id_1",
			},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  ctx,
			Req: []string{
				constant.StudentPackageID,
			},
			ExpectedResp: expectedStudentCourses,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedStudentCourses)).Run(func(args mock.Arguments) {
					rows.On("Scan", scanFields...).Once().Return(nil)
				}).Return(true)

				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "Failed case: Empty case",
			Req: []string{
				constant.StudentPackageID,
			},
			ExpectedResp: []*entities.BillItem{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "Failed case: Error when query",
			Req: []string{
				constant.StudentPackageID,
			},
			ExpectedResp: []*entities.BillItem{},
			ExpectedErr:  errors.New("error query"),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, errors.New("error query"))
			},
		},
		{
			Name: "scan failed case",
			Req: []string{
				constant.StudentPackageID,
			},
			ExpectedResp: []*entities.BillItem{},
			ExpectedErr:  status.Errorf(codes.Internal, "Err while scan student course"),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", scanFields...).Once().Return(status.Errorf(codes.Internal, "Err while scan student course"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			req := (testCase.Req).([]string)
			_, err := studentCourseRepoWithSqlMock.GetStudentCoursesByStudentPackageIDsForUpdate(ctx, db, req)
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentCourseRepo_UpdateTimeByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCourseRepoWithSqlMock, mockDB := StudentCourseRepoWithSqlMock()
	db := mockDB.DB

	args := []interface{}{
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	}

	testCases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         nil,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name:        "Failed case: Error when exec",
			ExpectedErr: fmt.Errorf("update time student course have error: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			err := studentCourseRepoWithSqlMock.UpdateTimeByID(ctx, db, "1", "1", time.Now())
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentCourseRepo_CancelByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCourseRepoWithSqlMock, mockDB := StudentCourseRepoWithSqlMock()
	db := mockDB.DB

	args := []interface{}{
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	}

	testCases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         nil,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.SuccessCommandTag, nil)
			},
		},
		{
			Name:        "Failed case: Error when exec",
			ExpectedErr: fmt.Errorf("cancel student course have error: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", args...).Once().Return(constant.FailCommandTag, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			err := studentCourseRepoWithSqlMock.CancelByStudentPackageIDAndCourseID(ctx, db, "1", "1")
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentCourseRepo_GetByStudentIDAndCourseIDAndLocationID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentCourseRepo, mockDB := StudentCourseRepoWithSqlMock()
	db := mockDB.DB

	mockEntity := &entities.StudentCourse{}
	_, fieldValues := mockEntity.FieldMap()

	testCases := []utils.TestCase{
		{
			Name: "Failed case: Error when scan",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
				constant.LocationID,
			},
			ExpectedErr: fmt.Errorf("err db.Exec StudentCourseRepo.GetByStudentIDAndCourseIDAndLocationID: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentID,
				constant.CourseID,
				constant.LocationID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", fieldValues...).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			studentID := testCase.Req.([]interface{})[0].(string)
			courseID := testCase.Req.([]interface{})[1].(string)
			locationID := testCase.Req.([]interface{})[1].(string)
			_, err := studentCourseRepo.GetByStudentIDAndCourseIDAndLocationID(testCase.Ctx, db, studentID, courseID, locationID)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}
