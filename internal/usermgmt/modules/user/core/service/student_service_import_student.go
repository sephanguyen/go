package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	nats2 "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc/importstudent"
	http_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/http"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	helper "github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gocarina/gocsv"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	studentCSVHeader string
)

const (
	emailStudentCSVHeader             studentCSVHeader = "email"
	enrollmentStatusStudentCSVHeader  studentCSVHeader = "enrollment_status"
	gradeStudentCSVHeader             studentCSVHeader = "grade"
	studentPhoneNumberCSVHeader       studentCSVHeader = "student_phone_number"
	studentHomePhoneNumberCSVHeader   studentCSVHeader = "home_phone_number"
	studentContactPreferenceCSVHeader studentCSVHeader = "contact_preference"
	phoneNumberStudentCSVHeader       studentCSVHeader = "phone_number"
	birthdayStudentCSVHeader          studentCSVHeader = "birthday"
	genderStudentCSVHeader            studentCSVHeader = "gender"
	locationStudentCSVHeader          studentCSVHeader = "location"
	firstNameStudentCSVHeader         studentCSVHeader = "first_name"
	lastNameStudentCSVHeader          studentCSVHeader = "last_name"
	prefectureStudentCSVHeader        studentCSVHeader = "prefecture"
	schoolStudentCSVHeader            studentCSVHeader = "school"
	schoolCourseStudentCSVHeader      studentCSVHeader = "school_course"
	startDateStudentCSVHeader         studentCSVHeader = "start_date"
	endDateStudentCSVHeader           studentCSVHeader = "end_date"
	tagStudentCSVHeader               studentCSVHeader = "student_tag"
	statusStartDateStudentCSVHeader   studentCSVHeader = "status_start_date"
)

type ImportStudentCSVField struct {
	FirstName                field.String `csv:"first_name"`
	LastName                 field.String `csv:"last_name"`
	FirstNamePhonetic        field.String `csv:"first_name_phonetic"`
	LastNamePhonetic         field.String `csv:"last_name_phonetic"`
	Email                    field.String `csv:"email"`
	EnrollmentStatus         field.String `csv:"enrollment_status"`
	Grade                    field.String `csv:"grade"`
	PhoneNumber              field.String `csv:"phone_number"`
	Birthday                 field.String `csv:"birthday"`
	Gender                   field.String `csv:"gender"`
	Location                 field.String `csv:"location"`
	PostalCode               field.String `csv:"postal_code"`
	Prefecture               field.String `csv:"prefecture"`
	City                     field.String `csv:"city"`
	FirstStreet              field.String `csv:"first_street"`
	SecondStreet             field.String `csv:"second_street"`
	StudentPhoneNumber       field.String `csv:"student_phone_number"`
	StudentHomePhoneNumber   field.String `csv:"home_phone_number"`
	StudentContactPreference field.String `csv:"contact_preference"`
	School                   field.String `csv:"school"`
	SchoolCourse             field.String `csv:"school_course"`
	StartDate                field.String `csv:"start_date"`
	EndDate                  field.String `csv:"end_date"`
	StudentTag               field.String `csv:"student_tag"`
	StatusStartDate          field.String `csv:"status_start_date"`
}

type MapEnrollmentStatusHistoryCSVValue struct {
	entity.DefaultDomainEnrollmentStatusHistory

	enrollmentStatusString string
	userID                 string
	resourcePath           string
	startDate              time.Time
}

func NewEnrollmentStatusHistoryCSV(enrollmentStatusString, userID, resourcePath string, startDate time.Time) entity.DomainEnrollmentStatusHistory {
	return &MapEnrollmentStatusHistoryCSVValue{
		enrollmentStatusString: enrollmentStatusString,
		userID:                 userID,
		startDate:              startDate,
		resourcePath:           resourcePath,
	}
}

func (e *MapEnrollmentStatusHistoryCSVValue) UserID() field.String {
	return field.NewString(e.userID)
}

func (e *MapEnrollmentStatusHistoryCSVValue) EnrollmentStatus() field.String {
	return field.NewString(e.enrollmentStatusString)
}

func (e *MapEnrollmentStatusHistoryCSVValue) StartDate() field.Time {
	return field.NewTime(e.startDate)
}

func (e *MapEnrollmentStatusHistoryCSVValue) OrganizationID() field.String {
	return field.NewString(e.resourcePath)
}

var mapStudentContactPreference = map[string]string{
	"1": entity.StudentPhoneNumber,
	"2": entity.StudentHomePhoneNumber,
	"3": entity.ParentPrimaryPhoneNumber,
	"4": entity.ParentSecondaryPhoneNumber,
}

func (s *StudentService) checkDuplicateData(ctx context.Context, students entity.LegacyStudents) ([]*pb.ImportStudentResponse_ImportStudentError, error) {
	var errorCSVs []*pb.ImportStudentResponse_ImportStudentError
	zapLogger := ctxzap.Extract(ctx).Sugar()

	// validate email in student data
	emails := students.Emails()
	existingUsers, err := s.UserRepo.GetByEmailInsensitiveCase(ctx, s.DB, emails)
	if err != nil {
		return errorCSVs, status.Errorf(codes.Internal, "s.UserRepo.GetByEmailInsensitiveCase: %v", err)
	}
	if len(existingUsers) > 0 {
		zapLogger.Warnf("email in student data validation failed email: %s", emails)
		for _, user := range existingUsers {
			rowNumber := helper.IndexOf(emails, user.Email.String)
			errorCSVs = append(errorCSVs, convertErrToErrResForEachLineCSV(fmt.Errorf("alreadyRegisteredRow"), emailStudentCSVHeader, rowNumber))
		}
	}

	// validate phone_number in student data
	phoneNumbers := students.PhoneNumbers()
	phoneNumberExistingStudents, err := s.UserRepo.GetByPhone(ctx, s.DB, database.TextArray(phoneNumbers))

	if err != nil {
		return errorCSVs, status.Errorf(codes.Internal, "s.UserRepo.GetByPhone: %v", err)
	}
	if len(phoneNumberExistingStudents) > 0 {
		zapLogger.Warnf("phone_number in student data validation failed phone_number: %v", phoneNumbers)
		for _, user := range phoneNumberExistingStudents {
			if user.PhoneNumber.String == "" {
				continue
			}
			rowNumber := helper.IndexOf(phoneNumbers, user.PhoneNumber.String)
			errorCSVs = append(errorCSVs, convertErrToErrResForEachLineCSV(fmt.Errorf("alreadyRegisteredRow"), phoneNumberStudentCSVHeader, rowNumber))
		}
	}
	return errorCSVs, nil
}

func (s *StudentService) checkDuplicateRow(students entity.LegacyStudents) []*pb.ImportStudentResponse_ImportStudentError {
	var errorCSVs []*pb.ImportStudentResponse_ImportStudentError

	// this var is for checking duplicate email and phone_number in CSV file
	knowEmailAndPhone := map[string]bool{}
	for i, student := range students {
		_, ok := knowEmailAndPhone[strings.ToLower(student.Email.String)]
		if !ok {
			knowEmailAndPhone[strings.ToLower(student.Email.String)] = true
		} else {
			errorCSVs = append(errorCSVs, convertErrToErrResForEachLineCSV(fmt.Errorf("duplicationRow"), emailStudentCSVHeader, i))
			continue
		}

		if student.PhoneNumber.String != "" {
			_, ok = knowEmailAndPhone[student.PhoneNumber.String]
			if !ok {
				knowEmailAndPhone[student.PhoneNumber.String] = true
			} else {
				errorCSVs = append(errorCSVs, convertErrToErrResForEachLineCSV(fmt.Errorf("duplicationRow"), phoneNumberStudentCSVHeader, i))
				continue
			}
		}
	}

	return errorCSVs
}

