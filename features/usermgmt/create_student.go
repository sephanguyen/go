package usermgmt

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc/importstudent"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gocarina/gocsv"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	amountSampleTestElement = 10
)

func (s *suite) cannotCreateThatAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return ctx, errors.New("expected response has err but actual is nil")
	}
	return ctx, nil
}

func (s *suite) createStudentSubscription(ctx context.Context, req interface{}) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 2)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleUpsertUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &pb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := req.(type) {
		case *pb.CreateStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *pb.EvtUser_CreateParent_:
				if req.StudentProfile.Name == msg.CreateParent.StudentName {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}

			case *pb.EvtUser_CreateStudent_:
				if req.StudentProfile.Name == msg.CreateStudent.StudentName {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		case *pb.ImportStudentRequest:
			studentCSVs := []importstudent.StudentCSV{}
			err := gocsv.UnmarshalBytes(req.Payload, &studentCSVs)
			if err != nil {
				return true, errors.Wrap(err, "gocsv.UnmarshalBytes")
			}
			switch msg := evtUser.Message.(type) {
			case *pb.EvtUser_CreateStudent_:
				for _, student := range studentCSVs {
					if student.FirstNameAttr.String() == msg.CreateStudent.StudentFirstName && student.LastNameAttr.String() == msg.CreateStudent.StudentLastName {
						locations := []string{}
						if !student.LocationAttr.IsEmpty() {
							locations = strings.Split(student.LocationAttr.String(), ";")
						}
						if len(locations) == len(msg.CreateStudent.LocationIds) {
							stepState.FoundChanForJetStream <- evtUser.Message
							return true, nil
						}
					}
				}
			case *pb.EvtUser_UpdateStudent_:
				for _, student := range studentCSVs {
					if student.FirstNameAttr.String() == msg.UpdateStudent.StudentFirstName && student.LastNameAttr.String() == msg.UpdateStudent.StudentLastName {
						locations := []string{}
						if !student.LocationAttr.IsEmpty() {
							locations = strings.Split(student.LocationAttr.String(), ";")
						}
						if len(locations) == len(msg.UpdateStudent.LocationIds) {
							stepState.FoundChanForJetStream <- evtUser.Message
							return true, nil
						}
					}
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createStudentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)

	subs, err = s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleUpsertUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createStudentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentWithFirstNameAndLastNameSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleCreateUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &pb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *pb.CreateStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *pb.EvtUser_CreateStudent_:
				if req.StudentProfile.FirstName == msg.CreateStudent.StudentFirstName && req.StudentProfile.LastName == msg.CreateStudent.StudentLastName {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleCreateUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createStudentSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createNewStudentAccount(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, account)

	ctx, err := s.createStudentSubscription(ctx, stepState.Request)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createStudentSubscription: %w", err)
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(ctx, stepState.Request.(*pb.CreateStudentRequest))

	if stepState.ResponseErr == nil {
		stepState.CurrentStudentID = stepState.
			Response.(*pb.CreateStudentResponse).
			GetStudentProfile().
			GetStudent().
			GetUserProfile().
			GetUserId()
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createNewStudentAccountWithInvalidResourcePath(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	ctx, err := s.createStudentSubscription(StepStateToContext(ctx, stepState), stepState.Requests)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createStudentSubscription: %w", err)
	}
	ctx = s.signedIn(ctx, constants.JPREPSchool, account)
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(contextWithToken(auth.InjectFakeJwtToken(ctx, "")), stepState.Request.(*pb.CreateStudentRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newStudentAccountCreatedSuccessWithStudentInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	// create student response fields must be equal with request fields
	ctx, err := s.validateCreateStudentResponse(ctx)
	if err != nil {
		return ctx, err
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return s.validateCreatedStudentInfo(ctx)
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) newStudentAccountCreatedSuccessWithStudentInfoAndFirstNameLastNamePhoneticName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	// create student response fields must be equal with request fields
	ctx, err := s.validateStudentResponseWithFirstNameAndLastNameAndPhoneticName(ctx)
	if err != nil {
		return ctx, err
	}
	select {
	case <-stepState.FoundChanForJetStream:
		return s.validateCreatedStudentInfo(ctx)
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) onlyStudentInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := newID()

	req := &pb.CreateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			Name:              fmt.Sprintf("user-%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func createStudentReqWithOnlyStudentInfo(schoolID int32, locationID string) *pb.CreateStudentRequest {
	randomID := newID()
	req := &pb.CreateStudentRequest{
		SchoolId: schoolID,
		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			Name:              fmt.Sprintf("user-%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{locationID},
		},
	}
	return req
}

func (s *suite) studentInfoWithFirstNameLastNameAndPhoneticName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := newID()
	req := &pb.CreateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			FirstName:         fmt.Sprintf("user-first-name-%v", randomID),
			LastName:          fmt.Sprintf("user-last-name-%v", randomID),
			FirstNamePhonetic: fmt.Sprintf("user-first-name-phonetic%v", randomID),
			LastNamePhonetic:  fmt.Sprintf("user-last-name-phonetic%v", randomID),
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	stepState.Request = req
	ctx, err := s.createStudentWithFirstNameAndLastNameSubscription(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createStudentWithFirstNameAndLastNameSubscription: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesStatusCode(ctx context.Context, expectedCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	if stt.Code().String() != expectedCode {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", expectedCode, stt.Code().String(), stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentDataMissing(ctx context.Context, missingField string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.onlyStudentInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*pb.CreateStudentRequest)

	switch missingField {
	case "username":
		req.StudentProfile.Email = ""
	case "password":
		req.StudentProfile.Password = ""
	case "name":
		req.StudentProfile.Name = ""
	case "enrollmentStatus":
		req.StudentProfile.EnrollmentStatus = pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE
	case "studentExternalId":
		req.StudentProfile.StudentExternalId = ""
	case "studentNote":
		req.StudentProfile.StudentNote = ""
	case "birthday":
		req.StudentProfile.Birthday = nil
	case "gender":
		req.StudentProfile.Gender = pb.Gender_NONE
	case "location_ids":
		req.StudentProfile.LocationIds = []string{}
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentDataWithUnknownStudentEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.onlyStudentInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*pb.CreateStudentRequest)
	req.StudentProfile.EnrollmentStatus = pb.StudentEnrollmentStatus(999999) // some unknown enrollment status

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInfoWithInvalidLocations(ctx context.Context, invalidType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.onlyStudentInfo(ctx)
	if err != nil {
		return ctx, err
	}
	req := stepState.Request.(*pb.CreateStudentRequest)

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

func (s *suite) validateCreatedStudentInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(ctx)
	currentResourcePath := golibs.ResourcePathFromCtx(ctx)
	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	stmt :=
		`
		SELECT 
			users.user_id,
			users.email,
			users.name,
			users.first_name,
			users.last_name,
			users.country,
			users.phone_number,
			users.birthday,
			users.gender,
			students.school_id,
			students.current_grade,
			students.enrollment_status,
			students.student_external_id,
			students.student_note,
			students.resource_path
		FROM
			users
		JOIN 
			students ON users.user_id = students.student_id  
		JOIN 
			users_groups ON users.user_id = users_groups.user_id
		WHERE 
			users.user_id = $1
			AND users.resource_path = $2
			AND students.resource_path = $2
		`

	row := s.BobDBTrace.QueryRow(
		ctx,
		stmt,
		resp.StudentProfile.Student.UserProfile.UserId,
		fmt.Sprint(req.SchoolId),
	)

	student := &entity.LegacyStudent{}
	if err := row.Scan(
		&student.ID,
		&student.Email,
		&student.GivenName,
		&student.FirstName,
		&student.LastName,
		&student.Country,
		&student.PhoneNumber,
		&student.Birthday,
		&student.Gender,
		&student.SchoolID,
		&student.CurrentGrade,
		&student.EnrollmentStatus,
		&student.StudentExternalID,
		&student.StudentNote,
		&student.ResourcePath,
	); err != nil {
		return ctx, err
	}

	if req.StudentProfile.FirstName != "" {
		switch {
		case req.StudentProfile.FirstName != student.FirstName.String:
			return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "first_name": %v but actual is %v`, req.StudentProfile.FirstName, student.FirstName)
		case req.StudentProfile.LastName != student.LastName.String:
			return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "last_name": %v but actual is %v`, req.StudentProfile.LastName, student.LastName)
		case helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName) != student.GivenName.String:
			return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "given_name": %v but actual is %v`, helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName), student.GivenName)
		}
	} else if req.StudentProfile.Name != student.GivenName.String {
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "given_name": %v but actual is %v`, req.StudentProfile.Name, student.GivenName)
	}

	switch {
	case req.StudentProfile.Email != student.Email.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "email": %v but actual is %v`, req.StudentProfile.Email, student.Email)
	case req.StudentProfile.CountryCode.String() != student.Country.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "country": %v but actual is %v`, req.StudentProfile.CountryCode.String(), student.Country.String)
	case req.StudentProfile.PhoneNumber != student.PhoneNumber.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "phone_number": %v but actual is %v`, req.StudentProfile.PhoneNumber, student.PhoneNumber.String)
	case req.SchoolId != student.SchoolID.Int:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "school_id": %v but actual is %v`, req.SchoolId, student.SchoolID.Int)
	case req.StudentProfile.Grade != int32(student.CurrentGrade.Int):
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "current_grade": %v but actual is %v`, req.StudentProfile.Grade, int32(student.CurrentGrade.Int))
	case req.StudentProfile.EnrollmentStatus.String() != student.EnrollmentStatus.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "enrollment_status": %v but actual is %v`, req.StudentProfile.EnrollmentStatus.String(), student.EnrollmentStatus.String)
	case req.StudentProfile.StudentExternalId != student.StudentExternalID.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "student_external_id": %v but actual is %v`, req.StudentProfile.StudentExternalId, student.StudentExternalID.String)
	case req.StudentProfile.StudentNote != student.StudentNote.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "student_note": %v but actual is %v`, req.StudentProfile.StudentNote, student.StudentNote.String)
	case req.StudentProfile.Gender == pb.Gender_NONE && student.LegacyUser.Gender.String != "":
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "gender": nil but actual is %v`, student.LegacyUser.Gender.String)
	case req.StudentProfile.Gender != pb.Gender_NONE && req.StudentProfile.Gender.String() != student.LegacyUser.Gender.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "gender": %v but actual is %v`, req.StudentProfile.Gender.String(), student.LegacyUser.Gender.String)
	case req.StudentProfile.Birthday == nil && student.LegacyUser.Birthday.Status != pgtype.Null:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "birthday": nil but actual is %v`, student.LegacyUser.Birthday.Time.Format(CommonDateLayout))
	case req.StudentProfile.Birthday != nil && req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) != student.LegacyUser.Birthday.Time.Format(CommonDateLayout):
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "birthday": %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), student.LegacyUser.Birthday.Time.Format(CommonDateLayout))
	case currentResourcePath != student.ResourcePath.String:
		return ctx, fmt.Errorf(`validateCreateStudentInfo: expected inserted "resource_path": %v but actual is %v`, currentResourcePath, student.ResourcePath.String)
	}

	// verify that student was created with assigned student user group and has student role
	if err := s.validateUsersHasUserGroupWithRole(ctx, []string{student.ID.String}, currentResourcePath, constant.RoleStudent); err != nil {
		return ctx, fmt.Errorf("s.userHasUserGroupWithRole: %v", err)
	}

	if err := s.validateLocationStored(ctx, student.ID.String, req.StudentProfile.LocationIds); err != nil {
		return ctx, fmt.Errorf(`validateCreateStudentInfo: %s`, err.Error())
	}

	if err := s.loginIdentityPlatform(ctx, auth.LocalTenants[constants.ManabieSchool], req.StudentProfile.Email, req.StudentProfile.Password); err != nil {
		return ctx, errors.Wrap(err, "loginIdentityPlatform")
	}

	if err := s.validateUserTags(ctx, student.ID.String, req.StudentProfile.TagIds); err != nil {
		return ctx, errors.Wrap(err, "validateUserTags")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateCreateStudentResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	switch {
	case req.StudentProfile.Email != resp.StudentProfile.Student.UserProfile.Email:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "email": %v but actual is %v`, req.StudentProfile.Email, resp.StudentProfile.Student.UserProfile.Email)
	case req.StudentProfile.Name != resp.StudentProfile.Student.UserProfile.Name:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "given_name": %v but actual is %v`, req.StudentProfile.Name, resp.StudentProfile.Student.UserProfile.Name)
	case req.StudentProfile.CountryCode != resp.StudentProfile.Student.UserProfile.CountryCode:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "country": %v but actual is %v`, req.StudentProfile.CountryCode.String(), resp.StudentProfile.Student.UserProfile.CountryCode.String())
	case req.StudentProfile.PhoneNumber != resp.StudentProfile.Student.UserProfile.PhoneNumber:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "phone_number": %v but actual is %v`, req.StudentProfile.PhoneNumber, resp.StudentProfile.Student.UserProfile.PhoneNumber)
	case req.SchoolId != resp.StudentProfile.Student.SchoolId:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "school_id": %v but actual is %v`, req.SchoolId, resp.StudentProfile.Student.SchoolId)
	case req.StudentProfile.Grade != resp.StudentProfile.Student.Grade:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "current_grade": %v but actual is %v`, req.StudentProfile.Grade, resp.StudentProfile.Student.Grade)
	case req.StudentProfile.Gender != resp.StudentProfile.Student.UserProfile.Gender:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "gender": %v but actual is %v`, req.StudentProfile.Gender.String(), resp.StudentProfile.Student.UserProfile.Gender.String())
	case req.StudentProfile.Birthday == nil && resp.StudentProfile.Student.UserProfile.Birthday != nil:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "birthday": nil but actual is %v`, resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout))
	case req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) != resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout):
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "birthday": %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout))
	case len(req.StudentProfile.LocationIds) != len(resp.StudentProfile.Student.UserProfile.LocationIds) && (!reflect.DeepEqual(req.StudentProfile.LocationIds, resp.StudentProfile.Student.UserProfile.LocationIds)):
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "locationIDs": %v but actual is %v `, req.StudentProfile.LocationIds, resp.StudentProfile.Student.UserProfile.LocationIds)
	}

	return StepStateToContext(ctx, stepState), nil
}

