package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

type UpsertCourseAccessPathsCommand struct {
	CourseAccessPaths []*domain.CourseAccessPath
}
