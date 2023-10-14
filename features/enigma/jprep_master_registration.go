package enigma

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

func (s *suite) requestMasterRegistration(ctx context.Context, course, lesson, class, academic_year int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.CurrentUserID = idutil.ULIDNow()
	now := time.Now()
	// courses
	courses := make([]dto.Course, 0, course)
	for i := 1; i <= course; i++ {
		courses = append(courses, dto.Course{
			ActionKind:         dto.ActionKindUpserted,
			CourseID:           i,
			CourseName:         "course-name-with-actionKind-upsert",
			CourseStudentDivID: rand.Intn(2) + 1,
		})
	}
	// lesson
	lessons := make([]dto.Lesson, 0, lesson)
	for i := 1; i <= lesson; i++ {
		lessons = append(lessons, dto.Lesson{
			ActionKind:    dto.ActionKindUpserted,
			LessonID:      i,
			LessonType:    "online",
			CourseID:      i,
			StartDatetime: int(now.Unix()),
			EndDatetime:   int(now.Unix()),
			ClassName:     "class name " + s.CurrentUserID,
			Week:          "1",
		})
	}
	// class
	classes := make([]dto.Class, 0, class)
	for i := 1; i <= class; i++ {
		classes = append(classes, dto.Class{
			ActionKind:     dto.ActionKindUpserted,
			ClassName:      "class name " + s.CurrentUserID,
			ClassID:        rand.Intn(999999999),
			CourseID:       i,
			StartDate:      now.Format("2006/01/02"),
			EndDate:        now.Format("2006/01/02"),
			AcademicYearID: i,
		})
	}
	// academic_year
	academicYears := make([]dto.AcademicYear, 0, academic_year)
	for i := 1; i <= academic_year; i++ {
		academicYears = append(academicYears, dto.AcademicYear{
			ActionKind:     dto.ActionKindUpserted,
			AcademicYearID: i,
			Name:           "name " + s.CurrentUserID,
			StartYearDate:  int64(i),
			EndYearDate:    int64(i),
		})
	}
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(now.Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Courses:       courses,
			Classes:       classes,
			Lessons:       lessons,
			AcademicYears: academicYears,
		},
	}

	s.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) requestMasterRegistrationInvalidPayload(ctx context.Context, course, lesson, class, academic_year int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.CurrentUserID = idutil.ULIDNow()
	now := time.Now()
	// courses
	courses := make([]dto.Course, 0, course)
	for i := 1; i <= course; i++ {
		courses = append(courses, dto.Course{
			ActionKind: dto.ActionKindUpserted,
			CourseName: "course-name-with-actionKind-upsert",
		})
	}
	// lesson
	lessons := make([]dto.Lesson, 0, lesson)
	for i := 1; i <= lesson; i++ {
		lessons = append(lessons, dto.Lesson{
			ActionKind:    dto.ActionKindUpserted,
			LessonType:    "online",
			CourseID:      i,
			StartDatetime: int(now.Unix()),
			EndDatetime:   int(now.Unix()),
			ClassName:     "class name " + s.CurrentUserID,
			Week:          "1",
		})
	}
	// class
	classes := make([]dto.Class, 0, class)
	for i := 1; i <= class; i++ {
		classes = append(classes, dto.Class{
			ActionKind:     dto.ActionKindUpserted,
			ClassName:      "class name " + s.CurrentUserID,
			ClassID:        rand.Intn(999999999),
			StartDate:      now.Format("2006/01/02"),
			EndDate:        now.Format("2006/01/02"),
			AcademicYearID: i,
		})
	}
	// academic_year
	academicYears := make([]dto.AcademicYear, 0, academic_year)
	for i := 1; i <= academic_year; i++ {
		academicYears = append(academicYears, dto.AcademicYear{
			ActionKind:    dto.ActionKindUpserted,
			Name:          "name " + s.CurrentUserID,
			StartYearDate: int64(i),
			EndYearDate:   int64(i),
		})
	}
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(now.Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Courses:       courses,
			Classes:       classes,
			Lessons:       lessons,
			AcademicYears: academicYears,
		},
	}

	s.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) stepPerformMasterRegistrationRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	url := fmt.Sprintf("%s/jprep/master-registration", s.EnigmaSrvURL)
	bodyBytes, err := s.makeHTTPRequest(http.MethodPut, url)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if bodyBytes == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("body is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) logCorrectLesson(ctx context.Context, payload []byte) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessons := []*npb.EventMasterRegistration_Lesson{}
	err := json.Unmarshal(payload, &lessons)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal lessons: %w", err)
	}
	for _, lesson := range lessons {
		found := false
		for _, lessonReq := range s.Request.(*dto.MasterRegistrationRequest).Payload.Lessons {
			if lesson.LessonId == fmt.Sprintf("JPREP_LESSON_%09d", lessonReq.LessonID) {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find lesson %s", lesson.LessonId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) logCorrectCourse(ctx context.Context, payload []byte) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courses := []*npb.EventMasterRegistration_Course{}
	err := json.Unmarshal(payload, &courses)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal courses: %w", err)
	}
	for _, course := range courses {
		found := false
		for _, courseReq := range s.Request.(*dto.MasterRegistrationRequest).Payload.Courses {
			if course.CourseId == fmt.Sprintf("JPREP_COURSE_%09d", courseReq.CourseID) {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find course %s", course.CourseId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) logCorrectClass(ctx context.Context, payload []byte) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	classes := []*npb.EventMasterRegistration_Class{}
	err := json.Unmarshal(payload, &classes)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal student lessons: %w", err)
	}
	for _, class := range classes {
		found := false
		for _, classReq := range s.Request.(*dto.MasterRegistrationRequest).Payload.Classes {
			if int(class.ClassId) == classReq.ClassID {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find class %d", class.ClassId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) logCorrectAcademicYear(ctx context.Context, payload []byte) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	academicYears := []*npb.EventMasterRegistration_AcademicYear{}
	err := json.Unmarshal(payload, &academicYears)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("json.Unmarshal academicYears: %w", err)
	}
	for _, academicYear := range academicYears {
		found := false
		for _, academicYearReq := range s.Request.(*dto.MasterRegistrationRequest).Payload.AcademicYears {
			if academicYear.AcademicYearId == fmt.Sprintf("JPREP_ACADEMIC_YEAR_%09d", academicYearReq.AcademicYearID) {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find academicYear %s", academicYear.AcademicYearId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
