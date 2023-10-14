package domain

import (
	"context"
	"fmt"
	"time"
)

type CourseAccessPath struct {
	ID         string
	CourseID   string
	LocationID string
	UpdatedAt  time.Time
	CreatedAt  time.Time
	DeletedAt  *time.Time
}

type CourseAccessPathBuilder struct {
	courseAccessPath *CourseAccessPath
}

func NewCourseAccessPath() *CourseAccessPathBuilder {
	return &CourseAccessPathBuilder{
		courseAccessPath: &CourseAccessPath{},
	}
}

func (c *CourseAccessPathBuilder) Build(ctx context.Context) (*CourseAccessPath, error) {
	if err := c.courseAccessPath.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid course_access_path: %w", err)
	}
	return c.courseAccessPath, nil
}

func (c *CourseAccessPathBuilder) WithLocationID(id string) *CourseAccessPathBuilder {
	c.courseAccessPath.LocationID = id
	return c
}

func (c *CourseAccessPathBuilder) WithID(id string) *CourseAccessPathBuilder {
	c.courseAccessPath.ID = id
	return c
}

func (c *CourseAccessPathBuilder) WithCourseID(id string) *CourseAccessPathBuilder {
	c.courseAccessPath.CourseID = id
	return c
}

func (c *CourseAccessPathBuilder) WithModificationTime(createdAt, updatedAt time.Time) *CourseAccessPathBuilder {
	c.courseAccessPath.CreatedAt = createdAt
	c.courseAccessPath.UpdatedAt = updatedAt
	return c
}

func (c *CourseAccessPathBuilder) GetCourseAccessPath() *CourseAccessPath {
	return c.courseAccessPath
}

func (c *CourseAccessPath) IsValid() error {
	if len(c.LocationID) == 0 {
		return fmt.Errorf("CourseAccessPath.LocationID cannot be empty")
	}

	if len(c.CourseID) == 0 {
		return fmt.Errorf("CourseAccessPath.CourseID cannot be empty")
	}

	if c.UpdatedAt.Before(c.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}
	return nil
}
