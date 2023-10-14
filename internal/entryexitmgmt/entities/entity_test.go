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
	sv, err := database.NewSchemaVerifier("entryexitmgmt")
	require.NoError(t, err)

	ents := []database.Entity{
		&StudentQR{},
		&StudentEntryExitRecords{},
		&StudentParent{},
		&EntryExitQueue{},
		&Student{},
		&User{},
		&Location{},
		&UserAccessPaths{},
		&UserBasicInfo{},
		&Grade{},
	}

	assert := assert.New(t)
	dir, err := os.Getwd()
	assert.NoError(err)

	count, err := database.CheckEntity(dir)
	assert.NoError(err)
	assert.Equalf(count, len(ents), "found %d entities in package, but only %d are being checked; please add new entities to the unit test", count, len(ents))

	for _, e := range ents {
		assert.NoError(database.CheckEntityDefinition(e))
		assert.NoError(sv.Verify(e))
	}
}

func TestEntities(t *testing.T) {
	t.Parallel()
	ents := []database.Entities{}

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