func generatedStudentsAndUserAccessPaths(studentCSVs []*StudentCSV) (entity.LegacyStudents, []*entity.UserAccessPath, error) {
	students := []*entity.LegacyStudent{}
	userAccessPaths := []*entity.UserAccessPath{}
	for _, studentCSV := range studentCSVs {
		students = append(students, &studentCSV.Student)
		if len(studentCSV.Locations) > 0 {
			generatedUserAccessPaths, err := toUserAccessPathEntities(studentCSV.Locations, []string{studentCSV.Student.ID.String})
			if err != nil {
				err = status.Errorf(codes.Internal, "otherErrorImport toUserAccessPathEntities: %v", err)
				return nil, nil, err
			}
			userAccessPaths = append(userAccessPaths, generatedUserAccessPaths...)
		}
	}
	return students, userAccessPaths, nil
}

func generatedEnrollmentStatusHistoriesAndUserAccessPaths(studentCSVs []*StudentCSV) []entity.DomainEnrollmentStatusHistories {
	enrollmentStatusHistories := []entity.DomainEnrollmentStatusHistories{}
	for _, studentCSV := range studentCSVs {
		if len(studentCSV.EnrollmentStatusHistories) > 0 {
			enrollmentStatusHistories = append(enrollmentStatusHistories, studentCSV.EnrollmentStatusHistories)
		}
	}
	return enrollmentStatusHistories
}

// generate usersWithTags, userPhoneNumbers, userAddresses from csv
func generateUsersDetailsFromCSV(studentCSVs []*StudentCSV) (map[entity.User][]entity.DomainTag, []*entity.UserPhoneNumber, []*entity.UserAddress) {
	userWithTags := map[entity.User][]entity.DomainTag{}
	userPhoneNumbers := []*entity.UserPhoneNumber{}
	userAddresses := []*entity.UserAddress{}

	for _, studentCSV := range studentCSVs {
		// extract userWithTags
		userProfile := grpc.NewUserProfileWithID(studentCSV.Student.GetUID())
		userWithTags[userProfile] = studentCSV.Tags

		// extract userPhoneNumbers
		userPhoneNumbers = append(userPhoneNumbers, studentCSV.StudentPhoneNumbers...)

		// extract UserAddress
		if studentCSV.UserAddress.UserAddressID.String != "" {
			userAddresses = append(userAddresses, &studentCSV.UserAddress)
		}
	}

	return userWithTags, userPhoneNumbers, userAddresses
}

func (s *StudentService) importWithSchoolHistories(ctx context.Context, tx pgx.Tx, studentCSVs []*StudentCSV) error {
	for _, studentCSV := range studentCSVs {
		if len(studentCSV.SchoolHistories) != 0 {
			if err := s.SchoolHistoryRepo.Upsert(ctx, tx, studentCSV.SchoolHistories); err != nil {
				return errorx.ToStatusError(err)
			}

			currentSchools, err := s.SchoolHistoryRepo.GetSchoolHistoriesByGradeIDAndStudentID(ctx, tx, studentCSV.Student.GradeID, database.Text(studentCSV.Student.ID.String), database.Bool(false))
			if err != nil {
				return err
			}
			if len(currentSchools) != 0 {
				err = s.SchoolHistoryRepo.SetCurrentSchoolByStudentIDAndSchoolID(ctx, tx, database.Text(currentSchools[0].SchoolID.String), database.Text(studentCSV.Student.ID.String))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *StudentService) importWithEnrollmentStatusHistories(ctx context.Context, db database.QueryExecer, enrollmentStatusHistories []entity.DomainEnrollmentStatusHistories) error {
	for _, domainEnrollmentStatusHistories := range enrollmentStatusHistories {
		for _, domainEnrollmentStatusHistory := range domainEnrollmentStatusHistories {
			userAccessPath := entity.UserAccessPathWillBeDelegated{
				HasUserID:         domainEnrollmentStatusHistory,
				HasLocationID:     domainEnrollmentStatusHistory,
				HasOrganizationID: domainEnrollmentStatusHistory,
			}
			if err := s.createEnrollmentStatusHistory(ctx, db, domainEnrollmentStatusHistory, userAccessPath); err != nil {
				return errorx.ToStatusError(err)
			}
		}
	}
	return nil
}

func (s *StudentService) createEnrollmentStatusHistory(ctx context.Context, db database.QueryExecer, enrolmentStatus entity.DomainEnrollmentStatusHistory, userAccessPath entity.DomainUserAccessPath) error {
	// check location of student is exists (new or not)
	enrollmentStatusHistories, err := s.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, db, enrolmentStatus.UserID().String(), enrolmentStatus.LocationID().String(), false)
	if err != nil {
		return errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID")
	}

	if len(enrollmentStatusHistories) == 0 {
		if err := s.EnrollmentStatusHistoryRepo.Create(ctx, db, enrolmentStatus); err != nil {
			return errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.Create")
		}
		if err := s.DomainUserAccessPathRepo.UpsertMultiple(ctx, db, []entity.DomainUserAccessPath{userAccessPath}...); err != nil {
			return errors.Wrap(err, "service.UserAccessPathRepo.UpsertMultiple")
		}
	} else { // location exist
		// Get current enrollment status (start date < current date , end date > current date)
		currentEnrollmentStatus, err := s.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, db, enrolmentStatus.UserID().String(), enrolmentStatus.LocationID().String(), true)
		if err != nil {
			return errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID")
		}
		// Have current enrollment status
		if len(currentEnrollmentStatus) != 0 {
			// if different Enrollment Status -> update Enrollment Status
			// INSERT new [Enrollment Status] with [Start_date] to StudentEnrollmentStatusHistory
			// UPDATE previous record’s [End_date] ([Start_date] - 1sec)
			if currentEnrollmentStatus[0].EnrollmentStatus().String() != enrolmentStatus.EnrollmentStatus().String() {
				err = s.EnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus(ctx, db,
					currentEnrollmentStatus[0],
					enrolmentStatus.StartDate().Time().Add(-1*time.Second),
				)
				if err != nil {
					return errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.DeactivateEnrollmentStatus")
				}
				if err := s.EnrollmentStatusHistoryRepo.Create(ctx, db, enrolmentStatus); err != nil {
					return errors.Wrap(err, "service.EnrollmentStatusHistoryRepo.Create")
				}
				if err := s.DomainUserAccessPathRepo.UpsertMultiple(ctx, db, []entity.DomainUserAccessPath{userAccessPath}...); err != nil {
					return errors.Wrap(err, "service.UserAccessPathRepo.UpsertMultiple")
				}
			}
			return nil
		}
	}

	return nil
}

func (s *StudentService) generatedStudentCSVs(ctx context.Context, importStudentData []*ImportStudentCSVField) ([]*StudentCSV, []*pb.ImportStudentResponse_ImportStudentError, error) {
	var errorCSVs []*pb.ImportStudentResponse_ImportStudentError
	currentUserID := interceptors.UserIDFromContext(ctx)
	currentUser, err := s.UserRepo.Get(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return nil, errorCSVs, fmt.Errorf("s.UserRepo.Get: %v", err)
	}
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	if err != nil {
		return nil, errorCSVs, fmt.Errorf("s.UserRepo.Get: %v", err)
	}
	studentCSVs := []*StudentCSV{}
	for idx, data := range importStudentData {
		var (
			errLineRes *pb.ImportStudentResponse_ImportStudentError
			studentCSV StudentCSV
		)

		studentCSV, errLineRes, err := s.convertLineCSVToStudentCSV(ctx, data, idx, currentUser.Country.String, resourcePath)
		if err != nil {
			err = status.Errorf(codes.Internal, "otherErrorImport s.convertLineCSVToStudentCSV: %v", err)
			return nil, errorCSVs, fmt.Errorf("s.convertLineCSVToStudentCSV: %v", err)
		}
		if errLineRes != nil {
			errorCSVs = append(errorCSVs, errLineRes)
			continue
		}
		studentCSVs = append(studentCSVs, &studentCSV)
	}
	return studentCSVs, errorCSVs, nil
}

func (s *StudentService) ImportStudentV2(ctx context.Context, req *pb.ImportStudentRequest) (*pb.UpsertStudentResponse, error) {
	service := importstudent.DomainStudentService{
		DomainStudent:  s.DomainStudentService,
		FeatureManager: s.FeatureManager,
	}
	return service.ImportStudentV2(ctx, req)
}

func (s *StudentService) ImportStudent(ctx context.Context, req *pb.ImportStudentRequest) (*pb.ImportStudentResponse, error) {
	res := &pb.ImportStudentResponse{}

	var errorCSVs []*pb.ImportStudentResponse_ImportStudentError

	if err := readAndValidatePayload(req.Payload); err != nil {
		return nil, err
	}
	importStudentData, err := convertPayloadToImportStudentData(req.Payload)
	if err != nil {
		return nil, err
	}
	if len(importStudentData) == 0 {
		return &pb.ImportStudentResponse{}, nil
	}

	if len(importStudentData) > constant.LimitRowsCSV {
		return nil, status.Error(codes.InvalidArgument, "invalidNumberRow")
	}

	studentCSVs, errorCSVs, err := s.generatedStudentCSVs(ctx, importStudentData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport s.generatedStudentCSVs: %v", err)
	}
	if len(errorCSVs) > 0 {
		res.Errors = errorCSVs
		return res, nil
	}

	if len(studentCSVs) == 0 {
		return res, nil
	}

	students, _, err := generatedStudentsAndUserAccessPaths(studentCSVs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport generatedStudentsAndUserAccessPaths without userAccessPaths: %v", err)
	}

	errorCSVs = s.checkDuplicateRow(students)
	if len(errorCSVs) > 0 {
		res.Errors = errorCSVs
		return res, nil
	}

	errorCSVs, err = s.checkDuplicateData(ctx, students)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport s.checkDuplicateData: %v", err)
	}

	if len(errorCSVs) > 0 {
		res.Errors = errorCSVs
		return res, nil
	}

	err = s.insertUsrEmailsFromStudents(ctx, studentCSVs)
	if err != nil {
		err = status.Errorf(codes.Internal, "otherErrorImport s.insertUsrEmailsFromStudents: %v", err)
		return nil, err
	}

	studentsAfterInsertedUsrEmails, userAccessPaths, err := generatedStudentsAndUserAccessPaths(studentCSVs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport generatedStudentsAndUserAccessPaths: %v", err)
	}

	enrollmentStatusHistories := generatedEnrollmentStatusHistoriesAndUserAccessPaths(studentCSVs)

	usersWithTags, userPhoneNumbers, userAddresses := generateUsersDetailsFromCSV(studentCSVs)

	importUserEvents, err := toImportStudentUserEvents(ctx, studentCSVs)
	if err != nil {
		err = status.Errorf(codes.Internal, "otherErrorImport toImportUserEvents: %v", err)
		return nil, err
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// find student user group id
		studentUserGroup, err := s.UserGroupV2Repo.FindUserGroupByRoleName(ctx, tx, constant.RoleStudent)
		if err != nil {
			return fmt.Errorf("error when finding student user group: %w", err)
		}

		err = s.createStudents(ctx, tx, studentsAfterInsertedUsrEmails)
		if err != nil {
			return fmt.Errorf("s.createStudents: %v", err)
		}

		if err := s.importWithSchoolHistories(ctx, tx, studentCSVs); err != nil {
			return fmt.Errorf("s.importWithSchoolHistories: %v", err)
		}

		if len(userAddresses) > 0 {
			if err := s.UserAddressRepo.Upsert(ctx, tx, userAddresses); err != nil {
				return fmt.Errorf("s.UserAddressRepo.Upsert: %v", err)
			}
		}

		if len(userPhoneNumbers) > 0 {
			if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
				return fmt.Errorf("s.UserPhoneNumberRepo.Upsert: %v", err)
			}
		}

		err = s.UserAccessPathRepo.Upsert(ctx, tx, userAccessPaths)
		if err != nil {
			return fmt.Errorf("s.UserAccessPathRepo.Upsert: %v", err)
		}

		// assign student user group to student
		if err := s.UserGroupsMemberRepo.AssignWithUserGroup(ctx, tx, students.Users(), studentUserGroup.UserGroupID); err != nil {
			return fmt.Errorf("error when assigning student user group to users: %w", err)
		}

		importUserEvents, err = s.ImportUserEventRepo.Upsert(ctx, tx, importUserEvents)
		if err != nil {
			return fmt.Errorf("s.ImportUserEventRepo.Upsert err: %v", err)
		}

		if err := s.UserModifierService.UpsertTaggedUsers(ctx, tx, usersWithTags, nil); err != nil {
			return errors.Wrap(err, "UpsertTaggedUsers")
		}

		if len(enrollmentStatusHistories) > 0 {
			if err := s.importWithEnrollmentStatusHistories(ctx, tx, enrollmentStatusHistories); err != nil {
				return fmt.Errorf("s.importWithEnrollmentStatusHistories: %v", err)
			}
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport database.ExecInTx: %v", err)
	}

	importUserEventIDs := importUserEvents.IDs()
	err = s.TaskQueue.Add(nats2.PublishImportUserEventsTask(ctx, s.DB, s.JSM, &nats2.PublishImportUserEventsTaskOptions{
		ImportUserEventIDs: importUserEventIDs,
		ResourcePath:       golibs.ResourcePathFromCtx(ctx),
	}))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport task.TaskQueue.Add: %v", err)
	}

	return res, nil
}

