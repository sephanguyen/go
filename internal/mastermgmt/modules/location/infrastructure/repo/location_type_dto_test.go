package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocationTypeFromEntity(t *testing.T) {
	t.Parallel()
	now := time.Time{}
	tcs := []struct {
		name         string
		locationType *domain.LocationType
		dto          *LocationType
	}{
		{
			name: "full fields",
			locationType: &domain.LocationType{
				LocationTypeID:       "location-id-1",
				Name:                 "location name 1",
				DisplayName:          "display 1",
				ParentName:           "location-id-parent-1",
				ParentLocationTypeID: "location-id-2",
				Level:                1,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			dto: &LocationType{
				LocationTypeID:       database.Text("location-id-1"),
				Name:                 database.Text("location name 1"),
				DisplayName:          database.Text("display 1"),
				ParentLocationTypeID: database.Text("location-id-2"),
				ParentName:           database.Text("location-id-parent-1"),
				Level:                database.Int4(1),
				IsArchived:           database.Bool(false),
				CreatedAt:            database.Timestamptz(now),
				UpdatedAt:            database.Timestamptz(now),
				DeletedAt:            pgtype.Timestamptz{Status: pgtype.Null},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewLocationTypeFromEntity(tc.locationType)
			require.NoError(t, err)
			assert.EqualValues(t, tc.dto, actual)
		})
	}
}
