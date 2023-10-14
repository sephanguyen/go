package usermgmt

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var STUDENT_ENROLLMENT_STATUS_STRING_EMPTY = ""

func (s *suite) updateStudentSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleUpdateUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &pb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *pb.UpdateStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *pb.EvtUser_UpdateStudent_:
				if req.StudentProfile.Id == msg.UpdateStudent.StudentId && req.StudentProfile.FirstName == msg.UpdateStudent.StudentFirstName && req.StudentProfile.LastName == msg.UpdateStudent.StudentLastName {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleUpdateUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("updateStudentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountDataToUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.ExistingStudents = []*entity.LegacyStudent{student}
	s.addUpdatingStudentProfileToUpdateStudentRequest(ctx, student, STUDENT_ENROLLMENT_STATUS_STRING_EMPTY)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountDataToUpdateWithFirstNameLastNameAndPhoneticName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.ExistingStudents = []*entity.LegacyStudent{student}
	stepState.Request = generateUpdateStudentRequestWithFirstNameAndLastName(student, STUDENT_ENROLLMENT_STATUS_STRING_EMPTY, []string{s.ExistingLocations[0].LocationID.String})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) cannotUpdateStudentAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting err but got nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateStudentAccount(ctx context.Context, signedInUser, ability string) (context.Context, error) {
	switch ability {
	case "can":
		return s.studentAccountIsUpdatedSuccess(ctx)
	case "cannot":
		return s.cannotUpdateStudentAccount(ctx)
	default:
		return ctx, nil
	}
}

func (s *suite) studentAccountIsUpdatedSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	req := stepState.Request.(*pb.UpdateStudentRequest)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	// Get updated student
	updatedStudent, err := (&repository.StudentRepo{}).Find(ctx, s.BobDBTrace, database.Text(req.StudentProfile.Id))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	user, err := (&repository.UserRepo{}).Get(ctx, s.BobDBTrace, database.Text(req.StudentProfile.Id))
	if err != nil {
		return StepStateToContext(ctx, stepState), status.Error(codes.InvalidArgument, "user id is not exists")
	}
	updatedStudent.LegacyUser = *user

	// update student response fields must be equal with request fields
	if err := s.validateUpdateStudentResponse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// Updated fields must be equal with requested fields
	if err := s.validateUpdatedStudentInfo(ctx, updatedStudent); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Updated student still can login with old password
	if err := s.loginIdentityPlatform(ctx, auth.LocalTenants[constants.ManabieSchool], updatedStudent.Email.String, s.ExistingStudents[0].Password); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "loginIdentityPlatform")
	}

	if err := s.validParentAccessPath(ctx, updatedStudent.ID.String); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return StepStateToContext(ctx, stepState), nil
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) studentAccountWithFirstNameLastNameAndPhoneticNameUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	req := stepState.Request.(*pb.UpdateStudentRequest)

	// Get updated student
	updatedStudent, err := (&repository.StudentRepo{}).Find(ctx, s.BobDBTrace, database.Text(req.StudentProfile.Id))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	user, err := (&repository.UserRepo{}).Get(ctx, s.BobDBTrace, database.Text(req.StudentProfile.Id))
	if err != nil {
		return StepStateToContext(ctx, stepState), status.Error(codes.InvalidArgument, "user id is not exists")
	}
	updatedStudent.LegacyUser = *user

	// update student response fields must be equal with request fields
	if err := s.validateUpdateStudentResponseWithFirstNameLastNameAndPhoneticName(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// Updated fields must be equal with requested fields
	if err := s.validateUpdatedStudentInfoWithFirstNameLastNameAndPhoneticName(ctx, updatedStudent); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Updated student still can login with old password
	if err := s.loginIdentityPlatform(ctx, auth.LocalTenants[constants.ManabieSchool], updatedStudent.Email.String, s.ExistingStudents[0].Password); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "loginIdentityPlatform")
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return StepStateToContext(ctx, stepState), nil
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) updateStudentAccount(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, account)
	req := stepState.Request.(*pb.UpdateStudentRequest)

	ctx, err := s.updateStudentSubscription(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.updateStudentSubscription: %w", err)
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = tryUpdateStudent(ctx, s.UserMgmtConn, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateStudentAccountThatHas(ctx context.Context, caller string, existingData string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateStudentRequest)

	switch existingData {
	case "unknown student enrollment status":
		req.StudentProfile.EnrollmentStatus = pb.StudentEnrollmentStatus(999999)
	}

	return s.updateStudentAccount(ctx, caller)
}

func (s *suite) updateStudentEmailThatExistInOurSystem(ctx context.Context, caller string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	req := stepState.Request.(*pb.UpdateStudentRequest)

	student, err := s.createStudent(ctx)
	if err != nil {
		return ctx, err
	}
	s.ExistingStudents = []*entity.LegacyStudent{student}
	req.StudentProfile.Email = student.Email.String

	return s.updateStudentAccount(ctx, caller)
}

//nolint:goconst
func (s *suite) updatesStudentAccountWithNewStudentDataMissingField(ctx context.Context, caller string, missingField string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateStudentRequest)

	switch missingField {
	case "name":
		req.StudentProfile.Name = ""
	case "enrollmentStatus":
		req.StudentProfile.EnrollmentStatus = pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE
	case "studentExternalId":
		req.StudentProfile.StudentExternalId = ""
	case "studentNote":
		req.StudentProfile.StudentNote = ""
	case "email":
		req.StudentProfile.Email = ""
	case "birthday":
		req.StudentProfile.Birthday = nil
	case "gender":
		req.StudentProfile.Gender = pb.Gender_NONE
	case "location_ids":
		req.StudentProfile.LocationIds = []string{}
	}

	return s.updateStudentAccount(ctx, caller)
}

