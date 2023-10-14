package jprep

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	yasuo_repo "github.com/manabie-com/backend/internal/yasuo/repositories"

	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func (s *suite) stepRequestNewLiveLessonWithLessonPayload() error {
	now := time.Now()
	s.CurrentLessonID = rand.Intn(1000)
	s.CurrentLessonGroup = idutil.ULIDNow()
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Lessons: []dto.Lesson{
				{
					ActionKind:    dto.ActionKindUpserted,
					LessonID:      s.CurrentLessonID,
					LessonType:    "online",
					CourseID:      s.CurrentCourseID,
					StartDatetime: int(now.Unix()),
					EndDatetime:   int(now.Unix()),
					ClassName:     "class name " + idutil.ULIDNow(),
					Week:          s.CurrentLessonGroup,
				},
			},
		},
	}

	s.Request = request
	return nil
}

func (s *suite) stepRequestNewLiveLessonWithLessonPayloadMissing(missingField string) error {
	now := time.Now()
	s.CurrentLessonID = rand.Intn(1000)
	s.CurrentLessonGroup = idutil.ULIDNow()
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Lessons: []dto.Lesson{
				{
					ActionKind:    dto.ActionKindUpserted,
					LessonID:      s.CurrentLessonID,
					LessonType:    "online",
					CourseID:      s.CurrentCourseID,
					StartDatetime: int(now.Unix()),
					EndDatetime:   int(now.Unix()),
					ClassName:     "class name " + idutil.ULIDNow(),
					Week:          s.CurrentLessonGroup,
				},
			},
		},
	}

	switch missingField {
	case "action kind":
		request.Payload.Lessons[0].ActionKind = ""
	case "lesson id":
		request.Payload.Lessons[0].LessonID = 0
	case "lesson name":
		request.Payload.Lessons[0].ClassName = ""
	case "lesson type":
		request.Payload.Lessons[0].LessonType = ""
	case "course id":
		request.Payload.Lessons[0].CourseID = 0
	case "start datetime":
		request.Payload.Lessons[0].StartDatetime = 0
	case "end datetime":
		request.Payload.Lessons[0].EndDatetime = 0
	}

	s.Request = request
	return nil
}

func (s *suite) stepRequestExistLiveLessonWithLessonPayload(actionKind string) error {
	now := time.Now()
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Lessons: []dto.Lesson{
				{
					ActionKind:    dto.Action(actionKind),
					LessonID:      s.CurrentLessonID,
					LessonType:    "online",
					CourseID:      s.CurrentCourseID,
					StartDatetime: int(now.Unix()),
					EndDatetime:   int(now.Unix()) + 3600,
					ClassName:     "update class name " + idutil.ULIDNow(),
					Week:          s.CurrentLessonGroup,
				},
			},
		},
	}

	s.Request = request
	return nil
}

