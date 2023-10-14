package auth

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/gcp"

	"github.com/pkg/errors"
)

const TestFirebaseAuthTenantID = ""

var (
	srcFirebaseCredentialsFile         string
	srcFirebaseProjectID               string
	srcIdentityPlatformCredentialFile  string
	srcIdentityPlatformProjectID       string
	srcIdentityPlatformTenantID        string
	destIdentityPlatformCredentialFile string
	destIdentityPlatformProjectID      string
	destIdentityPlatformTenantID       string
	exportReport                       bool
	test                               bool
)

func importBetweenTenants(ctx context.Context, srcTenantClient multitenant.TenantClient, destTenantClient multitenant.TenantClient) error {
	var successCount int
	var failureCount int

	now := time.Now()
	fmt.Println("Progress:")

	var successReportWriter, failureReportWriter *csv.Writer
	if exportReport {
		successReport, err := os.Create(fmt.Sprintf("%v_success.csv", now.Format("2006-01-02_15-04-05")))
		if err != nil {
			log.Fatalln("failed to open file", err)
		}
		successReportWriter = csv.NewWriter(successReport)
		defer func() {
			successReportWriter.Flush()
			_ = successReport.Close()
		}()

		failureReport, err := os.Create(fmt.Sprintf("%v_failure.csv", now.Format("2006-01-02_15-04-05")))
		if err != nil {
			log.Fatalln("failed to open file", err)
		}
		defer func() {
			_ = failureReport.Close()
		}()
		failureReportWriter = csv.NewWriter(failureReport)
		defer failureReportWriter.Flush()
	}

	pager := srcTenantClient.UserPager(ctx, "", 1000)
	for {
		users, nextPageToken, err := pager.NextPage()
		if err != nil {
			return err
		}
		if nextPageToken == "" {
			break
		}

		result, err := destTenantClient.ImportUsers(ctx, users, srcTenantClient.GetHashConfig())
		if err != nil {
			fmt.Println(nextPageToken)
			return err
		}

		if exportReport {
			for _, user := range result.UsersSuccessImport {
				if err := successReportWriter.Write([]string{destTenantClient.TenantID(), user.GetUID()}); err != nil {
					log.Fatalln("error writing record to file", err)
				}
			}
			for _, userFailedToImport := range result.UsersFailedToImport {
				if err := failureReportWriter.Write([]string{userFailedToImport.User.GetUID(), userFailedToImport.Err}); err != nil {
					log.Fatalln("error writing record to file", err)
				}
			}
		}

		fmt.Println(fmt.Sprintf("successfully imported %v user(s)", len(result.UsersSuccessImport)))

		if len(result.UsersFailedToImport) > 0 {
			fmt.Println(fmt.Sprintf("failed to import %v user(s)", len(result.UsersFailedToImport)))
		}

		successCount += len(result.UsersSuccessImport)
		failureCount += len(result.UsersFailedToImport)
	}

	fmt.Println("-----------------------------------------------------------")
	fmt.Println("Result:")
	fmt.Println(fmt.Sprintf("Total %v user(s) successfully imported", successCount))
	fmt.Println(fmt.Sprintf("Total %v user(s) fail imported", failureCount))

	return nil
}

func RunImportUsersFromFirebaseToIdentityPlatform(ctx context.Context) error {
	switch {
	case srcFirebaseCredentialsFile == "":
		return errors.New("source firebase credential file location is empty")
	case destIdentityPlatformCredentialFile == "":
		return errors.New("destination identity platform credential file location is empty")
	case destIdentityPlatformTenantID == "":
		return errors.New("destination identity platform tenant id is empty")
	}

	srcApp, err := gcp.NewApp(ctx, srcFirebaseCredentialsFile, srcFirebaseProjectID)
	if err != nil {
		return errors.Wrap(err, "gcp.NewApp")
	}
	srcFirebaseAuthClient, err := multitenant.NewFirebaseAuthClientFromGCP(ctx, srcApp)
	if err != nil {
		return errors.Wrap(err, "srcApp.Auth")
	}
	if test {
		srcTenantManager, err := multitenant.NewTenantManagerFromGCP(ctx, srcApp)
		if err != nil {
			return errors.Wrap(err, "auth.NewTenantManagerFromGCP")
		}
		srcFirebaseAuthClient, err = srcTenantManager.TenantClient(ctx, TestFirebaseAuthTenantID)
		if err != nil {
			return errors.Wrap(err, "srcApp.Auth")
		}
	}

	destApp, err := gcp.NewApp(ctx, destIdentityPlatformCredentialFile, destIdentityPlatformProjectID)
	if err != nil {
		return errors.Wrap(err, "gcp.NewApp")
	}
	destTenantManager, err := multitenant.NewTenantManagerFromGCP(ctx, destApp)
	if err != nil {
		return errors.Wrap(err, "auth.NewTenantManager")
	}
	destTenantClient, err := destTenantManager.TenantClient(ctx, destIdentityPlatformTenantID)
	if err != nil {
		return errors.Wrap(err, "destTenantManager.TenantClient")
	}

	if err := importBetweenTenants(ctx, srcFirebaseAuthClient, destTenantClient); err != nil {
		return err
	}

	return nil
}

func RunImportUsersBetweenTenants(ctx context.Context) error {
	switch {
	case srcIdentityPlatformCredentialFile == "":
		return errors.New("source identity platform credential file location is empty")
	case srcIdentityPlatformTenantID == "":
		return errors.New("source identity platform tenant id is empty")
	case destIdentityPlatformCredentialFile == "":
		return errors.New("destination identity platform credential file location is empty")
	case destIdentityPlatformTenantID == "":
		return errors.New("destination identity platform tenant id is empty")
	}

	srcApp, err := gcp.NewApp(ctx, srcIdentityPlatformCredentialFile, srcIdentityPlatformProjectID)
	if err != nil {
		return errors.Wrap(err, "gcp.NewApp")
	}
	srcTenantManager, err := multitenant.NewTenantManagerFromGCP(ctx, srcApp)
	if err != nil {
		return errors.Wrap(err, "auth.NewTenantManager")
	}
	srcTenantClient, err := srcTenantManager.TenantClient(ctx, srcIdentityPlatformTenantID)
	if err != nil {
		return errors.Wrap(err, "destTenantManager.TenantClient")
	}

	destApp, err := gcp.NewApp(ctx, destIdentityPlatformCredentialFile, destIdentityPlatformProjectID)
	if err != nil {
		return errors.Wrap(err, "gcp.NewApp")
	}
	destTenantManager, err := multitenant.NewTenantManagerFromGCP(ctx, destApp)
	if err != nil {
		return errors.Wrap(err, "auth.NewTenantManager")
	}
	destTenantClient, err := destTenantManager.TenantClient(ctx, destIdentityPlatformTenantID)
	if err != nil {
		return errors.Wrap(err, "destTenantManager.TenantClient")
	}

	if err := importBetweenTenants(ctx, srcTenantClient, destTenantClient); err != nil {
		return err
	}

	return nil
}
