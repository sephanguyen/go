package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCourseTypeFromEntity(t *testing.T) {
	t.Parallel()
	now := time.Time{}
	tcs := []struct {
		name             string
		courseTypeDomain *domain.CourseType
		dto              *CourseTypeDTO
	}{
		{
			name: "full fields",
			courseTypeDomain: &domain.CourseType{
				CourseTypeID: "id-1",
				Name:         "name-1",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			dto: &CourseTypeDTO{
				ID:        database.Text("id-1"),
				Name:      database.Text("name-1"),
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewCourseTypeFromEntity(tc.courseTypeDomain)
			require.NoError(t, err)
			assert.EqualValues(t, tc.dto, actual)
		})
	}
}
