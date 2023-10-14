package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ClassroomRepoWithSqlMock() (*ClassroomRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	mockRepo := &ClassroomRepo{}
	return mockRepo, mockDB
}

func TestClassroomRepo_CheckClassroomIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	classroomRepo, mockDB := ClassroomRepoWithSqlMock()
	classroomIDs := []string{"classroom-id1"}
	classroom := &Classroom{}
	fields, values := classroom.FieldMap()

	t.Run("successful check classroom ids", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, classroomIDs)
		mockDB.MockScanFields(nil, fields, values)
		err := classroomRepo.CheckClassroomIDs(ctx, mockDB.DB, classroomIDs)
		assert.NoError(t, err)
	})

	t.Run("failed get classroom ids", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, classroomIDs)
		mockDB.MockScanFields(pgx.ErrNoRows, fields, values)
		err := classroomRepo.CheckClassroomIDs(ctx, mockDB.DB, classroomIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})
}

func TestClassroomRepo_ExportAllClassrooms(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classroomRepo, mockDB := ClassroomRepoWithSqlMock()

	classroom := &ClassroomToExport{}
	fields, values := classroom.FieldMap()
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn: "location_id",
		},
		{
			DBColumn: "location_name",
		},
		{
			DBColumn: "classroom_id",
		},
		{
			DBColumn: "classroom_name",
		},
		{
			DBColumn: "remarks",
		},
		{
			DBColumn: "is_archived",
		},
	}
	t.Run("Get all classrooms successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)
		dateInfos, err := classroomRepo.ExportAllClassrooms(ctx, mockDB.DB, exportCols)
		assert.NoError(t, err)
		assert.NotNil(t, dateInfos)
	})

	t.Run("Fetch classrooms failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		dateInfos, err := classroomRepo.ExportAllClassrooms(ctx, mockDB.DB, exportCols)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, dateInfos)
	})
}
