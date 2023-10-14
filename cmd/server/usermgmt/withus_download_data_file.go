package usermgmt

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt/withus"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	userInterceptor "github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("usermgmt_withus_download_data_file", RunWithusDownloadDataFile).
		Desc("Cmd to download from data file withus")
	bootstrap.RegisterJob("usermgmt_itee_download_data_file", RunIteeDownloadDataFile).
		Desc("Cmd to download from data file itee")
}

func RunWithusDownloadDataFile(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	return runWithusDownloadDataFile(ctx, c, rsc, []string{fmt.Sprint(constants.ManagaraBase)})
}

func RunIteeDownloadDataFile(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	return runWithusDownloadDataFile(ctx, c, rsc, []string{fmt.Sprint(constants.ManagaraHighSchool)})
}

func runWithusDownloadDataFile(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources, orgIDs []string) error {
	zapLogger := rsc.Logger()
	sugarLogger := zapLogger.Sugar()
	ctx = ctxzap.ToContext(ctx, zapLogger)

	dbTrace := rsc.DBWith("bob")

	shamirConnection := rsc.GRPCDial("shamir")

	orgQuery := "SELECT organization_id, name FROM organizations WHERE organization_id = ANY($1)"
	orgIDTextArray := database.TextArray(orgIDs)

	rows, err := dbTrace.Query(ctx, orgQuery, &orgIDTextArray)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var orgID, orgName pgtype.Text

		if err := rows.Scan(&orgID, &orgName); err != nil {
			sugarLogger.Infof("failed to scan an orgs row: %s", err)
			continue
		}

		userIDs, err := GetUsermgmtUserIDByOrgID(ctx, dbTrace, orgID.String)
		if err != nil {
			zLogger.Sugar().Errorf("failed to get usermgmt user id: %s", err)
			continue
		}
		if len(userIDs) == 0 {
			zLogger.Sugar().Errorf("cannot find userID on organization: %s", orgID.String)
			continue
		}
		internalUserID := userIDs[0]

		var email, password string
		for _, account := range c.JobAccounts {
			if orgID.String == account.OrganizationID {
				email = account.Email
				password = account.Password
				break
			}
		}

		tenantID, err := new(repository.OrganizationRepo).GetTenantIDByOrgID(ctx, dbTrace, orgID.String)
		if err != nil {
			sugarLogger.Errorf("cannot GetTenantIDByOrgID :%s, err :%s", orgID.String, err.Error())
			continue
		}

		idToken, err := userInterceptor.LoginInAuthPlatform(ctx, c.FirebaseAPIKey, tenantID, email, password)
		if err != nil {
			sugarLogger.Errorf("cannot LoginInAuthPlatform :%s, err :%s", orgID.String, err.Error())
			continue
		}

		exchangedToken, err := userInterceptor.ExchangeToken(ctx, shamirConnection, c.JWTApplicant, internalUserID, idToken)
		if err != nil {
			sugarLogger.Errorf("cannot ExchangeToken :%s, err: %s", internalUserID, err.Error())
			continue
		}

		ctx = userInterceptor.GRPCContext(ctx, "token", exchangedToken)
		ctx = interceptors.ContextWithUserID(ctx, internalUserID)
		ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: orgID.String,
				UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
				UserID:       internalUserID,
			},
		})

		studentService, err := withus.NewStudentService(ctx, &c, rsc)
		if err != nil {
			return fmt.Errorf("failed to init student service: %s", err.Error())
		}

		tokyoLocation, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return fmt.Errorf("failed to init location: %s", err)
		}

		fileName := managaraFileName(orgID.String, time.Now().In(tokyoLocation))
		zapLogger.Info(fmt.Sprintf("-----Sync student from %s-----", orgName.String),
			zap.String("FileName", fileName),
		)

		listOfErrors := withus.ImportManagaraStudents(ctx, c.WithUsConfig.BucketName, fileName, studentService)
		if len(listOfErrors) != 0 {
			if err := withus.NotifySyncDataStatus(studentService.SlackClient, c, orgID.String, orgName.String, constant.StatusFailed); err != nil {
				zapLogger.Warn(errors.Wrap(err, "Send alert to slack failed").Error())
			}
			slackClientWithusWorkspace := &alert.SlackImpl{
				WebHookURL: c.WithUsConfig.WithusWebhookURL,
				HTTPClient: http.Client{Timeout: time.Duration(10) * time.Second},
			}
			if err := withus.NotifyWithusSyncDataStatus(slackClientWithusWorkspace, c, orgName.String, constant.StatusFailed, listOfErrors); err != nil {
				zapLogger.Warn(errors.Wrap(err, "Send alert to withus workspace failed").Error())
			}
			errors := []error{}
			for _, err := range listOfErrors {
				errors = append(errors, err)
			}

			zapLogger.Error(
				fmt.Sprintf("ImportManagaraStudents failed in %s, file name: %s", orgName.String, fileName),
				zap.Errors("list of errors", errors),
			)
			continue
		}

		zapLogger.Info(fmt.Sprintf("-----Sync student success %s, file name: %s-----", orgName.String, fileName))
		if err := withus.NotifySyncDataStatus(studentService.SlackClient, c, orgID.String, orgName.String, constant.StatusSuccess); err != nil {
			zapLogger.Warn(errors.Wrap(err, "Send alert to slack failed").Error())
		}
	}
	return nil
}

func managaraFileName(orgID string, uploadDate time.Time) string {
	switch orgID {
	case fmt.Sprint(constants.ManagaraBase):
		return fmt.Sprintf("/withus/W2-D6L_users%s.tsv", withus.DataFileNameSuffix(uploadDate))
	case fmt.Sprint(constants.ManagaraHighSchool):
		return fmt.Sprintf("/itee/N1-M1_users%s.tsv", withus.DataFileNameSuffix(uploadDate))
	default:
		return ""
	}
}