func toImportStudentUserEvents(ctx context.Context, studentCSVs []*StudentCSV) (entity.ImportUserEvents, error) {
	importUserEvents := make([]*entity.ImportUserEvent, 0, len(studentCSVs))

	for _, studentCSV := range studentCSVs {
		locationIDs := []string{}
		for _, location := range studentCSV.Locations {
			locationIDs = append(locationIDs, location.LocationID)
		}
		createStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_CreateStudent_{
				CreateStudent: &pb.EvtUser_CreateStudent{
					StudentId:   studentCSV.Student.ID.String,
					StudentName: studentCSV.Student.GetName(),
					SchoolId:    studentCSV.Student.ResourcePath.String,
					LocationIds: locationIDs,
				},
			},
		}
		importUserEvent := &entity.ImportUserEvent{}
		database.AllNullEntity(importUserEvent)

		payload, err := protojson.Marshal(createStudentEvent)
		if err != nil {
			return nil, fmt.Errorf("protojson.Marshal err: %v", err)
		}

		err = multierr.Combine(
			importUserEvent.ImporterID.Set(interceptors.UserIDFromContext(ctx)),
			importUserEvent.Status.Set(cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_WAITING.String()),
			importUserEvent.UserID.Set(studentCSV.Student.ID.String),
			importUserEvent.Payload.Set(payload),
			importUserEvent.ResourcePath.Set(golibs.ResourcePathFromCtx(ctx)),
		)
		if err != nil {
			return nil, fmt.Errorf("multierr.Combine err: %v", err)
		}

		importUserEvents = append(importUserEvents, importUserEvent)
	}

	return importUserEvents, nil
}

func (s *StudentService) insertUsrEmailsFromStudents(ctx context.Context, studentCSVs []*StudentCSV) error {
	// Insert UsrEmail
	users := make(entity.LegacyUsers, 0, len(studentCSVs))
	for _, studentCSV := range studentCSVs {
		users = append(users, &studentCSV.Student.LegacyUser)
	}
	usrEmails, err := s.UsrEmailRepo.CreateMultiple(ctx, s.DB, users)
	if err != nil {
		return fmt.Errorf("s.UsrEmailRepo.CreateMultiple: %v", err)
	}
	for i := range usrEmails {
		studentCSVs[i].Student.ID = usrEmails[i].UsrID
		studentCSVs[i].Student.LegacyUser.ID = usrEmails[i].UsrID
	}
	return nil
}

func readAndValidatePayload(payload []byte) error {
	sizeInMB := len(payload) / (1024 * 1024)
	if sizeInMB > 5 {
		return status.Error(codes.InvalidArgument, "invalidMaxSizeFile")
	}
	if len(payload) == 0 {
		return status.Error(codes.InvalidArgument, "emptyFile")
	}
	return nil
}

