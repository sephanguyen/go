package domain

import (
	"fmt"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

// StudentSubscriptions are not entity or aggregate, it's just is data type to contain list studentSubscriptions
type StudentSubscriptions []*StudentSubscription

func (s StudentSubscriptions) IsValid() error {
	for i := range s {
		if err := s[i].IsValid(); err != nil {
			return err
		}
	}
	return nil
}

type StudentSubscription struct {
	SubscriptionID        string
	StudentID             string
	CourseID              string
	ClassID               string
	Grade                 string
	LocationIDs           []string
	StartAt               time.Time
	EndAt                 time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
	GradeV2               string
	StudentSubscriptionID string
	CourseSlot            int32
	CourseSlotPerWeek     int32
	StudentFirstName      string
	StudentLastName       string
	PackageType           string
}

func (s *StudentSubscription) IsValid() error {
	if len(s.SubscriptionID) == 0 {
		return fmt.Errorf("SubscriptionID could not be empty")
	}

	if len(s.StudentID) == 0 {
		return fmt.Errorf("StudentID could not be empty")
	}

	if len(s.CourseID) == 0 {
		return fmt.Errorf("CourseID could not be empty")
	}

	if s.EndAt.Before(s.StartAt) {
		return fmt.Errorf("end time could not before start time")
	}

	if s.UpdatedAt.Before(s.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}

	return nil
}

func (s *StudentSubscription) StudentWithCourseID() string {
	return s.StudentID + "-" + s.CourseID
}

func (s *StudentSubscription) WithGrade(gradeV2 string) *StudentSubscription {
	s.GradeV2 = gradeV2
	return s
}

type StudentCoursesAndClasses struct {
	StudentID string
	Courses   []*StudentCoursesAndClassesCourses
	Classes   []*StudentCoursesAndClassesClasses
}

func (s *StudentCoursesAndClasses) ToGetStudentCoursesAndClassesResponse() *lpb.GetStudentCoursesAndClassesResponse {
	if s == nil {
		return nil
	}
	res := &lpb.GetStudentCoursesAndClassesResponse{
		StudentId: s.StudentID,
		Courses:   nil,
		Classes:   nil,
	}

	for _, course := range s.Courses {
		res.Courses = append(res.Courses, &cpb.Course{
			Info: &cpb.ContentBasicInfo{
				Id:   course.CourseID,
				Name: course.Name,
			},
		})
	}

	for _, class := range s.Classes {
		res.Classes = append(res.Classes, &lpb.GetStudentCoursesAndClassesResponse_Class{
			ClassId:  class.ClassID,
			Name:     class.Name,
			CourseId: class.CourseID,
		})
	}

	return res
}

type StudentCoursesAndClassesCourses struct {
	CourseID string
	Name     string
}

type StudentCoursesAndClassesClasses struct {
	ClassID  string
	Name     string
	CourseID string
}
