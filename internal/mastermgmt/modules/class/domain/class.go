package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgx/v4"
)

type Class struct {
	ClassID      string
	Name         string
	CourseID     string
	LocationID   string
	SchoolID     string
	UpdatedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
	ResourcePath string

	Persisted    bool
	Repo         ClassRepo
	LocationRepo LocationRepo
	CourseRepo   CourseRepo
}

type ClassBuilder struct {
	class *Class
}

func NewClass() *ClassBuilder {
	return &ClassBuilder{
		class: &Class{},
	}
}

func (c *ClassBuilder) IsValid(ctx context.Context, db database.Ext) error {
	if len(c.class.Name) == 0 {
		return fmt.Errorf("Class.Name cannot be empty")
	}
	if !utf8.ValidString(c.class.Name) {
		return fmt.Errorf("Class.Name is not valid UTF8 format")
	}
	if len(c.class.CourseID) == 0 {
		return fmt.Errorf("Class.CourseID cannot be empty")
	}
	if len(c.class.LocationID) == 0 {
		return fmt.Errorf("Class.LocationID cannot be empty")
	}
	if err := c.validateLocation(ctx, db); err != nil {
		return fmt.Errorf("Class.LocationID %w", err)
	}
	if err := c.validateCourse(ctx, db); err != nil {
		return fmt.Errorf("Class.CourseID %w", err)
	}
	return nil
}

func (c *ClassBuilder) validateLocation(ctx context.Context, db database.Ext) error {
	_, err := c.class.LocationRepo.GetLocationByID(ctx, db, c.class.LocationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("invalid")
	} else if err != nil {
		return fmt.Errorf("failed to get location by id: %w", err)
	}
	return nil
}

func (c *ClassBuilder) validateCourse(ctx context.Context, db database.Ext) error {
	_, err := c.class.CourseRepo.GetByID(ctx, db, c.class.CourseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("invalid")
	} else if err != nil {
		return fmt.Errorf("failed to get course by id: %w", err)
	}
	return nil
}

func (c *ClassBuilder) Build(ctx context.Context, db database.Ext) (*Class, error) {
	if err := c.IsValid(ctx, db); err != nil {
		return nil, fmt.Errorf("invalid class: %w", err)
	}
	if c.class.ClassID == "" {
		c.class.ClassID = idutil.ULIDNow()
	}
	return c.class, nil
}

func (c *ClassBuilder) WithClassRepo(repo ClassRepo) *ClassBuilder {
	c.class.Repo = repo
	return c
}
func (c *ClassBuilder) WithLocationRepo(locationRepo LocationRepo) *ClassBuilder {
	c.class.LocationRepo = locationRepo
	return c
}
func (c *ClassBuilder) WithCourseRepo(courseRepo CourseRepo) *ClassBuilder {
	c.class.CourseRepo = courseRepo
	return c
}

func (c *ClassBuilder) WithName(name string) *ClassBuilder {
	c.class.Name = name
	return c
}

func (c *ClassBuilder) WithCourseID(courseID string) *ClassBuilder {
	c.class.CourseID = courseID
	return c
}

func (c *ClassBuilder) WithLocationID(locationID string) *ClassBuilder {
	c.class.LocationID = locationID
	return c
}

func (c *ClassBuilder) WithSchoolID(schooldID string) *ClassBuilder {
	c.class.SchoolID = schooldID
	return c
}

func (c *ClassBuilder) WithModificationTime(createdAt, updatedAt time.Time) *ClassBuilder {
	c.class.CreatedAt = createdAt
	c.class.UpdatedAt = updatedAt
	return c
}

func (c *ClassBuilder) WithDeletedTime(deletedAt *time.Time) *ClassBuilder {
	c.class.DeletedAt = deletedAt
	return c
}

type ClassWithCourseStudent struct {
	CourseID  string
	StudentID string
	ClassID   string
}
