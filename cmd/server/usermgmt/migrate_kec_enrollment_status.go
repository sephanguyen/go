package usermgmt

import (
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"cloud.google.com/go/storage"
	"github.com/gocarina/gocsv"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

var (
	organizationID string
	bucketName     string
	objectName     string
	deleteType     string
)

func init() {
	bootstrap.RegisterJob("usermgmt_migrate_kec_enrollment_status", runMigrateKecEnrollmentStatus).
		Desc("Cmd migrate kec enrollment status").
		StringVar(&organizationID, "organizationID", "", "Migrate Locations For Users With organizationID").
		StringVar(&bucketName, "bucketName", "", "Bucket name migrate enrollment status").
		StringVar(&objectName, "objectName", "", "object name migrate enrollment status").
		StringVar(&deleteType, "deleteType", "soft", "delete type for enrollment status data")
}

const (
	JpTimeZone    = "Asia/Tokyo"
	KecFormatDate = "2006/01/02 15:04:05"
)

type KecData struct {
	StudentID        field.String `csv:"student_id"`
	LocationID       field.String `csv:"location"`
	EnrollmentStatus field.String `csv:"enrollment_status"`
	StartDate        field.String `csv:"start_date"`
	EndDate          field.String `csv:"end_date"`
}

type KecDatas []*KecData

func runMigrateKecEnrollmentStatus(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	return RunMigrateKecEnrollmentStatus(ctx, c, rsc, organizationID, bucketName, objectName)
}

func RunMigrateKecEnrollmentStatus(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources, organizationID, bucketName, objectName string) error {
	zapLogger := rsc.Logger()
	ctx = ctxzap.ToContext(ctx, zapLogger)

	dbPool := rsc.DBWith("bob")
	var data KecDatas
	err := getKecData(ctx, bucketName, objectName, &data)
	if err != nil {
		return fmt.Errorf("cannot get kec data: %s", err)
	}

	userIDs, err := GetUsermgmtUserIDByOrgID(ctx, dbPool, organizationID)
	if err != nil {
		zLogger.Sugar().Errorf("failed to get usermgmt user id: %s", err)
		return fmt.Errorf("failed to get usermgmt user id: %s", err)
	}
	if len(userIDs) == 0 {
		zLogger.Sugar().Errorf("cannot find userID on organization: %s", organizationID)
		return fmt.Errorf("failed to get usermgmt user id: %s", err)
	}
	internalUserID := userIDs[0]

	userRepo := repository.DomainUserRepo{}
	ctx = interceptors.ContextWithUserID(ctx, internalUserID)
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: organizationID,
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserID:       internalUserID,
		},
	})
	location, err := time.LoadLocation(JpTimeZone)
	if err != nil {
		return err
	}
	stmt := `UPDATE student_enrollment_status_history SET deleted_at = now() WHERE student_id = $1 AND location_id = $2;`
	if deleteType == "hard" {
		stmt = `DELETE FROM student_enrollment_status_history WHERE student_id = $1 AND location_id = $2;`
	}
	batchSize := 1000
	if err := database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		for i := 0; i < len(data); i += batchSize {
			end := i + batchSize
			if end > len(data) {
				end = len(data)
			}
			chunkedData := data[i:end]
			externalUserIDs := make([]string, 0, len(chunkedData))
			for _, row := range chunkedData {
				externalUserIDs = append(externalUserIDs, row.StudentID.String())
			}
			users, err := userRepo.GetByExternalUserIDs(ctx, tx, externalUserIDs)
			if err != nil {
				return err
			}
			userIDByUserExternalID := newMapToGetUserIDByUserExternalID(users, externalUserIDs)
			for _, row := range chunkedData {
				if userIDByUserExternalID[row.StudentID.String()] == "" {
					continue
				}
				_, err = tx.Exec(ctx, stmt, userIDByUserExternalID[row.StudentID.String()], row.LocationID.String())
				if err != nil {
					return err
				}
			}
		}

		for i := 0; i < len(data); i += batchSize {
			end := i + batchSize
			if end > len(data) {
				end = len(data)
			}
			chunkedData := data[i:end]
			externalUserIDs := make([]string, 0, len(chunkedData))
			for _, row := range chunkedData {
				externalUserIDs = append(externalUserIDs, row.StudentID.String())
			}
			users, err := userRepo.GetByExternalUserIDs(ctx, tx, externalUserIDs)
			if err != nil {
				return err
			}
			userIDByUserExternalID := newMapToGetUserIDByUserExternalID(users, externalUserIDs)
			for _, row := range chunkedData {
				if userIDByUserExternalID[row.StudentID.String()] == "" {
					continue
				}
				startDate, err := time.Parse(KecFormatDate, row.StartDate.String())
				if err != nil {
					return err
				}
				startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), startDate.Hour(), startDate.Minute(), startDate.Second(), 0, location)
				var endDate *time.Time
				if field.IsPresent(row.EndDate) {
					endDateVal, err := time.Parse(KecFormatDate, row.EndDate.String())
					if err != nil {
						return err
					}
					endDateVal = time.Date(endDateVal.Year(), endDateVal.Month(), endDateVal.Day(), endDateVal.Hour(), endDateVal.Minute(), endDateVal.Second(), 0, location)
					endDate = &endDateVal
				}
				enrollmentStatusQuery := `
					INSERT INTO student_enrollment_status_history (student_id, location_id, enrollment_status, start_date, end_date, resource_path)
					 VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT ON CONSTRAINT pk__student_enrollment_status_history DO UPDATE SET deleted_at = NULL, end_date = EXCLUDED.end_date;`
				_, err = tx.Exec(ctx, enrollmentStatusQuery, userIDByUserExternalID[row.StudentID.String()], row.LocationID.String(), row.EnrollmentStatus.String(), startDate, endDate, organizationID)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("insert kec data error : %s", err)
	}
	return nil
}

func newMapToGetUserIDByUserExternalID(users entity.Users, externalUserIDs []string) map[string]string {
	result := make(map[string]string, len(externalUserIDs))
	for _, externalUserID := range externalUserIDs {
		for _, user := range users {
			if externalUserID == user.ExternalUserID().String() {
				userID := user.UserID().String()
				result[externalUserID] = userID
				break
			}
		}
	}
	return result
}

func getKecData(ctx context.Context, bucketName, objectName string, dest interface{}) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()
	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()
	reader := csv.NewReader(rc)
	if err = gocsv.UnmarshalCSV(reader, dest); err != nil {
		return errors.Wrap(err, "gocsv.UnmarshalCSV error")
	}
	return nil
}
