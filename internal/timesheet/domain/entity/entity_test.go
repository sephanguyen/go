package entity

import (
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntity(t *testing.T) {
	t.Parallel()
	sv, err := database.NewSchemaVerifier("timesheet")
	require.NoError(t, err)

	entities := []database.Entity{
		&User{},
		&Timesheet{},
		&Staff{},
		&Course{},
		&CourseAccessPath{},
		&Lesson{},
		&LessonTeacher{},
		&OtherWorkingHours{},
		&TimesheetConfig{},
		&TimesheetLessonHours{},
		&AutoCreateFlagActivityLog{},
		&TransportationExpense{},
		&AutoCreateTimesheetFlag{},
		&StaffTransportationExpense{},
		&Location{},
		&TimesheetConfirmationCutOffDate{},
		&TimesheetConfirmationPeriod{},
		&TimesheetConfirmationInfo{},
		&PartnerAutoCreateTimesheetFlag{},
		&TimesheetActionLog{},
	}

	assertions := assert.New(t)
	dir, err := os.Getwd()
	assertions.NoError(err)

	count, err := database.CheckEntity(dir)
	assertions.NoError(err)
	assertions.Equalf(count, len(entities), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(entities))

	for _, e := range entities {
		assertions.NoError(database.CheckEntityDefinition(e))
		assertions.NoError(sv.Verify(e))
	}
}

func TestEntities(t *testing.T) {
	t.Parallel()
	ents := []database.Entities{
		&Timesheets{},
		&Users{},
		&Staffs{},
		&ListOtherWorkingHours{},
		&ListTimesheetLessonHours{},
		&Lessons{},
		&ListTransportationExpenses{},
		&AutoCreateTimesheetFlags{},
		&AutoCreateFlagActivityLogs{},
		&ListStaffTransportationExpense{},
		&Locations{},
		&TimesheetConfirmationPeriods{},
	}

	assertions := assert.New(t)
	dir, err := os.Getwd()
	assertions.NoError(err)

	count, err := database.CheckEntities(dir)
	assertions.NoError(err)
	assertions.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assertions.NoError(database.CheckEntitiesDefinition(e))
	}
}
