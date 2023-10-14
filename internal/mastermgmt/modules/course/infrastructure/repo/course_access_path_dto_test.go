package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCourseAccessPathFromEntity(t *testing.T) {
	t.Parallel()
	now := time.Time{}
	tcs := []struct {
		name string
		caps *domain.CourseAccessPath
		dto  *CourseAccessPath
	}{
		{
			name: "full fields",
			caps: &domain.CourseAccessPath{
				ID:         "cap_id",
				LocationID: "location-id-1",
				CourseID:   "course-id-1",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			dto: &CourseAccessPath{
				ID:         database.Varchar("cap_id"),
				LocationID: database.Text("location-id-1"),
				CourseID:   database.Text("course-id-1"),
				CreatedAt:  database.Timestamptz(now),
				UpdatedAt:  database.Timestamptz(now),
				DeletedAt:  pgtype.Timestamptz{Status: pgtype.Null},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewCourseAccessPathFromEntity(tc.caps)
			require.NoError(t, err)
			assert.EqualValues(t, tc.dto, actual)
		})
	}
}
