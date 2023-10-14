package withus

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/features"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"cloud.google.com/go/storage"
	"github.com/gocarina/gocsv"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	tsvHeaderGap = 2
)

func ImportManagaraStudents(ctx context.Context, bucketName, objectName string, domainStudentService *service.DomainStudent) []InternalError {
	errorsCollection := make([]InternalError, 0)
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		errorsCollection = append(errorsCollection, InternalError{
			RawErr: err,
		})
		return errorsCollection
	}

	fileReader, closeFileFunc, err := GetStudentDataFromFile(ctx, bucketName, objectName)
	if err != nil {
		errorsCollection = append(errorsCollection, InternalError{
			RawErr: err,
		})
		return errorsCollection
	}
	defer func() {
		_ = closeFileFunc()
	}()

	students, err := ToManagaraStudents(ctx, fileReader)
	if err != nil {
		errorsCollection = append(errorsCollection, InternalError{
			RawErr: err,
		})
		return errorsCollection
	}

	option := unleash.DomainStudentFeatureOption{}
	option = domainStudentService.FeatureManager.FeatureUsernameToStudentFeatureOption(ctx, organization, option)
	option = domainStudentService.FeatureManager.FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx, organization, option)
	option = domainStudentService.FeatureManager.FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(ctx, organization, option)
	option = domainStudentService.FeatureManager.FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(ctx, organization, option)

	mapIndexDomainStudent := make(map[int]aggregate.DomainStudentWithAssignedParent)
	for idx, student := range students {
		domainStudentAggregates, err := ManagaraStudentToDomainStudent(ctx, ManagaraStudents{student}, domainStudentService)
		if err != nil {
			errorsCollection = append(errorsCollection, InternalError{
				Index:  idx + tsvHeaderGap,
				RawErr: err,
				UserID: student.ExternalUserID().String(),
			})
			continue
		}
		domainStudentAggregates[0].IndexAttr = idx + tsvHeaderGap
		mapIndexDomainStudent[idx] = domainStudentAggregates[0]
	}

	studentIndex := make([]int, 0, len(mapIndexDomainStudent))
	for idx := range mapIndexDomainStudent {
		studentIndex = append(studentIndex, idx)
	}
	sort.Ints(studentIndex)
	for _, idx := range studentIndex {
		student := mapIndexDomainStudent[idx]
		if _, err := domainStudentService.UpsertMultipleWithAssignedParent(ctx, []aggregate.DomainStudentWithAssignedParent{student}, option); err != nil {
			errorsCollection = append(errorsCollection, InternalError{
				Index:  student.IndexAttr,
				RawErr: err,
				UserID: student.ExternalUserID().String(),
			})
		}
	}
	return errorsCollection
}

// StudentPortService represents a service in port layer
// This package import service package directly without decoupling with interface,
// so it causes an import cycle error if we define this port service in usermgmt port
// package and import this package's code
// TODO: Move this to grpc port when we move current code to cmd port
type StudentPortService struct {
	StudentService *service.DomainStudent
}

func (d *StudentPortService) ImportWithusManagaraBaseCSV(ctx context.Context, req *pb.ImportWithusManagaraBaseCSVRequest) (*pb.ImportWithusManagaraBaseCSVResponse, error) {
	logger := ctxzap.Extract(ctx)

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	domainErrs := ImportManagaraStudentV2(ctx, d.StudentService, bytes.NewReader(req.Payload))

	errs := make([]error, len(domainErrs))
	for i := range domainErrs {
		errs[i] = domainErrs[i]
	}

	if len(errs) > 0 {
		logger.Error("ImportManagaraStudentV2", zap.Errors("errs", errs))
		return nil, status.Error(codes.Internal, "")
	}
	return &pb.ImportWithusManagaraBaseCSVResponse{}, nil
}

