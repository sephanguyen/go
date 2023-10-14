package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentRepoWithSqlMock() (*StudentRepo, *testutil.MockDB) {
	repo := &StudentRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentRepo_GetStudentsByParentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	parentID := pgtype.Text{}
	parentID.Set(uuid.NewString())
	studentIDs := pgtype.TextArray{}
	_ = studentIDs.Set([]string{uuid.NewString()})

	_, userValues := (&entity.LegacyUser{}).FieldMap()
	argsUser := append([]interface{}{}, genSliceMock(len(userValues))...)
	_, studentValues := (&entity.LegacyStudent{}).FieldMap()
	argsStudent := append([]interface{}{}, genSliceMock(len(studentValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &parentID).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(studentIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsUser, argsStudent...)...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.GetStudentsByParentID(ctx, mockDB.DB, parentID)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &parentID).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		students, err := repo.GetStudentsByParentID(ctx, mockDB.DB, parentID)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &parentID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", append(argsUser, argsStudent...)...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.GetStudentsByParentID(ctx, mockDB.DB, parentID)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Err return error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &parentID).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(studentIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsUser, argsStudent...)...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(errors.New("mock-error"))
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.GetStudentsByParentID(ctx, mockDB.DB, parentID)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentRepo_Create(t *testing.T) {
	t.Parallel()
	userGroup := entity.UserGroup{}
	_, userGroupValues := userGroup.FieldMap()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		student := &entity.LegacyStudent{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()
		_, studentValues := user.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, student)
		assert.Nil(t, err)
	})
	t.Run("insert user group fail", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		student := &entity.LegacyStudent{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()
		_, studentValues := user.FieldMap()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(nil, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, student)
		assert.Equal(t, fmt.Errorf("err insert UserGroup: %w", pgx.ErrTxClosed).Error(), err.Error())
	})
	t.Run("no rows affect after insert user group", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		student := &entity.LegacyStudent{
			LegacyUser: user,
		}
		_, userValues := user.FieldMap()
		_, studentValues := user.FieldMap()

		cmdTag0 := pgconn.CommandTag([]byte(`0`))
		cmdTag1 := pgconn.CommandTag([]byte(`1`))

		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag1, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag1, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Return(cmdTag0, nil)

		err := repo.Create(ctx, mockDB.DB, student)
		assert.Equal(t, fmt.Errorf("insert users_groups: %d RowsAffected", cmdTag0.RowsAffected()).Error(), err.Error())
	})
	t.Run("resource path status is nil", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		student := &entity.LegacyStudent{
			LegacyUser: user,
		}

		_, userValues := user.FieldMap()
		_, studentValues := user.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, student)
		// assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, err)
	})
	t.Run("enrollmentStatus is null", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		student := &entity.LegacyStudent{
			LegacyUser:       user,
			EnrollmentStatus: pgtype.Text{Status: pgtype.Null},
		}

		_, userValues := user.FieldMap()
		_, studentValues := user.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, student)
		// assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, err)
	})
	t.Run("studentNote is null", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		student := &entity.LegacyStudent{
			LegacyUser:  user,
			StudentNote: pgtype.Text{Status: pgtype.Null},
		}

		_, userValues := user.FieldMap()
		_, studentValues := user.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, student)
		// assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, err)
	})
	t.Run("schoolID diff 0", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		user := entity.LegacyUser{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		student := &entity.LegacyStudent{
			LegacyUser: user,
			School: &entity.School{
				ID: pgtype.Int4{Int: 1111},
			},
		}

		_, userValues := user.FieldMap()
		_, studentValues := user.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
		argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)
		argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, student)
		// assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, err)
	})
}

