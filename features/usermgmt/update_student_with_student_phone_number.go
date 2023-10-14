package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) updateStudentInfoWithStudentPhoneNumberAndContactPreference(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomID := newID()

	req := &pb.UpdateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                stepState.CurrentStudentID,
			FirstName:         fmt.Sprintf("student-first-name-%v", randomID),
			LastName:          fmt.Sprintf("student-last-name-%v", randomID),
			Email:             fmt.Sprintf("%v@example.com", randomID),
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			StudentExternalId: fmt.Sprintf("student-external-id-%v", stepState.CurrentStudentID),
			StudentNote:       fmt.Sprintf("some random student note edited %v", stepState.CurrentStudentID),
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	switch condition {
	case "phone and contact preference":
		req.StudentProfile.StudentPhoneNumber = &pb.StudentPhoneNumber{
			PhoneNumber:       fmt.Sprintf("%v", randomNumericString(9)),
			HomePhoneNumber:   fmt.Sprintf("%v", randomNumericString(10)),
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
		}
	case "contact preference only":
		req.StudentProfile.StudentPhoneNumber = &pb.StudentPhoneNumber{
			PhoneNumber:       "",
			HomePhoneNumber:   "",
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
		}
	case "phone number with id and contact preference":
		userPhoneNumberRepo := repository.UserPhoneNumberRepo{}
		var currentStudentPhoneNumberID, currentStudentHomePhoneNumberID string
		userPhoneNumbers, err := userPhoneNumberRepo.FindByUserID(ctx, s.BobDBTrace, database.Text(stepState.CurrentStudentID))
		if err != nil {
			return nil, err
		}
		for _, userPhoneNumber := range userPhoneNumbers {
			switch userPhoneNumber.PhoneNumberType.String {
			case entity.StudentPhoneNumber:
				currentStudentPhoneNumberID = userPhoneNumber.ID.String
			case entity.StudentHomePhoneNumber:
				currentStudentHomePhoneNumberID = userPhoneNumber.ID.String
			}
		}
		req.StudentProfile.StudentPhoneNumbers = &pb.UpdateStudentPhoneNumber{
			StudentPhoneNumber: []*pb.StudentPhoneNumberWithID{
				{
					StudentPhoneNumberId: currentStudentPhoneNumberID,
					PhoneNumber:          fmt.Sprintf("%v", randomNumericString(9)),
					PhoneNumberType:      pb.StudentPhoneNumberType_PHONE_NUMBER,
				},
				{
					StudentPhoneNumberId: currentStudentHomePhoneNumberID,
					PhoneNumber:          fmt.Sprintf("%v", randomNumericString(9)),
					PhoneNumberType:      pb.StudentPhoneNumberType_HOME_PHONE_NUMBER,
				},
			},
			ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
		}
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountUpdatedSuccessWithStudentPhoneNumberIDAndContactPreference(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpdateStudentRequest)
	studentID := stepState.CurrentStudentID

	studentRepo := repository.StudentRepo{}

	studentInDB, err := studentRepo.Find(ctx, s.BobDB, database.Text(studentID))

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentAccountUpdatedSuccessWithStudentPhoneNumberAndContactPreference FindStudent: %v", err)
	}

	userPhoneNumber := &entity.UserPhoneNumber{}
	fields := database.GetFieldNames(userPhoneNumber)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), userPhoneNumber.TableName())

	rows, err := s.BobDB.Query(ctx, stmt, &studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentAccountUpdatedSuccessWithStudentPhoneNumberAndContactPreference s.BobDB.Query: %v", err)
	}

	defer rows.Close()

	userPhoneNumbers := make([]*entity.UserPhoneNumber, 0)
	for rows.Next() {
		userPhoneNumber := &entity.UserPhoneNumber{}
		if err := rows.Scan(database.GetScanFields(userPhoneNumber, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		userPhoneNumbers = append(userPhoneNumbers, userPhoneNumber)
	}
	userMainPhoneNumberInDB := &entity.UserPhoneNumber{}
	userHomePhoneNumberInDB := &entity.UserPhoneNumber{}

	for _, phoneNumber := range userPhoneNumbers {
		if phoneNumber.PhoneNumberType.String == entity.StudentPhoneNumber {
			userMainPhoneNumberInDB = phoneNumber
		}
		if phoneNumber.PhoneNumberType.String == entity.StudentHomePhoneNumber {
			userHomePhoneNumberInDB = phoneNumber
		}
	}
	userPhoneNumberInReq := &pb.StudentPhoneNumberWithID{}
	userHomePhoneNumberInReq := &pb.StudentPhoneNumberWithID{}
	for _, phoneNumber := range req.StudentProfile.StudentPhoneNumbers.StudentPhoneNumber {
		switch phoneNumber.PhoneNumberType {
		case pb.StudentPhoneNumberType_PHONE_NUMBER:
			userPhoneNumberInReq = phoneNumber
		case pb.StudentPhoneNumberType_HOME_PHONE_NUMBER:
			userHomePhoneNumberInReq = phoneNumber
		}
	}

	switch {
	case userPhoneNumberInReq.PhoneNumber != userMainPhoneNumberInDB.PhoneNumber.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number in DB failed, expect user phone number: %v, actual: %v", userPhoneNumberInReq.PhoneNumber, userMainPhoneNumberInDB.PhoneNumber.String)
	case userHomePhoneNumberInReq.PhoneNumber != userHomePhoneNumberInDB.PhoneNumber.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number in DB failed, expect user home phone number: %v, actual: %v", userHomePhoneNumberInReq.PhoneNumber, userHomePhoneNumberInDB.PhoneNumber.String)
	case req.StudentProfile.StudentPhoneNumbers.ContactPreference.String() != studentInDB.ContactPreference.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation contact preference in DB failed, expect contact preference: %v, actual: %v", req.StudentProfile.StudentPhoneNumbers.ContactPreference.String(), studentInDB.ContactPreference.String)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAccountUpdatedSuccessWithStudentPhoneNumberAndContactPreference(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpdateStudentRequest)
	resp := stepState.Response.(*pb.UpdateStudentResponse)
	studentID := stepState.CurrentStudentID

	studentRepo := repository.StudentRepo{}

	studentInDB, err := studentRepo.Find(ctx, s.BobDB, database.Text(studentID))

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentAccountUpdatedSuccessWithStudentPhoneNumberAndContactPreference FindStudent: %v", err)
	}

	userPhoneNumber := &entity.UserPhoneNumber{}
	fields := database.GetFieldNames(userPhoneNumber)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), userPhoneNumber.TableName())

	rows, err := s.BobDB.Query(ctx, stmt, &studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("studentAccountUpdatedSuccessWithStudentPhoneNumberAndContactPreference s.BobDB.Query: %v", err)
	}

	defer rows.Close()

	userPhoneNumbers := make([]*entity.UserPhoneNumber, 0)
	for rows.Next() {
		userPhoneNumber := &entity.UserPhoneNumber{}
		if err := rows.Scan(database.GetScanFields(userPhoneNumber, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		userPhoneNumbers = append(userPhoneNumbers, userPhoneNumber)
	}
	userMainPhoneNumberInDB := &entity.UserPhoneNumber{}
	userHomePhoneNumberInDB := &entity.UserPhoneNumber{}

	for _, phoneNumber := range userPhoneNumbers {
		if phoneNumber.PhoneNumberType.String == entity.StudentPhoneNumber {
			userMainPhoneNumberInDB = phoneNumber
		}
		if phoneNumber.PhoneNumberType.String == entity.StudentHomePhoneNumber {
			userHomePhoneNumberInDB = phoneNumber
		}
	}

	switch {
	case req.StudentProfile.StudentPhoneNumber.PhoneNumber != userMainPhoneNumberInDB.PhoneNumber.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number in DB failed, expect user phone number: %v, actual: %v", req.StudentProfile.StudentPhoneNumber.PhoneNumber, userMainPhoneNumberInDB.PhoneNumber.String)
	case req.StudentProfile.StudentPhoneNumber.HomePhoneNumber != userHomePhoneNumberInDB.PhoneNumber.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number in DB failed, expect user home phone number: %v, actual: %v", req.StudentProfile.StudentPhoneNumber.HomePhoneNumber, userHomePhoneNumberInDB.PhoneNumber.String)
	case req.StudentProfile.StudentPhoneNumber.ContactPreference.String() != studentInDB.ContactPreference.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number failed, expect contact preference: %v, actual: %v", req.StudentProfile.StudentPhoneNumber.ContactPreference.String(), studentInDB.ContactPreference.String)
	case req.StudentProfile.StudentPhoneNumber.PhoneNumber != resp.StudentProfile.StudentPhoneNumber.PhoneNumber:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number failed, expect phone number: %v, actual: %v", req.StudentProfile.StudentPhoneNumber.PhoneNumber, resp.StudentProfile.StudentPhoneNumber.PhoneNumber)
	case req.StudentProfile.StudentPhoneNumber.HomePhoneNumber != resp.StudentProfile.StudentPhoneNumber.HomePhoneNumber:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number failed, expect home phone number: %v, actual: %v", req.StudentProfile.StudentPhoneNumber.HomePhoneNumber, resp.StudentProfile.StudentPhoneNumber.HomePhoneNumber)
	case req.StudentProfile.StudentPhoneNumber.ContactPreference != resp.StudentProfile.StudentPhoneNumber.ContactPreference:
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation student phone number failed, expect contact preference: %v, actual: %v", req.StudentProfile.StudentPhoneNumber.ContactPreference, resp.StudentProfile.StudentPhoneNumber.ContactPreference)
	}

	return StepStateToContext(ctx, stepState), nil
}
