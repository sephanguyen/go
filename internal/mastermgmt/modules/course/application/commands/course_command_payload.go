package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

type UpdateCoursesCommand struct {
	Courses           []*domain.Course
	CourseAccessPaths []*domain.CourseAccessPath
	CourseIDs         []string
}
type ImportCoursesPayload struct {
	Courses []*domain.Course
}
