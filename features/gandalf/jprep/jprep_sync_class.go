package jprep

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func (s *suite) stepARequestNewClassWithRegularCoursePayload() error {
	now := time.Now().Format("2006/01/02")
	s.CurrentCourseID = rand.Intn(999999999)
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Classes: []dto.Class{
				{
					ActionKind:     dto.ActionKindUpserted,
					ClassName:      "class name " + idutil.ULIDNow(),
					ClassID:        rand.Intn(999999999),
					CourseID:       s.CurrentCourseID,
					StartDate:      now,
					EndDate:        now,
					AcademicYearID: s.CurrentAcademicYearID,
				},
			},

			Courses: []dto.Course{
				{
					ActionKind: dto.ActionKindUpserted,
					CourseID:   s.CurrentCourseID,
					CourseName: "course-name-with-actionKind-upsert",
				},
			},
		},
	}

	s.Request = request
	s.CurrentClassID = request.Payload.Classes[0].ClassID
	return nil
}

func (s *suite) stepARequestNewClassWithRegularCoursePayloadMissing(missingField string) error {
	now := time.Now().Format("2006/01/02")
	s.CurrentCourseID = rand.Intn(999999999)
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Classes: []dto.Class{
				{
					ActionKind:     dto.ActionKindUpserted,
					ClassName:      "class name " + idutil.ULIDNow(),
					ClassID:        rand.Intn(999999999),
					CourseID:       s.CurrentCourseID,
					StartDate:      now,
					EndDate:        now,
					AcademicYearID: s.CurrentAcademicYearID,
				},
			},

			Courses: []dto.Course{
				{
					ActionKind: dto.ActionKindUpserted,
					CourseID:   s.CurrentCourseID,
					CourseName: "course-name-with-actionKind-upsert",
				},
			},
		},
	}

	switch missingField {
	case "action kind":
		request.Payload.Classes[0].ActionKind = ""
	case "course id":
		request.Payload.Classes[0].CourseID = 0
	case "class id":
		request.Payload.Classes[0].ClassID = 0
	case "class name":
		request.Payload.Classes[0].ClassName = ""
	case "academic year id":
		request.Payload.Classes[0].AcademicYearID = 0
	case "start date":
		request.Payload.Classes[0].StartDate = ""
	case "end date":
		request.Payload.Classes[0].EndDate = ""
	}

	s.Request = request
	s.CurrentClassID = request.Payload.Classes[0].ClassID
	return nil
}

func (s *suite) stepARequestExistClassWithRegularCoursePayload(actionKind string) error {
	now := time.Now().Format("2006/01/02")
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Classes: []dto.Class{
				{
					ActionKind:     dto.Action(actionKind),
					ClassName:      "update class name " + idutil.ULIDNow(),
					ClassID:        s.CurrentClassID,
					CourseID:       s.CurrentCourseID,
					StartDate:      now,
					EndDate:        now,
					AcademicYearID: s.CurrentAcademicYearID,
				},
			},
		},
	}

	s.Request = request
	return nil
}

func (s *suite) validateClassResourcePath(ctx context.Context, class *entities.Class) error {
	var resourcePath string
	query := "SELECT resource_path FROM classes WHERE class_id = $1"
	if err := database.Select(ctx, s.bobDBTrace, query, class.ID).ScanFields(&resourcePath); err != nil {
		return err
	}

	if resourcePath != fmt.Sprint(constants.JPREPSchool) {
		return fmt.Errorf("resourcePath does not match, expected: %s, got: %s", fmt.Sprint(constants.JPREPSchool), resourcePath)
	}

	return nil
}

