package dto

import (
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
)

func TestStudyPlan_ToEntity(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		dto := StudyPlan{
			ID:           database.Text("id"),
			Name:         database.Text("name"),
			CourseID:     database.Text("course_id"),
			AcademicYear: database.Text("academic_year_id"),
			Status:       database.Text(string(domain.StudyPlanStatusActive)),
		}

		expected := domain.StudyPlan{
			ID:           "id",
			Name:         "name",
			CourseID:     "course_id",
			AcademicYear: "academic_year_id",
			Status:       domain.StudyPlanStatusActive,
		}

		// act
		actual, err := dto.ToEntity()

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})
}
