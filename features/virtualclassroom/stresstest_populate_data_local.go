package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) signedAdminAccountWithE2E(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.TestingSchool

	roleWithLocation := usermgmt.RoleWithLocation{
		RoleName:    constant.RoleSchoolAdmin,
		LocationIDs: []string{constants.E2EOrgLocation},
	}
	adminCtx := s.returnRootContextWithE2E(ctx)

	authInfo, err := usermgmt.SignIn(adminCtx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn,
		s.Cfg.JWTApplicant,
		s.CommonSuite.StepState.FirebaseAddress,
		s.Connections.UserMgmtConn,
		roleWithLocation,
		[]string{constants.E2EOrgLocation},
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = authInfo.UserID
	stepState.AuthToken = authInfo.Token
	stepState.LocationID = constants.E2EOrgLocation
	ctx = common.ValidContext(ctx, constants.TestingSchool, authInfo.UserID, authInfo.Token)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnRootContextWithE2E(ctx context.Context) context.Context {
	return common.ValidContext(ctx, constants.TestingSchool, s.RootAccount[constants.TestingSchool].UserID, s.RootAccount[constants.TestingSchool].Token)
}

func (s *suite) hasACenterForStressTest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CreateStressTestLocation = true

	return s.someCenters(StepStateToContext(ctx, stepState))
}

func (s *suite) hasACourseForStressTest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CourseIDs = []string{"VCSTRESSTESTCOURSE"}

	for _, id := range stepState.CourseIDs {
		if ctx, err := s.CommonSuite.UpsertLiveCourse(ctx, id, stepState.TeacherIDs, constants.TestingSchool); err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasAStudentForStressTest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	student, err := s.createUserWithRoleUsingE2E(ctx, constant.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentIds = append(stepState.StudentIds, student.UserID)

	if len(stepState.StudentIds) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no students were created")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasATeacherForStressTest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	teacher, err := s.createUserWithRoleUsingE2E(ctx, constant.RoleTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, teacher.UserID)

	if len(stepState.TeacherIDs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no teachers were created")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserWithRoleUsingE2E(ctx context.Context, role string) (*common.AuthInfo, error) {
	ctx = common.ValidContext(ctx, constants.TestingSchool, s.RootAccount[constants.TestingSchool].UserID, s.RootAccount[constants.TestingSchool].Token)

	roleWithLocation := usermgmt.RoleWithLocation{}
	roleWithLocation.RoleName = role
	roleWithLocation.LocationIDs = []string{constants.E2EOrgLocation}

	authInfo, err := usermgmt.SignIn(ctx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.CommonSuite.StepState.FirebaseAddress, s.Connections.UserMgmtConn, roleWithLocation, []string{constants.E2EOrgLocation})
	if err != nil {
		return nil, err
	}

	return &authInfo, nil
}

func (s *suite) anExistingSetOfNumberLessonsWithWait(ctx context.Context, count string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonsCount, err := strconv.Atoi(count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert string (%s) to integer err: %w", count, err)
	}

	// sleep to make sure NATS sync data successfully from bob to lessonmgmt data
	time.Sleep(5 * time.Second)

	now := time.Now()
	lessonStatus := int32(0)
	for i := 0; i <= lessonsCount; i++ {
		lessonTime := now.Add(time.Duration(i) * 24 * time.Hour)

		req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(StepStateToContext(ctx, stepState), cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE)
		req.StartTime = timestamppb.New(lessonTime)
		req.EndTime = timestamppb.New(lessonTime.Add(30 * time.Minute))
		req.SchedulingStatus = lpb.LessonStatus(lpb.LessonStatus_value[lpb.LessonStatus_name[lessonStatus]])

		ctx, err := s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(StepStateToContext(ctx, stepState), req)
		stepState = StepStateFromContext(ctx)
		if err != nil || stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create lesson, err: %w | response err: %w", err, stepState.ResponseErr)
		}

		// update lesson end at
		if i%4 == 0 {
			if err := s.modifyLessonAddEndAt(ctx, stepState.CurrentLessonID); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}

		// increment lesson status selection
		if lessonStatus == 3 {
			lessonStatus = 0
		} else {
			lessonStatus++
		}
	}

	time.Sleep(5 * time.Second)

	return StepStateToContext(ctx, stepState), nil
}