func ImportManagaraStudentV2(ctx context.Context, domainStudentService *service.DomainStudent, fileReader io.Reader) []InternalError {
	errorsCollection := make([]InternalError, 0)
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		errorsCollection = append(errorsCollection, InternalError{
			RawErr: err,
		})
		return errorsCollection
	}
	students, err := ToManagaraStudents(ctx, fileReader)
	if err != nil {
		errorsCollection = append(errorsCollection, InternalError{
			RawErr: err,
		})
		return errorsCollection
	}

	option := unleash.DomainStudentFeatureOption{}
	option = domainStudentService.FeatureManager.FeatureUsernameToStudentFeatureOption(ctx, organization, option)
	option = domainStudentService.FeatureManager.FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx, organization, option)
	option = domainStudentService.FeatureManager.FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(ctx, organization, option)
	option = domainStudentService.FeatureManager.FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(ctx, organization, option)

	mapIndexDomainStudent := make(map[int]aggregate.DomainStudentWithAssignedParent)
	for idx, student := range students {
		domainStudentAggregates, err := ManagaraStudentToDomainStudent(ctx, ManagaraStudents{student}, domainStudentService)
		if err != nil {
			errorsCollection = append(errorsCollection, InternalError{
				Index:  idx + tsvHeaderGap,
				RawErr: err,
				UserID: student.ExternalUserID().String(),
			})
			continue
		}
		domainStudentAggregates[0].IndexAttr = idx + tsvHeaderGap
		mapIndexDomainStudent[idx] = domainStudentAggregates[0]
	}

	studentIndex := make([]int, 0, len(mapIndexDomainStudent))
	for idx := range mapIndexDomainStudent {
		studentIndex = append(studentIndex, idx)
	}
	sort.Ints(studentIndex)
	for _, idx := range studentIndex {
		student := mapIndexDomainStudent[idx]
		if _, err := domainStudentService.UpsertMultipleWithAssignedParent(ctx, []aggregate.DomainStudentWithAssignedParent{student}, option); err != nil {
			errorsCollection = append(errorsCollection, InternalError{
				Index:  student.IndexAttr,
				RawErr: err,
				UserID: student.ExternalUserID().String(),
			})
		}
	}
	return errorsCollection
}