func convertPayloadToImportStudentData(payload []byte) ([]*ImportStudentCSVField, error) {
	var studentImportData []*ImportStudentCSVField
	if err := gocsv.UnmarshalBytes(payload, &studentImportData); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("gocsv.UnmarshalBytes: %w", err).Error())
	}
	return studentImportData, nil
}

func (s *StudentService) createStudents(ctx context.Context, db database.QueryExecer, students []*entity.LegacyStudent) error {
	// UserGroups need to be inserted with new parent
	userGroups := make([]*entity.UserGroup, 0, len(students))
	users := make(entity.LegacyUsers, 0, len(students))

	for _, student := range students {
		userGroup := &entity.UserGroup{}
		database.AllNullEntity(userGroup)
		err := multierr.Combine(
			userGroup.UserID.Set(student.ID.String),
			userGroup.GroupID.Set(entity.UserGroupStudent),
			userGroup.IsOrigin.Set(true),
			userGroup.Status.Set(entity.UserGroupStatusActive),
			userGroup.ResourcePath.Set(student.ResourcePath),
		)
		if err != nil {
			return fmt.Errorf("multierr.Combine: %v", err)
		}

		userGroups = append(userGroups, userGroup)
		users = append(users, &student.LegacyUser)
	}

	// Insert new students
	err := s.UserRepo.CreateMultiple(ctx, db, users)
	if err != nil {
		return fmt.Errorf("s.UserRepo.CreateMultiple: %v", errorx.ToStatusError(err))
	}

	err = s.StudentRepo.CreateMultiple(ctx, db, students)
	if err != nil {
		return fmt.Errorf("s.StudentRepo.CreateMultiple: %v", errorx.ToStatusError(err))
	}

	err = s.UserGroupRepo.CreateMultiple(ctx, db, userGroups)
	if err != nil {
		return fmt.Errorf("s.UserGroupRepo.CreateMultiple: %v", errorx.ToStatusError(err))
	}

	err = s.importToAuthPlatform(ctx, db, users)
	if err != nil {
		return status.Errorf(codes.Internal, "s.importToAuthPlatform: %v", err)
	}

	return nil
}

func (s *StudentService) importToAuthPlatform(ctx context.Context, db database.QueryExecer, identityPlatformAccounts entity.LegacyUsers) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	resourcePathInt, err := strconv.Atoi(resourcePath)
	if err != nil {
		return status.Errorf(codes.Internal, "strconv.Atoi: %v", err)
	}

	// Import to identity platform
	tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, db, resourcePath)
	if err != nil {
		zapLogger.Error(
			"cannot get tenant id",
			zap.Error(err),
			zap.String("organizationID", resourcePath),
		)
		switch err {
		case pgx.ErrNoRows:
			return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: resourcePath}.Error())
		default:
			return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
		}
	}

	err = s.UserModifierService.CreateUsersInIdentityPlatform(ctx, tenantID, identityPlatformAccounts, int64(resourcePathInt))
	if err != nil {
		zapLogger.Error(
			"failed to create users on identity platform",
			zap.Error(err),
			zap.String("organizationID", resourcePath),
			zap.String("tenantID", tenantID),
			zap.Strings("emails", identityPlatformAccounts.Limit(10).Emails()),
		)
		// convert to switch case to handle error types
		return status.Error(codes.Internal, errors.Wrap(err, "failed to create users on identity platform").Error())
	}

	return nil
}

type StudentCSV struct {
	Student                   entity.LegacyStudent
	UserAddress               entity.UserAddress
	Locations                 []*domain.Location
	StudentPhoneNumbers       []*entity.UserPhoneNumber
	SchoolHistories           []*entity.SchoolHistory
	Tags                      entity.DomainTags
	EnrollmentStatusHistories entity.DomainEnrollmentStatusHistories
	RowNumber                 int
}

