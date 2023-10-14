package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
)

type (
	CourseTeachingMethod string
)

const (
	CourseTeachingMethodNone       CourseTeachingMethod = "COURSE_TEACHING_METHOD_NONE"
	CourseTeachingMethodIndividual CourseTeachingMethod = "COURSE_TEACHING_METHOD_INDIVIDUAL"
	CourseTeachingMethodGroup      CourseTeachingMethod = "COURSE_TEACHING_METHOD_GROUP"
)

type Course struct {
	CourseID          string
	LocationID        string
	Name              string
	Country           string
	Subject           string
	Grade             int
	DisplayOrder      int
	SchoolID          int
	TeacherIDs        []string
	CourseType        string
	Icon              string
	PresetStudyPlanID string
	Status            string
	StartDate         time.Time
	EndDate           time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	TeachingMethod    CourseTeachingMethod
	CourseTypeID      string
	LocationIDs       []string
	IsArchived        bool
	Remarks           string
	PartnerID         string

	// relational fields
	SubjectIDs []string

	// port
	LocationRepo   infrastructure.LocationRepo
	CourseTypeRepo CourseTypeRepo
}

type CourseBuilder struct {
	course *Course
}

func NewCourse() *CourseBuilder {
	return &CourseBuilder{
		course: &Course{},
	}
}

func (c *CourseBuilder) Build(ctx context.Context, db database.Ext) (*Course, error) {
	if err := c.course.IsValid(ctx, db); err != nil {
		return nil, err
	}
	return c.course, nil
}

func (c *CourseBuilder) WithLocationRepo(repo infrastructure.LocationRepo) *CourseBuilder {
	c.course.LocationRepo = repo
	return c
}

func (c *CourseBuilder) WithCourseTypeRepo(repo CourseTypeRepo) *CourseBuilder {
	c.course.CourseTypeRepo = repo
	return c
}

func (c *CourseBuilder) WithCourseID(courseID string) *CourseBuilder {
	if len(courseID) == 0 {
		courseID = idutil.ULIDNow()
	}
	c.course.CourseID = courseID
	return c
}

func (c *CourseBuilder) WithLocationID(locationID string) *CourseBuilder {
	c.course.LocationID = locationID
	return c
}

func (c *CourseBuilder) WithName(name string) *CourseBuilder {
	c.course.Name = name
	return c
}

func (c *CourseBuilder) WithDisplayOrder(displayOrder int) *CourseBuilder {
	c.course.DisplayOrder = displayOrder
	return c
}

func (c *CourseBuilder) WithLocationIDs(locationIDs []string) *CourseBuilder {
	c.course.LocationIDs = locationIDs
	return c
}

func (c *CourseBuilder) WithSchoolID(schoolID int) *CourseBuilder {
	c.course.SchoolID = schoolID
	return c
}

func (c *CourseBuilder) WithCourseType(courseType string) *CourseBuilder {
	c.course.CourseTypeID = courseType
	return c
}

func (c *CourseBuilder) WithIcon(icon string) *CourseBuilder {
	c.course.Icon = icon
	return c
}

func (c *CourseBuilder) WithTeachingMethod(teachingMethod CourseTeachingMethod) *CourseBuilder {
	c.course.TeachingMethod = teachingMethod
	return c
}

func (c *CourseBuilder) WithModificationTime(createdAt, updatedAt time.Time) *CourseBuilder {
	c.course.CreatedAt = createdAt
	c.course.UpdatedAt = updatedAt
	return c
}

func (c *CourseBuilder) WithSubjects(subjects []string) *CourseBuilder {
	c.course.SubjectIDs = subjects
	return c
}

func (c *CourseBuilder) WithPartnerID(partner string) *CourseBuilder {
	c.course.PartnerID = partner
	return c
}

func (c *CourseBuilder) GetCourse() *Course {
	return c.course
}

func (c *Course) IsValid(ctx context.Context, db database.Ext) error {
	if len(c.CourseID) == 0 {
		return fmt.Errorf("courseID cannot be empty")
	}
	if len(c.Name) == 0 {
		return fmt.Errorf("course name cannot be empty")
	}
	if c.UpdatedAt.Before(c.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}
	// check locations
	if len(c.LocationIDs) > 0 {
		locations, err := c.LocationRepo.GetLocationsByLocationIDs(ctx, db, database.TextArray(c.LocationIDs), true)
		if err != nil {
			return err
		}
		if len(locations) != len(c.LocationIDs) {
			return fmt.Errorf("locationIDs invalid")
		}
	}
	if len(c.CourseTypeID) > 0 {
		courseTypes, err := c.CourseTypeRepo.GetByIDs(ctx, db, []string{c.CourseTypeID})
		if err != nil {
			return err
		}
		if len(courseTypes) == 0 {
			return fmt.Errorf("course type invalid")
		}
	}
	return nil
}
func (c *Course) IsValidTeachingMethod() error {
	// This field will be validated by FE, so we only need to whether it's valid to update.
	if c.TeachingMethod != CourseTeachingMethodIndividual &&
		c.TeachingMethod != CourseTeachingMethodGroup {
		return fmt.Errorf("invalid course teaching method")
	}
	return nil
}

func (c *Course) String() string {
	return fmt.Sprintf("ID: %s;Name: %s;Partner: %s;Subjects: [%s];Remarks: %s\n", c.CourseID, c.Name, c.PartnerID, strings.Join(c.SubjectIDs, ","), c.Remarks)
}

func ConvertTeachingMethodToString(teachingMethod string) string {
	teachingMethods := map[string]string{
		"COURSE_TEACHING_METHOD_GROUP":      "Group",
		"COURSE_TEACHING_METHOD_INDIVIDUAL": "Individual",
		"COURSE_TEACHING_METHOD_NONE":       "",
	}

	return teachingMethods[teachingMethod]
}