func NewStudentService(ctx context.Context, c *configurations.Config, rsc *bootstrap.Resources) (*service.DomainStudent, error) {
	dbPool := rsc.DBWith("bob")
	zapLogger := rsc.Logger()
	unleash := rsc.WithUnleashC(&c.UnleashClientConfig).Unleash()
	jsm := rsc.NATS()

	firebaseProject := c.Common.FirebaseProject
	if firebaseProject == "" {
		firebaseProject = c.Common.GoogleCloudProject
	}
	singleTenantGCPApp, err := gcp.NewApp(ctx, "", firebaseProject)
	if err != nil {
		zapLogger.Fatal("failed to initialize gcp app for single tenant env", zap.Error(err))
	}
	firebaseAuthClient, err := multitenant.NewFirebaseAuthClientFromGCP(ctx, singleTenantGCPApp)
	if err != nil {
		zapLogger.Fatal("failed to initialize firebase auth client for single tenant env", zap.Error(err))
	}

	identityPlatformProject := c.Common.IdentityPlatformProject
	if identityPlatformProject == "" {
		identityPlatformProject = c.Common.GoogleCloudProject
	}
	multiTenantGCPApp, err := gcp.NewApp(ctx, "", identityPlatformProject)
	if err != nil {
		zapLogger.Fatal("failed to initialize gcp app for multi tenant env", zap.Error(err))
	}

	tenantManager, err := multitenant.NewTenantManagerFromGCP(ctx, multiTenantGCPApp)
	if err != nil {
		zapLogger.Fatal("failed to initialize identity platform tenant manager for multi tenant env", zap.Error(err))
	}
	fatimaConn := rsc.GRPCDial("fatima")
	if err != nil {
		return nil, fmt.Errorf("grpc.Dia fatima service: %w", err)
	}
	subscriptionModifierServiceClient := fpb.NewSubscriptionModifierServiceClient(fatimaConn)

	slackClient := &alert.SlackImpl{
		WebHookURL: c.SlackWebhook,
		HTTPClient: http.Client{Timeout: time.Duration(10) * time.Second},
	}
	configurationClient := clients.InitConfigurationClient(rsc.GRPCDial("mastermgmt"))

	userRepo := &repository.DomainUserRepo{}
	userGroupRepo := &repository.DomainUserGroupRepo{}
	userAddressRepo := &repository.DomainUserAddressRepo{}
	userPhoneNumberRepo := &repository.DomainUserPhoneNumberRepo{}
	schoolHistoryRepo := &repository.DomainSchoolHistoryRepo{}
	legacyUserGroup := &repository.LegacyUserGroupRepo{}
	userAccessPathRepo := &repository.DomainUserAccessPathRepo{}
	userGroupMemberRepo := &repository.DomainUserGroupMemberRepo{}
	locationRepo := &repository.DomainLocationRepo{}
	gradeRepo := &repository.DomainGradeRepo{}
	schoolRepo := &repository.DomainSchoolRepo{}
	schoolCourseRepo := &repository.DomainSchoolCourseRepo{}
	prefectureRepo := &repository.DomainPrefectureRepo{}
	usrEmailRepo := &repository.DomainUsrEmailRepo{}
	organizationRepo := (&repository.OrganizationRepo{}).WithDefaultValue(c.Common.Environment)
	enrollmentStatusHistoryRepo := &repository.DomainEnrollmentStatusHistoryRepo{}
	tagRepo := &repository.DomainTagRepo{}
	taggedUserRepo := &repository.DomainTaggedUserRepo{}
	courseRepo := &repository.DomainCourseRepo{}
	studentPackageRepo := &repository.DomainStudentPackageRepo{}
	studentParentRepo := &repository.DomainStudentParentRelationshipRepo{}
	internalConfigurationRepo := &repository.DomainInternalConfigurationRepo{}
	OrganizationRepoWithDefaultValue := (&repository.OrganizationRepo{}).WithDefaultValue(c.Common.Environment)
	domainParentService := &service.DomainParent{
		DB:                 dbPool,
		JSM:                jsm,
		FirebaseAuthClient: firebaseAuthClient,
		TenantManager:      tenantManager,
		UnleashClient:      unleash,
		Env:                c.Common.Environment,
		UserRepo:           userRepo,
		UserGroupRepo:      userGroupRepo,
		ParentRepo: &repository.DomainParentRepo{
			UserRepo:            userRepo,
			LegacyUserGroupRepo: legacyUserGroup,
			UserAccessPathRepo:  userAccessPathRepo,
			UserGroupMemberRepo: userGroupMemberRepo,
		},
		UserPhoneNumberRepo: userPhoneNumberRepo,
		UsrEmailRepo:        usrEmailRepo,
		OrganizationRepo:    organizationRepo,
		TaggedUserRepo:      taggedUserRepo,
		AuthUserUpserter:    service.NewAuthUserUpserter(userRepo, OrganizationRepoWithDefaultValue, firebaseAuthClient, tenantManager),
	}
	studentService := &service.DomainStudent{
		DB:                  dbPool,
		JSM:                 jsm,
		FirebaseAuthClient:  firebaseAuthClient,
		TenantManager:       tenantManager,
		UnleashClient:       unleash,
		Env:                 c.Common.Environment,
		DomainParentService: domainParentService,
		StudentRepo: &repository.DomainStudentRepo{
			UserRepo:            userRepo,
			LegacyUserGroupRepo: legacyUserGroup,
			UserAccessPathRepo:  userAccessPathRepo,
			UserGroupMemberRepo: userGroupMemberRepo,
		},
		UserRepo:                         userRepo,
		UserGroupRepo:                    userGroupRepo,
		UserAddressRepo:                  userAddressRepo,
		UserPhoneNumberRepo:              userPhoneNumberRepo,
		SchoolHistoryRepo:                schoolHistoryRepo,
		SchoolRepo:                       schoolRepo,
		SchoolCourseRepo:                 schoolCourseRepo,
		LocationRepo:                     locationRepo,
		GradeRepo:                        gradeRepo,
		PrefectureRepo:                   prefectureRepo,
		UsrEmailRepo:                     usrEmailRepo,
		OrganizationRepo:                 organizationRepo,
		TagRepo:                          tagRepo,
		TaggedUserRepo:                   taggedUserRepo,
		EnrollmentStatusHistoryRepo:      enrollmentStatusHistoryRepo,
		UserAccessPathRepo:               userAccessPathRepo,
		FatimaClient:                     subscriptionModifierServiceClient,
		CourseRepo:                       courseRepo,
		StudentPackage:                   studentPackageRepo,
		ConfigurationClient:              configurationClient,
		SlackClient:                      slackClient,
		StudentParentRepo:                studentParentRepo,
		StudentParentRelationshipManager: service.NewStudentParentRelationshipManager(&repository.DomainStudentParentRelationshipRepo{}),
		AuthUserUpserter:                 service.NewAuthUserUpserter(userRepo, OrganizationRepoWithDefaultValue, firebaseAuthClient, tenantManager),
		InternalConfigurationRepo:        internalConfigurationRepo,
		StudentValidationManager: &service.StudentValidationManager{
			UserRepo:                    &repository.DomainUserRepo{},
			UserGroupRepo:               &repository.DomainUserGroupRepo{},
			LocationRepo:                &repository.DomainLocationRepo{},
			GradeRepo:                   &repository.DomainGradeRepo{},
			SchoolRepo:                  &repository.DomainSchoolRepo{},
			SchoolCourseRepo:            &repository.DomainSchoolCourseRepo{},
			PrefectureRepo:              &repository.DomainPrefectureRepo{},
			TagRepo:                     &repository.DomainTagRepo{},
			InternalConfigurationRepo:   &repository.DomainInternalConfigurationRepo{},
			EnrollmentStatusHistoryRepo: &repository.DomainEnrollmentStatusHistoryRepo{},
			StudentRepo:                 &repository.DomainStudentRepo{},
		},
		FeatureManager: &features.FeatureManager{
			UnleashClient:             unleash,
			Env:                       c.Common.Environment,
			DB:                        dbPool,
			InternalConfigurationRepo: internalConfigurationRepo,
		},
	}

	return studentService, nil
}

