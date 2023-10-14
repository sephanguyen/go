package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	UsermgmtName                                       = "Usermgmt Schedule Job"
	featureToggleAutoDeactivateAndReactivateStudentsV2 = "User_StudentManagement_DeactivateStudent_V2"
)

func init() {
	bootstrap.RegisterJob("usermgmt_cronjob_check_enrollment_status_date", RunCronJobCheckEnrollmentStatusEndDate).
		Desc("Cron job check enrollment status date")
}

func RunCronJobCheckEnrollmentStatusEndDate(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zLogger := rsc.Logger()
	zLogger.Sugar().Info("-----Migration remove access path of student for location-----")

	dbPool := rsc.DBWith("bob")
	unleashClient := rsc.WithUnleashC(&c.UnleashClientConfig).Unleash()

	orgQuery := "SELECT organization_id, name FROM organizations" //nolint:goconst
	organizations, err := dbPool.Query(ctx, orgQuery)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %s", err)
	}
	defer organizations.Close()

	for organizations.Next() {
		var organizationID, name pgtype.Text

		if err := organizations.Scan(&organizationID, &name); err != nil {
			zLogger.Sugar().Infof("failed to scan an orgs row: %s", err)
			continue
		}

		userIDs, err := GetUsermgmtUserIDByOrgID(ctx, dbPool, organizationID.String)
		if err != nil {
			zLogger.Sugar().Errorf("failed to get usermgmt user id: %s", err)
			continue
		}
		if len(userIDs) == 0 {
			zLogger.Sugar().Errorf("cannot find userID on organization: %s", organizationID.String)
			continue
		}
		internalUserID := userIDs[0]
		tenantWithInternalUserContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: organizationID.String,
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
				UserID:       internalUserID,
			},
		})

		err = migrateSetAccessPathForStudent(tenantWithInternalUserContext, dbPool, organizationID.String)
		if err != nil {
			zLogger.Sugar().Errorf("remove access path for student outdate enrollment status error: %s, org: %s", zap.Error(err), organizationID.String)
		}
		zLogger.Sugar().Info("-----Migration of deactivate and reactivate student based on enrollment status history-----")

		enrollmentStatusHistoryRepo := &repository.DomainEnrollmentStatusHistoryRepo{}
		currentCursor := ""
		for {
			studentIDs, err := GetStudentsToCheckActivation(tenantWithInternalUserContext, dbPool, currentCursor)
			if err != nil {
				zLogger.Sugar().Errorf("get students to check activation error: %s, org: %s", zap.Error(err), organizationID.String)
				break
			}
			if len(studentIDs) == 0 {
				break
			}
			currentCursor = studentIDs[len(studentIDs)-1]
			enrollmentStatuses := []string{
				upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String(),
			}
			isEnableAutoDeactivateAndReactivateStudentsV2, err := unleashClient.IsFeatureEnabledOnOrganization(featureToggleAutoDeactivateAndReactivateStudentsV2, c.Common.Environment, organizationID.String)
			if err != nil {
				zLogger.Sugar().Errorf("failed to get feature toggle:%s, org: %s", zap.Error(err), organizationID.String)
				isEnableAutoDeactivateAndReactivateStudentsV2 = false
			}
			if isEnableAutoDeactivateAndReactivateStudentsV2 {
				enrollmentStatuses = []string{
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String(),
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
				}
			}
			if err := enrollmentStatusHistoryRepo.UpdateStudentStatusBasedEnrollmentStatus(tenantWithInternalUserContext, dbPool, studentIDs, enrollmentStatuses); err != nil {
				zLogger.Sugar().Errorf("deactivate and reactivate students error: %s, org: %s", zap.Error(err), organizationID.String)
				break
			}
		}
		zLogger.Sugar().Infof("Done migration for %s. Migrate success", name)
	}
	return nil
}

func migrateSetAccessPathForStudent(ctx context.Context, db database.QueryExecer, organizationID string) error {
	DomainEnrollmentStatusHistories, err := (&repository.DomainEnrollmentStatusHistoryRepo{}).GetOutDateEnrollmentStatus(ctx, db, organizationID)
	if err != nil {
		return errorx.ToStatusError(err)
	}

	if len(DomainEnrollmentStatusHistories) == 0 {
		return nil
	}

	mapStudentAssignLocations := make(map[string][]string)
	for _, enrollmentStatusHistory := range DomainEnrollmentStatusHistories {
		currentEnrollment, err := (&repository.DomainEnrollmentStatusHistoryRepo{}).GetByStudentIDAndLocationID(ctx, db, enrollmentStatusHistory.UserID().String(), enrollmentStatusHistory.LocationID().String(), true)
		if err != nil {
			return errorx.ToStatusError(err)
		}
		if len(currentEnrollment) != 0 {
			continue
		}
		if _, ok := mapStudentAssignLocations[enrollmentStatusHistory.UserID().String()]; ok {
			mapStudentAssignLocations[enrollmentStatusHistory.UserID().String()] = append(mapStudentAssignLocations[enrollmentStatusHistory.UserID().String()], enrollmentStatusHistory.LocationID().String())
		} else {
			mapStudentAssignLocations[enrollmentStatusHistory.UserID().String()] = []string{enrollmentStatusHistory.LocationID().String()}
		}
	}

	userAccessPathRepo := &repository.DomainUserAccessPathRepo{}

	for studentID, locations := range mapStudentAssignLocations {
		err := userAccessPathRepo.SoftDeleteByUserIDAndLocationIDs(ctx, db, studentID, organizationID, locations)
		if err != nil {
			return errorx.ToStatusError(err)
		}
	}

	return nil
}

func GetUsermgmtUserIDByOrgID(ctx context.Context, db database.QueryExecer, orgID string) ([]string, error) {
	tenantWithInternalUserContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: orgID,
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserID:       "",
		},
	})
	query := `SELECT user_id FROM user_basic_info WHERE name = $1 AND resource_path = $2 AND deleted_at IS NULL`
	rows, err := db.Query(
		tenantWithInternalUserContext,
		query,
		UsermgmtName,
		orgID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string

		err = rows.Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func GetStudentsToCheckActivation(ctx context.Context, db database.QueryExecer, currentCursor string) ([]string, error) {
	query := fmt.Sprintf(`SELECT student_id from student_enrollment_status_history 
		WHERE start_date BETWEEN (NOW() - INTERVAL '2 day') AND NOW() AND deleted_at IS NULL AND student_id > '%s'
	 	GROUP BY student_id ORDER BY student_id LIMIT 1000`,
		currentCursor)
	rows, err := db.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	var studentIDs []string
	for rows.Next() {
		var studentID string
		err := rows.Scan(&studentID)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		studentIDs = append(studentIDs, studentID)
	}

	return studentIDs, nil
}
