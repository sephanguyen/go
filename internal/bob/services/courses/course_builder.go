package courses

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type CourseBuilder struct {
	db database.Ext
	repo.CourseRepo
}

func NewCourseBuilder(db database.Ext, repo repo.CourseRepo) *CourseBuilder {
	return &CourseBuilder{
		db:         db,
		CourseRepo: repo,
	}
}

func (c *CourseBuilder) UpdateCourseAvailableRanges(ctx context.Context, availableRanges *entities.CourseAvailableRanges) error {
	courseIDs := availableRanges.GetIDs()
	courses, err := c.CourseRepo.FindByIDs(ctx, c.db, courseIDs)
	if err != nil {
		return fmt.Errorf("CourseRepo.FindByIDs: %s", err)
	}

	updatedCourses := make([]*entities.Course, 0)
	for i, course := range courses {
		if v := availableRanges.Get(course.ID); v != nil {
			courses[i].StartDate = v.StartDate
			courses[i].EndDate = v.EndDate
			err = courses[i].DeletedAt.Set(nil)
			if err != nil {
				return fmt.Errorf("courses[%s].DeletedAt.Set(nil): %s", i.String, err)
			}
			updatedCourses = append(updatedCourses, courses[i])
		}
	}

	err = c.CourseRepo.Upsert(ctx, c.db, updatedCourses)
	if err != nil {
		return fmt.Errorf("CourseRepo.Upsert: %s", err)
	}

	return nil
}
