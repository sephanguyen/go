package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) createStudentByGRPC(ctx context.Context, typeFields string, conditionFields string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	uid := idutil.ULIDNow()
	orgID := OrgIDFromCtx(ctx)
	studentProfile := &pb.StudentProfileV2{
		FirstName: "FirstName" + uid,
		LastName:  "LastName" + uid,
		Email:     uid + "student@email.com",
		Username:  uid + "username",
		GradeId:   fmt.Sprintf("%d_grade_01", orgID),
		EnrollmentStatusHistories: []*pb.EnrollmentStatusHistory{
			{
				LocationId:       fmt.Sprintf("%d_location-id-2", orgID),
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				StartDate:        timestamppb.New(time.Now()),
			},
		},
		Password: "123456",
		StudentPhoneNumbers: &pb.StudentPhoneNumbers{
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
		},
	}

	switch typeFields {
	case "general info":
		if _, err := s.modifyGeneralInfo(ctx, conditionFields, studentProfile, uid); err != nil {
			return ctx, fmt.Errorf("s.modifyGeneralInfo err:%v", err)
		}
	case "address":
		if _, err := s.modifyAddress(ctx, conditionFields, studentProfile); err != nil {
			return ctx, fmt.Errorf("s.modifyAddress err:%v", err)
		}
	case "student phone number":
		modifyStudentPhoneNumber(conditionFields, studentProfile)
	case "school history":
		modifySchoolHistory(ctx, conditionFields, studentProfile)
	case "enrollment status history":
		if err := s.modifyEnrollmentStatusHistory(ctx, conditionFields, studentProfile); err != nil {
			return ctx, errors.Wrap(err, "s.modifyEnrollmentStatusHistory")
		}
	}

	studentProfiles := []*pb.StudentProfileV2{studentProfile}

	req := &pb.UpsertStudentRequest{
		StudentProfiles: studentProfiles,
	}

	ctx, err := s.createSubscriptionForCreatedStudentByGRPC(ctx, req)
	if err != nil {
		return ctx, errors.Wrap(err, "s.createSubscriptionForCreatedStudentByGRPC")
	}

	stepState.RequestSentAt = time.Now()
	resp, err := pb.NewStudentServiceClient(s.UserMgmtConn).UpsertStudent(ctx, req)
	if err != nil {
		errorMessages := make([]*pb.ErrorMessage, 0)
		for _, detail := range status.Convert(err).Details() {
			errorMessages = append(errorMessages, detail.(*pb.ErrorMessage))
		}
		resp = &pb.UpsertStudentResponse{Messages: errorMessages}
	}
	stepState.Request = req
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsWereUpsertedSuccessfullyByGRPC(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.UpsertStudentResponse)
	req := stepState.Request.(*pb.UpsertStudentRequest)

	for _, message := range resp.Messages {
		if message.Code != 20000 {
			return ctx, fmt.Errorf("message: %s, code: %d, field name: %s", message.Error, message.Code, message.FieldName)
		}
	}
	if len(resp.StudentProfiles) == 0 {
		return ctx, fmt.Errorf("resp.StudentProfiles is empty")
	}

	mapEmailAndUserID := make(map[string]string)

	for _, user := range resp.StudentProfiles {
		mapEmailAndUserID[user.Email] = user.Id
	}

	for i := range req.StudentProfiles {
		req.StudentProfiles[i].Id = mapEmailAndUserID[req.StudentProfiles[i].Email]
		for j := range req.StudentProfiles[i].EnrollmentStatusHistories {
			req.StudentProfiles[i].EnrollmentStatusHistories[j].StudentId = mapEmailAndUserID[req.StudentProfiles[i].Email]
		}
	}
	students := grpc.ToDomainStudents(req.StudentProfiles, true)

	if _, err := s.verifyStudentsInBD(ctx, students); err != nil {
		return ctx, fmt.Errorf("s.verifyStudentsInBD err: %v", err)
	}

	// if _, err := s.verifyLocationInNatsEvent(ctx, userIDs); err != nil {
	// 	return ctx, fmt.Errorf("s.verifyNatsEvent err: %v", err)
	// }

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) modifyGeneralInfo(ctx context.Context, conditionFields string, studentProfile *pb.StudentProfileV2, uid string) (context.Context, error) {
	switch conditionFields {
	case "all fields":
		prefectureRepo := &repository.PrefectureRepo{}
		prefecture, err := prefectureRepo.GetByPrefectureCode(ctx, s.BobDBTrace, database.Text("01"))
		if err != nil {
			return ctx, fmt.Errorf("prefectureRepo.GetByPrefectureCode err:%v", err)
		}
		studentProfile.ExternalUserId = "ExternalUserId" + uid
		studentProfile.Birthday = timestamppb.Now()
		studentProfile.Gender = pb.Gender_MALE
		studentProfile.LastNamePhonetic = "LastNamePhonetic" + uid
		studentProfile.FirstNamePhonetic = "FirstNamePhonetic" + uid
		studentProfile.SchoolHistories = []*pb.SchoolHistory{
			{
				SchoolId:       fmt.Sprintf("%d_school_id_01", OrgIDFromCtx(ctx)),
				SchoolCourseId: fmt.Sprintf("%d_school_course_id_01", OrgIDFromCtx(ctx)),
			},
		}
		studentProfile.StudentNote = "StudentNote" + uid
		// studentProfile.TagIds TODO: need to init tags default when run IT
		studentProfile.UserAddresses = []*pb.UserAddress{
			{
				AddressType:  pb.AddressType_HOME_ADDRESS,
				PostalCode:   "70000",
				Prefecture:   prefecture.ID.String,
				City:         "HCM",
				FirstStreet:  "Binh Thanh",
				SecondStreet: "Binh Tan",
			},
		}
		studentProfile.StudentPhoneNumbers = &pb.StudentPhoneNumbers{
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
			StudentPhoneNumberWithIds: []*pb.StudentPhoneNumberWithID{
				{
					PhoneNumber:     "123456789",
					PhoneNumberType: pb.StudentPhoneNumberType_PHONE_NUMBER,
				},
				{
					PhoneNumber:     "0987654321",
					PhoneNumberType: pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
				},
			},
		}
	case "duplicated external_user_id":
		if _, err := s.createStudentByGRPC(ctx, "general info", "all fields"); err != nil {
			return ctx, fmt.Errorf("s.createStudentByGRPC err:%v", err)
		}
		stepState := StepStateFromContext(ctx)
		resp := stepState.Response.(*pb.UpsertStudentResponse)
		externalUserID := resp.StudentProfiles[0].ExternalUserId
		studentProfile.ExternalUserId = externalUserID
	case "empty phonetic name":
		studentProfile.LastNamePhonetic = ""
		studentProfile.FirstNamePhonetic = ""
	case "external user id with spaces":
		externalUserIDWithSpaces := studentProfile.ExternalUserId + "     "
		studentProfile.ExternalUserId = externalUserIDWithSpaces
	case "available username":
		studentProfile.Username = "username" + idutil.ULIDNow()
	case "username was used by other":
		username, err := s.getUsernameByUsingOther(ctx)
		if err != nil {
			return ctx, errors.Wrap(err, "getUsernameByUsingOther")
		}
		studentProfile.Username = username
	case "username was used by other with upper case":
		username, err := s.getUsernameByUsingOther(ctx)
		if err != nil {
			return ctx, errors.Wrap(err, "getUsernameByUsingOther")
		}
		studentProfile.Username = strings.ToUpper(username)
	case "empty username":
		studentProfile.Username = ""
	case "username has special characters":
		studentProfile.Username = ":))" + idutil.ULIDNow()
	}

	return ctx, nil
}