func (s *StudentService) convertLineCSVToStudentCSV(
	ctx context.Context,
	importStudentCSVField *ImportStudentCSVField,
	order int,
	countryCode,
	resourcePath string,
) (StudentCSV, *pb.ImportStudentResponse_ImportStudentError, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	studentCSV := StudentCSV{
		RowNumber: order,
	}
	student := &studentCSV.Student
	userAddress := &studentCSV.UserAddress
	database.AllNullEntity(student)
	database.AllNullEntity(userAddress)
	database.AllNullEntity(&student.LegacyUser)
	var (
		id                 string
		enrollmentStatuses []string
		phoneNumber        string
		grade              int32
		gradeID            string
		gender             string
		birthday           time.Time
		locations          []string
		tags               []string
	)

	if err := validateRequiredFieldImportStudentCSV(importStudentCSVField, order); err != nil {
		return studentCSV, err, nil
	}
	errResForEachLineCSV, err := importStudentCSVValidateEmail(importStudentCSVField.Email.String(), order)
	if errResForEachLineCSV != nil || err != nil {
		return studentCSV, errResForEachLineCSV, err
	}

	// validate enrollment_status
	enrollmentStatus := importStudentCSVField.EnrollmentStatus.String()
	enrollmentStatuses = strings.Split(enrollmentStatus, ";")

	enrollmentStatuses, errResForEachLineCSV = importStudentCSVValidateEnrollmentStatus(enrollmentStatuses, order)
	if errResForEachLineCSV != nil {
		return studentCSV, errResForEachLineCSV, nil
	}

	grade, gradeID, errResForEachLineCSV = s.importStudentGetGradeFromGradeMaster(ctx, order, importStudentCSVField)
	if errResForEachLineCSV != nil {
		return studentCSV, errResForEachLineCSV, nil
	}

	if field.IsPresent(importStudentCSVField.Gender) {
		gender, errResForEachLineCSV = importStudentCSVValidateGender(importStudentCSVField.Gender.String(), order)
		if errResForEachLineCSV != nil {
			return studentCSV, errResForEachLineCSV, nil
		}
		err := student.LegacyUser.Gender.Set(gender)
		if err != nil {
			return studentCSV, nil, fmt.Errorf("student.User.Gender.Set: %v, row: %v", err, order)
		}
	}

	if field.IsPresent(importStudentCSVField.Birthday) {
		birthday, errResForEachLineCSV = importStudentCSVValidateBirthday(importStudentCSVField.Birthday.String(), order)
		if errResForEachLineCSV != nil {
			return studentCSV, errResForEachLineCSV, nil
		}
		if err := multierr.Combine(
			student.Birthday.Set(birthday),
			student.LegacyUser.Birthday.Set(birthday),
		); err != nil {
			return studentCSV, nil, fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
	}

	if field.IsPresent(importStudentCSVField.PhoneNumber) {
		phoneNumber = importStudentCSVField.PhoneNumber.String()
		errResForEachLineCSV = importStudentCSVValidatePhoneNumber(phoneNumber, order, countryCode)
		if errResForEachLineCSV != nil {
			return studentCSV, errResForEachLineCSV, nil
		}
		err := student.LegacyUser.PhoneNumber.Set(phoneNumber)
		if err != nil {
			return studentCSV, nil, fmt.Errorf("student.User.PhoneNumber.Set: %v, row: %v", err, order)
		}
	}

	if field.IsPresent(importStudentCSVField.Location) {
		location := importStudentCSVField.Location.String()
		locations = strings.Split(location, ";")

		studentCSV.Locations, err = s.UserModifierService.GetLocationsByPartnerInternalIDs(ctx, locations)
		if err != nil {
			zapLogger.Warnf("get location by partner err: %v, row: %v", err, order)
			return studentCSV, convertErrToErrResForEachLineCSV(errNotFollowTemplate, locationStudentCSVHeader, order), nil
		}

		errResForEachLineCSV, err = importStudentCSVValidateLocation(studentCSV.Locations, order)
		if errResForEachLineCSV != nil || err != nil {
			return studentCSV, errResForEachLineCSV, err
		}
	}

	if field.IsPresent(importStudentCSVField.StudentTag) {
		tagSequence := importStudentCSVField.StudentTag.String()
		tags = strings.Split(tagSequence, ";")

		if isDuplicatedIDs(tags) {
			return studentCSV, convertErrToErrResForEachLineCSV(errNotFollowTemplate, tagStudentCSVHeader, order), nil
		}

		studentCSV.Tags, err = s.DomainTagRepo.GetByPartnerInternalIDs(ctx, s.DB, tags)
		if err != nil {
			zapLogger.Warnf("get tag by ids err: %v, row: %v", err, order)
			return studentCSV, convertErrToErrResForEachLineCSV(errNotFollowTemplate, tagStudentCSVHeader, order), err
		}

		if errResForEachLineCSV := importCsvValidateTag(constant.RoleStudent, tags, studentCSV.Tags); errResForEachLineCSV != nil {
			return studentCSV, convertErrToErrResForEachLineCSV(errResForEachLineCSV, tagStudentCSVHeader, order), nil
		}
	}

	schoolID, err := strconv.Atoi(resourcePath)
	if err != nil {
		return studentCSV, nil, fmt.Errorf("strconv.Atoi: %v, row: %v", err, order)
	}

	id = idutil.ULIDNow()
	if err = multierr.Combine(
		student.LegacyUser.ID.Set(id),
		student.LegacyUser.Email.Set(importStudentCSVField.Email),
		student.LegacyUser.FullName.Set(CombineFirstNameAndLastNameToFullName(importStudentCSVField.FirstName.String(), importStudentCSVField.LastName.String())),
		student.LegacyUser.FirstName.Set(importStudentCSVField.FirstName),
		student.LegacyUser.LastName.Set(importStudentCSVField.LastName),
		student.LegacyUser.Group.Set(entity.UserGroupStudent),
		student.LegacyUser.Country.Set(countryCode),
		student.LegacyUser.ResourcePath.Set(resourcePath),

		student.ID.Set(id),
		student.EnrollmentStatus.Set(enrollmentStatuses[0]), // backward compatible select first enrollment of first location -> first index
		student.StudentNote.Set(""),
		student.CurrentGrade.Set(grade),
		student.SchoolID.Set(schoolID),
		student.ResourcePath.Set(resourcePath),
	); err != nil {
		return studentCSV, nil, fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
	}
	if gradeID != "" {
		if err := student.GradeID.Set(gradeID); err != nil {
			return studentCSV, nil, fmt.Errorf("student.GradeID.Set: %v, row: %v", err, order)
		}
	}
	if err := importStudentCSVNormalizePhoneticName(student, importStudentCSVField.FirstNamePhonetic, importStudentCSVField.LastNamePhonetic, order); err != nil {
		return studentCSV, nil, err
	}
	// school histories epic
	schoolHistories, errResForEachLineCSV, err := s.importStudentNormalizeSchoolHistory(ctx, order, student.LegacyUser.ID.String, resourcePath, importStudentCSVField)
	if errResForEachLineCSV != nil || err != nil {
		return studentCSV, errResForEachLineCSV, err
	}
	studentCSV.SchoolHistories = schoolHistories
	if importStudentError, err := s.importStudentNormalizeStudentAddress(ctx, order, student.LegacyUser.ID.String, resourcePath, importStudentCSVField, userAddress); importStudentError != nil || err != nil {
		return studentCSV, importStudentError, err
	}
	// phone number epic
	studentPhoneNumbers, errResForEachLineCSV, err := importStudentCSVValidateAndGetUserPhoneNumbers(order, student.LegacyUser.ID.String, resourcePath, importStudentCSVField, student)
	if errResForEachLineCSV != nil || err != nil {
		return studentCSV, errResForEachLineCSV, err
	}
	studentCSV.StudentPhoneNumbers = studentPhoneNumbers

	enrollmentStatusHistories, errResForEachLineCSV, err := s.importStudentNormalizeEnrollmentStatusHistories(ctx, student.LegacyUser.ID.String, resourcePath, order, importStudentCSVField)
	if errResForEachLineCSV != nil || err != nil {
		return studentCSV, errResForEachLineCSV, err
	}
	studentCSV.EnrollmentStatusHistories = enrollmentStatusHistories

	return studentCSV, nil, nil
}

func validateRequiredFieldImportStudentCSV(
	importStudentCSVField *ImportStudentCSVField,
	order int,
) *pb.ImportStudentResponse_ImportStudentError {

	switch {
	case !field.IsPresent(importStudentCSVField.FirstName):
		return convertErrToErrResForEachLineCSV(fmt.Errorf("missingMandatory"), firstNameStudentCSVHeader, order)
	case !field.IsPresent(importStudentCSVField.LastName):
		return convertErrToErrResForEachLineCSV(fmt.Errorf("missingMandatory"), lastNameStudentCSVHeader, order)
	case !field.IsPresent(importStudentCSVField.Email):
		return convertErrToErrResForEachLineCSV(fmt.Errorf("missingMandatory"), emailStudentCSVHeader, order)
	case !field.IsPresent(importStudentCSVField.EnrollmentStatus):
		return convertErrToErrResForEachLineCSV(fmt.Errorf("missingMandatory"), enrollmentStatusStudentCSVHeader, order)
	case !field.IsPresent(importStudentCSVField.Grade):
		return convertErrToErrResForEachLineCSV(fmt.Errorf("missingMandatory"), gradeStudentCSVHeader, order)
	case !field.IsPresent(importStudentCSVField.Location):
		return convertErrToErrResForEachLineCSV(fmt.Errorf("missingMandatory"), locationStudentCSVHeader, order)
	}
	return nil
}

func (s *StudentService) importStudentNormalizeEnrollmentStatusHistories(
	ctx context.Context,
	userID string,
	resourcePath string,
	order int,
	importStudentCSVField *ImportStudentCSVField,
) (entity.DomainEnrollmentStatusHistories, *pb.ImportStudentResponse_ImportStudentError, error) {
	var locations, enrollmentStatuses, statusStartDates []string
	getConfigReq := &mpb.GetConfigurationByKeyRequest{Key: constant.KeyEnrollmentStatusHistoryConfig}

	location := importStudentCSVField.Location.String()
	locations = strings.Split(location, ";")

	enrollmentStatus := importStudentCSVField.EnrollmentStatus.String()
	enrollmentStatuses = strings.Split(enrollmentStatus, ";")
	resp, err := s.ConfigurationClient.GetConfigurationByKey(ctx, getConfigReq)
	if err != nil {
		return nil, nil, err
	}

	// Get configuration to detect logic lms or erp
	configurationResp := resp.GetConfiguration()
	if configurationResp == nil {
		return nil, nil, fmt.Errorf("not found config for org: %s", resourcePath)
	}
	if configurationResp.GetConfigValue() == constant.ConfigValueOff {
		if err := validateEnrollmentStatusCSVCreateRequest(enrollmentStatuses, order); err != nil {
			return nil, err, nil
		}
	}

	if field.IsPresent(importStudentCSVField.StatusStartDate) {
		statusStartDate := importStudentCSVField.StatusStartDate.String()
		statusStartDates = strings.Split(statusStartDate, ";")
	}

	if len(statusStartDates) != 0 {
		if len(locations) != len(enrollmentStatuses) ||
			len(enrollmentStatuses) != len(statusStartDates) {
			studentCSVHeaderError := checkInvalidEnrollmentStatusHistoriesDataImport(len(locations), len(enrollmentStatuses), len(statusStartDates))
			if studentCSVHeaderError != "" {
				return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, studentCSVHeaderError, order), nil
			}
		}
	} else {
		if len(locations) != len(enrollmentStatuses) {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, enrollmentStatusStudentCSVHeader, order), nil
		}
	}

	locationsIDs, err := s.DomainLocationRepo.GetByPartnerInternalIDs(ctx, s.DB, locations)
	if err != nil {
		return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, locationStudentCSVHeader, order), nil
	}
	enrollmentStatusHistoryWillBeDelegated := []entity.EnrollmentStatusHistoryWillBeDelegated{}

	for idx, domainLocation := range locationsIDs {
		startDate := time.Now()
		if len(statusStartDates) != 0 {
			startDate, err = time.Parse(constant.DateLayout, statusStartDates[idx])
			if err != nil {
				return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, startDateStudentCSVHeader, order), nil
			}
		}

		enrollmentStatusInt, err := strconv.Atoi(enrollmentStatuses[idx])
		if err != nil {
			return nil, nil, fmt.Errorf("strconv.Atoi %v", enrollmentStatuses[idx])
		}
		domainEnrollmentStatusHistory := NewEnrollmentStatusHistoryCSV(http_port.StudentEnrollmentStatusMap[enrollmentStatusInt], userID, resourcePath, startDate)
		enrollmentStatusHistoryWillBeDelegated = append(enrollmentStatusHistoryWillBeDelegated, entity.EnrollmentStatusHistoryWillBeDelegated{
			EnrollmentStatusHistory: domainEnrollmentStatusHistory,
			HasLocationID:           domainLocation,
			HasUserID:               domainEnrollmentStatusHistory,
			HasOrganizationID:       domainEnrollmentStatusHistory,
		})
	}

	enrollmentStatusHistories := []entity.DomainEnrollmentStatusHistory{}
	for j := range enrollmentStatusHistoryWillBeDelegated {
		enrollmentStatusHistories = append(enrollmentStatusHistories, &enrollmentStatusHistoryWillBeDelegated[j])
	}

	return enrollmentStatusHistories, nil, nil
}