func validateCreatedStudentReqAndResp(req *pb.CreateStudentRequest, resp *pb.CreateStudentResponse) error {
	switch {
	case req.StudentProfile.Email != resp.StudentProfile.Student.UserProfile.Email:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "email": %v but actual is %v`, req.StudentProfile.Email, resp.StudentProfile.Student.UserProfile.Email)
	case req.StudentProfile.Name != resp.StudentProfile.Student.UserProfile.Name:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "given_name": %v but actual is %v`, req.StudentProfile.Name, resp.StudentProfile.Student.UserProfile.Name)
	case req.StudentProfile.CountryCode != resp.StudentProfile.Student.UserProfile.CountryCode:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "country": %v but actual is %v`, req.StudentProfile.CountryCode.String(), resp.StudentProfile.Student.UserProfile.CountryCode.String())
	case req.StudentProfile.PhoneNumber != resp.StudentProfile.Student.UserProfile.PhoneNumber:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "phone_number": %v but actual is %v`, req.StudentProfile.PhoneNumber, resp.StudentProfile.Student.UserProfile.PhoneNumber)
	case req.SchoolId != resp.StudentProfile.Student.SchoolId:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "school_id": %v but actual is %v`, req.SchoolId, resp.StudentProfile.Student.SchoolId)
	case req.StudentProfile.Grade != resp.StudentProfile.Student.Grade:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "current_grade": %v but actual is %v`, req.StudentProfile.Grade, resp.StudentProfile.Student.Grade)
	case req.StudentProfile.Gender != resp.StudentProfile.Student.UserProfile.Gender:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "gender": %v but actual is %v`, req.StudentProfile.Gender.String(), resp.StudentProfile.Student.UserProfile.Gender.String())
	case req.StudentProfile.Birthday == nil && resp.StudentProfile.Student.UserProfile.Birthday != nil:
		return fmt.Errorf(`validateCreateStudentResponse: expected response "birthday": nil but actual is %v`, resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout))
	case req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) != resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout):
		return fmt.Errorf(`validateCreateStudentResponse: expected response "birthday": %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout))
	case len(req.StudentProfile.LocationIds) != len(resp.StudentProfile.Student.UserProfile.LocationIds) && (!reflect.DeepEqual(req.StudentProfile.LocationIds, resp.StudentProfile.Student.UserProfile.LocationIds)):
		return fmt.Errorf(`validateCreateStudentResponse: expected response "locationIDs": %v but actual is %v `, req.StudentProfile.LocationIds, resp.StudentProfile.Student.UserProfile.LocationIds)
	}
	return nil
}