func (s *suite) assignCoursePackageWithToExistStudent(ctx context.Context, studentPackageStatus string) (context.Context, error) {
	s.MapExistingPackageAndCourses = make(map[string]string)

	var (
		startAt, endAt *timestamppb.Timestamp
		location       string
	)
	switch studentPackageStatus {
	case "active":
		startAt = timestamppb.New(time.Now())
		endAt = timestamppb.New(time.Now().Add(30 * 24 * time.Hour))
		location = constants.ManabieOrgLocation
	case "inactive":
		startAt = timestamppb.New(time.Now().Add(-30 * 24 * time.Hour))
		endAt = timestamppb.New(time.Now().Add(-24 * time.Hour))
		location = constants.ManabieOrgLocation
	case "active with valid location":
		startAt = timestamppb.New(time.Now())
		endAt = timestamppb.New(time.Now().Add(30 * 24 * time.Hour))
		location = s.ExistingLocations[0].LocationID.String
	}

	coursePackages := []*fpb.AddStudentPackageCourseRequest{
		{
			StudentId:   s.ExistingStudents[0].ID.String,
			CourseIds:   []string{"existing-course-id-1"},
			StartAt:     startAt,
			EndAt:       endAt,
			LocationIds: []string{location},
		},
		{
			StudentId:   s.ExistingStudents[0].ID.String,
			CourseIds:   []string{"existing-course-id-2"},
			StartAt:     startAt,
			EndAt:       endAt,
			LocationIds: []string{location},
		},
	}

	for _, req := range coursePackages {
		resp, err := tryAddStudentPackage(ctx, s.FatimaConn, req)
		if err != nil {
			return ctx, err
		}

		s.MapExistingPackageAndCourses[resp.StudentPackageId] = req.CourseIds[0]
	}

	return ctx, nil
}

func (s *suite) updateStudentAccountDoesNotExist(ctx context.Context, caller string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateStudentRequest)
	req.StudentProfile.Id = fmt.Sprintf("%s-non-exist", s.ExistingStudents[0].ID.String)

	return s.updateStudentAccount(ctx, caller)
}

