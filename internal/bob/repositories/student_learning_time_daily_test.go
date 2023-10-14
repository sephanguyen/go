package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentLearningTimeDailyRepo_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &StudentLearningTimeDailyRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	rows := mockDB.Rows

	var studentID pgtype.Text
	e := &entities.StudentLearningTimeDaily{}
	from := database.Timestamptz(time.Now())
	to := database.Timestamptz(time.Now().Add(time.Hour))

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &studentID)
		conversationMember, err := r.Retrieve(ctx, db, studentID, nil, nil)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversationMember)
	})
	t.Run("err when have from, to arguments", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &studentID, &from, &to)
		conversationMember, err := r.Retrieve(ctx, db, studentID, &from, &to)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, conversationMember)
	})
	t.Run("success with from, to argument", func(t *testing.T) {
		expectedMap := make(map[pgtype.Text][]*entities.StudentLearningTimeDaily)
		expectedMap[pgtype.Text{}] = []*entities.StudentLearningTimeDaily{{}}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID, &from, &to)

		_, values := e.FieldMap()
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", values...).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)
		rows.On("Close").Once().Return(nil)

		studentLearningTimeDaily, err := r.Retrieve(ctx, db, studentID, &from, &to)

		assert.NoError(t, err)
		assert.Equal(t, len(expectedMap), len(studentLearningTimeDaily), "the length student learning time daily not equal")
		mockDB.RawStmt.AssertSelectedFields(t, "learning_time_id", "student_id", "learning_time", "day", "sessions", "created_at", "updated_at", "assignment_learning_time", "assignment_submission_ids")
		mockDB.RawStmt.AssertSelectedTable(t, "student_learning_time_by_daily", "")
	})
	t.Run("success", func(t *testing.T) {
		expectedMap := make(map[pgtype.Text][]*entities.StudentLearningTimeDaily)
		expectedMap[pgtype.Text{}] = []*entities.StudentLearningTimeDaily{{}}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)

		_, values := e.FieldMap()
		mockDB.DB.On("Query").Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", values...).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)
		rows.On("Close").Once().Return(nil)

		studentLearningTimeDaily, err := r.Retrieve(ctx, db, studentID, nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, len(expectedMap), len(studentLearningTimeDaily), "the length student learning time daily not equal")
		mockDB.RawStmt.AssertSelectedFields(t, "learning_time_id", "student_id", "learning_time", "day", "sessions", "created_at", "updated_at", "assignment_learning_time", "assignment_submission_ids")
		mockDB.RawStmt.AssertSelectedTable(t, "student_learning_time_by_daily", "")
	})
}

func TestStudentLearningTimeDailyRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *entities.StudentLearningTimeDaily
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &entities.StudentLearningTimeDaily{},
			SetUp: func(ctx context.Context) {
				e := &entities.StudentLearningTimeDaily{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name: "err when upsert",
			Ent:  &entities.StudentLearningTimeDaily{},
			Err:  errors.Wrap(puddle.ErrClosedPool, "tx.ExecEx"),
			SetUp: func(ctx context.Context) {
				e := &entities.StudentLearningTimeDaily{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), puddle.ErrClosedPool)
			},
		},
	}

	repo := &StudentLearningTimeDailyRepo{}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.Upsert(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err.Error(), err.Error())
			}
		})
	}
}

func TestStudentLearningTimeDailyRepo_UpsertTaskAssignment(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	testCases := []struct {
		Name  string
		Ent   *entities.StudentLearningTimeDaily
		Err   error
		SetUp func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Ent:  &entities.StudentLearningTimeDaily{},
			SetUp: func(ctx context.Context) {
				e := &entities.StudentLearningTimeDaily{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
			},
		},
		{
			Name: "err when upsert",
			Ent:  &entities.StudentLearningTimeDaily{},
			Err:  fmt.Errorf("StudentLearningTimeDailyRepo.UpsertTaskAssignment %w", puddle.ErrClosedPool),
			SetUp: func(ctx context.Context) {
				e := &entities.StudentLearningTimeDaily{}
				fields := database.GetFieldNames(e)
				values := database.GetScanFields(e, fields)
				mockValues := make([]interface{}, 0, len(values)+2)
				mockValues = append(mockValues, mock.Anything)
				mockValues = append(mockValues, mock.AnythingOfType("string"))
				for range values {
					mockValues = append(mockValues, mock.Anything)
				}
				db.On("Exec", mockValues...).Once().Return(pgconn.CommandTag([]byte(`1`)), puddle.ErrClosedPool)
			},
		},
	}

	repo := &StudentLearningTimeDailyRepo{}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.SetUp(ctx)
			err := repo.UpsertTaskAssignment(ctx, db, testCase.Ent)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err.Error(), err.Error())
			}
		})
	}
}
