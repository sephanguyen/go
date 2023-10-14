package domain

import (
	"time"
)

type Course struct {
	ID             string
	Name           string
	Icon           string
	DisplayOrder   int
	LocationID     string
	SchoolID       int
	CourseType     string
	TeachingMethod string

	Country           string
	SubjectIDs        []string
	Grade             int
	TeacherIDs        []string
	PresetStudyPlanID string
	Status            string

	CourseTypeID string
	LocationIDs  []string
	IsArchived   bool
	Remarks      string
	PartnerID    string

	StartDate time.Time
	EndDate   time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	BookID string
}

func NewCourse(courseID, name, icon string, displayOrder int, schoolID int, locationIds []string, courseTypeID string, bookID string, teachingMethod string, subjectIDs []string) Course {
	c := Course{
		ID:             courseID,
		Name:           name,
		Icon:           icon,
		DisplayOrder:   displayOrder,
		SchoolID:       schoolID,
		LocationIDs:    locationIds,
		CourseTypeID:   courseTypeID,
		TeachingMethod: teachingMethod,
		SubjectIDs:     subjectIDs,
	}
	if bookID != "" {
		c.BookID = bookID
	}
	return c
}
