package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func AssignedStudentRepoWithSqlMock() (*AssignedStudentRepo, *testutil.MockDB) {
	r := &AssignedStudentRepo{}
	return r, testutil.NewMockDB()
}

func TestAssignedStudentRepo_GetAssignedStudentList(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := AssignedStudentRepoWithSqlMock()
	args := &payloads.GetAssignedStudentListArg{
		CourseIDs:      []string{"course-1"},
		StudentIDs:     []string{"student-1"},
		KeyWord:        "student name",
		LocationIDs:    []string{"center-1"},
		PurchaseMethod: string(domain.PurchaseMethodSlot),
		Timezone:       "timezone",
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		asgStudents, _, _, _, err := r.GetAssignedStudentList(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, asgStudents)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &AssignedStudentDTO{}
		selectFields := []string{"student_id", "course_id", "location_id", "start_date", "end_date", "duration", "purchased_slot", "assigned_slot", "slot_gap", "status", "student_subscription_id"}

		value := append(database.GetScanFields(e, selectFields))
		selectFields = append(selectFields)
		_ = e.StudentID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		asgStudents, _, _, _, err := r.GetAssignedStudentList(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.EqualValues(t, []*domain.AssignedStudent{
			{StudentID: "id"},
		}, asgStudents)

		assert.NoError(t, err)
		assert.NotNil(t, asgStudents)
	})
}
