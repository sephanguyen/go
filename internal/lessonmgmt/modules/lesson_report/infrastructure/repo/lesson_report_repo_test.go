package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonReportRepoWithSqlMock() (*LessonReportRepo, *testutil.MockDB) {
	r := &LessonReportRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonReportRepo_DeleteReportsBelongToLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lessonIDs := []string{"lesson-id-1"}
	r, mockDB := LessonReportRepoWithSqlMock()
	t.Run("delete successfully", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.AnythingOfType("string"), &lessonIDs)

		err := r.DeleteReportsBelongToLesson(ctx, mockDB.DB, lessonIDs)
		require.NoError(t, err)
	})

	t.Run("delete failed", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &lessonIDs)

		err := r.DeleteReportsBelongToLesson(ctx, mockDB.DB, lessonIDs)
		require.Error(t, err)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
}

func TestLessonReportRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonReportRepoWithSqlMock()
	id := "id"
	e := &LessonReportDTO{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &id)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		formConfig, err := l.FindByID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, formConfig)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &id)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := l.FindByID(ctx, mockDB.DB, id)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestLessonReportRepo_FindByLessonID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonReportRepoWithSqlMock()
	lessonID := "id"
	e := &LessonReportDTO{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		formConfig, err := l.FindByLessonID(ctx, mockDB.DB, lessonID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, formConfig)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := l.FindByLessonID(ctx, mockDB.DB, lessonID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestLessonReportRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonReportRepoWithSqlMock()
	e := &LessonReportDTO{}
	now := time.Now()
	lessonReport := &domain.LessonReport{
		LessonReportID:         "lesson-report-1",
		ReportSubmittingStatus: lesson_report_consts.ReportSubmittingStatusSaved,
		FormConfigID:           "form-config-1",
		LessonID:               "lesson-1",
		CreatedAt:              now,
		UpdatedAt:              now,
		FormConfig: &domain.FormConfig{
			FormConfigID: "form-config-1",
		},
	}
	t.Run("err insert", func(t *testing.T) {
		lessonReportDTO, err := NewLessonReportDTOFromDomain(lessonReport)
		_, values := lessonReportDTO.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		_, err = l.Create(ctx, mockDB.DB, lessonReport)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		lessonReportDTO, err := NewLessonReportDTOFromDomain(lessonReport)
		_, values := lessonReportDTO.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		_, err = l.Create(ctx, mockDB.DB, lessonReport)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
	})
}

func TestLessonReportRepo_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonReportRepoWithSqlMock()
	now := time.Now()
	lessonReport := &domain.LessonReport{
		LessonReportID:         "lesson-report-1",
		ReportSubmittingStatus: lesson_report_consts.ReportSubmittingStatusSaved,
		FormConfigID:           "form-config-1",
		LessonID:               "lesson-1",
		UpdatedAt:              now,
		FormConfig: &domain.FormConfig{
			FormConfigID: "form-config-1",
		},
	}
	t.Run("err update", func(t *testing.T) {

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		_, err := l.Update(ctx, mockDB.DB, lessonReport)
		assert.NotEmpty(t, err)
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		_, err := l.Update(ctx, mockDB.DB, lessonReport)
		assert.Nil(t, err)
	})
}

func TestPartnerFormConfigRepo_FindByResourcePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e := &LessonReportDTO{}
	fields, value := e.FieldMap()
	p, mockDB := LessonReportRepoWithSqlMock()
	resourcePath := "id1"
	limit := 1
	offset := 1
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &resourcePath, &limit, &offset)
		mockDB.MockScanFields(nil, fields, value)
		_, err := p.FindByResourcePath(ctx, mockDB.DB, resourcePath, limit, offset)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &resourcePath, &limit, &offset)
		_, err := p.FindByResourcePath(ctx, mockDB.DB, resourcePath, limit, offset)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
