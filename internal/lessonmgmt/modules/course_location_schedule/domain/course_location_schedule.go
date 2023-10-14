package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type (
	ProductTypeSchedule string
)

var ErrUniqCourseLocationSchedule = errors.New("duplicate key value violates unique constraint unique_course_location_schedule")
var ErrNotExistsFKCourseLocationSchedule = errors.New("violates foreign key constraint course_location_schedule_fk")

type ImportCourseLocationScheduleError struct {
	Index int
	Err   error
}

const (
	OneTime   ProductTypeSchedule = "PACKAGE_TYPE_ONE_TIME"
	Scheduled ProductTypeSchedule = "PACKAGE_TYPE_SCHEDULED"
	SlotBase  ProductTypeSchedule = "PACKAGE_TYPE_SLOT_BASED"
	Frequency ProductTypeSchedule = "PACKAGE_TYPE_FREQUENCY"
)

var ProductTypeScheduleMap = map[string]ProductTypeSchedule{
	"1": OneTime,
	"2": Scheduled,
	"3": SlotBase,
	"4": Frequency,
}

var MapStringToProductTypeSchedule = map[string]ProductTypeSchedule{
	"PACKAGE_TYPE_ONE_TIME":   OneTime,
	"PACKAGE_TYPE_SCHEDULED":  Scheduled,
	"PACKAGE_TYPE_SLOT_BASED": SlotBase,
	"PACKAGE_TYPE_FREQUENCY":  Frequency,
}

var MapStringToProductTypeScheduleNumber = map[string]string{
	"PACKAGE_TYPE_ONE_TIME":   "1",
	"PACKAGE_TYPE_SCHEDULED":  "2",
	"PACKAGE_TYPE_SLOT_BASED": "3",
	"PACKAGE_TYPE_FREQUENCY":  "4",
}

const IDLabel = "course_location_schedule_id"
const CourseIDLabel = "course_id"
const LocationIDLabel = "location_id"
const AcademicWeekLabel = "academic_week"
const ProductTypeScheduleLabel = "product_type_schedule"
const FrequencyLabel = "frequency"
const TotalNoLessonLabel = "total_no_lessons"

type CourseLocationSchedule struct {
	ID                  string
	CourseID            string
	LocationID          string
	AcademicWeeks       []string
	ProductTypeSchedule ProductTypeSchedule
	Frequency           *int
	TotalNoLesson       *int
	CreatedAt           *time.Time
	UpdatedAt           *time.Time
	DeletedAt           *time.Time
	Persisted           bool // true: already exists in db
}

type CourseLocationScheduleBuilder struct {
	CourseLocationSchedule *CourseLocationSchedule
}

func NewCourseLocationScheduleBuilder() *CourseLocationScheduleBuilder {
	return &CourseLocationScheduleBuilder{
		CourseLocationSchedule: &CourseLocationSchedule{},
	}
}

func (c *CourseLocationScheduleBuilder) WithID(id string) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.ID = id
	c.CourseLocationSchedule.Persisted = id != ""
	if id == "" {
		c.CourseLocationSchedule.ID = idutil.ULIDNow()
	}
	return c
}

func (c *CourseLocationScheduleBuilder) WithCourseID(courseID string) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.CourseID = courseID
	return c
}

func (c *CourseLocationScheduleBuilder) WithCreateAt(createAt *time.Time) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.CreatedAt = createAt
	return c
}

func (c *CourseLocationScheduleBuilder) WithUpdatedAt(updatedAt *time.Time) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.UpdatedAt = updatedAt
	return c
}

func (c *CourseLocationScheduleBuilder) WithLocationID(locationID string) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.LocationID = locationID
	return c
}

func (c *CourseLocationScheduleBuilder) WithAcademicWeek(academicWeeks []string) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.AcademicWeeks = academicWeeks
	return c
}

func (c *CourseLocationScheduleBuilder) WithProductTypeSchedule(productTypeSchedule ProductTypeSchedule) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.ProductTypeSchedule = productTypeSchedule
	return c
}

func (c *CourseLocationScheduleBuilder) WithFrequency(frequency *int) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.Frequency = frequency
	return c
}

func (c *CourseLocationScheduleBuilder) WithTotalNoLesson(totalNoLesson *int) *CourseLocationScheduleBuilder {
	c.CourseLocationSchedule.TotalNoLesson = totalNoLesson
	return c
}

func (c *CourseLocationScheduleBuilder) Build() (*CourseLocationSchedule, error) {
	if err := c.CourseLocationSchedule.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid course location schedule detail: %w", err)
	}
	return c.CourseLocationSchedule, nil
}

func (cls *CourseLocationSchedule) IsValid() error {
	if cls.CourseID == "" {
		return fmt.Errorf("courseId not be empty")
	}
	if cls.LocationID == "" {
		return fmt.Errorf("locationId not be empty")
	}
	return nil
}

func (cls *CourseLocationSchedule) IsFrequency() bool {
	return cls.ProductTypeSchedule == Frequency
}

func (cls *CourseLocationSchedule) IsSchedule() bool {
	return cls.ProductTypeSchedule == Scheduled
}

func (cls *CourseLocationSchedule) IsOneTime() bool {
	return cls.ProductTypeSchedule == OneTime
}

func (cls *CourseLocationSchedule) IsSlotBased() bool {
	return cls.ProductTypeSchedule == SlotBase
}

func (cls *CourseLocationSchedule) IsDefinedByOrder() bool {
	return cls.ProductTypeSchedule == SlotBase || cls.ProductTypeSchedule == Frequency
}

func (cls *CourseLocationSchedule) IsDefinedByCourse() bool {
	return cls.ProductTypeSchedule == OneTime || cls.ProductTypeSchedule == Scheduled
}

func (cls *CourseLocationSchedule) IsScheduleWeekly() bool {
	return cls.ProductTypeSchedule == Frequency || cls.ProductTypeSchedule == Scheduled
}
