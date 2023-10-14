package domain

import "github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"

// Assessment represents a Test that students need to take.
type Assessment struct {
	ID string // known as activity_id.

	// The compound IDs is alias of ID.
	CourseID           string
	LearningMaterialID string

	// LearningMaterialType helps us to indicate which type of learning belongs to.
	LearningMaterialType constants.LearningMaterialType

	// Virtual fields.
	ManualGrading bool
}

func (a *Assessment) Validate() error {
	if a.ID == "" {
		return ErrIDRequired
	}
	if a.CourseID == "" {
		return ErrCourseIDRequired
	}
	if a.LearningMaterialID == "" {
		return ErrLearningMaterialIDRequired
	}

	switch a.LearningMaterialType {
	case constants.LearningObjective, constants.ExamLO, constants.FlashCard, constants.GeneralAssignment, constants.TaskAssignment:
	default:
		return ErrInvalidLearningMaterialType
	}

	return nil
}