func (s *suite) createStudent(ctx context.Context) (*entity.LegacyStudent, error) {
	resp, err := CreateStudent(ctx, s.UserMgmtConn, nil, getChildrenLocation(OrgIDFromCtx(ctx)))
	if err != nil {
		return nil, fmt.Errorf("createStudent: %w", err)
	}

	student, err := (&repository.StudentRepo{}).Find(ctx, s.BobDBTrace, database.Text(resp.StudentProfile.Student.UserProfile.UserId))
	if err != nil {
		return nil, fmt.Errorf("createStudent.Find: %w", err)
	}

	user, err := (&repository.UserRepo{}).Get(ctx, s.BobDBTrace, student.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id is not exists")
	}
	student.LegacyUser = *user
	student.LegacyUser.UserAdditionalInfo.Password = student.GetUID()

	// We don't auth with phone number fow now
	userToImport := student.LegacyUser
	if err := userToImport.PhoneNumber.Set(nil); err != nil {
		return nil, fmt.Errorf("userToImport.PhoneNumber.Set(nil): %w", err)
	}

	err = s.importUserToFirebaseAndIdentityPlatform(ctx, &student.LegacyUser, []byte(student.LegacyUser.UserAdditionalInfo.Password), []byte(newID()))
	if err != nil {
		return nil, errors.Wrap(err, "importUserToFirebaseAndIdentityPlatform")
	}

	return student, nil
}

func (s *suite) addUpdatingStudentProfileToUpdateStudentRequest(ctx context.Context, student *entity.LegacyStudent, enrollmentStatus string) {
	stepState := StepStateFromContext(ctx)
	uid := newID()

	req := &pb.UpdateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                  student.ID.String,
			Name:                fmt.Sprintf("updated-%s", student.GetName()),
			Grade:               int32(1),
			EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			EnrollmentStatusStr: enrollmentStatus,
			StudentExternalId:   fmt.Sprintf("student-external-id-%v", uid),
			StudentNote:         fmt.Sprintf("some random student note edited %v", uid),
			Email:               fmt.Sprintf("student-email-edited-%s@example.com", uid),
			Birthday:            timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:              pb.Gender_MALE,
			LocationIds:         []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	stepState.Request = req
}

