package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	pbu "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const (
	countEnrollmentStatusOutDateQuery = `SELECT COUNT(*) 
			FROM user_access_paths 
                WHERE user_id = $1 
                  AND location_id = $2 
                  AND deleted_at IS NOT NULL`
)

func (s *suite) enrollmentStatusOutdateInOurSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	userIDs, err := usermgmt.GetUsermgmtUserIDByOrgID(ctx, s.BobDBTrace, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get usermgmt account: %s", err)
	}
	if len(userIDs) == 0 {
		return StepStateToContext(ctx, stepState), errors.Errorf("no account founded with org:%d", constants.ManabieSchool)
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserID:       userIDs[0],
		},
	})

	studentID := idutil.ULIDNow()
	locationID := s.ExistingLocations[0].LocationID.String // stepState.ExistingLocations = {0:manabie, 1:jprep, 2:manabie}
	ctx, err = s.aValidStudentInDB(ctx, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	orderEventLog := service.OrderEventLog{
		OrderStatus:      pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
		OrderType:        pb.OrderType_ORDER_TYPE_NEW.String(),
		StudentID:        studentID,
		LocationID:       locationID,
		EnrollmentStatus: pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
		StartDate:        time.Now().Add(-240 * time.Hour),
		EndDate:          time.Now().Add(-12 * time.Hour),
	}

	stepState.Request = orderEventLog

	orderLog := service.NewOrderLogRequest(&orderEventLog)

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	enrollmentStatusHistory := entity.EnrollmentStatusHistoryWillBeDelegated{
		EnrollmentStatusHistory: orderLog,
		HasUserID:               orderLog,
		HasLocationID:           orderLog,
		HasOrganizationID:       organization,
	}

	err = (&repository.DomainEnrollmentStatusHistoryRepo{}).Create(ctx, s.BobDBTrace, enrollmentStatusHistory)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = s.validUserAccessPath(ctx, studentID, locationID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validUserAccessPath(ctx context.Context, userID, locationID string) error {
	userAccessPathRepo := &repository.UserAccessPathRepo{}
	userAccessPathEnt := &entity.UserAccessPath{}
	database.AllNullEntity(userAccessPathEnt)

	if err := multierr.Combine(
		userAccessPathEnt.UserID.Set(userID),
		userAccessPathEnt.LocationID.Set(locationID),
	); err != nil {
		return err
	}

	if err := userAccessPathRepo.Upsert(ctx, s.BobDBTrace, []*entity.UserAccessPath{userAccessPathEnt}); err != nil {
		return errors.Wrap(err, "userAccessPathRepo.Upsert")
	}

	return nil
}

func (s *suite) systemRunJobToDisableAccessPathLocationForOutdateEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsc := bootstrap.NewResources().WithLoggerC(&s.Cfg.Common).WithDatabaseC(ctx, s.Cfg.PostgresV2.Databases)
	defer rsc.Cleanup() //nolint:errcheck

	err := usermgmt.RunCronJobCheckEnrollmentStatusEndDate(ctx, configurations.Config{Common: s.Cfg.Common, PostgresV2: s.Cfg.PostgresV2, UnleashClientConfig: s.Cfg.UnleashClientConfig}, rsc)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("usermgmt.RunCronJobCheckEnrollmentStatusEndDate: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentNoLongerAccessLocationWhenAccessPathRemoved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	orderEventLog := stepState.Request.(service.OrderEventLog)

	count := database.Int8(0)
	err := s.BobPostgresDBTrace.QueryRow(ctx, countEnrollmentStatusOutDateQuery,
		database.Text(orderEventLog.StudentID),
		database.Text(orderEventLog.LocationID)).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentNoLongerAccessLocationWhenAccessPathRemoved: query error %s", err.Error())
	}

	if count.Int == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("migrate update student enrollment status fail")
	}

	return StepStateToContext(ctx, stepState), nil
}