func ToManagaraStudents(ctx context.Context, fileReader io.Reader) (ManagaraStudents, error) {
	csvReader := SJISReaderToUnicodeUTF8Reader(fileReader)

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "OrganizationFromContext")
	}

	switch organization.OrganizationID().String() {
	case fmt.Sprint(constants.ManagaraBase):
		students := make([]ManagaraBaseStudent, 0)
		if err := gocsv.UnmarshalCSV(csvReader, &students); err != nil {
			return nil, errors.Wrap(err, "gocsv.UnmarshalCSV error")
		}

		manaragaStudents := make(ManagaraStudents, 0, len(students))
		for _, student := range students {
			manaragaStudents = append(manaragaStudents, student.toManagaraStudent())
		}

		return manaragaStudents, nil
	case fmt.Sprint(constants.ManagaraHighSchool):
		students := make([]ManagaraHSStudent, 0)
		if err := gocsv.UnmarshalCSV(csvReader, &students); err != nil {
			return nil, errors.Wrap(err, "gocsv.UnmarshalCSV error")
		}

		manaragaStudents := make(ManagaraStudents, 0, len(students))
		for _, student := range students {
			manaragaStudents = append(manaragaStudents, student.toManagaraStudent())
		}

		return manaragaStudents, nil

	default:
		return nil, fmt.Errorf("invalid organization: %s", organization.OrganizationID().String())
	}
}

