package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/stretchr/testify/assert"
)

func TestToImportLog(t *testing.T) {
	t.Parallel()
	now := time.Now()
	testcases := []struct {
		name      string
		importLog *domain.ImportLog
		dto       *ImportLog
	}{
		{
			name: "empty payload",
			importLog: &domain.ImportLog{
				ID:         "log-1",
				UserID:     "user-1",
				ImportType: "location",
				Payload:    "{upserted_ids: ['location-1']}",
				CreatedAt:  now,
				DeletedAt:  &now,
			},
			dto: &ImportLog{
				ID:         database.Text("log-1"),
				UserID:     database.Text("user-1"),
				ImportType: database.Text("location"),
				Payload:    database.JSONB("{upserted_ids: ['location-1']}"),
				CreatedAt:  database.Timestamptz(now),
				DeletedAt:  database.Timestamptz(now),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualDto, err := ToImportLog(tc.importLog)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.dto, actualDto)
		})
	}
}