func validateEnrollmentStatusCSVCreateRequest(enrollmentStatuses []string, order int) *pb.ImportStudentResponse_ImportStudentError {
	allowList := []string{
		pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
		pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
		pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
	}

	for _, enrollmentStatus := range enrollmentStatuses {
		enrollmentStatusInt, err := strconv.Atoi(enrollmentStatus)
		if err != nil {
			return convertErrToErrResForEachLineCSV(errNotFollowTemplate, enrollmentStatusStudentCSVHeader, order)
		}
		if !golibs.InArrayString(http_port.StudentEnrollmentStatusMap[enrollmentStatusInt], allowList) {
			return convertErrToErrResForEachLineCSV(errNotFollowTemplate, enrollmentStatusStudentCSVHeader, order)
		}
	}

	return nil
}

func (s *StudentService) importStudentGetGradeFromGradeMaster(
	ctx context.Context,
	order int,
	importStudentCSVField *ImportStudentCSVField,
) (int32, string, *pb.ImportStudentResponse_ImportStudentError) {
	gradeMaster, err := s.UserModifierService.GetGradeMaster(ctx, importStudentCSVField.Grade.String())
	var grade int32
	var gradeID string
	if err != nil || len(gradeMaster) == 0 {
		gradeInt, errResForEachLineCSV := importStudentCSVValidateGrade(importStudentCSVField.Grade.String(), order)
		if errResForEachLineCSV != nil {
			return 0, "", errResForEachLineCSV
		}
		grade = int32(gradeInt)
		gradeOrgs, err := s.GradeOrganizationRepo.GetByGradeValues(ctx, s.DB, []int32{grade})
		if err != nil {
			return 0, "", convertErrToErrResForEachLineCSV(errNotFollowTemplate, gradeStudentCSVHeader, order)
		}

		if len(gradeOrgs) > 0 {
			gradeID = gradeOrgs[0].GradeID().String()
		}
	} else if len(gradeMaster) > 0 {
		for k, v := range gradeMaster {
			if !(field.IsNull(v) && field.IsUndefined(v)) {
				grade = v.Int32()
			}
			gradeID = k.GradeID().String()
		}
	}
	return grade, gradeID, nil
}

func (s *StudentService) importStudentNormalizeStudentAddress(
	ctx context.Context,
	order int,
	userId string,
	resourcePath string,
	importStudentCSVField *ImportStudentCSVField,
	userAddress *entity.UserAddress,
) (*pb.ImportStudentResponse_ImportStudentError, error) {
	if !field.IsPresent(importStudentCSVField.City) &&
		!field.IsPresent(importStudentCSVField.PostalCode) &&
		!field.IsPresent(importStudentCSVField.Prefecture) &&
		!field.IsPresent(importStudentCSVField.FirstStreet) &&
		!field.IsPresent(importStudentCSVField.SecondStreet) {
		return nil, nil
	}
	prefectureID := ""
	if field.IsPresent(importStudentCSVField.Prefecture) {
		prefectureEnt, err := s.PrefectureRepo.GetByPrefectureCode(ctx, s.DB, database.Text(importStudentCSVField.Prefecture.String()))
		if err != nil {
			return convertErrToErrResForEachLineCSV(errNotFollowTemplate, prefectureStudentCSVHeader, order), nil
		}
		prefectureID = prefectureEnt.ID.String
	}
	if prefectureID == "" {
		_ = userAddress.PrefectureID.Set(sql.NullString{})
	} else {
		_ = userAddress.PrefectureID.Set(prefectureID)
	}
	if err := multierr.Combine(
		userAddress.UserAddressID.Set(idutil.ULIDNow()),
		userAddress.AddressType.Set(pb.AddressType_HOME_ADDRESS),
		userAddress.UserID.Set(userId),
		userAddress.PostalCode.Set(importStudentCSVField.PostalCode),
		userAddress.City.Set(importStudentCSVField.City),
		userAddress.FirstStreet.Set(importStudentCSVField.FirstStreet),
		userAddress.SecondStreet.Set(importStudentCSVField.SecondStreet),
		userAddress.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
	}
	return nil, nil
}