func ManagaraStudentToDomainStudent(ctx context.Context, students ManagaraStudents, domainStudentService *service.DomainStudent) (aggregate.DomainStudentWithAssignedParents, error) {
	domainStudentAggs := []aggregate.DomainStudentWithAssignedParent{}
	userIDs, err := toUserIDs(ctx, domainStudentService, students.externalUserIDs())
	if err != nil {
		return nil, err
	}

	for idx, student := range students {
		if userIDs[idx] != "" {
			student.UserIDAttr = userIDs[idx]
		}
		domainStudentAgg := aggregate.DomainStudent{
			DomainStudent: &entity.StudentWillBeDelegated{
				DomainStudentProfile: student,
				HasUserID:            student,
				HasGradeID:           student,
				HasLoginEmail: &entity.UserProfileLoginEmailDelegate{
					Email: student.Email().String(),
				},
			},
		}

		if !student.partnerGradeID().IsEmpty() {
			grade, err := mapToManabieGradeID(ctx, []string{student.partnerGradeID().String()}, domainStudentService)
			if err != nil {
				return nil, err
			}
			domainStudentAgg.DomainStudent = &entity.StudentWillBeDelegated{
				DomainStudentProfile: student,
				HasUserID:            student,
				HasGradeID:           grade,
				HasLoginEmail: &entity.UserProfileLoginEmailDelegate{
					Email: student.Email().String(),
				},
			}
		}

		locations := entity.DomainLocations{}
		if partnerLocationIDs := student.partnerLocationIDs(); len(partnerLocationIDs) > 0 {
			locations, err = mapToDomainLocations(ctx, partnerLocationIDs, domainStudentService)
			if err != nil {
				return nil, err
			}
			domainStudentAgg.UserAccessPaths = locations.ToUserAccessPath(student)
		}

		domainStudentCourses, err := mapToStudentCourses(ctx, domainStudentService, student.partnerCourseIDs(), domainStudentAgg)
		if err != nil {
			return nil, err
		}
		domainStudentAgg.Courses = domainStudentCourses

		if partnerTagIDs := student.partnerTagIDs(); len(partnerTagIDs) > 0 {
			domainTags, err := mapToDomainTags(ctx, domainStudentService, partnerTagIDs)
			if err != nil {
				return nil, err
			}
			domainStudentAgg.TaggedUsers = domainTags.ToTaggedUser(student)
		}

		enrollmentStatus, err := mapToDomainEnrollmentStatus(ctx, domainStudentService, domainStudentAgg, student.EnrollmentStatus().String())
		if err != nil {
			return nil, err
		}
		domainStudentAgg.EnrollmentStatusHistories = enrollmentStatus

		domainParents, err := toDomainParent(ctx, domainStudentService, student.Parent, locations)
		if err != nil {
			return nil, fmt.Errorf("toDomainParent error: %s", err.Error())
		}

		domainStudentAggs = append(domainStudentAggs, aggregate.DomainStudentWithAssignedParent{
			DomainStudent: domainStudentAgg,
			Parents:       domainParents,
		})
	}

	return domainStudentAggs, nil
}

func GetStudentDataFromFile(ctx context.Context, bucketName, objectName string) (io.Reader, func() error, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, nil, err
	}

	return rc, rc.Close, nil
}

func SJISReaderToUnicodeUTF8Reader(shiftJISReader io.Reader) *csv.Reader {
	UTF8reader := transform.NewReader(shiftJISReader, japanese.ShiftJIS.NewDecoder())

	reader := csv.NewReader(UTF8reader)
	reader.Comma = '\t'

	return reader
}

