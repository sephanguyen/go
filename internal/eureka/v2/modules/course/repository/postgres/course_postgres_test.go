package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCourseRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := &mock_database.Ext{}
	courseRepo := &CourseRepo{
		DB: mockDB,
	}

	t.Run("successfully", func(t *testing.T) {
		rows := &mock_database.Rows{}

		mockDB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(rows, nil)

		fields := database.GetFieldNames(&dto.CourseDto{})
		fieldDescriptions := make([]pgproto3.FieldDescription, 0, len(fields))
		for _, f := range fields {
			fieldDescriptions = append(fieldDescriptions, pgproto3.FieldDescription{Name: []byte(f)})
		}
		rows.On("FieldDescriptions").Return(fieldDescriptions)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", database.GetScanFields(&dto.CourseDto{}, fields)...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		courses, err := courseRepo.RetrieveByIDs(ctx, []string{"course-id-1", "course-id-2"})

		require.Nil(t, err)
		require.NotNil(t, courses)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
	t.Run("error no rows", func(t *testing.T) {
		mockDB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(nil, pgx.ErrNoRows)

		courses, err := courseRepo.RetrieveByIDs(ctx, []string{"course-id-3", "course-id-4"})
		require.Error(t, err)
		require.Nil(t, courses)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
}