func TestStudentRepo_FindStudentProfilesByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentIDs := pgtype.TextArray{}
	_ = studentIDs.Set([]string{uuid.NewString()})

	_, userValues := (&entity.LegacyUser{}).FieldMap()
	argsUser := append([]interface{}{}, genSliceMock(len(userValues))...)
	_, studentValues := (&entity.LegacyStudent{}).FieldMap()
	argsStudent := append([]interface{}{}, genSliceMock(len(studentValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(studentIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsUser, argsStudent...)...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.FindStudentProfilesByIDs(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		students, err := repo.FindStudentProfilesByIDs(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", append(argsUser, argsStudent...)...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.FindStudentProfilesByIDs(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Err return error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(studentIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsUser, argsStudent...)...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(errors.New("mock-error"))
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.FindStudentProfilesByIDs(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentID := database.Text("id")
	_, studentValues := (&entity.LegacyStudent{}).FieldMap()
	argsStudent := append([]interface{}{}, genSliceMock(len(studentValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(nil)
		students, err := repo.Find(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(puddle.ErrClosedPool)
		student, err := repo.Find(ctx, mockDB.DB, studentID)
		assert.Nil(t, student)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentRepo_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	student := &entity.LegacyStudent{}
	_, userValues := (&entity.LegacyUser{}).FieldMap()
	_, studentValues := (&entity.LegacyStudent{}).FieldMap()

	argsUser := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)
	argsStudent := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, student)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update user fail: connection closed", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Exec", argsUser...).Once().Return(nil, puddle.ErrClosedPool)

		err := repo.Update(ctx, mockDB.DB, student)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update user fail: rows affect not equal", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, student)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update student fail", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		mockDB.DB.On("Exec", argsStudent...).Once().Return(nil, puddle.ErrClosedPool)

		err := repo.Update(ctx, mockDB.DB, student)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update student fail: rows affect not equal", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsUser...).Once().Return(cmdTag, nil)
		cmdTag = pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", argsStudent...).Once().Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, student)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studentRepo := &StudentRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.LegacyStudent{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
					LegacyUser: entity.LegacyUser{
						ResourcePath: pgtype.Text{Status: pgtype.Null},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: create multiple students",
			req: []*entity.LegacyStudent{
				{
					ID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.LegacyStudent{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "send batch return ",
			req: []*entity.LegacyStudent{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("student is not inserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentRepo.CreateMultiple(ctx, db, testCase.req.([]*entity.LegacyStudent))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestStudentRepo_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentIDs := pgtype.TextArray{}
	_ = studentIDs.Set([]string{uuid.NewString()})

	_, studentValues := (&entity.LegacyStudent{}).FieldMap()
	argsStudent := append([]interface{}{}, genSliceMock(len(studentValues))...)

	_, userValues := (&entity.LegacyUser{}).FieldMap()
	argsUser := append([]interface{}{}, genSliceMock(len(userValues))...)

	_, schoolValues := (&entity.School{}).FieldMap()
	argsSchool := append([]interface{}{}, genSliceMock(len(schoolValues))...)

	_, cityValues := (&entity.City{}).FieldMap()
	argsCity := append([]interface{}{}, genSliceMock(len(cityValues))...)

	_, districtValues := (&entity.District{}).FieldMap()
	argsDistrict := append([]interface{}{}, genSliceMock(len(districtValues))...)

	_, gradeValues := (&GradeEntity{}).FieldMap()
	argsGrade := append([]interface{}{}, genSliceMock(len(gradeValues))...)

	scanFields := append(argsStudent, argsUser...)
	scanFields = append(scanFields, argsSchool...)
	scanFields = append(scanFields, argsCity...)
	scanFields = append(scanFields, argsDistrict...)
	scanFields = append(scanFields, argsGrade...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(studentIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)

			mockDB.Rows.On("Scan", scanFields...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.Retrieve(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		students, err := repo.Retrieve(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)

		mockDB.Rows.On("Scan", scanFields...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.Retrieve(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Err return error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(studentIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)

			mockDB.Rows.On("Scan", scanFields...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(errors.New("mock-error"))
		mockDB.Rows.On("Close").Once().Return(nil)

		students, err := repo.Retrieve(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	repo, mockDB := StudentRepoWithSqlMock()
	studentIDs := database.TextArray([]string{"student-1", "student-2"})

	testCases := []TestCase{
		{
			name:        "error cannot delete student",
			expectedErr: fmt.Errorf("cannot delete student"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(nil, fmt.Errorf("cannot delete student"))
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`2`))
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(cmdTag, nil)
				mockDB.DB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.SoftDelete(ctx, mockDB.DB, studentIDs)
		assert.Equal(t, testCase.expectedErr, err)
	}
}
