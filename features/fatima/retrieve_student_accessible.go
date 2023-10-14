package fatima

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
)

func (s *suite) aSignedAs(arg string) error {
	s.UserID = idutil.ULIDNow()
	var userGroup string
	switch arg {
	case "teacher":
		userGroup = cpb.UserGroup_USER_GROUP_TEACHER.String()
	case "student":
		userGroup = cpb.UserGroup_USER_GROUP_STUDENT.String()
	case "admin":
		userGroup = cpb.UserGroup_USER_GROUP_ADMIN.String()
	case "school admin":
		userGroup = cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String()
	}
	token, err := s.generateValidAuthenticationToken(s.UserID, userGroup)
	if err != nil {
		return fmt.Errorf("unable to generate token: %w", err)
	}
	s.AuthToken = token
	return nil
}
func (s *suite) theUserRetrieveStudentAccessibleCourse() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s.Response, s.ResponseErr = fpb.NewAccessibilityReadServiceClient(s.Conn).RetrieveStudentAccessibility(contextWithToken(s, ctx), &fpb.RetrieveStudentAccessibilityRequest{UserId: s.UserID})
	return nil
}

func (s *suite) aStudentHasPackageIs(pkgs, rawStatuses string) error {
	s.UserID = idutil.ULIDNow()
	return s.userPackage(s.UserID, pkgs, rawStatuses)
}

func (s *suite) returnsAllCourseAccessibleResponseOfThisStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*fpb.RetrieveStudentAccessibilityResponse)
	courses := make(map[string]*fpb.RetrieveAccessibilityResponse_CourseAccessibility)
	for key, c := range resp.Courses {
		courses[key] = convCourseAccessibilityCommonToSpecific(c)

	}
	return StepStateToContext(ctx, stepState), s.returnAllCourseAccessibleWithUserID(ctx, courses, s.UserID)
}

func convCourseAccessibilityCommonToSpecific(course *cpb.CourseAccessibility) *fpb.RetrieveAccessibilityResponse_CourseAccessibility {
	return &fpb.RetrieveAccessibilityResponse_CourseAccessibility{
		CanDoQuiz:         course.CanDoQuiz,
		CanWatchVideo:     course.CanWatchVideo,
		CanViewStudyGuide: course.CanViewStudyGuide,
	}
}