func (s *suite) studentAccountDataToUpdateHasInvalidLocations(ctx context.Context, invalidType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.ExistingStudents = []*entity.LegacyStudent{student}
	s.addUpdatingStudentProfileToUpdateStudentRequest(ctx, student, STUDENT_ENROLLMENT_STATUS_STRING_EMPTY)
	req := stepState.Request.(*pb.UpdateStudentRequest)

	switch invalidType {
	case "empty":
		req.StudentProfile.LocationIds = []string{""}
	case "not found":
		req.StudentProfile.LocationIds = []string{"location-not-existed"}
	case "invalid resource_path":
		req.StudentProfile.LocationIds = []string{s.ExistingLocations[1].LocationID.String}
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateUpdatedStudentInfoWithFirstNameLastNameAndPhoneticName(ctx context.Context, updatedStudent *entity.LegacyStudent) error {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateStudentRequest)

	switch {
	case updatedStudent.FullName.String != helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName):
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student name, expected: "%s", actual: "%s"`, helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName), updatedStudent.GetName())
	case updatedStudent.FirstName.String != req.StudentProfile.FirstName:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student first name, expected: "%s", actual: "%s"`, req.StudentProfile.FirstName, updatedStudent.FirstName.String)
	case updatedStudent.LastName.String != req.StudentProfile.LastName:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student last name, expected: "%s", actual: "%s"`, req.StudentProfile.LastName, updatedStudent.LastName.String)
	case updatedStudent.FirstNamePhonetic.String != req.StudentProfile.FirstNamePhonetic:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student first name phonetic, expected: "%s", actual: "%s"`, req.StudentProfile.FirstNamePhonetic, updatedStudent.FirstNamePhonetic.String)
	case updatedStudent.LastNamePhonetic.String != req.StudentProfile.LastNamePhonetic:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student last name phonetic, expected: "%s", actual: "%s"`, req.StudentProfile.LastNamePhonetic, updatedStudent.LastNamePhonetic.String)
	case updatedStudent.FullNamePhonetic.String != helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic):
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student full name phonetic, expected: "%s", actual: "%s"`, helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic), updatedStudent.FullNamePhonetic.String)
	case int32(updatedStudent.CurrentGrade.Int) != req.StudentProfile.Grade:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student grade, expected: "%v", actual: "%v"`, req.StudentProfile.Grade, updatedStudent.CurrentGrade.Int)
	case updatedStudent.EnrollmentStatus.String != req.StudentProfile.EnrollmentStatus.String():
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student enrollment status, expected: "%v", actual: "%v"`, req.StudentProfile.EnrollmentStatus.String(), updatedStudent.EnrollmentStatus.String)
	case updatedStudent.StudentExternalID.String != req.StudentProfile.StudentExternalId:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student external id, expected: "%v", actual: "%v"`, req.StudentProfile.StudentExternalId, updatedStudent.StudentExternalID.String)
	case updatedStudent.StudentNote.String != req.StudentProfile.StudentNote:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student note, expected: "%v", actual: "%v"`, req.StudentProfile.StudentNote, updatedStudent.StudentNote.String)
	case updatedStudent.Email.String != req.StudentProfile.Email:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student email, expected: "%v", actual: "%v"`, req.StudentProfile.Email, updatedStudent.Email.String)
	case updatedStudent.LegacyUser.Gender.String != "" && req.StudentProfile.Gender == pb.Gender_NONE:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student gender: nil but actual is %v`, updatedStudent.LegacyUser.Gender.String)
	case updatedStudent.LegacyUser.Gender.String != req.StudentProfile.Gender.String() && req.StudentProfile.Gender != pb.Gender_NONE:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student gender, expected: %v but actual is %v`, req.StudentProfile.Gender.String(), updatedStudent.LegacyUser.Gender.String)
	case updatedStudent.LegacyUser.Birthday.Status != pgtype.Null && req.StudentProfile.Birthday == nil:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student birthday, expected: nil but actual is %v`, updatedStudent.LegacyUser.Birthday.Time.Format(CommonDateLayout))
	case updatedStudent.LegacyUser.Birthday.Time.Format(CommonDateLayout) != req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) && req.StudentProfile.Birthday != nil:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student birthday, expected: %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), updatedStudent.LegacyUser.Birthday.Time.Format(CommonDateLayout))
	}

	// Other student fields must not be changed
	preInitializedStudent := s.ExistingStudents[0]
	switch {
	case updatedStudent.PhoneNumber.String != preInitializedStudent.PhoneNumber.String:
		return fmt.Errorf(`validateUpdatedStudentInfo: expected student's phone number: "%s", actual "%s"`, preInitializedStudent.PhoneNumber.String, updatedStudent.PhoneNumber.String)
	case updatedStudent.Country.String != preInitializedStudent.Country.String:
		return fmt.Errorf(`validateUpdatedStudentInfo: expected student's country: "%s", actual "%s"`, preInitializedStudent.Country.String, updatedStudent.Country.String)
	case updatedStudent.Avatar.String != preInitializedStudent.Avatar.String:
		return fmt.Errorf(`validateUpdatedStudentInfo: expected student's avatart: "%s", actual "%s"`, preInitializedStudent.Avatar.String, updatedStudent.Avatar.String)
	}

	if err := s.validateLocationStored(ctx, updatedStudent.ID.String, req.StudentProfile.LocationIds); err != nil {
		return fmt.Errorf(`validateCreateStudentInfo: %s`, err.Error())
	}

	return nil
}

