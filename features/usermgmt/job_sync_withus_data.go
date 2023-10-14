package usermgmt

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/cmd/server/usermgmt/withus"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"

	"cloud.google.com/go/storage"
	"github.com/gocarina/gocsv"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const (
	iteeDomain   = "itee"
	withusDomain = "withus"
)

type ManagaraBaseStudent struct {
	CustomerNumber    string `csv:"顧客番号"`
	StudentNumber     string `csv:"生徒番号"`
	Name              string `csv:"氏名（ニックネーム）"`
	PasswordRaw       string `csv:"パスワード"`
	StudentEmail      string `csv:"生徒メール"`
	ParentNumber      string `csv:"保護者番号"`
	ParentName        string `csv:"保護者氏名"`
	ParentRawPassword string `csv:"保護者パスワード"`
	ParentEmail       string `csv:"保護者メール"`
	Locations         string `csv:"G1（所属）"`
	TagG2             string `csv:"G2（セグメント）"`
	TagG3             string `csv:"G3（生徒区分）"`
	TagG4             string `csv:"G4（本校）"`
	TagG5             string `csv:"G5（学年）"`
	Courses           string `csv:"所持商品"`
	DeleteFlag        string `csv:"削除フラグ"`
}

type ManagaraHSStudent struct {
	CustomerNumber         string `csv:"顧客番号"`
	StudentNumber          string `csv:"生徒番号"`
	Name                   string `csv:"氏名"`
	PasswordRaw            string `csv:"パスワード"`
	StudentEmail           string `csv:"生徒メール"`
	ParentNumber           string `csv:"保護者番号"`
	ParentName             string `csv:"保護者氏名"`
	ParentRawPassword      string `csv:"保護者パスワード"`
	ParentEmail            string `csv:"保護者メール"`
	Locations              string `csv:"G1（所属）"`
	TagG2                  string `csv:"G2（コース）"`
	TagG3                  string `csv:"G3（高校生区分）"`
	TagG4                  string `csv:"G4（本校）"`
	TagG5                  string `csv:"G5（学年）"`
	Courses                string `csv:"所持商品"`
	DeleteFlag             string `csv:"削除フラグ"`
	GraduationExpectedDate string `csv:"卒業予定日"`
}