func (s *suite) theClassesMustBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		classRepo := &repositories.ClassRepo{}
		courseClassRepo := &repositories.CourseClassRepo{}

		c := s.Request.(*dto.MasterRegistrationRequest).Payload.Classes[0]
		class, err := classRepo.FindByID(ctx, s.bobDBTrace, database.Int4(int32(c.ClassID)))
		if err != nil {
			return fmt.Errorf("err find class: %w", err)
		}

		switch c.ActionKind {
		case dto.ActionKindUpserted:
			if class == nil {
				return fmt.Errorf("class does not existed")
			}

			if class.Name.String != c.ClassName {
				return fmt.Errorf("class name does not match, expected: %s, got: %s", c.ClassName, class.Name.String)
			}

			if class.Country.String != "COUNTRY_JP" {
				return fmt.Errorf("class country does not match, expected: %s, got: %s", "COUNTRY_JP", class.Country.String)
			}

			err := s.validateClassResourcePath(ctx, class)
			if err != nil {
				return err
			}

			// check course_class
			mapByClass, err := courseClassRepo.Find(ctx, s.bobDBTrace, database.Int4Array([]int32{int32(c.ClassID)}))
			if err != nil {
				return fmt.Errorf("err find Class Course")
			}

			courses, ok := mapByClass[class.ID]
			if !ok {
				return fmt.Errorf("not found any courses")
			}

			found := false
			for _, course := range courses.Elements {
				if course.String == toJprepCourseID(c.CourseID) {
					found = true
				}
			}

			if !found {
				return fmt.Errorf("course id not found")
			}

			count := 0
			query := `SELECT COUNT(*) FROM courses_academic_years WHERE course_id = $1 AND academic_year_id = $2 AND resource_path = $3 AND deleted_at is NULL`
			s.bobDBTrace.QueryRow(
				ctx,
				query,
				toJprepCourseID(c.CourseID),
				toJprepAcedemicYearID(c.AcademicYearID),
				database.Text(fmt.Sprint(constants.JPREPSchool)),
			).Scan(&count)

			if count == 0 {
				return fmt.Errorf("academicYearId does not match, expected %v", c.AcademicYearID)
			}

		case dto.ActionKindDeleted:
			// check course_class deleted
			if class.Status.String == entities.ClassStatusActive {
				return fmt.Errorf("class does not deleted, still active")
			}

			// check course_class
			mapByClass, err := courseClassRepo.Find(ctx, s.bobDBTrace, database.Int4Array([]int32{int32(c.ClassID)}))
			if err != nil {
				return fmt.Errorf("err find Class Course")
			}

			if len(mapByClass) != 0 {
				return fmt.Errorf("course class does not deleted")
			}
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theCoursesClassesMustNotBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		c := s.Request.(*dto.MasterRegistrationRequest).Payload.Classes[0]

		query := `SELECT count(*) 
			FROM courses_classes
			WHERE class_id = $1
			AND course_id = $2`

		count := 1
		err := s.bobDB.QueryRow(
			ctx,
			query,
			database.Int4(int32(c.ClassID)),
			database.Text(toJprepCourseID(c.CourseID)),
		).Scan(&count)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("courses_classes must not be store")
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theCoursesAcademicYearsMustNotBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		c := s.Request.(*dto.MasterRegistrationRequest).Payload.Classes[0]

		query := `SELECT count(*) 
			FROM courses_academic_years
			WHERE academic_year_id = $1
			AND course_id = $2`

		count := 1
		err := s.bobDB.QueryRow(
			ctx,
			query,
			database.Text(toJprepAcedemicYearID(c.AcademicYearID)),
			database.Text(toJprepCourseID(c.CourseID)),
		).Scan(&count)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf(
				"courses_academic_years must not be store with course_id = %s, academic_year_id= %s",
				toJprepCourseID(c.CourseID),
				toJprepAcedemicYearID(c.AcademicYearID),
			)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theClassesMustNotBeStoreInOutSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		c := s.Request.(*dto.MasterRegistrationRequest).Payload.Classes[0]

		query := `SELECT count(*) 
			FROM classes
			WHERE class_id = $1`
		count := 1
		err := s.bobDB.QueryRow(ctx, query, database.Int4(int32(c.ClassID))).Scan(&count)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("class must not be store")
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) tomMustStoreConversationWithStatus(status string) error {
	mainProcess := func() error {
		query := `select count(conversation_id)
			from conversations
			where conversation_id = $1
			and status = $2`
		rows, err := s.tomDB.Query(context.Background(), query, s.ConversationID, status)
		if err != nil {
			return err
		}

		defer rows.Close()
		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count == 0 {
			return fmt.Errorf("cannot find any conversation with conversation_id = %s, status = %s ", s.ConversationID, status)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) tomMustRecordMessageCreated() error {
	mainProcess := func() error {
		query := `SELECT count(message_id) 
				FROM messages 
				WHERE conversation_id = $1 
				AND message = $2 
				AND type = $3 
				AND deleted_at IS NULL`
		rows, err := s.tomDB.Query(context.Background(), query, s.ConversationID, "CODES_MESSAGE_TYPE_CREATED_CLASS", "MESSAGE_TYPE_SYSTEM")
		if err != nil {
			return err
		}
		defer rows.Close()
		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count == 0 {
			return fmt.Errorf("tom cannot create message with conversation_id = %v, message =%v, type = %v", s.ConversationID, "CODES_MESSAGE_TYPE_CREATED_CLASS", "MESSAGE_TYPE_SYSTEM")
		}

		return nil
	}
	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}
