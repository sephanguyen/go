package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonReportDetailRepoWithSqlMock() (*LessonReportDetailRepo, *testutil.MockDB) {
	r := &LessonReportDetailRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonReportDetailRepo_GetByLessonReportID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	l, mockDB := LessonReportDetailRepoWithSqlMock()
	e := &LessonReportDetailDTO{}
	fields, value := e.FieldMap()
	lessonReportID := "lesson-report-id"
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonReportID)
		details, err := l.GetByLessonReportID(ctx, mockDB.DB, lessonReportID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, details)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonReportID)
		mockDB.MockScanFields(nil, fields, value)
		_ = e.StudentID.Set("student-1")
		details, err := l.GetByLessonReportID(ctx, mockDB.DB, lessonReportID)
		assert.NoError(t, err)
		assert.NotNil(t, details)
	})
}

func TestLessonReportDetailRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	detailRepo, mockDB := LessonReportDetailRepoWithSqlMock()
	details := domain.LessonReportDetails{
		{
			LessonReportDetailID: "lesson-report-detail-1",
			LessonReportID:       "lesson-report-id-1",
			StudentID:            "user-1",
			CreatedAt:            now,
			UpdatedAt:            now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Twice().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := detailRepo.Upsert(ctx, mockDB.DB, "lessonReportID", details)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)
		err := detailRepo.Upsert(ctx, mockDB.DB, "lessonReportID", details)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLessonReportDetailRepo_UpsertFieldValues(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	detailRepo, mockDB := LessonReportDetailRepoWithSqlMock()
	fields := []*domain.PartnerDynamicFormFieldValue{
		{
			DynamicFormFieldValueID: "id-1",
			FieldID:                 "field-id-1",
			LessonReportDetailID:    "lesson-report-detail-id-1",
			ValueType:               "1",
			CreatedAt:               now,
			UpdatedAt:               now,
		},
		{
			DynamicFormFieldValueID: "id-2",
			FieldID:                 "field-id-2",
			LessonReportDetailID:    "lesson-report-detail-id-2",
			ValueType:               "1",
			CreatedAt:               now,
			UpdatedAt:               now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := detailRepo.UpsertFieldValues(ctx, mockDB.DB, fields)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)
		err := detailRepo.UpsertFieldValues(ctx, mockDB.DB, fields)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLessonReportDetailRepo_UpsertWithVersion(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	detailRepo, mockDB := LessonReportDetailRepoWithSqlMock()
	const (
		lessonReportDetailID1 = "lesson-report-detail-1"
		lessonReportDetailID2 = "lesson-report-detail-2"
		lessonReportID        = "lesson-report-1"
		studentID1            = "student-1"
		studentID2            = "student-2"
		reportVersion         = 1
	)
	details := domain.LessonReportDetails{
		{
			LessonReportDetailID: lessonReportDetailID1,
			LessonReportID:       lessonReportID,
			StudentID:            studentID1,
			CreatedAt:            now,
			UpdatedAt:            now,
			ReportVersion:        reportVersion,
		},
		{
			LessonReportDetailID: lessonReportDetailID2,
			LessonReportID:       lessonReportID,
			StudentID:            studentID2,
			CreatedAt:            now,
			UpdatedAt:            now,
			ReportVersion:        1,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`2`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Times(3).Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := detailRepo.UpsertWithVersion(ctx, mockDB.DB, "lessonReportID", details)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Return(nil)
		err := detailRepo.UpsertWithVersion(ctx, mockDB.DB, "lessonReportID", details)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)
		err := detailRepo.UpsertWithVersion(ctx, mockDB.DB, "lessonReportID", details)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLessonReportDetailRepo_GetReportVersionByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	l, mockDB := LessonReportDetailRepoWithSqlMock()
	e := &LessonReportDetailDTO{}
	fields, value := e.FieldMap()
	lessonID := "lesson-report-id"
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonID)
		details, err := l.GetReportVersionByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, details)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockScanFields(nil, fields, value)
		_ = e.StudentID.Set("student-1")
		details, err := l.GetReportVersionByLessonID(ctx, mockDB.DB, lessonID)
		assert.NoError(t, err)
		assert.NotNil(t, details)
	})
}