func toUserIDs(ctx context.Context, domainStudentService *service.DomainStudent, externalUserIDs []string) ([]string, error) {
	for _, userID := range externalUserIDs {
		if strings.TrimSpace(userID) == "" {
			return nil, errcode.Error{
				FieldName: "external_user_id",
				Code:      errcode.MissingMandatory,
			}
		}
	}

	existingUsers, err := domainStudentService.GetUsersByExternalIDs(ctx, externalUserIDs)
	if err != nil {
		return nil, err
	}
	userIDs := []string{}
	for _, externalUserID := range externalUserIDs {
		userID := ""
		for _, user := range existingUsers {
			if externalUserID == user.ExternalUserID().String() {
				userID = user.UserID().String()
			}
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func mapToDomainLocations(ctx context.Context, locationIDs []string, domainStudentService *service.DomainStudent) (entity.DomainLocations, error) {
	locations, err := domainStudentService.LocationRepo.GetByPartnerInternalIDs(ctx, domainStudentService.DB, locationIDs)
	if err != nil {
		return nil, err
	}
	if len(locations) != len(locationIDs) {
		return nil, fmt.Errorf("invalid locations: %s", locationIDs)
	}

	return locations, nil
}

func mapToManabieGradeID(ctx context.Context, gradeIDs []string, domainStudentService *service.DomainStudent) (entity.Grade, error) {
	grades, err := domainStudentService.GradeRepo.GetByPartnerInternalIDs(ctx, domainStudentService.DB, gradeIDs)
	if err != nil {
		return nil, err
	}
	if len(grades) > 0 {
		return grades[0], nil
	}

	return nil, fmt.Errorf("invalid grade %s", gradeIDs)
}

func mapToStudentCourses(ctx context.Context, domainStudentService *service.DomainStudent, coursePartnerIDs []string, studentAggregate aggregate.DomainStudent) (entity.DomainStudentCourses, error) {
	if len(coursePartnerIDs) == 0 {
		if err := softDeleteStudentPackages(ctx, domainStudentService, studentAggregate); err != nil {
			return nil, err
		}
		return nil, nil
	}

	courses, err := domainStudentService.CourseRepo.GetByCoursePartnerIDs(ctx, domainStudentService.DB, coursePartnerIDs)
	if err != nil {
		return nil, err
	}
	if len(courses) != len(coursePartnerIDs) {
		return nil, fmt.Errorf("invalid course ids: %s", coursePartnerIDs)
	}
	if err := softDeleteStudentPackages(ctx, domainStudentService, studentAggregate); err != nil {
		return nil, err
	}

	domainStudentCourses := entity.DomainStudentCourses{}
	for _, course := range courses {
		for _, location := range studentAggregate.UserAccessPaths {
			studentPackages, err := domainStudentService.StudentPackage.GetByStudentCourseAndLocationIDs(
				ctx, domainStudentService.DB, studentAggregate.UserID().String(), course.CourseID().String(), []string{location.LocationID().String()})
			if err != nil {
				return nil, err
			}
			if len(studentPackages) == 0 {
				withUsCourse := withUsCourse{
					courseID: course.CourseID(),
					startAt:  field.NewTime(startTime),
					endAt:    field.NewTime(endTime),
				}

				domainStudentCourses = append(domainStudentCourses, entity.StudentCourseWillBeDelegated{
					DomainStudentCourseAttribute: withUsCourse,
					HasUserID:                    studentAggregate,
					HasLocationID:                location,
				})
			} else {
				for _, studentPackage := range studentPackages {
					withUsCourse := withUsCourse{
						studentPackageID: studentPackage.StudentPackageID(),
						courseID:         course.CourseID(),
						startAt:          field.NewTime(startTime),
						endAt:            field.NewTime(endTime),
					}

					domainStudentCourses = append(domainStudentCourses, entity.StudentCourseWillBeDelegated{
						DomainStudentCourseAttribute: withUsCourse,
						HasUserID:                    studentAggregate,
						HasLocationID:                location,
					})
				}
			}
		}
	}

	return domainStudentCourses, nil
}

func mapToDomainTags(ctx context.Context, domainStudentService *service.DomainStudent, partnerTagIds []string) (entity.DomainTags, error) {
	domainTags, err := domainStudentService.TagRepo.GetByPartnerInternalIDs(ctx, domainStudentService.DB, partnerTagIds)
	if err != nil {
		return nil, err
	}
	if len(domainTags) != len(partnerTagIds) {
		return nil, fmt.Errorf("invalid tag ids: %s", partnerTagIds)
	}
	return domainTags, nil
}

func mapToDomainEnrollmentStatus(ctx context.Context, domainStudentService *service.DomainStudent, studentAggregate aggregate.DomainStudent, enrollmentStatus string) (entity.DomainEnrollmentStatusHistories, error) {
	enrollmentStatusHistories := entity.DomainEnrollmentStatusHistories{}

	if len(studentAggregate.UserAccessPaths) > 0 {
		withusEnrollmentStatus := &withUsEnrollmentStatusImpl{
			EnrollmentStatusAttr: field.NewString(enrollmentStatus),
			LocationAttr:         studentAggregate.UserAccessPaths[0].LocationID(),
		}

		latestEnrollmentStatus, err := domainStudentService.EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation(
			ctx, domainStudentService.DB, studentAggregate.UserID().String(), studentAggregate.UserAccessPaths[0].LocationID().String())
		if err != nil {
			return nil, InternalError{
				RawErr: errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetLatestEnrollmentStudentOfLocation"),
			}
		}
		switch {
		case len(latestEnrollmentStatus) == 0:
			withusEnrollmentStatus.StartDateAttr = field.NewTime(time.Now())
		case latestEnrollmentStatus[0].EnrollmentStatus().String() != enrollmentStatus:
			withusEnrollmentStatus.StartDateAttr = field.NewTime(time.Now())
		case latestEnrollmentStatus[0].EnrollmentStatus().String() == enrollmentStatus:
			withusEnrollmentStatus.StartDateAttr = field.NewNullTime()
		}

		enrollmentStatusHistories = append(enrollmentStatusHistories, withusEnrollmentStatus)
	}

	return enrollmentStatusHistories, nil
}

func softDeleteStudentPackages(ctx context.Context, domainStudentService *service.DomainStudent, studentAggregate aggregate.DomainStudent) error {
	studentPackages, err := domainStudentService.StudentPackage.GetByStudentIDs(ctx, domainStudentService.DB, []string{studentAggregate.UserID().String()})
	if err != nil {
		return err
	}

	for _, studentPackage := range studentPackages {
		// don't update end_date if end_date in past
		if studentPackage.EndDate().Time().Before(time.Now()) {
			continue
		}

		updateStudentPackageCourseReq := &fpb.EditTimeStudentPackageRequest{
			StudentPackageId: studentPackage.StudentPackageID().String(),
			StartAt:          timestamppb.New(studentPackage.StartDate().Time()),
			EndAt:            timestamppb.New(time.Now()),
			LocationIds:      field.ToSliceString(studentPackage.LocationIDs()),
		}

		_, err := domainStudentService.FatimaClient.EditTimeStudentPackage(ctx, updateStudentPackageCourseReq)
		if err != nil {
			return InternalError{
				RawErr: errors.Wrap(err, "domainStudentService.FatimaClient.EditTimeStudentPackage"),
			}
		}
	}

	return nil
}

func toDomainParent(ctx context.Context, domainStudentService *service.DomainStudent, parent Parent, locations entity.DomainLocations) (aggregate.DomainParents, error) {
	if !field.IsPresent(parent.ParentNumber) {
		return nil, nil
	}
	parentIDs, err := toUserIDs(ctx, domainStudentService, []string{parent.ExternalUserID().String()})
	if err != nil {
		return nil, InternalError{
			RawErr: errors.Wrap(err, "invalid parent"),
		}
	}
	if len(parentIDs) > 0 {
		if parentIDs[0] != "" {
			parent.UserIDAttr = parentIDs[0]
		}
	}
	domainParent := aggregate.DomainParent{
		DomainParent: &entity.ParentWillBeDelegated{
			DomainParentProfile: parent,
			HasUserID:           parent,
			HasLoginEmail: &entity.UserProfileLoginEmailDelegate{
				Email: parent.Email().String(),
			},
		},
	}
	if len(locations) > 0 {
		domainParent.UserAccessPaths = locations.ToUserAccessPath(parent)
	}

	return []aggregate.DomainParent{domainParent}, nil
}

func DataFileNameSuffix(uploadDate time.Time) string {
	return uploadDate.Format("20060102")
}
