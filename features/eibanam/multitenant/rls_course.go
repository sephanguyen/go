package multitenant

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuo_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bob_pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	com_pbv1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	"github.com/pkg/errors"
)

func (s *suite) loginsCMSApp(ctx context.Context, resourcePath string) (context.Context, error) {
	time.Sleep(2 * time.Second) // wait to avoid locking the course table
	s.StepState.ResourcePath = resourcePath

	return ctx, s.signedInAsAccountWithResourcePath("school admin", resourcePath)
}

func (s *suite) loginsTeacherApp(ctx context.Context, resourcePath string) (context.Context, error) {
	s.StepState.ResourcePath = resourcePath

	return ctx, s.signedInAsAccountWithResourcePath("teacher", resourcePath)
}

func (s *suite) schoolAdminCreatesANewCourse(ctx context.Context) error {
	nCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	s.StepState.UserGroupInContext = constant.UserGroupSchoolAdmin
	nCtx = contextWithTokenForGrpcCall(s, nCtx)

	req, err := addCourseReq(s)
	if err != nil {
		return errors.Wrap(err, "addCourseReq()")
	}
	resp, err := yasuo_pb.NewCourseServiceClient(s.yasuoConn).UpsertCourses(nCtx, req)
	if err != nil {
		return errors.Wrap(err, "UpsertCourses()")
	}
	s.StepState.ResponseStack.Push(resp)
	s.StepState.RequestStack.Push(req)

	return nil
}

func addCourseReq(s *suite) (*yasuo_pb.UpsertCoursesRequest, error) {
	id := idutil.ULIDNow()
	country := bob_pb.COUNTRY_VN
	grade, err := i18n.ConvertIntGradeToString(country, 1)
	if err != nil {
		return nil, errors.Wrap(err, "ConvertIntGradeToString")
	}

	req := &yasuo_pb.UpsertCoursesRequest{
		Courses: []*yasuo_pb.UpsertCoursesRequest_Course{
			{
				Id:           id,
				Name:         id,
				Country:      country,
				Subject:      bob_pb.SUBJECT_MATHS,
				Grade:        grade,
				DisplayOrder: 1,
				SchoolId:     int32(s.getSchoolId()),
				Icon:         "https://example-url.com",
			},
		},
	}
	return req, nil
}

func (s *suite) enableRLSCourseTable(ctx context.Context, table string) (context.Context, error) {
	nCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`ALTER TABLE %s ENABLE ROW LEVEL security;`, table)
	_, err := s.bobPostgresDB.Exec(nCtx, stmt)
	if err != nil {
		return ctx, err
	}

	stmt = fmt.Sprintf(`ALTER TABLE %s FORCE ROW LEVEL security;`, table)
	_, err = s.bobPostgresDB.Exec(nCtx, stmt)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) disableRLSCourseTable(ctx context.Context, table string) (context.Context, error) {
	nCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stmt := fmt.Sprintf(`ALTER TABLE %s DISABLE ROW LEVEL security;`, table)
	_, err := s.bobPostgresDB.Exec(nCtx, stmt)
	if err != nil {
		return ctx, err
	}

	stmt = fmt.Sprintf(`ALTER TABLE %s NO FORCE ROW LEVEL security;`, table)
	_, err = s.bobPostgresDB.Exec(nCtx, stmt)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) teacherSeeNewCourseWithResourcePath(ctx context.Context, resutl string) (context.Context, error) {
	nCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	s.StepState.UserGroupInContext = constant.UserGroupTeacher
	nCtx = contextWithTokenForGrpcCall(s, nCtx)

	upsertCourseReq := s.RequestStack.Requests[0].(*yasuo_pb.UpsertCoursesRequest)
	req := &bob_pbv1.ListCoursesRequest{
		Paging: nil,
		Filter: &com_pbv1.CommonFilter{
			Ids: []string{upsertCourseReq.Courses[0].Id},
		},
	}
	resp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).ListCourses(nCtx, req)
	if err != nil {
		return ctx, errors.Wrap(err, "ListCourses()")
	}

	if resutl == "can see" {
		if len(resp.Items) == 0 {
			return ctx, errors.New("Teacher cannot see a new course")
		}
		return ctx, canSeeANewCourse(upsertCourseReq, resp)

	} else if resutl == "cannot see" {
		if len(resp.Items) != 0 {
			return ctx, errors.New("Teacher can see a new course")
		}
	}

	return ctx, nil
}

func canSeeANewCourse(upsertCourseReq *yasuo_pb.UpsertCoursesRequest, resp *bob_pbv1.ListCoursesResponse) error {
	requestedToCreateCourse := upsertCourseReq.Courses[0]
	requestedToCreateCourseGrade, err := i18n.ConvertStringGradeToInt(requestedToCreateCourse.Country, requestedToCreateCourse.Grade)
	if err != nil {
		return errors.Wrap(err, "ConvertStringGradeToInt()")
	}
	createdCourse := resp.Items[0]

	switch {
	case createdCourse.Info.Name != requestedToCreateCourse.Name:
		return fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.Name, createdCourse.Info.Name)
	case createdCourse.Info.IconUrl != requestedToCreateCourse.Icon:
		return fmt.Errorf(`expect created course has "icon": %v but actual is %v`, requestedToCreateCourse.Icon, createdCourse.Info.IconUrl)
	case createdCourse.Info.DisplayOrder != requestedToCreateCourse.DisplayOrder:
		return fmt.Errorf(`expect created course has "display_order": %v but actual is %v`, requestedToCreateCourse.DisplayOrder, createdCourse.Info.DisplayOrder)
	case createdCourse.Info.Subject.String() != requestedToCreateCourse.Subject.String():
		return fmt.Errorf(`expect created course has "subject": %v but actual is %v`, requestedToCreateCourse.Subject.String(), createdCourse.Info.Subject)
	case createdCourse.Info.Country.String() != requestedToCreateCourse.Country.String():
		return fmt.Errorf(`expect created course has "name": %v but actual is %v`, requestedToCreateCourse.Name, createdCourse.Info.Name)
	case createdCourse.Info.Grade != int32(requestedToCreateCourseGrade):
		return fmt.Errorf(`expect created course has "grade": %v but actual is %v`, requestedToCreateCourseGrade, createdCourse.Info.Grade)
	}

	return nil
}
