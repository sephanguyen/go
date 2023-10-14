package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

var (
	studentIDsScanQuery = `SELECT uap.user_id
	FROM user_access_paths uap
	INNER JOIN locations l ON uap.location_id = l.location_id 
	INNER JOIN location_types lt ON l.location_type = lt.location_type_id 
	WHERE lt.name = $1 AND uap.deleted_at IS NULL
	LIMIT $2;
	`
	limit = 100
)

func (s *suite) studentsWithLocationTypeOrgInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < 5; i++ {
		if ctx, err := s.createStudentWithDefaultLocation(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToMigrateDeleteStudentLocationOrgExistedInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	usermgmt.RunMigrateDeleteStudentLocationOrg(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existingStudentsHaveDefaultLocationAreRemovedLocationTypeOrg(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rows, err := s.BobPostgresDBTrace.Query(
		ctx,
		studentIDsScanQuery,
		domain.DefaultLocationType,
		limit,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("existingStudentsHaveDefaultLocationAreRemovedLocationTypeOrg: query error %s", err.Error())
	}
	defer rows.Close()

	studentIDs := []string{}

	for rows.Next() {
		studentID := ""
		if err := rows.Scan(&studentID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		studentIDs = append(studentIDs, studentID)
	}

	if len(studentIDs) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("migrate delete student location org fail")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentWithDefaultLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.onlyStudentInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	req.StudentProfile.LocationIds = []string{constants.ManabieOrgLocation}

	if ctx, err := s.createNewStudentAccount(ctx, schoolAdminType); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
