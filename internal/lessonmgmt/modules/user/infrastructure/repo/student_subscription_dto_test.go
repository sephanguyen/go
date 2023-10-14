package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/stretchr/testify/assert"
)

func TestStudentSubscriptions_ToListStudentSubscriptionEntities(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name        string
		dto         StudentSubscriptions
		locationIDs map[string][]string
		gradeIds    map[string]string
		expected    domain.StudentSubscriptions
	}{
		{
			name: "successfully",
			dto: StudentSubscriptions{
				{
					StudentSubscriptionID: database.Text("sub-id-1"),
					CourseID:              database.Text("course-id-1"),
					StudentID:             database.Text("student-id-1"),
					SubscriptionID:        database.Text("alter-id-1"),
					StartAt:               database.Timestamptz(now.Add(2 * time.Hour)),
					EndAt:                 database.Timestamptz(now.Add(3 * time.Hour)),
					CreatedAt:             database.Timestamptz(now),
					UpdatedAt:             database.Timestamptz(now),
					DeletedAt:             database.Timestamptz(now),
				},
				{
					StudentSubscriptionID: database.Text("sub-id-2"),
					CourseID:              database.Text("course-id-2"),
					StudentID:             database.Text("student-id-2"),
					SubscriptionID:        database.Text("alter-id-2"),
					StartAt:               database.Timestamptz(now.Add(2 * time.Hour)),
					EndAt:                 database.Timestamptz(now.Add(5 * time.Hour)),
					CreatedAt:             database.Timestamptz(now),
					UpdatedAt:             database.Timestamptz(now),
					DeletedAt:             database.Timestamptz(now),
				},
				{
					StudentSubscriptionID: database.Text("sub-id-3"),
					CourseID:              database.Text("course-id-2"),
					StudentID:             database.Text("student-id-1"),
					SubscriptionID:        database.Text("alter-id-3"),
					StartAt:               database.Timestamptz(now.Add(2 * time.Hour)),
					EndAt:                 database.Timestamptz(now.Add(3 * time.Hour)),
					CreatedAt:             database.Timestamptz(now),
					UpdatedAt:             database.Timestamptz(now),
					DeletedAt:             database.Timestamptz(now),
				},
			},
			gradeIds: map[string]string{
				"sub-id-1": "grade-1",
			},
			locationIDs: map[string][]string{
				"sub-id-1": {
					"location-id-1",
					"location-id-2",
				},
				"sub-id-2": {
					"location-id-1",
					"location-id-3",
					"location-id-5",
				},
			},
			expected: domain.StudentSubscriptions{
				{
					SubscriptionID: "sub-id-1",
					StudentID:      "student-id-1",
					CourseID:       "course-id-1",
					LocationIDs:    []string{"location-id-1", "location-id-2"},
					StartAt:        now.Add(2 * time.Hour),
					EndAt:          now.Add(3 * time.Hour),
					CreatedAt:      now,
					UpdatedAt:      now,
					GradeV2:        "grade-1",
				},
				{
					SubscriptionID: "sub-id-2",
					StudentID:      "student-id-2",
					CourseID:       "course-id-2",
					LocationIDs:    []string{"location-id-1", "location-id-3", "location-id-5"},
					StartAt:        now.Add(2 * time.Hour),
					EndAt:          now.Add(5 * time.Hour),
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					SubscriptionID: "sub-id-3",
					StudentID:      "student-id-1",
					CourseID:       "course-id-2",
					StartAt:        now.Add(2 * time.Hour),
					EndAt:          now.Add(3 * time.Hour),
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.dto.ToListStudentSubscriptionEntities(tc.locationIDs, tc.gradeIds)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