func (s *suite) modifyAddress(ctx context.Context, conditionFields string, studentProfile *pb.StudentProfileV2) (context.Context, error) {
	switch conditionFields {
	case "city only":
		studentProfile.UserAddresses = []*pb.UserAddress{
			{
				City: "HCM",
			},
		}
	case "prefecture only":
		prefectureRepo := &repository.PrefectureRepo{}
		prefecture, err := prefectureRepo.GetByPrefectureCode(ctx, s.BobDBTrace, database.Text("01"))
		if err != nil {
			return ctx, fmt.Errorf("prefectureRepo.GetByPrefectureCode err:%v", err)
		}
		studentProfile.UserAddresses = []*pb.UserAddress{
			{
				Prefecture: prefecture.ID.String,
			},
		}

	case "postal code only":
		studentProfile.UserAddresses = []*pb.UserAddress{
			{
				PostalCode: "70000",
			},
		}

	case "first street only":
		studentProfile.UserAddresses = []*pb.UserAddress{
			{
				FirstStreet: "Binh Thanh",
			},
		}

	case "second street only":
		studentProfile.UserAddresses = []*pb.UserAddress{
			{
				FirstStreet: "Binh Tan",
			},
		}
	}
	return ctx, nil
}

func modifyStudentPhoneNumber(conditionFields string, studentProfile *pb.StudentProfileV2) {
	switch conditionFields {
	case "student phone number only":
		studentProfile.StudentPhoneNumbers = &pb.StudentPhoneNumbers{
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
			StudentPhoneNumberWithIds: []*pb.StudentPhoneNumberWithID{
				{
					PhoneNumberType: pb.StudentPhoneNumberType_PHONE_NUMBER,
					PhoneNumber:     "123456789",
				},
			},
		}
	case "student home phone number only":
		studentProfile.StudentPhoneNumbers = &pb.StudentPhoneNumbers{
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
			StudentPhoneNumberWithIds: []*pb.StudentPhoneNumberWithID{
				{
					PhoneNumberType: pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
					PhoneNumber:     "0987654321",
				},
			},
		}
	}
}

