package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

type ImportCourseTypesPayload struct {
	CourseTypes []*domain.CourseType
}
