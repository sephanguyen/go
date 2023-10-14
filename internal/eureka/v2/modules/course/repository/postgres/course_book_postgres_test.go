package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCourseBookRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := &mock_database.Ext{}
	courseBookRepo := &CourseBookRepo{
		DB: mockDB,
	}
	upsertCourseBookRequest := []*dto.CourseBookDto{
		{
			CourseID: pgtype.Text{String: "course-id-1", Status: pgtype.Present},
			BookID:   pgtype.Text{String: "book-id-1", Status: pgtype.Present},
		},
		{
			CourseID: pgtype.Text{String: "course-id-2", Status: pgtype.Present},
			BookID:   pgtype.Text{String: "book-id-2", Status: pgtype.Present},
		},
	}

	t.Run("successfully", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		var fields []interface{}
		for i := 0; i < len(upsertCourseBookRequest); i++ {
			_, field := upsertCourseBookRequest[i].FieldMap()
			fields = append(fields, field...)
		}

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
		mockDB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := courseBookRepo.Upsert(ctx, upsertCourseBookRequest)
		require.Nil(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
	t.Run("error", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		var fields []interface{}
		for i := 0; i < len(upsertCourseBookRequest); i++ {
			_, field := upsertCourseBookRequest[i].FieldMap()
			fields = append(fields, field...)
		}

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
		mockDB.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))

		err := courseBookRepo.Upsert(ctx, upsertCourseBookRequest)
		require.Error(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
}

func TestCourseBookRepo_RetrieveAssociateBook(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := &mock_database.Ext{}
	courseBookRepo := &CourseBookRepo{
		DB: mockDB,
	}

	t.Run("successfully", func(t *testing.T) {
		rows := &mock_database.Rows{}

		mockDB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(rows, nil)

		fields := database.GetFieldNames(&dto.CourseBookDto{})
		fieldDescriptions := make([]pgproto3.FieldDescription, 0, len(fields))
		for _, f := range fields {
			fieldDescriptions = append(fieldDescriptions, pgproto3.FieldDescription{Name: []byte(f)})
		}
		rows.On("FieldDescriptions").Return(fieldDescriptions)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", database.GetScanFields(&dto.CourseBookDto{}, fields)...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		cb, err := courseBookRepo.RetrieveAssociatedBook(ctx, "book-id-1")
		require.Nil(t, err)
		require.NotNil(t, cb)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
	t.Run("error no rows", func(t *testing.T) {
		mockDB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(nil, pgx.ErrNoRows)

		cbs, err := courseBookRepo.RetrieveAssociatedBook(ctx, "book-id-2")
		require.Error(t, err)
		require.Nil(t, cbs)

		mock.AssertExpectationsForObjects(
			t,
			mockDB,
		)
	})
}