func modifySchoolHistory(ctx context.Context, conditionFields string, studentProfile *pb.StudentProfileV2) {
	// should change to  switch case if there are 2 conditions
	if conditionFields == "school only" {
		studentProfile.SchoolHistories = []*pb.SchoolHistory{
			{
				SchoolId: fmt.Sprintf("%d_school_id_01", OrgIDFromCtx(ctx)),
			},
		}
	}
}

func (s *suite) modifyEnrollmentStatusHistory(ctx context.Context, conditionFields string, studentProfile *pb.StudentProfileV2) error {
	orgID := OrgIDFromCtx(ctx)
	switch conditionFields {
	case "potential and temporary status":
		studentProfile.EnrollmentStatusHistories = []*pb.EnrollmentStatusHistory{
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[1]),
				StartDate:        timestamppb.Now(),
			},
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[2]),
				StartDate:        timestamppb.Now(),
			},
		}

	case "potential and enrolled status":
		studentProfile.EnrollmentStatusHistories = []*pb.EnrollmentStatusHistory{
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[1]),
				StartDate:        timestamppb.Now(),
			},
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[2]),
				StartDate:        timestamppb.Now(),
			},
		}

	case "potential and withdrawal status":
		studentProfile.EnrollmentStatusHistories = []*pb.EnrollmentStatusHistory{
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[1]),
				StartDate:        timestamppb.Now(),
			},
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[2]),
				StartDate:        timestamppb.Now(),
			},
		}
	case "potential on future start date":
		studentProfile.EnrollmentStatusHistories = []*pb.EnrollmentStatusHistory{
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[1]),
				StartDate:        timestamppb.New(time.Now().AddDate(0, 0, 1)),
			},
		}
	case "non potential on future start date":
		studentProfile.EnrollmentStatusHistories = []*pb.EnrollmentStatusHistory{
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[1]),
				StartDate:        timestamppb.New(time.Now().AddDate(0, 0, 1)),
			},
		}
	case "temporary on future date":
		studentProfile.EnrollmentStatusHistories = []*pb.EnrollmentStatusHistory{
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[1]),
				StartDate:        timestamppb.Now(),
			},
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[2]),
				StartDate:        timestamppb.New(time.Now().AddDate(0, 0, 1)),
			},
		}
	case "temporary with start date less than end date":
		studentProfile.EnrollmentStatusHistories = []*pb.EnrollmentStatusHistory{
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[1]),
				StartDate:        timestamppb.Now(),
			},
			{
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[2]),
				StartDate:        timestamppb.New(time.Now()),
				EndDate:          timestamppb.New(time.Now().AddDate(0, 0, -1)),
			},
		}
	default:
		return fmt.Errorf("conditionFields %s not supported", conditionFields)
	}

	return nil
}

