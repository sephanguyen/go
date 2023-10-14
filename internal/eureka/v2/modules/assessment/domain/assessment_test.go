package domain

import (
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/stretchr/testify/assert"
)

func TestAssessment_Validate(t *testing.T) {
	t.Parallel()

	t.Run("return nil when no fields miss", func(t *testing.T) {
		// arrange
		sut := Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             idutil.ULIDNow(),
			LearningMaterialID:   idutil.ULIDNow(),
			LearningMaterialType: constants.LearningObjective,
		}
		// act
		err := sut.Validate()

		// assert
		assert.Nil(t, err)
	})

	t.Run("return ErrIDRequired when there is no ID", func(t *testing.T) {
		// arrange
		sut := Assessment{
			ID:                   "",
			CourseID:             idutil.ULIDNow(),
			LearningMaterialID:   idutil.ULIDNow(),
			LearningMaterialType: constants.LearningObjective,
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrIDRequired, err)
	})

	t.Run("return ErrCourseIDRequired when there is no Course ID", func(t *testing.T) {
		// arrange
		sut := Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             "",
			LearningMaterialID:   idutil.ULIDNow(),
			LearningMaterialType: constants.LearningObjective,
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrCourseIDRequired, err)
	})

	t.Run("return ErrLearningMaterialIDRequired when there is no LM ID", func(t *testing.T) {
		// arrange
		sut := Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             idutil.ULIDNow(),
			LearningMaterialID:   "",
			LearningMaterialType: constants.LearningObjective,
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrLearningMaterialIDRequired, err)
	})

	t.Run("return ErrInvalidLearningMaterialType when there is no ID", func(t *testing.T) {
		// arrange
		sut := Assessment{
			ID:                   idutil.ULIDNow(),
			CourseID:             idutil.ULIDNow(),
			LearningMaterialID:   idutil.ULIDNow(),
			LearningMaterialType: "",
		}
		// act
		err := sut.Validate()

		// assert
		assert.Equal(t, ErrInvalidLearningMaterialType, err)
	})

}
