package usermgmt

import (
	"context"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt/withus"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc/importstudent"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"cloud.google.com/go/storage"
	"github.com/gocarina/gocsv"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	groupSize     int
	sleepDuration int
)

type CSVError struct {
	CVSRow      int    `csv:"csv_row"`
	Field       string `csv:"field"`
	RawError    string `csv:"raw_error"`
	DomainError string `csv:"domain_error"`
	Code        int    `csv:"code"`
}

func init() {
	bootstrap.RegisterJob("usermgmt_migrate_bulk_insert_students", runMigrateBulkInsertStudents).
		Desc("Cmd usermgmt migrate bulk insert students").
		StringVar(&organizationID, "organizationID", "", "Migrate Locations For Users With organizationID").
		StringVar(&bucketName, "bucketName", "", "Bucket name migrate bulk insert students").
		StringVar(&objectName, "objectName", "", "object name migrate bulk insert students").
		IntVar(&groupSize, "groupSize", 250, "groupSize migrate bulk insert students").
		IntVar(&sleepDuration, "sleepDuration", 10, "sleepDuration (seconds) migrate bulk insert students")
}

func runMigrateBulkInsertStudents(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	return RunMigrateBulkInsertStudents(ctx, c, rsc, organizationID, bucketName, objectName, groupSize, sleepDuration)
}

func RunMigrateBulkInsertStudents(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources, organizationID, bucketName, objectName string, groupSize, sleepDuration int) error {
	zapLogger := rsc.Logger()
	ctx = ctxzap.ToContext(ctx, zapLogger)
	now := time.Now()
	zapLogger.Sugar().Info("-----START: Migrate Bulk Insert Students-----")
	connBob := rsc.DBWith("bob")

	userIDs, err := GetUsermgmtUserIDByOrgID(ctx, connBob, organizationID)

	if err != nil {
		zLogger.Sugar().Errorf("failed to get usermgmt user id: %s", err)
		return fmt.Errorf("failed to get usermgmt user id: %s", err)
	}
	if len(userIDs) == 0 {
		zLogger.Sugar().Errorf("cannot find userID on organization: %s", organizationID)
		return fmt.Errorf("failed to get usermgmt user id: %s", err)
	}
	internalUserID := userIDs[0]

	ctx = interceptors.ContextWithUserID(ctx, internalUserID)
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: organizationID,
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserID:       internalUserID,
		},
	})

	studentService, err := withus.NewStudentService(ctx, &c, rsc)
	if err != nil {
		return fmt.Errorf("failed to init student service: %s", err.Error())
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Close()
	}()

	importStudentCSVFields, err := getStudentsFormStorage(ctx, client, bucketName, objectName)
	if err != nil {
		return fmt.Errorf("error getStudentsFormStorage: %s", err.Error())
	}
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return fmt.Errorf("error interceptors.OrganizationFromContext: %s", err.Error())
	}

	option := unleash.DomainStudentFeatureOption{}
	option = studentService.FeatureManager.FeatureUsernameToStudentFeatureOption(ctx, organization, option)
	option = studentService.FeatureManager.FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx, organization, option)

	bulkSizeInsertStudents := len(importStudentCSVFields)
	numberGroup := (bulkSizeInsertStudents + groupSize - 1) / groupSize
	for i := 0; i < numberGroup; i++ {
		start := i * groupSize
		end := (i + 1) * groupSize
		if end > bulkSizeInsertStudents {
			end = bulkSizeInsertStudents
		}
		domainStudents := make(aggregate.DomainStudents, 0)
		collectionErrors := make([]error, 0)
		for idx, student := range importStudentCSVFields[start:end] {
			student.UserNameAttr = student.EmailAttr
			domainStudent, err := importstudent.ToDomainStudentsV2(student, idx+start, option.EnableUsername)
			if err != nil {
				collectionErrors = append(collectionErrors, err)
				continue
			}
			domainStudents = append(domainStudents, domainStudent)
		}
		respStudents, listOfErrors := studentService.UpsertMultipleWithErrorCollection(ctx, domainStudents, option)
		collectionErrors = append(collectionErrors, listOfErrors...)
		if len(collectionErrors) > 0 {
			name := filepath.Base(objectName)
			errorName := filepath.Join(filepath.Dir(objectName), fmt.Sprintf("[%d-%d] error-%s", start, end, name))
			if err := writeCSVToBucket(ctx, client, bucketName, errorName, collectionErrors); err != nil {
				return errors.Wrap(err, "failed to write csv to bucket")
			}

			zapLogger.Error(fmt.Sprintf("some students was failed when from [%d - %d] in group %d", start, end, numberGroup),
				zap.Errors("collectionErrors", collectionErrors),
			)
		}
		if len(respStudents) > 0 {
			zapLogger.Sugar().Infof("--inserted %d students [%d - %d]--", len(respStudents), start, end)
		}
		zapLogger.Sugar().Infof("---Start sleep %d seconds after run from [%d - %d]", sleepDuration, start, end)
		duration := time.Duration(sleepDuration) * time.Second
		time.Sleep(duration)
		zapLogger.Sugar().Infof("---End sleep %d seconds after run from [%d - %d]", sleepDuration, start, end)
	}
	zapLogger.Sugar().Infof("---Done: Migrate bulk insert students with end time: %d", time.Since(now).Milliseconds())
	return nil
}

func getStudentsFormStorage(ctx context.Context, client *storage.Client, bucketName, objectName string) ([]*importstudent.StudentCSV, error) {
	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	reader := csv.NewReader(rc)
	var studentImportData []*importstudent.StudentCSV
	if err = gocsv.UnmarshalCSV(reader, &studentImportData); err != nil {
		return nil, errors.Wrap(err, "gocsv.UnmarshalCSV error")
	}
	return studentImportData, nil
}

func writeCSVToBucket(ctx context.Context, client *storage.Client, bucketName, objectName string, collectionErrors []error) error {
	listOfCSVError := make([]*CSVError, 0)
	for _, err := range collectionErrors {
		switch e := err.(type) {
		case errcode.DomainError:
			index := grpc.GetIndexFromMessageError(e.DomainError())
			field := grpc.GetFieldFromMessageError(e.DomainError())
			csv := CSVError{
				CVSRow:      index + 2,
				Field:       field,
				RawError:    e.Error(),
				DomainError: e.DomainError(),
				Code:        e.DomainCode(),
			}
			listOfCSVError = append(listOfCSVError, &csv)
		default:
			csv := CSVError{
				RawError: e.Error(),
			}
			listOfCSVError = append(listOfCSVError, &csv)
		}
	}
	csvContent, err := gocsv.MarshalString(&listOfCSVError)
	if err != nil {
		return fmt.Errorf("error marshal csv: %s", err.Error())
	}

	obj := client.Bucket(bucketName).Object(objectName)
	writer := obj.NewWriter(ctx)
	if _, err := writer.Write([]byte(csvContent)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("error write csv to storage: %s", err.Error())
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("error closing writer: %s", err.Error())
	}
	return nil
}