func (s *suite) theLessonMustBeStoreInOurSystemWithAction(action string) error {
	mainProcess := func() error {
		lessonID := toJprepLessonID(s.CurrentLessonID)
		courseID := toJprepCourseID(s.CurrentCourseID)
		query := `SELECT count(lesson_id)
			FROM lessons
			WHERE lesson_id = $1
			AND course_id = $2
			AND lesson_group_id = $3
			AND resource_path = $4
			AND deleted_at IS NULL`

		if action == "deleted" {
			query = `SELECT count(lesson_id)
				FROM lessons
				WHERE lesson_id = $1
				AND course_id = $2
				AND lesson_group_id = $3
				AND resource_path = $4
				AND deleted_at IS NOT NULL`
		}

		rows, err := s.bobDB.Query(
			context.Background(),
			query,
			lessonID,
			courseID,
			s.CurrentLessonGroup,
			database.Text(fmt.Sprint(constants.JPREPSchool)),
		)
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
			return fmt.Errorf("cannot find any lesson with lesson_id = %s and course_id = %s and lesson_group_id = %s", lessonID, courseID, s.CurrentLessonGroup)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theLessonGroupMustBeStoreInOurSystem() error {
	mainProcess := func() error {
		courseID := toJprepCourseID(s.CurrentCourseID)
		query := `SELECT count(lesson_group_id)
			FROM lesson_groups
			WHERE lesson_group_id = $1
			AND course_id = $2
			AND resource_path = $3`

		rows, err := s.bobDB.Query(
			context.Background(),
			query,
			s.CurrentLessonGroup,
			courseID,
			database.Text(fmt.Sprint(constants.JPREPSchool)),
		)
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
			return fmt.Errorf("cannot find any lesson group with lesson_group_id = %s and course_id = %s", s.CurrentLessonGroup, courseID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theTopicMustBeStoreInOurSystemWithAction(action string) error {
	mainProcess := func() error {
		query := `SELECT count(topic_id)
			FROM topics
			WHERE topic_id = $1
			AND resource_path = $2
			AND deleted_at IS NULL`
		if action == "deleted" {
			query = `SELECT count(topic_id)
				FROM topics
				WHERE topic_id = $1
				AND resource_path = $2
				AND deleted_at IS NOT NULL`
		}

		rows, err := s.bobDB.Query(
			context.Background(),
			query,
			s.CurrentTopicID,
			database.Text(fmt.Sprint(constants.JPREPSchool)),
		)
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
			return fmt.Errorf("cannot find any topic with topic_id = %s", s.CurrentTopicID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) thePresetStudyPlanMustBeStoreInOurSystem() error {
	mainProcess := func() error {
		query := `SELECT count(preset_study_plan_id)
			FROM preset_study_plans
			WHERE preset_study_plan_id = $1
			AND resource_path = $2
			AND deleted_at IS NULL`

		rows, err := s.bobDB.Query(
			context.Background(),
			query, s.CurrentPresetStudyPlanID,
			database.Text(fmt.Sprint(constants.JPREPSchool)),
		)
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
			return fmt.Errorf("cannot find any preset_study_plans with preset_study_plan_id = %s", s.CurrentPresetStudyPlanID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) thePresetStudyPlanWeeklyMustBeStoreInOurSystem() error {
	mainProcess := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		presetStudyPlanWeeklyRepo := &yasuo_repo.PresetStudyPlanWeeklyRepo{}
		lessondID := toJprepLessonID(s.CurrentLessonID)
		lesson := s.Request.(*dto.MasterRegistrationRequest).Payload.Lessons[0]

		p, err := presetStudyPlanWeeklyRepo.FindByLessonID(ctx, s.bobDBTrace, database.Text(lessondID))
		if err != nil {
			return err
		}

		if p == nil {
			return fmt.Errorf("cannot find any preset_study_plans_weekly with lesson_id = %s", lessondID)
		}

		if p.StartDate.Time.Unix() != int64(lesson.StartDatetime) {
			return fmt.Errorf("startDate does not match, expected: %d, got: %d", p.StartDate.Time.Unix(), lesson.StartDatetime)
		}

		if p.EndDate.Time.Unix() != int64(lesson.EndDatetime) {
			return fmt.Errorf("endDate does not match, expected: %d, got: %d", p.EndDate.Time.Unix(), lesson.EndDatetime)
		}

		if err := s.validatePresetStudyPlanWeeklyResourcePath(ctx, p); err != nil {
			return err
		}

		s.CurrentTopicID = p.TopicID.String
		s.CurrentPresetStudyPlanID = p.PresetStudyPlanID.String

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theLessonMustNotBeStoreInOurSystem() error {
	mainProcess := func() error {
		lessonID := toJprepLessonID(s.CurrentLessonID)
		courseID := toJprepCourseID(s.CurrentCourseID)
		query := `SELECT count(lesson_id)
			FROM lessons
			WHERE lesson_id = $1
			AND course_id = $2
			AND lesson_group_id = $3`

		rows, err := s.bobDB.Query(context.Background(), query, lessonID, courseID, s.CurrentLessonGroup)
		if err != nil {
			return err
		}

		defer rows.Close()
		count := 1
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count != 0 {
			return fmt.Errorf("lesson must not be store with lesson_id = %s and course_id = %s and lesson_group_id = %s", lessonID, courseID, s.CurrentLessonGroup)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theLessonGroupMustNotBeStoreInOurSystem() error {
	mainProcess := func() error {
		courseID := toJprepCourseID(s.CurrentCourseID)
		query := `SELECT count(lesson_group_id)
			FROM lesson_groups
			WHERE lesson_group_id = $1
			AND course_id = $2`

		rows, err := s.bobDB.Query(context.Background(), query, s.CurrentLessonGroup, courseID)
		if err != nil {
			return err
		}

		defer rows.Close()
		count := 1
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count != 0 {
			return fmt.Errorf("lesson group must not be store with lesson_group_id = %s and course_id = %s", s.CurrentLessonGroup, courseID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theTopicMustNotBeStoreInOurSystem() error {
	mainProcess := func() error {
		query := `SELECT count(topic_id)
			FROM topics
			WHERE topic_id = $1`

		rows, err := s.bobDB.Query(context.Background(), query, s.CurrentTopicID)
		if err != nil {
			return err
		}

		defer rows.Close()
		count := 1
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count != 0 {
			return fmt.Errorf("topic must not be store with topic_id = %s", s.CurrentTopicID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) thePresetStudyPlanMustNotBeStoreInOurSystem() error {
	mainProcess := func() error {
		query := `SELECT count(preset_study_plan_id)
			FROM preset_study_plans
			WHERE preset_study_plan_id = $1`

		rows, err := s.bobDB.Query(context.Background(), query, s.CurrentPresetStudyPlanID)
		if err != nil {
			return err
		}

		defer rows.Close()
		count := 1
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count != 0 {
			return fmt.Errorf("preset_study_plans must not be store with preset_study_plan_id = %s", s.CurrentPresetStudyPlanID)
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) thePresetStudyPlanWeeklyMustNotBeStoreInOurSystem() error {
	mainProcess := func() error {
		lessondID := toJprepLessonID(s.CurrentLessonID)
		lesson := s.Request.(*dto.MasterRegistrationRequest).Payload.Lessons[0]

		query := `SELECT count(*)
			FROM preset_study_plans_weekly
			WHERE lesson_id = $1
			AND EXTRACT(epoch FROM start_date) = $2
			and EXTRACT(epoch FROM end_date) = $3`

		rows, err := s.bobDB.Query(
			context.Background(),
			query,
			database.Text(lessondID),
			lesson.StartDatetime,
			lesson.EndDatetime,
		)
		if err != nil {
			return err
		}

		defer rows.Close()
		count := 1
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err
			}
		}

		if count != 0 {
			return fmt.Errorf("preset_study_plans_weekly must not be store with lesson_id = %s", toJprepLessonID(s.CurrentLessonID))
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) thePresetStudyPlanWeeklyMustBeDeleteInOurSystem() error {
	mainProcess := func() error {
		presetStudyPlanWeeklyRepo := &yasuo_repo.PresetStudyPlanWeeklyRepo{}
		lessondID := toJprepLessonID(s.CurrentLessonID)

		p, err := presetStudyPlanWeeklyRepo.FindByLessonID(context.Background(), s.bobDBTrace, database.Text(lessondID))
		if err != nil {
			return err
		}

		if p == nil {
			return fmt.Errorf("cannot find any preset_study_plans_weekly with lesson_id = %s", lessondID)
		}

		if p.DeletedAt.Time.IsZero() {
			return fmt.Errorf("cannot delete preset_study_plans_weekly with lesson_id = %s", lessondID)
		}
		s.CurrentTopicID = p.TopicID.String

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) tomMustStoreConversationLesson() error {
	mainProcess := func() error {
		lessonID := toJprepLessonID(s.CurrentLessonID)
		query := `SELECT conversation_id 
			FROM conversation_lesson
			WHERE lesson_id = $1`

		rows, err := s.tomDB.Query(context.Background(), query, lessonID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var conversationID string
		for rows.Next() {
			err = rows.Scan(&conversationID)
			if err != nil {
				return err
			}
		}

		if conversationID == "" {
			return fmt.Errorf("not found any conversation_lesson with lesson_id = %s", lessonID)
		}

		s.ConversationID = conversationID
		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) theCourseMustBeUpdateInOurSystem() error {
	mainProcess := func() error {
		courseRepo := &bob_repo.CourseRepo{}
		courseID := toJprepCourseID(s.CurrentCourseID)
		lesson := s.Request.(*dto.MasterRegistrationRequest).Payload.Lessons[0]

		course, err := courseRepo.FindByID(context.Background(), s.bobDBTrace, database.Text(courseID))
		if err != nil {
			return err
		}

		if course == nil {
			return fmt.Errorf("cannot find any course with course_id = %s", courseID)
		}

		if course.StartDate.Time.Unix() != int64(lesson.StartDatetime) {
			return fmt.Errorf("startDate does not match, expected: %d, got: %d", course.StartDate.Time.Unix(), lesson.StartDatetime)
		}

		if course.EndDate.Time.Unix() != int64(lesson.EndDatetime) {
			return fmt.Errorf("endDate does not match, expected: %d, got: %d", course.EndDate.Time.Unix(), lesson.EndDatetime)
		}

		if course.PresetStudyPlanID.String != s.CurrentPresetStudyPlanID {
			return fmt.Errorf("preset_study_plan_id is stored in course not match with lesson, preset_study_plan_id = %s", s.CurrentPresetStudyPlanID)
		}

		if err := s.validateCourseResourcePath(context.Background(), course); err != nil {
			return err
		}

		return nil
	}

	return s.ExecuteWithRetry(mainProcess, time.Second*2, 5)
}

func (s *suite) validatePresetStudyPlanWeeklyResourcePath(ctx context.Context, presetStudyPlanWeekly *entities.PresetStudyPlanWeekly) error {
	var resourcePath string
	query := "SELECT resource_path FROM preset_study_plans_weekly WHERE preset_study_plan_weekly_id = $1"
	if err := database.Select(ctx, s.bobDBTrace, query, presetStudyPlanWeekly.ID).ScanFields(&resourcePath); err != nil {
		return err
	}

	if resourcePath != fmt.Sprint(constants.JPREPSchool) {
		return fmt.Errorf("resourcePath does not match, expected: %s, got: %s", fmt.Sprint(constants.JPREPSchool), resourcePath)
	}

	return nil
}

func (s *suite) validateCourseResourcePath(ctx context.Context, course *entities.Course) error {
	var resourcePath string
	query := "SELECT resource_path FROM courses WHERE course_id = $1"
	if err := database.Select(ctx, s.bobDBTrace, query, course.ID).ScanFields(&resourcePath); err != nil {
		return err
	}

	if resourcePath != fmt.Sprint(constants.JPREPSchool) {
		return fmt.Errorf("resourcePath does not match, expected: %s, got: %s", fmt.Sprint(constants.JPREPSchool), resourcePath)
	}

	return nil
}
