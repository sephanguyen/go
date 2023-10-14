package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func CourseStudentAccessPathRepoWithSqlMock() (*CourseStudentAccessPathRepo, *testutil.MockDB) {
	r := &CourseStudentAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestBulkUpsertCourseStudentAccessPath(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseStudentAccessPathRepo := &CourseStudentAccessPathRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.CourseStudentsAccessPath{
				{
					CourseStudentID: database.Text("course-student-id"),
					LocationID:      database.Text("location-id"),
					CourseID:        database.Text("course-id"),
					StudentID:       database.Text("student-id"),
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
			req: []*entities.CourseStudentsAccessPath{
				{
					CourseStudentID: database.Text("course-student-id"),
					LocationID:      database.Text("location-id"),
					CourseID:        database.Text("course-id"),
					StudentID:       database.Text("student-id"),
				},
				{
					CourseStudentID: database.Text("course-student-id-1"),
					LocationID:      database.Text("location-id-1"),
					CourseID:        database.Text("course-id-1"),
					StudentID:       database.Text("student-id-1"),
				},
				{
					CourseStudentID: database.Text("course-student-id=2"),
					LocationID:      database.Text("location-id-2"),
					CourseID:        database.Text("course-id-2"),
					StudentID:       database.Text("student-id-2"),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := courseStudentAccessPathRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.CourseStudentsAccessPath))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestCourseStudentAccessPath_GetByLocationsAndStudents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseStudentAccessPathRepoWithSqlMock()
	locationIDs := database.TextArray([]string{"id1", "id2"})
	studentIDs := database.TextArray([]string{"id1", "id2"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&locationIDs,
			&studentIDs,
		)

		courseStudentAccessPath, err := r.GetByLocationsAndStudents(ctx, mockDB.DB, locationIDs, studentIDs)

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseStudentAccessPath)
	})
}

func TestCourseStudentAccessPath_GetByLocationsStudentsAndCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := CourseStudentAccessPathRepoWithSqlMock()
	locationIDs := database.TextArray([]string{"id1", "id2"})
	studentIDs := database.TextArray([]string{"id1", "id2"})
	courseIDs := database.TextArray([]string{"id1", "id2"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			puddle.ErrClosedPool,
			mock.Anything,
			mock.AnythingOfType("string"),
			&locationIDs,
			&studentIDs,
			&courseIDs,
		)

		courseStudentAccessPath, err := r.GetByLocationsStudentsAndCourse(ctx, mockDB.DB, locationIDs, studentIDs, courseIDs)

		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, courseStudentAccessPath)
	})
}