func (s *StudentService) importStudentNormalizeSchoolHistory(
	ctx context.Context,
	order int,
	userId string,
	resourcePath string,
	importStudentCSVField *ImportStudentCSVField,
) ([]*entity.SchoolHistory, *pb.ImportStudentResponse_ImportStudentError, error) {
	if !field.IsPresent(importStudentCSVField.School) &&
		!field.IsPresent(importStudentCSVField.SchoolCourse) &&
		!field.IsPresent(importStudentCSVField.StartDate) &&
		!field.IsPresent(importStudentCSVField.EndDate) {
		return nil, nil, nil
	}

	schoolHistories := []*entity.SchoolHistory{}
	schoolCoursesEntities := []*entity.SchoolCourse{}
	var schools, schoolCourses, startDates, endDates []string

	// require
	school := importStudentCSVField.School.String()
	schools = strings.Split(school, ";")

	// optional
	schoolCourse := importStudentCSVField.SchoolCourse.String()
	if schoolCourse != "" {
		schoolCourses = strings.Split(schoolCourse, ";")
	}

	// optional
	startDate := importStudentCSVField.StartDate.String()
	if startDate != "" {
		startDates = strings.Split(startDate, ";")
	}

	// optional
	endDate := importStudentCSVField.EndDate.String()
	if endDate != "" {
		endDates = strings.Split(endDate, ";")
	}

	if len(schools) != len(schoolCourses) ||
		len(schoolCourses) != len(startDates) ||
		len(startDates) != len(endDates) {
		studentCSVHeaderError := checkInvalidSchoolHistoryDataImport(len(schools), len(schoolCourses), len(startDates), len(endDates))
		if studentCSVHeaderError != "" {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, studentCSVHeaderError, order), nil
		}
	}

	schoolInfos, err := s.SchoolInfoRepo.GetBySchoolPartnerIDs(ctx, s.DB, database.TextArray(schools))
	if err != nil {
		return nil, nil, fmt.Errorf("s.SchoolInfoRepo.GetBySchoolPartnerIDs: %v", err)
	}
	if len(schoolInfos) != len(schools) {
		return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, schoolStudentCSVHeader, order), nil
	}

	schoolIDs := []string{}
	levelWithSchoolInfo := map[string]string{}
	for _, schoolInfo := range schoolInfos {
		if schoolInfo.IsArchived.Bool {
			return nil, nil, fmt.Errorf("school_info %v is archived", schoolInfo.ID.String)
		}
		if _, ok := levelWithSchoolInfo[schoolInfo.LevelID.String]; ok {
			return nil, nil, fmt.Errorf("duplicate school_level_id in school_info %v", schoolInfo.ID.String)
		}
		levelWithSchoolInfo[schoolInfo.LevelID.String] = schoolInfo.ID.String
		schoolIDs = append(schoolIDs, schoolInfo.ID.String)
	}

	if len(schoolCourses) != 0 && !checkSchoolCourseEmptyValue(schoolCourses) {
		schoolCoursesEntities, err = s.SchoolCourseRepo.GetBySchoolCoursePartnerIDsAndSchoolIDs(ctx, s.DB, database.TextArray(schoolCourses), database.TextArray(schoolIDs))
		if err != nil {
			return nil, nil, fmt.Errorf("schoolCourseRepo.GetBySchoolCoursePartnerIDsAndSchoolIDs: %v", err)
		}
		if len(schoolCoursesEntities) != len(schoolCourses) {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, schoolCourseStudentCSVHeader, order), nil
		}

		for _, schoolCourse := range schoolCoursesEntities {
			if schoolCourse.IsArchived.Bool {
				return nil, nil, fmt.Errorf("school_course %v is archived", schoolCourse.ID.String)
			}
		}
	}

	for i, schoolInfo := range schoolInfos {
		var err error
		schoolHistoryEntity := &entity.SchoolHistory{}
		database.AllNullEntity(schoolHistoryEntity)
		if schools[i] == "" {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, schoolStudentCSVHeader, order), nil
		}
		if len(startDates) != 0 && len(endDates) != 0 {
			var startDate, endDate time.Time
			switch {
			case startDates[i] != "" && endDates[i] != "":
				startDate, err = time.Parse(constant.DateLayout, startDates[i])
				if err != nil {
					return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, startDateStudentCSVHeader, order), nil
				}
				endDate, err = time.Parse(constant.DateLayout, endDates[i])
				if err != nil {
					return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, endDateStudentCSVHeader, order), nil
				}
				if startDate.After(endDate) {
					return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, endDateStudentCSVHeader, order), nil
				}
				if err := schoolHistoryEntity.StartDate.Set(startDate); err != nil {
					return nil, nil, fmt.Errorf("importStudentNormalizeSchoolHistory schoolHistoryEntity.StartDate.Set: %v", err)
				}
				if err := schoolHistoryEntity.EndDate.Set(endDate); err != nil {
					return nil, nil, fmt.Errorf("importStudentNormalizeSchoolHistory schoolHistoryEntity.EndDate.Set: %v", err)
				}
			case startDates[i] == "" && endDates[i] != "":
				endDate, err = time.Parse(constant.DateLayout, endDates[i])
				if err != nil {
					return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, endDateStudentCSVHeader, order), nil
				}
				if err := schoolHistoryEntity.EndDate.Set(endDate); err != nil {
					return nil, nil, fmt.Errorf("importStudentNormalizeSchoolHistory schoolHistoryEntity.StartDate.Set: %v", err)
				}
			case startDates[i] != "" && endDates[i] == "":
				startDate, err = time.Parse(constant.DateLayout, startDates[i])
				if err != nil {
					return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, startDateStudentCSVHeader, order), nil
				}
				if err := schoolHistoryEntity.StartDate.Set(startDate); err != nil {
					return nil, nil, fmt.Errorf("importStudentNormalizeSchoolHistory schoolHistoryEntity.EndDate.Set: %v", err)
				}
			}
		}

		if len(schoolCoursesEntities) != 0 {
			if err := schoolHistoryEntity.SchoolCourseID.Set(schoolCoursesEntities[i].ID.String); err != nil {
				return nil, nil, fmt.Errorf("importStudentNormalizeSchoolHistory schoolHistoryEntity.SchoolCourseID.Set: %v", err)
			}
		}
		if err := multierr.Combine(
			schoolHistoryEntity.StudentID.Set(userId),
			schoolHistoryEntity.ResourcePath.Set(resourcePath),
			schoolHistoryEntity.IsCurrent.Set(false),
			schoolHistoryEntity.SchoolID.Set(schoolInfo.ID.String),
		); err != nil {
			return nil, nil, fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
		schoolHistories = append(schoolHistories, schoolHistoryEntity)
	}

	return schoolHistories, nil, nil
}

func checkSchoolCourseEmptyValue(schoolCourses []string) bool {
	for _, schoolCourse := range schoolCourses {
		if schoolCourse != "" {
			return false
		}
	}
	return true
}

func checkInvalidSchoolHistoryDataImport(schools, schoolCourses, startDates, endDates int) studentCSVHeader {
	var arrCheck = []int{schools, schoolCourses, startDates, endDates}

	// Count Number of Element Occurrences
	keys := countElementOccurrencesInSlice(arrCheck)
	uniqueValue := findUniqueValue(keys, len(arrCheck))

	switch {
	case uniqueValue == schools && schoolCourses != 0 && startDates != 0 && endDates != 0:
		return schoolStudentCSVHeader
	case uniqueValue == schools && schoolCourses != 0 && startDates != 0:
		return schoolStudentCSVHeader
	case uniqueValue == schools && schoolCourses != 0 && schools != schoolCourses:
		return schoolStudentCSVHeader
	case uniqueValue == schoolCourses && schoolCourses != 0 && schools != schoolCourses:
		return schoolCourseStudentCSVHeader
	case uniqueValue == startDates && startDates != 0 && schools != startDates:
		return startDateStudentCSVHeader
	case uniqueValue == endDates && endDates != 0 && schools != endDates:
		return endDateStudentCSVHeader
	}
	return ""
}

func checkInvalidEnrollmentStatusHistoriesDataImport(locations, enrollmentStatuses, statusStartDates int) studentCSVHeader {
	var arrCheck = []int{locations, enrollmentStatuses, statusStartDates}

	// Count Number of Element Occurrences
	keys := countElementOccurrencesInSlice(arrCheck)
	uniqueValue := findUniqueValue(keys, len(arrCheck))

	switch {
	case uniqueValue == locations && enrollmentStatuses != 0 && statusStartDates != 0:
		return locationStudentCSVHeader
	case uniqueValue == locations && enrollmentStatuses != 0 && locations != enrollmentStatuses:
		return locationStudentCSVHeader
	case uniqueValue == enrollmentStatuses && enrollmentStatuses != 0 && locations != enrollmentStatuses:
		return enrollmentStatusStudentCSVHeader
	case uniqueValue == statusStartDates && statusStartDates != 0 && locations != statusStartDates:
		return statusStartDateStudentCSVHeader
	}
	return ""
}

func importStudentGetUserPhoneNumberEntity(userID string, phoneNumber string, phoneNumberType string, resourcePath string) (*entity.UserPhoneNumber, error) {
	studentPhoneNumber := &entity.UserPhoneNumber{}
	database.AllNullEntity(studentPhoneNumber)
	if err := multierr.Combine(
		studentPhoneNumber.ID.Set(idutil.ULIDNow()),
		studentPhoneNumber.UserID.Set(userID),
		studentPhoneNumber.PhoneNumber.Set(phoneNumber),
		studentPhoneNumber.PhoneNumberType.Set(phoneNumberType),
		studentPhoneNumber.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, err
	}
	return studentPhoneNumber, nil
}