func (s *suite) tsvDataFileInBucket(ctx context.Context, domainName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	orgID := orgIDFromWithusDomainName(domainName)
	ctx = s.signedIn(ctx, orgID, UsermgmtScheduleJob)
	s.OrganizationID = fmt.Sprint(orgID)

	csvData, err := os.ReadFile(tsvFilePath(s.OrganizationID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("os.ReadFile err: %v", err)
	}

	csvWriter, objWriter, err := initManagaraWriter(ctx, s.OrganizationID, s.Cfg.WithUsConfig.BucketName)
	if err != nil {
		return ctx, err
	}
	switch domainName {
	case iteeDomain:
		err = uploadManagaraHighSchoolTSVData(csvWriter, csvData)
	case withusDomain:
		err = uploadManagaraBaseTSVData(csvWriter, csvData)
	}

	if err != nil {
		return ctx, errors.Wrap(err, "upload tsv data failed")
	}

	defer func() {
		_ = objWriter.Close()
		zapLogger.Info(
			"uploaded file",
			zap.String("destFilePath", objWriter.Name),
		)
	}()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToSyncDataFromBucket(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsc := bootstrap.NewResources().WithDatabaseC(ctx, s.Cfg.PostgresV2.Databases).WithLoggerC(&s.Cfg.Common).WithNATS(s.JSM)
	usermgmtCfg := configurations.Config{
		Common:              s.Cfg.Common,
		PostgresV2:          s.Cfg.PostgresV2,
		FirebaseAPIKey:      s.Cfg.FirebaseAPIKey,
		JWTApplicant:        s.Cfg.JWTApplicant,
		UnleashClientConfig: s.Cfg.UnleashClientConfig,

		JobAccounts:  s.Cfg.JobAccounts,
		WithUsConfig: s.Cfg.WithUsConfig,
	}
	switch s.OrganizationID {
	case fmt.Sprint(constants.ManagaraHighSchool):
		if err := usermgmt.RunIteeDownloadDataFile(ctx, usermgmtCfg, rsc); err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
	case fmt.Sprint(constants.ManagaraBase):
		if err := usermgmt.RunWithusDownloadDataFile(ctx, usermgmtCfg, rsc); err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) dataWereSyncedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	tokyoTime, err := timeInTokyo()
	if err != nil {
		return ctx, err
	}
	objectName := managaraFileName(s.OrganizationID, tokyoTime)

	fileReader, closeFileFunc, err := withus.GetStudentDataFromFile(ctx, s.Cfg.WithUsConfig.BucketName, objectName)
	if err != nil {
		return ctx, err
	}
	defer func() {
		_ = closeFileFunc()
	}()

	students, err := withus.ToManagaraStudents(ctx, fileReader)
	if err != nil {
		return ctx, err
	}
	rsc := bootstrap.NewResources().WithDatabaseC(ctx, s.Cfg.PostgresV2.Databases).WithLoggerC(&s.Cfg.Common).WithNATS(s.JSM)

	cfg := &configurations.Config{
		Common: configs.CommonConfig{
			FirebaseProject:         s.Cfg.Common.FirebaseProject,
			GoogleCloudProject:      s.Cfg.Common.GoogleCloudProject,
			IdentityPlatformProject: s.Cfg.Common.IdentityPlatformProject,
			Environment:             s.Cfg.Common.Environment,
		},
		UnleashClientConfig: s.Cfg.UnleashClientConfig,
	}

	service, err := withus.NewStudentService(ctx, cfg, rsc)
	if err != nil {
		return ctx, err
	}
	domainStudents, err := withus.ManagaraStudentToDomainStudent(ctx, students, service)
	if err != nil {
		return ctx, err
	}

	for i := range domainStudents {
		student := domainStudents[i]
		users, err := (&repository.DomainUserRepo{}).GetByExternalUserIDs(ctx, s.BobDBTrace, []string{student.ExternalUserID().String()})
		if err != nil {
			return ctx, err
		}
		if len(users) == 0 {
			return ctx, fmt.Errorf("can't find student with ExternalUserID: %s in our system", student.ExternalUserID().String())
		}
		user := users[0]

		if len(student.Parents) > 0 {
			parent := student.Parents[0]
			if field.IsPresent(parent.ExternalUserID()) {
				if err := s.validStudentParent(ctx, s.BobDBTrace, user.UserID().String(), parent); err != nil {
					return ctx, fmt.Errorf("validStudentParent failed: %s", err.Error())
				}
			}
		}
		if len(student.Courses) > 0 {
			if err := validStudentCourses(ctx, s.BobDBTrace, user.UserID().String(), student.Courses); err != nil {
				return ctx, fmt.Errorf("validStudentCourses failed: %s", err.Error())
			}
		}

		student.DomainStudent.DomainStudent = entity.StudentWillBeDelegated{
			DomainStudentProfile: student.DomainStudent,
			HasCountry:           student.DomainStudent,
			HasOrganizationID:    student.DomainStudent,
			HasSchoolID:          student.DomainStudent,
			HasGradeID:           student.DomainStudent,
			HasUserID:            user,
		}

		for idx, enrollmentStatusHistory := range student.DomainStudent.EnrollmentStatusHistories {
			enrollmentStatusHistoriesBD, err := (&repository.DomainEnrollmentStatusHistoryRepo{}).GetByStudentIDAndLocationID(ctx, s.BobDBTrace, user.UserID().String(), enrollmentStatusHistory.LocationID().String(), false)
			if err != nil {
				return ctx, err
			}
			if len(enrollmentStatusHistoriesBD) == 0 {
				return ctx, fmt.Errorf("can't find enrollmentStatusHistory %s in our system, enrollment_status: %s, studentID: %s", enrollmentStatusHistory.LocationID().String(), enrollmentStatusHistory.EnrollmentStatus().String(), user.UserID().String())
			}
			var enrollmentStatusHistoryBD entity.DomainEnrollmentStatusHistory

			for _, e := range enrollmentStatusHistoriesBD {
				if e.EnrollmentStatus() == enrollmentStatusHistory.EnrollmentStatus() {
					enrollmentStatusHistoryBD = e
					break
				}
			}
			student.DomainStudent.EnrollmentStatusHistories[idx] = entity.EnrollmentStatusHistoryWillBeDelegated{
				EnrollmentStatusHistory: enrollmentStatusHistoryBD,
				HasUserID:               user,
				HasLocationID:           enrollmentStatusHistoryBD,
				HasOrganizationID:       enrollmentStatusHistoryBD,
			}
		}
	}

	if _, err := s.verifyStudentsInBD(ctx, domainStudents.Students()); err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

// func (s *suite) tsvDataFileInBucketToUpdate(ctx context.Context, domainName string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	s.OrganizationID = fmt.Sprint(orgIDFromWithusDomainName(domainName))

// 	client, err := storage.NewClient(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	bucket := client.Bucket(s.Cfg.WithUsConfig.BucketName)

// 	tokyoTime, err := timeInTokyo()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// objectName := managaraFileName(fmt.Sprint(s.OrganizationID), tokyoTime)
// 	// objectReader, err := bucket.Object(managarafilePath(s.OrganizationID) + objectName).NewReader(ctx)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// defer objectReader.Close()

// 	// read tsv file test from bucket
// 	csvData, err := os.ReadFile(tsvFilePath(s.OrganizationID))
// 	if err != nil {
// 		return StepStateToContext(ctx, stepState), fmt.Errorf("os.ReadFile err: %v", err)
// 	}

// 	// upload file with correct name into bucket
// 	objectWriteName := managaraFileName(fmt.Sprint(s.OrganizationID), tokyoTime)
// 	objWriter := bucket.Object(objectWriteName).NewWriter(ctx)
// 	defer func() {
// 		_ = objWriter.Close()
// 		zapLogger.Info(
// 			"uploaded file",
// 			zap.String("destFilePath", objectWriteName),
// 		)
// 	}()

// 	// if _, err = io.Copy(objWriter, objectReader); err != nil {
// 	// 	return nil, errors.Wrap(err, "failed to write data to obj")
// 	// }
// 	// if err := objWriter.Close(); err != nil {
// 	// 	return nil, errors.Wrap(err, "failed to close obj write")
// 	// }

// 	UTF8writer := transform.NewWriter(objWriter, japanese.ShiftJIS.NewEncoder())
// 	_, err = UTF8writer.Write(csvData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return StepStateToContext(ctx, stepState), nil
// }

// func (s *suite) systemDoesNotHaveTsvDataFileInBucket(ctx context.Context, domainName string) (context.Context, error) {
// 	stepState := StepStateFromContext(ctx)
// 	s.OrganizationID = orgIDFromWithusDomainName(domainName)

// 	client, err := storage.NewClient(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	tokyoTime, err := timeInTokyo()
// 	if err != nil {
// 		return nil, err
// 	}

// 	bucket := client.Bucket(s.Cfg.WithUsConfig.BucketName)
// 	// delete current file if existed
// 	objectDeleteName := managaraFileName(s.OrganizationID, tokyoTime)
// 	if err := bucket.Object(objectDeleteName).Delete(ctx); err != nil {
// 		// skip error if object doesn't exist
// 		if err.Error() == "storage: object doesn't exist" {
// 			return StepStateToContext(ctx, stepState), nil
// 		}
// 		return nil, errors.Wrap(err, "failed to delete object")
// 	}

// 	return StepStateToContext(ctx, stepState), nil
// }

func uploadManagaraBaseTSVData(csvWriter *csv.Writer, data []byte) error {
	reader := bytes.NewReader(data)
	csvReader := csv.NewReader(reader)
	csvReader.Comma = '\t'

	students := []*ManagaraBaseStudent{}
	if err := gocsv.UnmarshalCSV(csvReader, &students); err != nil {
		return err
	}

	for _, v := range students {
		uid := idutil.ULIDNow()
		v.CustomerNumber = fmt.Sprintf("%s-%s", v.CustomerNumber, uid)
		v.StudentNumber = fmt.Sprintf("%s-%s", v.StudentNumber, uid)
		if v.ParentName != "" {
			v.ParentNumber = fmt.Sprintf("%s-%s", v.ParentNumber, uid)
		}
	}
	if err := gocsv.MarshalCSV(students, csvWriter); err != nil {
		return fmt.Errorf("gocsv.MarshalBytes: %s", err)
	}

	return nil
}

func uploadManagaraHighSchoolTSVData(csvWriter *csv.Writer, data []byte) error {
	reader := bytes.NewReader(data)
	csvReader := csv.NewReader(reader)
	csvReader.Comma = '\t'

	students := []*ManagaraHSStudent{}
	if err := gocsv.UnmarshalCSV(csvReader, &students); err != nil {
		return err
	}

	for _, v := range students {
		uid := idutil.ULIDNow()
		v.CustomerNumber = fmt.Sprintf("%s-%s", v.CustomerNumber, uid)
		v.StudentNumber = fmt.Sprintf("%s-%s", v.StudentNumber, uid)
		if v.ParentName != "" {
			v.ParentNumber = fmt.Sprintf("%s-%s", v.ParentNumber, uid)
		}
	}
	if err := gocsv.MarshalCSV(students, csvWriter); err != nil {
		return fmt.Errorf("gocsv.MarshalBytes: %s", err)
	}

	return nil
}

func initManagaraWriter(ctx context.Context, orgID string, bucketName string) (*csv.Writer, *storage.Writer, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	bucket := client.Bucket(bucketName)
	tokyoTime, err := timeInTokyo()
	if err != nil {
		return nil, nil, err
	}

	// upload file with correct name into bucket
	objectWriteName := managaraFileName(orgID, tokyoTime)
	objWriter := bucket.Object(objectWriteName).NewWriter(ctx)

	UTF8writer := transform.NewWriter(objWriter, japanese.ShiftJIS.NewEncoder())
	csvWriter := csv.NewWriter(UTF8writer)
	csvWriter.Comma = '\t'

	return csvWriter, objWriter, nil
}

func managaraFileName(orgID string, uploadDate time.Time) string {
	switch orgID {
	case fmt.Sprint(constants.ManagaraBase):
		return fmt.Sprintf("/withus/W2-D6L_users%s.tsv", withusDataFileNameSuffix(uploadDate))
	case fmt.Sprint(constants.ManagaraHighSchool):
		return fmt.Sprintf("/itee/N1-M1_users%s.tsv", withusDataFileNameSuffix(uploadDate))
	default:
		return ""
	}
}

func withusDataFileNameSuffix(uploadDate time.Time) string {
	return uploadDate.Format("20060102")
}

func timeInTokyo() (time.Time, error) {
	tokyoLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to init location: %s", err)
	}

	return time.Now().In(tokyoLocation), nil
}

func orgIDFromWithusDomainName(domain string) int {
	switch domain {
	case iteeDomain:
		return constants.ManagaraHighSchool
	case withusDomain:
		return constants.ManagaraBase
	default:
		return 0
	}
}

func tsvFilePath(orgID string) string {
	switch orgID {
	case fmt.Sprint(constants.ManagaraBase):
		return "usermgmt/testdata/tsv/itee/bdd_test.tsv"
	case fmt.Sprint(constants.ManagaraHighSchool):
		return "usermgmt/testdata/tsv/withus/bdd_test.tsv"
	default:
		return ""
	}
}

func (s *suite) validStudentParent(ctx context.Context, db database.QueryExecer, studentID string, parent aggregate.DomainParent) error {
	parentDBs, err := (&repository.DomainUserRepo{}).GetByExternalUserIDs(ctx, db, []string{parent.ExternalUserID().String()})
	if err != nil {
		return err
	}
	if len(parentDBs) == 0 {
		return fmt.Errorf("expect sync parentDBs %s success but don't existed in our system", parent.ExternalUserID().String())
	}
	parentInDB := parentDBs[0]
	stmt := `SELECT count(*) FROM student_parents WHERE student_id = $1 AND parent_id = $2 AND deleted_at IS NULL`
	count := 0
	if err := db.QueryRow(ctx, stmt, database.Text(studentID), database.Text(parentInDB.UserID().String())).Scan(&count); err != nil {
		return fmt.Errorf("validStudentParent failed: %s", err.Error())
	}
	if count != 1 {
		return fmt.Errorf("validStudentParent expected 1 but got %d", count)
	}
	if parentInDB.UserRole().String() != string(constant.UserRoleParent) {
		return fmt.Errorf("expected user_role is %v, actual external_user_id is %v", constant.UserRoleParent, parentInDB.UserRole().String())
	}

	isEnableUsername, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
	if err != nil {
		return errors.Wrap(err, "isFeatureToggleEnabled failed")
	}
	return assertUpsertUsername(isEnableUsername,
		assertUsername{
			requestUsername:    parent.UserName().String(),
			requestEmail:       parent.Email().String(),
			databaseUsername:   parentInDB.UserName().String(),
			requestLoginEmail:  parent.UserID().String() + constant.LoginEmailPostfix,
			databaseLoginEmail: parentInDB.LoginEmail().String(),
		},
	)
}

func validStudentCourses(ctx context.Context, db database.QueryExecer, studentID string, courseIDs entity.DomainStudentCourses) error {
	for _, course := range courseIDs {
		err := try.Do(func(attempt int) (bool, error) {
			studentPackage, err := (&repository.DomainStudentPackageRepo{}).GetByStudentIDAndCourseID(ctx, db, studentID, course.CourseID().String())
			if err != nil {
				return true, err
			}
			if len(studentPackage) == 0 {
				return true, fmt.Errorf("sync student package was not success, student: %s, course: %s", studentID, course.CourseID().String())
			}

			if attempt < retryTimes {
				time.Sleep(time.Second)
				return true, nil
			}

			return false, fmt.Errorf("expect sync student course %s success but don't existed in our system", course.CourseID().String())
		})

		if err != nil {
			return fmt.Errorf("validStudentCourses failed: %s", err.Error())
		}
	}

	return nil
}
