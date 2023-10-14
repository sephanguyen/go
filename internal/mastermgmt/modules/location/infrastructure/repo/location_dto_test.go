package repo

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocationFromEntity(t *testing.T) {
	t.Parallel()
	now := time.Time{}
	tcs := []struct {
		name     string
		location *domain.Location
		dto      *Location
	}{
		{
			name: "full fields",
			location: &domain.Location{
				LocationID:              "location-id-1",
				Name:                    "location name 1",
				LocationType:            "center",
				ParentLocationID:        "location-id-2",
				PartnerInternalID:       "partner-1",
				PartnerInternalParentID: "partner-2",
				IsArchived:              false,
				AccessPath:              "location-id-1",
				CreatedAt:               now,
				UpdatedAt:               now,
				ResourcePath:            fmt.Sprint(constants.ManabieSchool),
			},
			dto: &Location{
				LocationID:              database.Text("location-id-1"),
				Name:                    database.Text("location name 1"),
				LocationType:            database.Text("center"),
				PartnerInternalID:       database.Text("partner-1"),
				PartnerInternalParentID: database.Text("partner-2"),
				IsArchived:              database.Bool(false),
				AccessPath:              database.Text("location-id-1"),
				CreatedAt:               database.Timestamptz(now),
				UpdatedAt:               database.Timestamptz(now),
				DeletedAt:               pgtype.Timestamptz{Status: pgtype.Null},
				ParentLocationID:        database.Text("location-id-2"),
				ResourcePath:            database.Text(fmt.Sprint(constants.ManabieSchool)),
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewLocationFromEntity(tc.location)
			require.NoError(t, err)
			assert.EqualValues(t, tc.dto, actual)
		})
	}
}
