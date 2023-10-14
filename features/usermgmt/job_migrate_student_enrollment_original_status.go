package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

const (
	countUpdatedEnrollmentStatusQuery = `
	SELECT
	    count(*)
	FROM
	    public.students 
	WHERE
	    enrollment_status = $1
	    AND resource_path = $2
	`
	ManabieSchool             = "MANABIE_SCHOOL"
	ManabieSchoolResourcePath = "-2147483648"
)

func (s *suite) studentsHaveEnrollmentStatusAreUpdatedToWith(ctx context.Context, originStatus, newStatus, resourcePath string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if resourcePath == ManabieSchool {
		resourcePath = ManabieSchoolResourcePath
	}

	count := database.Int8(0)
	err := s.BobPostgresDBTrace.QueryRow(ctx, countUpdatedEnrollmentStatusQuery, originStatus, resourcePath).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentsHaveEnrollmentStatusAreUpdatedToWithdrawn: query error %s", err.Error())
	}

	if count.Int > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("migrate update student enrollment status fail")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsWithEnrollmentOriginalStatusInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < 5; i++ {
		if ctx, err := s.createStudentWithQuitEnrollmentStatus(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToMigrateUpdateStudentEnrollmentStatusToInOurSystem(ctx context.Context, originStatus, newStatus, resourcePath string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if resourcePath == ManabieSchool {
		resourcePath = ManabieSchoolResourcePath
	}

	usermgmt.RunMigrateStudentEnrollmentOriginalStatus(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, newStatus, originStatus, resourcePath)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentWithQuitEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.onlyStudentInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	req.StudentProfile.LocationIds = []string{constants.ManabieOrgLocation}
	req.StudentProfile.EnrollmentStatus = pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA

	if ctx, err := s.createNewStudentAccount(ctx, schoolAdminType); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