func importStudentCSVValidateAndGetUserPhoneNumbers(
	order int,
	userID string,
	resourcePath string,
	importStudentCSVField *ImportStudentCSVField,
	student *entity.LegacyStudent,
) ([]*entity.UserPhoneNumber, *pb.ImportStudentResponse_ImportStudentError, error) {
	var userPhoneNumbers []*entity.UserPhoneNumber
	studentPhoneNumberCSV, studentHomePhoneNumberCSV := importStudentCSVField.StudentPhoneNumber, importStudentCSVField.StudentHomePhoneNumber
	studentContactPreferenceCSV := importStudentCSVField.StudentContactPreference
	if field.IsPresent(studentPhoneNumberCSV) {
		err := MatchingRegex(PhoneNumberPattern, studentPhoneNumberCSV.String())
		if err != nil {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, studentPhoneNumberCSVHeader, order), nil
		}
		studentPhoneNumber, err := importStudentGetUserPhoneNumberEntity(userID, studentPhoneNumberCSV.String(), entity.StudentPhoneNumber, resourcePath)
		if err != nil {
			return nil, nil, err
		}
		userPhoneNumbers = append(userPhoneNumbers, studentPhoneNumber)
	}
	if field.IsPresent(studentHomePhoneNumberCSV) {
		err := MatchingRegex(PhoneNumberPattern, studentHomePhoneNumberCSV.String())
		if err != nil {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, studentHomePhoneNumberCSVHeader, order), nil
		}
		studentHomePhoneNumber, err := importStudentGetUserPhoneNumberEntity(userID, studentHomePhoneNumberCSV.String(), entity.StudentHomePhoneNumber, resourcePath)
		if err != nil {
			return nil, nil, err
		}
		userPhoneNumbers = append(userPhoneNumbers, studentHomePhoneNumber)
	}
	if field.IsPresent(studentPhoneNumberCSV) && field.IsPresent(studentHomePhoneNumberCSV) {
		if studentPhoneNumberCSV.String() == studentHomePhoneNumberCSV.String() {
			return nil, convertErrToErrResForEachLineCSV(fmt.Errorf("duplicationRow"), studentPhoneNumberCSVHeader, order), nil
		}
	}
	if field.IsPresent(studentContactPreferenceCSV) {
		contactPreference, ok := mapStudentContactPreference[studentContactPreferenceCSV.String()]
		if !ok {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, studentContactPreferenceCSVHeader, order), nil
		}
		if err := student.ContactPreference.Set(contactPreference); err != nil {
			return nil, nil, err
		}
	}
	return userPhoneNumbers, nil, nil
}

func importStudentCSVNormalizePhoneticName(student *entity.LegacyStudent, firstNamePhoneticCSV field.String, lastNamePhoneticCSV field.String, order int) error {
	// if firstNamePhoneticCSV != nil && firstNamePhoneticCSV.Exist {
	if field.IsPresent(firstNamePhoneticCSV) {
		if err := student.LegacyUser.FirstNamePhonetic.Set(firstNamePhoneticCSV); err != nil {
			return fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
	}
	// if lastNamePhoneticCSV != nil && lastNamePhoneticCSV.Exist {
	if field.IsPresent(lastNamePhoneticCSV) {
		if err := student.LegacyUser.LastNamePhonetic.Set(lastNamePhoneticCSV); err != nil {
			return fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
	}
	fullNamePhonetic := CombineFirstNamePhoneticAndLastNamePhoneticToFullName(firstNamePhoneticCSV.String(), lastNamePhoneticCSV.String())
	if fullNamePhonetic != "" {
		if err := student.LegacyUser.FullNamePhonetic.Set(fullNamePhonetic); err != nil {
			return fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
	}
	return nil
}

func convertErrToErrResForEachLineCSV(err error, fieldName studentCSVHeader, i int) *pb.ImportStudentResponse_ImportStudentError {
	return &pb.ImportStudentResponse_ImportStudentError{
		RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
		Error:     err.Error(),
		FieldName: string(fieldName),
	}
}

func importStudentCSVValidateEmail(email string, order int) (*pb.ImportStudentResponse_ImportStudentError, error) {
	match, err := regexp.MatchString(emailPattern, email)
	if err != nil {
		return nil, fmt.Errorf("regexp.MatchString: %v, row: %v", err, order)
	}
	if !match {
		return convertErrToErrResForEachLineCSV(errNotFollowTemplate, emailStudentCSVHeader, order), nil
	}
	return nil, nil
}

func importStudentCSVValidateEnrollmentStatus(enrollmentStatusCSV []string, order int) ([]string, *pb.ImportStudentResponse_ImportStudentError) {
	enrollmentStatusStrings := []string{}
	for _, enrollmentStatus := range enrollmentStatusCSV {
		enrollmentStatusString, ok := studentEnrollmentStatusMap[enrollmentStatus]
		if !ok {
			return nil, convertErrToErrResForEachLineCSV(errNotFollowTemplate, enrollmentStatusStudentCSVHeader, order)
		}
		enrollmentStatusStrings = append(enrollmentStatusStrings, enrollmentStatusString)
	}

	return enrollmentStatusStrings, nil
}

func importStudentCSVValidateGrade(gradeCSV string, order int) (int, *pb.ImportStudentResponse_ImportStudentError) {
	grade, err := strconv.Atoi(gradeCSV)
	if err != nil {
		return 0, convertErrToErrResForEachLineCSV(errNotFollowTemplate, gradeStudentCSVHeader, order)
	}
	if grade < 0 || grade > 16 {
		return 0, convertErrToErrResForEachLineCSV(errNotFollowTemplate, gradeStudentCSVHeader, order)
	}
	return grade, nil
}

func importStudentCSVValidateGender(genderCSV string, order int) (string, *pb.ImportStudentResponse_ImportStudentError) {
	genderInt, err := strconv.Atoi(genderCSV)
	gender := ""
	if err != nil {
		return "", convertErrToErrResForEachLineCSV(errNotFollowTemplate, genderStudentCSVHeader, order)
	}
	switch g := pb.Gender(genderInt); g {
	case pb.Gender_MALE,
		pb.Gender_FEMALE:
		gender = g.String()
	default:
		return "", convertErrToErrResForEachLineCSV(errNotFollowTemplate, genderStudentCSVHeader, order)
	}
	return gender, nil
}

func importStudentCSVValidateBirthday(birthdayCSV string, order int) (time.Time, *pb.ImportStudentResponse_ImportStudentError) {
	birthday, err := time.Parse(constant.DateLayout, birthdayCSV)
	if err != nil {
		return time.Time{}, convertErrToErrResForEachLineCSV(errNotFollowTemplate, birthdayStudentCSVHeader, order)
	}
	return birthday, nil
}

func importStudentCSVValidatePhoneNumber(phoneNumberCSV string, order int, countryCode string) *pb.ImportStudentResponse_ImportStudentError {
	num, err := phonenumbers.Parse(phoneNumberCSV, strings.Split(countryCode, "_")[1])
	if err != nil || !phonenumbers.IsValidNumber(num) {
		return convertErrToErrResForEachLineCSV(errNotFollowTemplate, phoneNumberStudentCSVHeader, order)
	}
	return nil
}

func importStudentCSVValidateLocation(locations []*domain.Location, order int) (*pb.ImportStudentResponse_ImportStudentError, error) {
	for _, l := range locations {
		if l.IsArchived {
			return convertErrToErrResForEachLineCSV(errNotFollowTemplate, locationStudentCSVHeader, order), nil
		}
	}
	return nil, nil
}

func importCsvValidateTag(role string, partnerInternalIDs []string, tags entity.DomainTags) error {
	handleError := func(role string) error {
		err := errNotFollowTemplate
		if role == constant.RoleParent {
			return errNotFollowParentTemplate
		}
		return err
	}

	if ok := tags.ContainPartnerInternalIDs(partnerInternalIDs...); !ok {
		return handleError(role)
	}

	for _, tag := range tags {
		if tag.IsArchived().Boolean() {
			return handleError(role)
		}

		if role == constant.RoleParent && !entity.IsParentTag(tag) {
			return errNotFollowParentTemplate
		}

		if role == constant.RoleStudent && !entity.IsStudentTag(tag) {
			return errNotFollowTemplate
		}
	}

	return nil
}

func isDuplicatedIDs(ids []string) bool {
	mapIDs := map[string]struct{}{}
	for _, ids := range ids {
		mapIDs[ids] = struct{}{}
	}

	return len(ids) != len(mapIDs)
}