func (s *suite) validateStudentResponseWithFirstNameAndLastNameAndPhoneticName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	switch {
	case req.StudentProfile.Email != resp.StudentProfile.Student.UserProfile.Email:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "email": %v but actual is %v`, req.StudentProfile.Email, resp.StudentProfile.Student.UserProfile.Email)
	case req.StudentProfile.FirstName != resp.StudentProfile.Student.UserProfile.FirstName:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "first_name": %v but actual is %v`, req.StudentProfile.FirstName, resp.StudentProfile.Student.UserProfile.FirstName)
	case req.StudentProfile.LastName != resp.StudentProfile.Student.UserProfile.LastName:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "last_name": %v but actual is %v`, req.StudentProfile.LastName, resp.StudentProfile.Student.UserProfile.LastName)
	case helper.CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName) != resp.StudentProfile.Student.UserProfile.Name:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "given_name": %v but actual is %v`, req.StudentProfile.Name, resp.StudentProfile.Student.UserProfile.Name)
	case req.StudentProfile.FirstNamePhonetic != resp.StudentProfile.Student.UserProfile.FirstNamePhonetic:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "first_name_phonetic": %v but actual is %v`, req.StudentProfile.FirstNamePhonetic, resp.StudentProfile.Student.UserProfile.FirstNamePhonetic)
	case req.StudentProfile.LastNamePhonetic != resp.StudentProfile.Student.UserProfile.LastNamePhonetic:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "last_name_phonetic": %v but actual is %v`, req.StudentProfile.LastNamePhonetic, resp.StudentProfile.Student.UserProfile.LastNamePhonetic)
	case helper.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic) != resp.StudentProfile.Student.UserProfile.FullNamePhonetic:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "full_name_phonetic": %v but actual is %v`, helper.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic), resp.StudentProfile.Student.UserProfile.FullNamePhonetic)
	case req.StudentProfile.CountryCode != resp.StudentProfile.Student.UserProfile.CountryCode:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "country": %v but actual is %v`, req.StudentProfile.CountryCode.String(), resp.StudentProfile.Student.UserProfile.CountryCode.String())
	case req.StudentProfile.PhoneNumber != resp.StudentProfile.Student.UserProfile.PhoneNumber:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "phone_number": %v but actual is %v`, req.StudentProfile.PhoneNumber, resp.StudentProfile.Student.UserProfile.PhoneNumber)
	case req.SchoolId != resp.StudentProfile.Student.SchoolId:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "school_id": %v but actual is %v`, req.SchoolId, resp.StudentProfile.Student.SchoolId)
	case req.StudentProfile.Grade != resp.StudentProfile.Student.Grade:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "current_grade": %v but actual is %v`, req.StudentProfile.Grade, resp.StudentProfile.Student.Grade)
	case req.StudentProfile.Gender != resp.StudentProfile.Student.UserProfile.Gender:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "gender": %v but actual is %v`, req.StudentProfile.Gender.String(), resp.StudentProfile.Student.UserProfile.Gender.String())
	case req.StudentProfile.Birthday == nil && resp.StudentProfile.Student.UserProfile.Birthday != nil:
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "birthday": nil but actual is %v`, resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout))
	case req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout) != resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout):
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "birthday": %v but actual is %v`, req.StudentProfile.Birthday.AsTime().Format(CommonDateLayout), resp.StudentProfile.Student.UserProfile.Birthday.AsTime().Format(CommonDateLayout))
	case len(req.StudentProfile.LocationIds) != len(resp.StudentProfile.Student.UserProfile.LocationIds) && (!reflect.DeepEqual(req.StudentProfile.LocationIds, resp.StudentProfile.Student.UserProfile.LocationIds)):
		return ctx, fmt.Errorf(`validateCreateStudentResponse: expected response "locationIDs": %v but actual is %v `, req.StudentProfile.LocationIds, resp.StudentProfile.Student.UserProfile.LocationIds)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateLocationStored(ctx context.Context, studentID string, locationIDs []string) error {
	stepState := StepStateFromContext(ctx)

	schoolID := fmt.Sprint(constants.ManabieSchool)
	orgID := constants.ManabieSchool
	switch req := stepState.Request.(type) {
	case *pb.CreateStudentRequest:
		schoolID = fmt.Sprint(req.SchoolId)
		orgID = int(req.SchoolId)
	case *pb.UpdateStudentRequest:
		schoolID = fmt.Sprint(req.SchoolId)
		orgID = int(req.SchoolId)
	case *pb.CreateStaffRequest:
		schoolID = fmt.Sprint(constants.ManabieSchool)
	}

	ctx = s.signedIn(ctx, orgID, StaffRoleSchoolAdmin)

	stmt :=
		`
	SELECT 
		uap.user_id,
		uap.location_id,
		uap.access_path,
		uap.resource_path
	FROM
		user_access_paths uap
	WHERE 
		uap.user_id = $1
		AND uap.location_id = ANY ($2)
		AND uap.resource_path = $3
	`
	ids := pgtype.TextArray{}
	if err := ids.Set(locationIDs); err != nil {
		return err
	}

	rows, err := s.BobDBTrace.Query(
		ctx,
		stmt,
		studentID,
		ids,
		schoolID,
	)
	if err != nil {
		return fmt.Errorf("validateLocationStored: query locations stored fail %s", err.Error())
	}
	defer rows.Close()
	userAccessPaths := []*entity.UserAccessPath{}

	for rows.Next() {
		uap := &entity.UserAccessPath{}
		if err := rows.Scan(
			&uap.UserID,
			&uap.LocationID,
			&uap.AccessPath,
			&uap.ResourcePath,
		); err != nil {
			return err
		}

		userAccessPaths = append(userAccessPaths, uap)
	}

	if len(userAccessPaths) != len(locationIDs) {
		return fmt.Errorf("validateLocationStored fail: expect stored %d user_access_paths, but actual %d", len(locationIDs), len(userAccessPaths))
	}

	for _, uap := range userAccessPaths {
		if uap.UserID.String != studentID {
			return fmt.Errorf("validateLocationStored fail: user_id stored not equal, expected: %s but actual: %s", studentID, uap.UserID.String)
		}

		if !golibs.InArrayString(uap.LocationID.String, locationIDs) {
			return fmt.Errorf("validateLocationStored fail: location_id %s stored not in locationIDs request", uap.LocationID.String)
		}
	}

	return nil
}

