package entities

import (
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntity(t *testing.T) {
	t.Parallel()
	sv, err := database.NewSchemaVerifier("fatima") // use fatima db for payment now
	require.NoError(t, err)

	entities := []database.Entity{
		&BillItem{},
		&Discount{},
		&Product{},
		&ProductGroup{},
		&ProductGroupMapping{},
		&StudentDiscountTracker{},
		&StudentProduct{},
		&UserDiscountTag{},
		&StudentParent{},
		&DiscountTag{},
		&PackageDiscountSetting{},
		&PackageDiscountCourseMapping{},
		&OrderItem{},
		&User{},
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
	var entities []database.Entities

	assertions := assert.New(t)
	dir, err := os.Getwd()
	assertions.NoError(err)

	count, err := database.CheckEntities(dir)
	assertions.NoError(err)
	assertions.Equalf(count, len(entities), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(entities))

	for _, entity := range entities {
		assertions.NoError(database.CheckEntitiesDefinition(entity))
	}
}