func (s *suite) updateStudentByGRPC(ctx context.Context, conditions string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpsertStudentRequest)
	resp := stepState.Response.(*pb.UpsertStudentResponse)

	mapEmailAndStudentID := make(map[string]string, len(resp.StudentProfiles))

	for _, student := range resp.StudentProfiles {
		mapEmailAndStudentID[student.Email] = student.Id
	}

	for _, student := range req.StudentProfiles {
		student.Id = mapEmailAndStudentID[student.Email]

		switch conditions {
		case "edit external_user_id":
			student.ExternalUserId += student.ExternalUserId + idutil.ULIDNow()
		case "update end-date temporary enrollment status history":
			for idx, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
				if enrollmentStatusHistory.EnrollmentStatus != pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY {
					enrollmentStatusAfterSync, ok := s.ExpectedData.(enrollmentAssertion)
					if !ok {
						continue
					}
					syncedEnrollmentStatus := enrollmentStatusAfterSync.expectNewEnrollmentStatus
					student.EnrollmentStatusHistories[idx].EnrollmentStatus = pb.StudentEnrollmentStatus(pb.StudentEnrollmentStatus_value[syncedEnrollmentStatus.EnrollmentStatus().String()])
					student.EnrollmentStatusHistories[idx].StartDate = timestamppb.New(syncedEnrollmentStatus.StartDate().Time())
				} else {
					// random end-date from 1 to 10 days
					newEndDate := time.Now().AddDate(0, 0, rand.Intn(10)+1)
					student.EnrollmentStatusHistories[idx].EndDate = timestamppb.New(newEndDate)
				}
			}
		case "update potential status to temporary status":
			for idx, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
				if enrollmentStatusHistory.EnrollmentStatus != pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL {
					continue
				}

				student.EnrollmentStatusHistories[idx].StartDate = timestamppb.Now()
				student.EnrollmentStatusHistories[idx].EnrollmentStatus = pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY
			}
		case "editing to non-existing grade":
			student.GradeId = "non-existing-grade"
		case "editing to non-existing school":
			student.SchoolHistories = []*pb.SchoolHistory{
				{
					SchoolId:       "school_id_01_non_existing",
					SchoolCourseId: "school_course_id_01",
				},
			}
		case "editing to non-existing school_course":
			student.SchoolHistories = []*pb.SchoolHistory{
				{
					SchoolId:       fmt.Sprintf("%d_school_id_01", OrgIDFromCtx(ctx)),
					SchoolCourseId: "school_course_id_01_non_existing",
				},
			}
		case "editing to empty first_name":
			student.FirstName = ""
		case "editing to empty last_name":
			student.LastName = ""
		case "another available username":
			student.Username = "username" + idutil.ULIDNow()
		case "username was used by other":
			username, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return ctx, errors.Wrap(err, "getUsernameByUsingOther")
			}
			student.Username = username
		case "username was used by other with upper case":
			username, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return ctx, errors.Wrap(err, "getUsernameByUsingOther")
			}
			student.Username = strings.ToUpper(username)
		case "empty username":
			student.Username = ""
		case "username has special characters":
			student.Username = ":))" + idutil.ULIDNow()
		case "editing to empty enrollment_status and location":
			student.EnrollmentStatusHistories = nil
		case "adding one more enrollment_status and location":
			student.EnrollmentStatusHistories = append(student.EnrollmentStatusHistories, &pb.EnrollmentStatusHistory{
				StudentId:        student.Id,
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				LocationId:       fmt.Sprintf("%d_%s", OrgIDFromCtx(ctx), s.LocationIDs[2]),
				StartDate:        timestamppb.Now(),
			})
		case "changing to enrollment status potential and new date":
			student.EnrollmentStatusHistories[0].EnrollmentStatus = pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL
			student.EnrollmentStatusHistories[0].StartDate = timestamppb.Now()
		case "changing to enrollment status withdraw and new date":
			student.EnrollmentStatusHistories[0].EnrollmentStatus = pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN
			student.EnrollmentStatusHistories[0].StartDate = timestamppb.Now()
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid condition: %s", conditions)
		}
	}
	stepState.RequestSentAt = time.Now()
	resp, err := pb.NewStudentServiceClient(s.UserMgmtConn).UpsertStudent(ctx, req)
	if err != nil {
		errorMessages := make([]*pb.ErrorMessage, 0)
		for _, detail := range status.Convert(err).Details() {
			errorMessages = append(errorMessages, detail.(*pb.ErrorMessage))
		}
		resp = &pb.UpsertStudentResponse{Messages: errorMessages}
	}
	stepState.Request = req
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsWereUpsertedUnsuccessfullyByGRPCWithCodeAndField(ctx context.Context, stringCode string, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.UpsertStudentResponse)
	req := stepState.Request.(*pb.UpsertStudentRequest)

	code, err := strconv.ParseInt(stringCode, 10, 64)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("strconv.Atoi err: %v", err)
	}
	codeInt32 := int32(code)

	for _, message := range resp.Messages {
		if message.Code != codeInt32 {
			return ctx, fmt.Errorf("code is incorrect: %s, code: %d, field name: %s", message.Error, message.Code, message.FieldName)
		}
		if message.FieldName != field {
			return ctx, fmt.Errorf("fieldName is incorrect: %s, code: %d, field name: %s", message.Error, message.Code, message.FieldName)
		}
	}

	students := grpc.ToDomainStudents(req.StudentProfiles, true)

	emails := make([]string, 0, len(students))

	for _, student := range resp.StudentProfiles {
		emails = append(emails, student.Email)
	}

	if _, err := s.verifyUsersNotInBD(ctx, emails); err != nil {
		return ctx, fmt.Errorf("s.verifyStudentsInBD err: %v", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getUsernameByUsingOther(ctx context.Context) (string, error) {
	_, err := s.createStudentByGRPC(ctx, "general info", "available username")
	if err != nil {
		return "", fmt.Errorf("s.createStudentByGRPC err:%v", err)
	}

	resp := s.Response.(*pb.UpsertStudentResponse)
	if len(resp.StudentProfiles) == 0 {
		return "", errors.New("resp.StudentProfiles is empty")
	}

	student := resp.StudentProfiles[0]
	return student.Username, nil
}