func (s *suite) validateUpdatedStudentInfo(ctx context.Context, updatedStudent *entity.LegacyStudent) error {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateStudentRequest)

	switch {
	case updatedStudent.FullName.String != req.StudentProfile.Name:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student name, expected: "%s", actual: "%s"`, req.StudentProfile.Name, updatedStudent.GetName())
	case int32(updatedStudent.CurrentGrade.Int) != req.StudentProfile.Grade:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student grade, expected: "%v", actual: "%v"`, req.StudentProfile.Grade, updatedStudent.CurrentGrade.Int)
	case updatedStudent.EnrollmentStatus.String != req.StudentProfile.EnrollmentStatus.String():
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student enrollment status, expected: "%v", actual: "%v"`, req.StudentProfile.EnrollmentStatus.String(), updatedStudent.EnrollmentStatus.String)
	case updatedStudent.StudentExternalID.String != req.StudentProfile.StudentExternalId:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student external id, expected: "%v", actual: "%v"`, req.StudentProfile.StudentExternalId, updatedStudent.StudentExternalID.String)
	case updatedStudent.StudentNote.String != req.StudentProfile.StudentNote:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student note, expected: "%v", actual: "%v"`, req.StudentProfile.StudentNote, updatedStudent.StudentNote.String)
	case updatedStudent.Email.String != req.StudentProfile.Email:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student email, expected: "%v", actual: "%v"`, req.StudentProfile.Email, updatedStudent.Email.String)
	case updatedStudent.LegacyUser.Gender.String != "" && req.StudentProfile.Gender == pb.Gender_NONE:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student gender: nil but actual is %v`, updatedStudent.LegacyUser.Gender.String)
	case updatedStudent.LegacyUser.Gender.String != req.StudentProfile.Gender.String() && req.StudentProfile.Gender != pb.Gender_NONE:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student gender, expected: %v but actual is %v`, req.StudentProfile.Gender.String(), updatedStudent.LegacyUser.Gender.String)
	case updatedStudent.LegacyUser.Birthday.Status != pgtype.Null && req.StudentProfile.Birthday == nil:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student birthday, expected: nil but actual is %v`, updatedStudent.LegacyUser.Birthday.Time.Format(CommonDateLayout))
	case updatedStudent.LegacyUser.Birthday.Time.Format(CommonDateLayout) != req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) && req.StudentProfile.Birthday != nil:
		return fmt.Errorf(`validateUpdatedStudentInfo: failed to update student birthday, expected: %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), updatedStudent.LegacyUser.Birthday.Time.Format(CommonDateLayout))
	}

	// Other student fields must not be changed
	preInitializedStudent := s.ExistingStudents[0]
	switch {
	case updatedStudent.PhoneNumber.String != preInitializedStudent.PhoneNumber.String:
		return fmt.Errorf(`validateUpdatedStudentInfo: expected student's phone number: "%s", actual "%s"`, preInitializedStudent.PhoneNumber.String, updatedStudent.PhoneNumber.String)
	case updatedStudent.Country.String != preInitializedStudent.Country.String:
		return fmt.Errorf(`validateUpdatedStudentInfo: expected student's country: "%s", actual "%s"`, preInitializedStudent.Country.String, updatedStudent.Country.String)
	case updatedStudent.Avatar.String != preInitializedStudent.Avatar.String:
		return fmt.Errorf(`validateUpdatedStudentInfo: expected student's avatart: "%s", actual "%s"`, preInitializedStudent.Avatar.String, updatedStudent.Avatar.String)
	}

	if err := s.validateLocationStored(ctx, updatedStudent.LegacyUser.ID.String, req.StudentProfile.LocationIds); err != nil {
		return fmt.Errorf(`validateCreateStudentInfo: %s`, err.Error())
	}

	return nil
}

