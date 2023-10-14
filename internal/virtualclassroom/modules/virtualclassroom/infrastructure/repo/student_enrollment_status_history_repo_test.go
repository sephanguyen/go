package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentEnrollmentStatusHistoryRepoWithSqlMock() (*StudentEnrollmentStatusHistoryRepo, *testutil.MockDB) {
	r := &StudentEnrollmentStatusHistoryRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentEnrollmentStatusHistoryRepo_GetStatusHistoryByStudentIDsAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentEnrollmentRepo, mockDB := StudentEnrollmentStatusHistoryRepoWithSqlMock()
	studentEnrollmentHistory := &StudentEnrollmentStatusHistory{}
	fields, values := studentEnrollmentHistory.FieldMap()

	studentIDs := []string{"user_id1", "user_id2"}
	locationID := "location_id1"

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, studentIDs, locationID)
		mockDB.MockScanFields(nil, fields, values)

		enrollmentInfo, err := studentEnrollmentRepo.GetStatusHistoryByStudentIDsAndLocationID(ctx, mockDB.DB, studentIDs, locationID)
		assert.NoError(t, err)
		assert.NotNil(t, enrollmentInfo)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, studentIDs, locationID)

		enrollmentInfo, err := studentEnrollmentRepo.GetStatusHistoryByStudentIDsAndLocationID(ctx, mockDB.DB, studentIDs, locationID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, enrollmentInfo)
	})
}
