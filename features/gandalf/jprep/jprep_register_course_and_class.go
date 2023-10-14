package jprep

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
)

func (s *suite) stepARequestWithCourseNamePayload() error {
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Courses: []dto.Course{
				{
					ActionKind: dto.ActionKindUpserted,
					CourseID:   rand.Intn(10000),
					CourseName: "course-name-with-actionKind-upsert",
				},
				{
					ActionKind: dto.ActionKindUpserted,
					CourseID:   rand.Intn(10000),
					CourseName: "course-name-with-actionKind-upsert",
				},
			},
		},
	}
	s.Request = request

	for _, course := range request.Payload.Courses {
		s.CurrentCourseIDs = append(s.CurrentCourseIDs, course.CourseID)
	}

	return nil
}

func (s *suite) stepRequestWithCurrentCourseNamePayloadAndAction(actionKind string) error {
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Courses: []dto.Course{
				{
					ActionKind: dto.Action(actionKind),
					CourseID:   s.CurrentCourseIDs[0],
					CourseName: "course-name-with-actionKind-upsert",
				},
				{
					ActionKind: dto.Action(actionKind),
					CourseID:   s.CurrentCourseIDs[1],
					CourseName: "course-name-with-actionKind-upsert",
				},
			},
		},
	}
	s.Request = request

	return nil
}

func (s *suite) stepARequestWithCourseNamePayloadMissing(missingField string) error {
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Courses: []dto.Course{
				{
					ActionKind: dto.ActionKindUpserted,
					CourseID:   rand.Intn(10000),
					CourseName: "course-name-with-actionKind-upsert",
				},
			},
		},
	}

	switch missingField {
	case "action kind":
		request.Payload.Courses[0].ActionKind = ""
	case "course id":
		request.Payload.Courses[0].CourseID = 0
	case "course name":
		request.Payload.Courses[0].CourseName = ""
	}

	s.Request = request

	for _, course := range request.Payload.Courses {
		s.CurrentCourseIDs = append(s.CurrentCourseIDs, course.CourseID)
	}

	return nil
}

func (s *suite) stepYasuoMustCreateCourse() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		courseIDs := []string{}
		courses := s.Request.(*dto.MasterRegistrationRequest).Payload.Courses
		for _, course := range courses {
			courseIDs = append(courseIDs, toJprepCourseID(course.CourseID))
		}

		count := 0
		query := fmt.Sprintln("SELECT count(*) FROM courses WHERE course_id = ANY($1) AND resource_path = $2 AND deleted_at IS NULL")
		err := s.bobDB.QueryRow(ctx, query, courseIDs, database.Text(fmt.Sprint(constants.JPREPSchool))).Scan(&count)
		if err != nil {
			return err
		}
		if count != len(courses) {
			return fmt.Errorf("yasuo does not upsert course correctly")
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) stepYasuoMustNotCreateCourse() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		courseIDs := []string{}
		courses := s.Request.(*dto.MasterRegistrationRequest).Payload.Courses
		for _, course := range courses {
			courseIDs = append(courseIDs, toJprepCourseID(course.CourseID))
		}

		count := 1
		query := fmt.Sprintln("SELECT count(*) FROM courses WHERE course_id = ANY($1)")
		err := s.bobDB.QueryRow(ctx, query, courseIDs).Scan(&count)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("yasuo must not upsert course")
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) stepYasuoMustCreateCourseWithAction(actionKind string) error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		courseIDs := []string{}
		courses := s.Request.(*dto.MasterRegistrationRequest).Payload.Courses
		for _, course := range courses {
			courseIDs = append(courseIDs, toJprepCourseID(course.CourseID))
		}

		count := 1
		query := fmt.Sprintln("SELECT count(*) FROM courses WHERE course_id = ANY($1) AND resource_path = $2 AND deleted_at IS NULL")
		if actionKind == "deleted" {
			query = fmt.Sprintln("SELECT count(*) FROM courses WHERE course_id = ANY($1) AND resource_path = $2 AND deleted_at IS NOT NULL")
		}
		err := s.bobDB.QueryRow(ctx, query, courseIDs, database.Text(fmt.Sprint(constants.JPREPSchool))).Scan(&count)
		if err != nil {
			return err
		}
		if count != len(courses) {
			return fmt.Errorf("yasuo does not upsert course correctly with action = %s", actionKind)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}