func (s *suite) validateUpdateStudentResponseWithFirstNameLastNameAndPhoneticName(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateStudentRequest)
	resp := stepState.Response.(*pb.UpdateStudentResponse)

	switch {
	case req.StudentProfile.FirstName != resp.StudentProfile.FirstName:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student first name, expected: "%s", actual: "%s"`, req.StudentProfile.FirstName, resp.StudentProfile.FirstName)
	case req.StudentProfile.LastName != resp.StudentProfile.LastName:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student last name, expected: "%s", actual: "%s"`, req.StudentProfile.LastName, resp.StudentProfile.LastName)
	case helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName) != resp.StudentProfile.Name:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student name, expected: "%s", actual: "%s"`, helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName), resp.StudentProfile.Name)
	case req.StudentProfile.FirstNamePhonetic != resp.StudentProfile.FirstNamePhonetic:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student first name phonetic, expected: "%s", actual: "%s"`, req.StudentProfile.FirstNamePhonetic, resp.StudentProfile.FirstNamePhonetic)
	case req.StudentProfile.LastNamePhonetic != resp.StudentProfile.LastNamePhonetic:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student last name phonetic, expected: "%s", actual: "%s"`, req.StudentProfile.LastNamePhonetic, resp.StudentProfile.LastNamePhonetic)
	case helper.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic) != resp.StudentProfile.FullNamePhonetic:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student full name phonetic, expected: "%s", actual: "%s"`, helper.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic), resp.StudentProfile.FullNamePhonetic)
	case req.StudentProfile.Grade != resp.StudentProfile.Grade:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student grade, expected: "%v", actual: "%v"`, req.StudentProfile.Grade, resp.StudentProfile.Grade)
	case req.StudentProfile.EnrollmentStatus != resp.StudentProfile.EnrollmentStatus:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student enrollment status, expected: "%v", actual: "%v"`, req.StudentProfile.EnrollmentStatus.String(), resp.StudentProfile.EnrollmentStatus.String())
	case req.StudentProfile.StudentExternalId != resp.StudentProfile.StudentExternalId:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student external id, expected: "%v", actual: "%v"`, req.StudentProfile.StudentExternalId, resp.StudentProfile.StudentExternalId)
	case req.StudentProfile.StudentNote != resp.StudentProfile.StudentNote:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student note, expected: "%v", actual: "%v"`, req.StudentProfile.StudentNote, resp.StudentProfile.StudentNote)
	case req.StudentProfile.Email != resp.StudentProfile.Email:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student email, expected: "%v", actual: "%v"`, req.StudentProfile.Email, resp.StudentProfile.Email)
	case req.StudentProfile.Gender != resp.StudentProfile.Gender:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student gender, expected: %v but actual is %v`, req.StudentProfile.Gender.String(), resp.StudentProfile.Gender.String())
	case req.StudentProfile.Birthday == nil && resp.StudentProfile.Birthday != nil:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student birthday, expected: nil but actual is %v`, resp.StudentProfile.Birthday.AsTime().Format(CommonDateLayout))
	case req.StudentProfile.Birthday != nil && req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) != resp.StudentProfile.Birthday.AsTime().Format(CommonDateLayout):
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student birthday, expected: %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), resp.StudentProfile.Birthday.AsTime().Format(CommonDateLayout))
	case len(req.StudentProfile.LocationIds) != len(resp.StudentProfile.LocationIds) && (!reflect.DeepEqual(req.StudentProfile.LocationIds, resp.StudentProfile.LocationIds)):
		return fmt.Errorf(`validateUpdateStudentResponse: expected response "locationIDs": %v but actual is %v `, req.StudentProfile.LocationIds, resp.StudentProfile.LocationIds)
	}

	return nil
}

func (s *suite) validateUpdateStudentResponse(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateStudentRequest)
	resp := stepState.Response.(*pb.UpdateStudentResponse)

	switch {
	case req.StudentProfile.Name != resp.StudentProfile.Name:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student name, expected: "%s", actual: "%s"`, req.StudentProfile.Name, resp.StudentProfile.Name)
	case req.StudentProfile.Grade != resp.StudentProfile.Grade:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student grade, expected: "%v", actual: "%v"`, req.StudentProfile.Grade, resp.StudentProfile.Grade)
	case req.StudentProfile.EnrollmentStatus != resp.StudentProfile.EnrollmentStatus:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student enrollment status, expected: "%v", actual: "%v"`, req.StudentProfile.EnrollmentStatus.String(), resp.StudentProfile.EnrollmentStatus.String())
	case req.StudentProfile.StudentExternalId != resp.StudentProfile.StudentExternalId:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student external id, expected: "%v", actual: "%v"`, req.StudentProfile.StudentExternalId, resp.StudentProfile.StudentExternalId)
	case req.StudentProfile.StudentNote != resp.StudentProfile.StudentNote:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student note, expected: "%v", actual: "%v"`, req.StudentProfile.StudentNote, resp.StudentProfile.StudentNote)
	case req.StudentProfile.Email != resp.StudentProfile.Email:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student email, expected: "%v", actual: "%v"`, req.StudentProfile.Email, resp.StudentProfile.Email)
	case req.StudentProfile.Gender != resp.StudentProfile.Gender:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student gender, expected: %v but actual is %v`, req.StudentProfile.Gender.String(), resp.StudentProfile.Gender.String())
	case req.StudentProfile.Birthday == nil && resp.StudentProfile.Birthday != nil:
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student birthday, expected: nil but actual is %v`, resp.StudentProfile.Birthday.AsTime().Format(CommonDateLayout))
	case req.StudentProfile.Birthday != nil && req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) != resp.StudentProfile.Birthday.AsTime().Format(CommonDateLayout):
		return fmt.Errorf(`validateUpdateStudentResponse: failed to update student birthday, expected: %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), resp.StudentProfile.Birthday.AsTime().Format(CommonDateLayout))
	case len(req.StudentProfile.LocationIds) != len(resp.StudentProfile.LocationIds) && (!reflect.DeepEqual(req.StudentProfile.LocationIds, resp.StudentProfile.LocationIds)):
		return fmt.Errorf(`validateUpdateStudentResponse: expected response "locationIDs": %v but actual is %v `, req.StudentProfile.LocationIds, resp.StudentProfile.LocationIds)
	}

	return nil
}

