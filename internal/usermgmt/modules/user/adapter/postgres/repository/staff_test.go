package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StaffRepoWithSqlMock() (*StaffRepo, *testutil.MockDB) {
	repo := &StaffRepo{}
	return repo, testutil.NewMockDB()
}

func TestStaffRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	teacherRepo, db := StaffRepoWithSqlMock()
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.Staff{
				{
					ID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: create multiple teachers",
			req: []*entity.Staff{
				{
					ID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.Staff{
				{
					ID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := teacherRepo.CreateMultiple(ctx, db.DB, testCase.req.([]*entity.Staff))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStaffRepo_Retrieve(t *testing.T) {
	t.Parallel()
	teacherRepo, _ := StaffRepoWithSqlMock()
	teacherIDs := database.TextArray([]string{uuid.NewString()})
	_, teacherValues := new(entity.Staff).FieldMap()
	argsTeacher := append([]interface{}{}, genSliceMock(len(teacherValues))...)
	_, userValues := new(entity.LegacyUser).FieldMap()
	argsUser := append([]interface{}{}, genSliceMock(len(userValues))...)

	testCases := []struct {
		name        string
		req         interface{}
		expectedErr error
		setup       func(ctx context.Context) *testutil.MockDB
	}{
		{
			name:        "happy case",
			req:         teacherIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) *testutil.MockDB {
				_, mockDB := StaffRepoWithSqlMock()
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(mockDB.Rows, nil)
				for i := 0; i < len(teacherIDs.Elements); i++ {
					mockDB.Rows.On("Next").Once().Return(true)
					mockDB.Rows.On("Scan", append(argsTeacher, argsUser...)...).Once().Return(nil)
				}
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
				return mockDB
			},
		},
		{
			name:        "db Query returns error",
			req:         teacherIDs,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) *testutil.MockDB {
				_, mockDB := TeacherRepoWithSqlMock()
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(nil, pgx.ErrTxClosed)
				return mockDB
			},
		},
		{
			name:        "rows Scan return err",
			req:         teacherIDs,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) *testutil.MockDB {
				_, mockDB := TeacherRepoWithSqlMock()
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(mockDB.Rows, nil)
				for i := 0; i < len(teacherIDs.Elements); i++ {
					mockDB.Rows.On("Next").Once().Return(true)
					mockDB.Rows.On("Scan", append(argsTeacher, argsUser...)...).Once().Return(pgx.ErrTxClosed)
				}
				mockDB.Rows.On("Close").Once().Return(nil)
				return mockDB
			},
		},
		{
			name:        "rows Err return error",
			req:         teacherIDs,
			expectedErr: fmt.Errorf("error"),
			setup: func(ctx context.Context) *testutil.MockDB {
				_, mockDB := TeacherRepoWithSqlMock()
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(mockDB.Rows, nil)
				for i := 0; i < len(teacherIDs.Elements); i++ {
					mockDB.Rows.On("Next").Once().Return(true)
					mockDB.Rows.On("Scan", append(argsTeacher, argsUser...)...).Once().Return(nil)
				}
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(fmt.Errorf("error"))
				mockDB.Rows.On("Close").Once().Return(nil)
				return mockDB
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		mockDB := testCase.setup(ctx)
		staff, err := teacherRepo.Retrieve(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
		if err == nil {
			assert.NotNil(t, staff)
		} else {
			assert.Nil(t, staff)
		}
	}
}

func TestStaffRepo_UpdateStaffOnly(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	staff := &entity.Staff{}
	_, staffValues := (&entity.Staff{}).FieldMap()
	argsStaff := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(staffValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsStaff...).Once().Return(cmdTag, nil)

		err := repo.UpdateStaffOnly(ctx, mockDB.DB, staff)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update user fail: connection closed", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		mockDB.DB.On("Exec", argsStaff...).Once().Return(nil, puddle.ErrClosedPool)

		err := repo.UpdateStaffOnly(ctx, mockDB.DB, staff)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestStaffRepo_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	staff := &entity.Staff{}
	_, userValues := (&entity.LegacyUser{}).FieldMap()
	_, staffValues := (&entity.Staff{}).FieldMap()

	argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
	argsStaff := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(staffValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStaff...).Once().Return(cmdTag, nil)

		_, err := repo.Update(ctx, mockDB.DB, staff)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update user fail: connection closed", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		mockDB.DB.On("Exec", argsUser...).Once().Return(nil, puddle.ErrClosedPool)

		_, err := repo.Update(ctx, mockDB.DB, staff)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update user fail: rows affect not equal", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)

		_, err := repo.Update(ctx, mockDB.DB, staff)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update staff fail", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStaff...).Once().Return(nil, puddle.ErrClosedPool)

		_, err := repo.Update(ctx, mockDB.DB, staff)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update staff fail: rows affect not equal", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		cmdTag = pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", argsStaff...).Once().Return(cmdTag, nil)

		_, err := repo.Update(ctx, mockDB.DB, staff)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStaffRepo_Create(t *testing.T) {
	t.Parallel()
	userGroup := entity.UserGroup{}
	_, userGroupValues := userGroup.FieldMap()
	user := entity.LegacyUser{
		ResourcePath: pgtype.Text{
			String: fmt.Sprint(constants.JPREPSchool),
			Status: pgtype.Null,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		staff := &entity.Staff{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()
		_, staffValues := staff.FieldMap()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStaff := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(staffValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStaff...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, staff)
		assert.Nil(t, err)
	})
	t.Run("insert user group fail", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		staff := &entity.Staff{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()
		_, staffValues := staff.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))

		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStaff := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(staffValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStaff...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(nil, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, staff)
		assert.Equal(t, fmt.Errorf("err insert UserGroup: %w", pgx.ErrTxClosed).Error(), err.Error())
	})
	t.Run("insert user fail", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		staff := &entity.Staff{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()

		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(nil, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, staff)
		assert.Equal(t, fmt.Errorf("err insert user: %w", pgx.ErrTxClosed).Error(), err.Error())
	})
	t.Run("insert staff fail", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		staff := &entity.Staff{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()
		_, staffValues := staff.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))

		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsStaff := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(staffValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStaff...).Once().Return(nil, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, staff)
		assert.Equal(t, fmt.Errorf("err insert staff: %w", pgx.ErrTxClosed).Error(), err.Error())
	})
	t.Run("no rows affect after insert user group", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		staff := &entity.Staff{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()
		_, staffValues := staff.FieldMap()

		cmdTag0 := pgconn.CommandTag([]byte(`0`))
		cmdTag1 := pgconn.CommandTag([]byte(`1`))
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStaff := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(staffValues))...)
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag1, nil)
		mockDB.DB.On("Exec", argsStaff...).Once().Return(cmdTag1, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Return(cmdTag0, nil)

		err := repo.Create(ctx, mockDB.DB, staff)
		assert.Equal(t, fmt.Errorf("%d RowsAffected: %w", cmdTag0.RowsAffected(), ErrUnAffected).Error(), err.Error())
	})
	t.Run("resource path status is nil", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		staff := &entity.Staff{
			LegacyUser: user,
		}

		cmdTag := pgconn.CommandTag([]byte(`1`))
		_, values := user.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(values))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)

		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, staff)
		assert.Nil(t, err)
	})
}

func TestStaffRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	repo, mockDB := StaffRepoWithSqlMock()
	staffIDs := database.TextArray([]string{"staff-1", "staff-2"})

	testCases := []TestCase{
		{
			name:        "error cannot delete staff",
			expectedErr: fmt.Errorf("cannot delete staff"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &staffIDs).Once().Return(nil, fmt.Errorf("cannot delete staff"))
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`2`))
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &staffIDs).Once().Return(cmdTag, nil)
				mockDB.DB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.SoftDelete(ctx, mockDB.DB, staffIDs)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestStaffRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	staffID := database.Text("id")
	_, staffValues := (&entity.Staff{}).FieldMap()
	argsStaff := append([]interface{}{}, genSliceMock(len(staffValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &staffID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsStaff...).Once().Return(nil)
		staff, err := repo.Find(ctx, mockDB.DB, staffID)
		assert.Nil(t, err)
		assert.NotNil(t, staff)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := StaffRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &staffID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsStaff...).Once().Return(puddle.ErrClosedPool)
		staff, err := repo.Find(ctx, mockDB.DB, staffID)
		assert.Nil(t, staff)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
