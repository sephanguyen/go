package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
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

func TeacherRepoWithSqlMock() (*TeacherRepo, *testutil.MockDB) {
	repo := &TeacherRepo{}
	return repo, testutil.NewMockDB()
}

func TestTeacherRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	teacherRepo := &TeacherRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.Teacher{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			name: "happy case: create multiple teachers",
			req: []*entity.Teacher{
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
			req: []*entity.Teacher{
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
			req: []*entity.Teacher{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("teacher not inserted"),
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
		err := teacherRepo.CreateMultiple(ctx, db, testCase.req.([]*entity.Teacher))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestTeacherRepo_Upsert(t *testing.T) {
	t.Parallel()
	teacherRepo := &TeacherRepo{}
	db := &mock_database.Ext{}
	uid := idutil.ULIDNow()
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entity.Teacher{
				ID: database.Text(uid),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}

				cmdTag := pgconn.CommandTag(successTag)
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "connection closed",
			req: &entity.Teacher{
				ID: database.Text(uid),
			},
			expectedErr: puddle.ErrClosedPool,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, puddle.ErrClosedPool)
			},
		},
		{
			name: "no rows affected",
			req: &entity.Teacher{
				ID: database.Text(uid),
			},
			expectedErr: fmt.Errorf("cannot upsert teacher %s", uid),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag(failedTag)
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := teacherRepo.Upsert(ctx, db, testCase.req.(*entity.Teacher))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestTeacherRepo_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	teacherID := pgtype.Text{}
	teacherID.Set(uuid.NewString())
	teacherIDs := pgtype.TextArray{}
	_ = teacherIDs.Set([]string{uuid.NewString()})

	_, teacherValues := (&entity.Teacher{}).FieldMap()
	argsTeacher := append([]interface{}{}, genSliceMock(len(teacherValues))...)
	_, userValues := (&entity.LegacyUser{}).FieldMap()
	argsUser := append([]interface{}{}, genSliceMock(len(userValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(teacherIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsTeacher, argsUser...)...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		teachers, err := repo.Retrieve(ctx, mockDB.DB, teacherIDs)
		assert.Nil(t, err)
		assert.NotNil(t, teachers)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(nil, pgx.ErrTxClosed)

		teachers, err := repo.Retrieve(ctx, mockDB.DB, teacherIDs)
		assert.NotNil(t, err)
		assert.Nil(t, teachers)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Scan return err", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(teacherIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsTeacher, argsUser...)...).Once().Return(pgx.ErrTxClosed)
		}
		mockDB.Rows.On("Close").Once().Return(nil)

		teachers, err := repo.Retrieve(ctx, mockDB.DB, teacherIDs)
		assert.NotNil(t, err)
		assert.Nil(t, teachers)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("rows Err return error", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &teacherIDs).Once().Return(mockDB.Rows, nil)
		for i := 0; i < len(teacherIDs.Elements); i++ {
			mockDB.Rows.On("Next").Once().Return(true)
			mockDB.Rows.On("Scan", append(argsTeacher, argsUser...)...).Once().Return(nil)
		}
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(fmt.Errorf("======="))
		mockDB.Rows.On("Close").Once().Return(nil)

		teachers, err := repo.Retrieve(ctx, mockDB.DB, teacherIDs)
		assert.NotNil(t, err)
		assert.Nil(t, teachers)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTeacherRepo_SoftDelete(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	teacherRepo := &TeacherRepo{}
	returnAmount := 1

	testCases := []TestCase{
		{
			name:        "error cannot delete teachers",
			expectedErr: fmt.Errorf("error"),
			setup: func(ctx context.Context) {
				mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(fmt.Sprint(returnAmount)))
				mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
				mockDB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := teacherRepo.SoftDelete(ctx, mockDB, database.Text(idutil.ULIDNow()))
		assert.Equal(t, err, testCase.expectedErr)
	}
}

func TestTeacherRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	teacherRepo := &TeacherRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.Teacher{
				{ID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(successTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: create multiple teachers",
			req: []*entity.Teacher{
				{ID: database.Text(idutil.ULIDNow())},
				{ID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(successTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.Teacher{
				{ID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %v", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "teacher not upserted",
			req: []*entity.Teacher{
				{
					ID: database.Text(idutil.ULIDNow()),
				},
			},
			expectedErr: fmt.Errorf("teacher not upserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(failedTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := teacherRepo.UpsertMultiple(ctx, db, testCase.req.([]*entity.Teacher))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestTeacherRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	teacherID := database.Text("id")
	_, teacherValues := (&entity.Teacher{}).FieldMap()
	argsTeacher := append([]interface{}{}, genSliceMock(len(teacherValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &teacherID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsTeacher...).Once().Return(nil)
		Teacher, err := repo.Find(ctx, mockDB.DB, teacherID)
		assert.Nil(t, err)
		assert.NotNil(t, Teacher)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &teacherID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsTeacher...).Once().Return(puddle.ErrClosedPool)
		Teacher, err := repo.Find(ctx, mockDB.DB, teacherID)
		assert.Nil(t, Teacher)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTeacherRepo_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	teacher := &entity.Teacher{}
	_, teacherValues := (&entity.Teacher{}).FieldMap()

	argsTeacher := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(teacherValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsTeacher...).Once().Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, teacher)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update teacher fail", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		mockDB.DB.On("Exec", argsTeacher...).Once().Return(nil, puddle.ErrClosedPool)

		err := repo.Update(ctx, mockDB.DB, teacher)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update teacher fail: rows affect not equal", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", argsTeacher...).Once().Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, teacher)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTeacherRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		teacher := &entity.Teacher{}
		_, teacherValues := teacher.FieldMap()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsTeacher := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(teacherValues))...)
		mockDB.DB.On("Exec", argsTeacher...).Once().Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, teacher)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert teacher fail", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		teacher := &entity.Teacher{}
		_, teacherValues := teacher.FieldMap()

		argsTeacher := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(teacherValues))...)
		mockDB.DB.On("Exec", argsTeacher...).Once().Return(nil, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, teacher)
		assert.Equal(t, fmt.Errorf("err insert teacher: %w", pgx.ErrTxClosed).Error(), err.Error())
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("resource path status is nil", func(t *testing.T) {
		repo, mockDB := TeacherRepoWithSqlMock()
		teacher := &entity.Teacher{}
		_, teacherValues := teacher.FieldMap()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		argsTeacher := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(teacherValues))...)

		mockDB.DB.On("Exec", argsTeacher...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, teacher)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
