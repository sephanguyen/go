package repo

import (
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntity(t *testing.T) {
	t.Parallel()
	sv, err := database.NewSchemaVerifier("bob")
	require.NoError(t, err)

	ents := []database.Entity{
		&Lesson{},
		&LessonGroup{},
		&LessonMember{},
		&Location{},
		&LessonRoomState{},
		&LessonMember{},
		&LessonTeacher{},
		&Course{},
		&Class{},
		&Reallocation{},
		&CourseTypeDTO{},
		&Classroom{},
		&LessonClassroom{},
		&GrantedPermission{},
		&CourseTeachingTime{},
	}

	assert := assert.New(t)
	dir, err := os.Getwd()
	assert.NoError(err)

	count, err := database.CheckEntity(dir)
	/* reduce the counter because:
	- the Classroom entity is counted twice on same table and have some join columns (Classroom & ClassroomToExport)
	- the Lesson entity is counted twice on same table and have some join columns (Lesson & LessonToExport)
	- the CourseTeachingTime entity is counted twice on same table and have some join columns (CourseTeachingTime & CourseTeachingTimeToExport)
	*/
	count = count - 3

	assert.NoError(err)
	assert.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assert.NoError(database.CheckEntityDefinition(e))
		assert.NoError(sv.Verify(e))
	}
}

func TestEntities(t *testing.T) {
	t.Parallel()
	ents := []database.Entities{
		&LessonMembers{},
		&Users{},
		&Lessons{},
		&LessonTeachers{},
	}

	assert := assert.New(t)
	dir, err := os.Getwd()
	assert.NoError(err)

	count, err := database.CheckEntities(dir)
	assert.NoError(err)
	assert.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assert.NoError(database.CheckEntitiesDefinition(e))
	}
}
