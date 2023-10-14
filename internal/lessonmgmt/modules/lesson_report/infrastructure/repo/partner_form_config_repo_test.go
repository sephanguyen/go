package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func RepoWithSqlMock() (*PartnerFormConfigRepo, *testutil.MockDB) {
	r := &PartnerFormConfigRepo{}
	return r, testutil.NewMockDB()
}

func TestPartnerFormConfigRepo_FindByPartnerAndFeatureName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p, mockDB := RepoWithSqlMock()
	partnerID := 1
	featureName := "group"
	e := &PartnerFormConfigDTO{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &partnerID, &featureName)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		formConfig, err := p.FindByPartnerAndFeatureName(ctx, mockDB.DB, partnerID, featureName)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, formConfig)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &partnerID, &featureName)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := p.FindByPartnerAndFeatureName(ctx, mockDB.DB, partnerID, featureName)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}
func TestPartnerFormConfigRepo_DeleteByLessonReportDetailIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p, mockDB := RepoWithSqlMock()
	ids := []string{
		"id1", "id2",
	}
	t.Run("success", func(t *testing.T) {
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, database.TextArray(ids)).Once().Return(nil, nil)
		err := p.DeleteByLessonReportDetailIDs(ctx, mockDB.DB, ids)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, database.TextArray(ids)).Once().Return(nil, puddle.ErrClosedPool)
		err := p.DeleteByLessonReportDetailIDs(ctx, mockDB.DB, ids)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestPartnerFormConfigRepo_GetMapStudentFieldValuesByDetailID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	e := &PartnerDynamicFormFieldValueWithStudentIdDTO{}
	fields, value := e.FieldMap()

	p, mockDB := RepoWithSqlMock()
	id := "id1"
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &id)
		mockDB.MockScanFields(nil, fields, value)
		_, err := p.GetMapStudentFieldValuesByDetailID(ctx, mockDB.DB, id)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &id)
		_, err := p.GetMapStudentFieldValuesByDetailID(ctx, mockDB.DB, id)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestPartnerFormConfigRepo_CreatePartnerFormConfig(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p, mockDB := RepoWithSqlMock()
	partnerFormConfig := &domain.PartnerFormConfig{
		FormConfigID: "1",
		PartnerID:    1,
		FeatureName:  "1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	dto, _ := NewPartnerFormConfigFromEntity(partnerFormConfig)
	_, values := dto.FieldMap()

	t.Run("success", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := p.CreatePartnerFormConfig(ctx, mockDB.DB, partnerFormConfig)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, cmdTag, puddle.ErrNotAvailable, args...)
		err := p.CreatePartnerFormConfig(ctx, mockDB.DB, partnerFormConfig)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