func (s *suite) studentAccountDataToUpdateWithEnrollmentStatusString(ctx context.Context, enrollmentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	if enrollmentStatus == "STUDENT_ENROLLMENT_STATUS_STRING_EMPTY" {
		enrollmentStatus = STUDENT_ENROLLMENT_STATUS_STRING_EMPTY
	}
	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.ExistingStudents = []*entity.LegacyStudent{student}
	s.addUpdatingStudentProfileToUpdateStudentRequest(ctx, student, enrollmentStatus)
	return StepStateToContext(ctx, stepState), nil
}

func generateUpdateStudentRequestWithFirstNameAndLastName(student *entity.LegacyStudent, enrollmentStatus string, locationIDs []string) *pb.UpdateStudentRequest {
	uid := newID()

	return &pb.UpdateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                  student.ID.String,
			Grade:               int32(1),
			FirstName:           fmt.Sprintf("updated-first-name-%s", student.GetName()),
			LastName:            fmt.Sprintf("updated-last-name-%s", student.GetName()),
			FirstNamePhonetic:   fmt.Sprintf("updated-first-name-phonetic-%s", student.GetName()),
			LastNamePhonetic:    fmt.Sprintf("updated-last-name-phonetic-%s", student.GetName()),
			EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			EnrollmentStatusStr: enrollmentStatus,
			StudentExternalId:   fmt.Sprintf("student-external-id-%v", uid),
			StudentNote:         fmt.Sprintf("some random student note edited %v", uid),
			Email:               fmt.Sprintf("student-email-edited-%s@example.com", uid),
			Birthday:            timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:              pb.Gender_MALE,
			LocationIds:         locationIDs,
		},
	}
}

