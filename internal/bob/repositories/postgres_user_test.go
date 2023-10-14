package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPostgresUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()

	mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)

	r := PostgresUserRepo{}

	e := entities.PostgresUser{}
	fields, values := e.FieldMap()

	mockDB.MockScanArray(nil, fields, [][]interface{}{
		values,
	})

	data, err := r.Get(ctx, mockDB.DB)
	assert.Equal(t, err, nil)
	assert.NotNil(t, data)
}

func TestGetPostgresNamespace(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := testutil.NewMockDB()
	mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
	r := PostgresNamespaceRepo{}
	e := entities.PostgresNamespace{}
	fields, values := e.FieldMap()

	mockDB.MockScanArray(nil, fields, [][]interface{}{
		values,
	})

	data, err := r.Get(ctx, mockDB.DB)

	assert.Equal(t, err, nil)
	assert.NotNil(t, data)
}
