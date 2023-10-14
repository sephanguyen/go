package usermanagement

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/eibanam"
	"github.com/manabie-com/backend/internal/enigma/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	eureka_pbv1 "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/pkg/errors"
)

func (s *suite) systemHasSyncedCourseFromPartner() error {
	courseId := rand.Intn(999) + 1
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
					ActionKind:         dto.ActionKindUpserted,
					CourseID:           courseId,
					CourseName:         "course-name-with-actionKind-upsert",
					CourseStudentDivID: dto.CourseIDKid,
				},
			},
		},
	}

	if err := s.attachValidJPREPSignature(request); err != nil {
		return err
	}

	if err := s.performMasterRegistrationRequest(request); err != nil {
		return errors.Wrap(err, "makeJPREPHTTPRequest")
	}

	s.RequestStack.Push(request)

	return nil
}

func (s *suite) teacherSeesClassAvailableInClassFilterOnTeacherApp() error {
	// Setup context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.UserGroupInContext = constant.UserGroupTeacher
	ctx = contextWithTokenForGrpcCall(s, ctx)

	req1, err := s.RequestStack.Pop()
	if err != nil {
		return err
	}
	classID := req1.(*dto.MasterRegistrationRequest).Payload.Classes[0].ClassID
	courseID := req1.(*dto.MasterRegistrationRequest).Payload.Classes[0].CourseID

	req := &eureka_pbv1.ListClassByCourseRequest{
		CourseId: toJprepCourseID(courseID),
	}
	err = eibanam.TryUntilSuccess(ctx, 100*time.Millisecond, func(ctx context.Context) (bool, error) {
		// Check if teacher sees class available in class filter using
		// eureka ListClassByCourse api
		resp, err := eureka_pbv1.NewCourseReaderServiceClient(s.eurekaConn).ListClassByCourse(ctx, req)
		if err != nil {
			return true, err
		}

		if len(resp.ClassIds) == 0 {
			return true, fmt.Errorf("teacher does not see any class in filter")
		}
		for _, id := range resp.ClassIds {
			if strconv.Itoa(classID) == id {
				return false, nil
			}
		}

		return false, fmt.Errorf("teacher does not see class id %v in filter", classID)
	})
	return err
}

func (s *suite) systemSyncsClassToCourseWhichCurrentAcademicYear(status string) error {
	req, err := s.RequestStack.Peek()
	if err != nil {
		return err
	}
	coursesRequested := req.(*dto.MasterRegistrationRequest).Payload.Courses
	classes := []dto.Class{
		{
			ActionKind: dto.ActionKindUpserted,
			ClassName:  "class name " + idutil.ULIDNow(),
			ClassID:    rand.Intn(99999999),
			CourseID:   coursesRequested[0].CourseID,
			StartDate:  time.Now().Add(-48 * time.Hour).Format("2006/01/02"),
			EndDate:    time.Now().Add(48 * time.Hour).Format("2006/01/02"),
		},
	}
	t := time.Now()
	year, _, _ := t.Date()
	request := &dto.MasterRegistrationRequest{
		Timestamp: int(time.Now().Unix()),
		Payload: struct {
			Courses       []dto.Course       `json:"m_course_name"`
			Classes       []dto.Class        `json:"m_regular_course"`
			Lessons       []dto.Lesson       `json:"m_lesson"`
			AcademicYears []dto.AcademicYear `json:"m_academic_year"`
		}{
			Classes: classes,
			Courses: coursesRequested,
			AcademicYears: []dto.AcademicYear{
				{
					ActionKind:     dto.ActionKindUpserted,
					AcademicYearID: year,
					Name:           "Year " + idutil.ULIDNow(),
					StartYearDate:  time.Date(year, time.January, 1, 0, 0, 0, 0, t.Location()).Unix(),
					EndYearDate:    time.Date(year+1, time.January, 1, 0, 0, 0, 0, t.Location()).Unix() - 1,
				},
				{
					ActionKind:     dto.ActionKindUpserted,
					AcademicYearID: year - 1,
					Name:           "Year " + idutil.ULIDNow(),
					StartYearDate:  time.Date(year-1, time.January, 1, 0, 0, 0, 0, t.Location()).Unix(),
					EndYearDate:    time.Date(year, time.January, 1, 0, 0, 0, 0, t.Location()).Unix() - 1,
				},
			},
		},
	}

	switch status {
	case "in":
		(&request.Payload).Classes[0].AcademicYearID = year
	default:
		return fmt.Errorf("this status %v is not supported for testing", status)
	}

	if err := s.attachValidJPREPSignature(request); err != nil {
		return err
	}

	if err := s.performMasterRegistrationRequest(request); err != nil {
		return errors.Wrap(err, "makeJPREPHTTPRequest")
	}

	s.RequestStack.Push(request)

	return nil
}
