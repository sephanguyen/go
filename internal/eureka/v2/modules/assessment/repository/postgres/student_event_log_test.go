package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentEventLogRepo_GetManyByEventTypesAndLMs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userID := idutil.ULIDNow()
	courseID := idutil.ULIDNow()
	eventTypes := []string{"type_1", "type_2"}
	lmIDs := []string{idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow()}

	t.Run("query failed returns DB error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &StudentEventLogRepo{}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"),
			userID,
			courseID,
			eventTypes[0],
			eventTypes[1],
			lmIDs[0],
			lmIDs[1],
			lmIDs[2],
			lmIDs[3],
			lmIDs[4],
		)
		expectedErr := errors.NewDBError("StudentEventLogRepo.GetManyByEventTypesAndLMs", puddle.ErrClosedPool)

		// act
		actual, err := repo.GetManyByEventTypesAndLMs(ctx, mockDB.DB, courseID, userID, eventTypes, lmIDs)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed returns Conversion error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &StudentEventLogRepo{}
		holder := dto.StudentEventLog{}
		_, val := holder.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"),
				userID,
				courseID,
				eventTypes[0],
				eventTypes[1],
				lmIDs[0],
				lmIDs[1],
				lmIDs[2],
				lmIDs[3],
				lmIDs[4]).
			Once().
			Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("StudentEventLogRepo.scanStudentEventLogs", scanErr)

		// act
		actual, err := repo.GetManyByEventTypesAndLMs(ctx, mockDB.DB, courseID, userID, eventTypes, lmIDs)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded returns nil err", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &StudentEventLogRepo{}
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			userID,
			courseID,
			eventTypes[0],
			eventTypes[1],
			lmIDs[0],
			lmIDs[1],
			lmIDs[2],
			lmIDs[3],
			lmIDs[4])

		e := &dto.StudentEventLog{
			Payload: pgtype.JSONB{
				Status: pgtype.Present,
				Bytes:  json.RawMessage(`{}`),
			},
		}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyByEventTypesAndLMs(ctx, mockDB.DB, courseID, userID, eventTypes, lmIDs)

		// assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}