func tryUpdateStudent(ctx context.Context, client *grpc.ClientConn, req *pb.UpdateStudentRequest) (*pb.UpdateStudentResponse, error) {
	var (
		resp = &pb.UpdateStudentResponse{}
		err  error
	)

	err = try.Do(func(attempt int) (bool, error) {
		resp, err = pb.NewUserModifierServiceClient(client).UpdateStudent(ctx, req)
		if err == nil {
			return false, nil
		}
		if attempt < retryTimes {
			time.Sleep(time.Second)
			return true, err
		}
		return false, err
	})

	return resp, err
}

func (s *suite) existedStudentProfileWithTags(ctx context.Context, tagType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	ctx, err := s.studentProfileWithTags(ctx, tagType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	resp, err := pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.TagIDs = req.StudentProfile.GetTagIds()
	s.ExistingStudents = []*entity.LegacyStudent{{
		ID: database.Text(resp.GetStudentProfile().Student.UserProfile.GetUserId()),
	}}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateStudentTags(ctx context.Context, typeerwfr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := generateUpdateStudentRequestWithFirstNameAndLastName(s.ExistingStudents[0], STUDENT_ENROLLMENT_STATUS_STRING_EMPTY, []string{s.ExistingLocations[0].LocationID.String})

	// re-assign tags from created student
	req.StudentProfile.TagIds = s.TagIDs

	newTagIDs, _, err := s.createAmountTags(ctx, amountSampleTestElement, pb.UserTagType_USER_TAG_TYPE_STUDENT.String(), fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch typeerwfr {
	case "add more":
		req.GetStudentProfile().TagIds = append(req.GetStudentProfile().TagIds, newTagIDs...)

	case "remove one":
		req.GetStudentProfile().TagIds = req.GetStudentProfile().TagIds[1:]

	case "remove one & add more":
		req.GetStudentProfile().TagIds = append(req.GetStudentProfile().TagIds[1:], newTagIDs...)

	case "remove all":
		req.GetStudentProfile().TagIds = []string{}

	case "not found":
		req.GetStudentProfile().TagIds = []string{idutil.ULIDNow()}

	case "tag for only parent":
		tagIDs, _, err := s.createTagsType(ctx, parentType)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.GetStudentProfile().TagIds = tagIDs
	}

	// assign for checking result
	s.TagIDs = req.GetStudentProfile().TagIds
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentUpdateSuccessWithUserTags(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	mapExistTags := map[string]struct{}{}
	updateStudentRequest := stepState.Request.(*pb.UpdateStudentRequest)
	taggedUserRepo := new(repository.DomainTaggedUserRepo)
	taggedUsers, err := taggedUserRepo.GetByUserIDs(ctx, s.BobDB, []string{updateStudentRequest.StudentProfile.Id})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, taggedUser := range taggedUsers {
		mapExistTags[taggedUser.TagID().String()] = struct{}{}
	}

	for _, tagID := range s.TagIDs {
		if _, ok := mapExistTags[tagID]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("user %s is missing %s tag", s.ExistingStudents[0].GetUID(), tagID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountDataToUpdateWithParentInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	existingStudent, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := NewCreateParentReqWithOnlyParentInfo(existingStudent)

	_, err = pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.importUserToFirebaseAndIdentityPlatform(ctx, &existingStudent.LegacyUser, []byte(existingStudent.LegacyUser.UserAdditionalInfo.Password), []byte(newID()))
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "importUserToFirebaseAndIdentityPlatform")
	}

	s.ExistingStudents = []*entity.LegacyStudent{existingStudent}
	s.addUpdatingStudentProfileToUpdateStudentRequest(ctx, existingStudent, STUDENT_ENROLLMENT_STATUS_STRING_EMPTY)
	return StepStateToContext(ctx, stepState), nil
}