func (s *suite) onlyStudentInfoWithEnrollmentStatusString(ctx context.Context, enrollmentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := newID()

	if enrollmentStatus == "STUDENT_ENROLLMENT_STATUS_STRING_EMPTY" {
		enrollmentStatus = STUDENT_ENROLLMENT_STATUS_STRING_EMPTY
	}

	req := &pb.CreateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.CreateStudentRequest_StudentProfile{
			Email:               fmt.Sprintf("%v@example.com", randomID),
			Password:            fmt.Sprintf("password-%v", randomID),
			Name:                fmt.Sprintf("user-%v", randomID),
			CountryCode:         cpb.Country_COUNTRY_VN,
			EnrollmentStatus:    pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			EnrollmentStatusStr: enrollmentStatus,
			PhoneNumber:         fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId:   fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:         fmt.Sprintf("some random student note %v", randomID),
			Grade:               5,
			Birthday:            timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:              pb.Gender_MALE,
			LocationIds:         []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) inOrganizationCreateUser(ctx context.Context, signedUser string, orgOrdinal int, userOrdinal int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := createStudentReqWithOnlyStudentInfo(constants.ManabieSchool, stepState.ExistingLocations[0].LocationID.String)
	stepState.Request1 = req

	ctx = s.signedIn(ctx, constants.ManabieSchool, signedUser)
	resp, err := pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response1 = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) inOrganizationCreateUserWithTheSameAsUser(ctx context.Context, signedUser string, orgOrdinal int, user1 int, userAttribute string, user2 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := createStudentReqWithOnlyStudentInfo(constants.JPREPSchool, stepState.ExistingLocations[1].LocationID.String)
	switch userAttribute {
	case "email":
		req.StudentProfile.Email = stepState.Request1.(*pb.CreateStudentRequest).StudentProfile.Email
	case "phone number":
		req.StudentProfile.PhoneNumber = stepState.Request1.(*pb.CreateStudentRequest).StudentProfile.PhoneNumber
	}
	stepState.Request2 = req

	ctx = s.signedIn(ctx, constants.JPREPSchool, signedUser)
	response2, err := pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response2 = response2

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userWillBeCreatedSuccessfullyAndBelongedToOrganization(ctx context.Context, orgOrdinal int, userOrdinal int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp1 := stepState.Response1.(*pb.CreateStudentResponse)
	err := validateCreatedStudentReqAndResp(stepState.Request1.(*pb.CreateStudentRequest), resp1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = LoginIdentityPlatform(ctx, s.Cfg.FirebaseAPIKey, auth.LocalTenants[constants.ManabieSchool], resp1.StudentProfile.Student.UserProfile.Email, resp1.StudentProfile.StudentPassword)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp2 := stepState.Response2.(*pb.CreateStudentResponse)
	err = validateCreatedStudentReqAndResp(stepState.Request2.(*pb.CreateStudentRequest), stepState.Response2.(*pb.CreateStudentResponse))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = LoginIdentityPlatform(ctx, s.Cfg.FirebaseAPIKey, auth.LocalTenants[constants.JPREPSchool], resp2.StudentProfile.Student.UserProfile.Email, resp2.StudentProfile.StudentPassword)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentProfileWithTags(ctx context.Context, tagType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	tagIDs, _, err := s.createAmountTags(ctx, amountSampleTestElement, tagType, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := createStudentReq([]string{s.ExistingLocations[0].LocationID.String})
	req.StudentProfile.TagIds = tagIDs

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentProfileInvalidWithTags(ctx context.Context, tagType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.onlyStudentInfo(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var tagIDs []string
	switch tagType {
	case "not found":
		tagIDs = []string{idutil.ULIDNow()}

	case "invalid resource_path":
		ctx = s.signedIn(ctx, constants.JPREPSchool, StaffRoleSchoolAdmin)
		tagIDs, _, err = s.createAmountTags(ctx, amountSampleTestElement, pb.UserTagType_USER_TAG_TYPE_STUDENT.String(), fmt.Sprint(constants.JPREPSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

	case "tag for only parent":
		ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
		tagIDs, _, err = s.createTagsType(ctx, parentType)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	req.StudentProfile.TagIds = tagIDs

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createAmountTags(ctx context.Context, amount int, tagType string, resourcePath string) ([]string, []string, error) {
	tagIDs := []string{}
	tagPartnerIDs := []string{}
	batch := new(pgx.Batch)
	for i := 0; i < amount; i++ {
		tagID := idutil.ULIDNow()
		tagIDs = append(tagIDs, tagID)

		tagPartnerID := fmt.Sprintf("partner-id-%s", tagID)
		tagPartnerIDs = append(tagPartnerIDs, tagPartnerID)

		ut := repository.NewTag(&repository.Tag{
			TagAttribute: repository.TagAttribute{
				TagID:          field.NewString(tagID),
				TagName:        field.NewString(tagID),
				TagType:        field.NewString(tagType),
				TagPartnerID:   field.NewString(tagPartnerID),
				IsArchived:     field.NewBoolean(false),
				OrganizationID: field.NewString(resourcePath),
			},
		})

		fieldNames := database.GetFieldNames(ut)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(
			`INSERT INTO %s (%s) VALUES (%s)`,
			ut.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		batch.Queue(stmt, database.GetScanFields(ut, fieldNames)...)
	}

	batchResults := s.BobDBTrace.SendBatch(ctx, batch)
	defer batchResults.Close()
	for i := 0; i < amount; i++ {
		if _, err := batchResults.Exec(); err != nil {
			return nil, nil, err
		}
	}

	return tagIDs, tagPartnerIDs, nil
}
