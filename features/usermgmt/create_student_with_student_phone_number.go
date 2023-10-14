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
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) studentInfoWithStudentPhoneNumberAndContactPreference(ctx context.Context, condition string) (context.Context, error) {
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

	switch condition {
	case "phone and contact preference":
		req.StudentProfile.StudentPhoneNumber = &pb.StudentPhoneNumber{
			PhoneNumber:       fmt.Sprintf("%v", randomNumericString(9)),
			HomePhoneNumber:   fmt.Sprintf("%v", randomNumericString(10)),
			ContactPreference: pb.StudentContactPreference_STUDENT_PHONE_NUMBER,
		}
	case "contact preference only":
		req.StudentProfile.StudentPhoneNumber = &pb.StudentPhoneNumber{
			PhoneNumber:       "",
			HomePhoneNumber:   "",
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
		}
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newStudentCreatedSuccessfullyWithStudentPhoneNumberAndContactPreference(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	studentID := resp.StudentProfile.Student.UserProfile.UserId
	if err := validateStudentPhoneNumber(ctx, s.BobPostgresDBTrace, req.StudentProfile.StudentPhoneNumber, studentID); err != nil {
		return ctx, fmt.Errorf("validateStudentPhoneNumber err: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func validateStudentPhoneNumber(ctx context.Context, db database.QueryExecer, pbUserPhoneNumber *pb.StudentPhoneNumber, studentID string) error {
	studentRepo := repository.StudentRepo{}

	studentInDB, err := studentRepo.Find(ctx, db, database.Text(studentID))
	if err != nil {
		return fmt.Errorf("newStudentCreatedSuccessfullyWithStudentPhoneNumberAndContactPreference FindStudent: %v, student_id %s", err, studentID)
	}

	userPhoneNumber := &entity.UserPhoneNumber{}
	fields := database.GetFieldNames(userPhoneNumber)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), userPhoneNumber.TableName())

	rows, err := db.Query(ctx, stmt, &studentID)
	if err != nil {
		return fmt.Errorf("newStudentCreatedSuccessfullyWithStudentPhoneNumberAndContactPreference s.BobDB.Query: %v", err)
	}

	defer rows.Close()
	userPhoneNumbers := make([]*entity.UserPhoneNumber, 0)
	for rows.Next() {
		userPhoneNumber := &entity.UserPhoneNumber{}
		if err := rows.Scan(database.GetScanFields(userPhoneNumber, fields)...); err != nil {
			return fmt.Errorf("row.Scan: %w", err)
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
	case pbUserPhoneNumber.PhoneNumber != userMainPhoneNumberInDB.PhoneNumber.String:
		return fmt.Errorf("validation student phone number in DB failed, expect user phone number: %v, actual: %v", pbUserPhoneNumber.PhoneNumber, userMainPhoneNumberInDB.PhoneNumber.String)
	case pbUserPhoneNumber.HomePhoneNumber != userHomePhoneNumberInDB.PhoneNumber.String:
		return fmt.Errorf("validation student phone number in DB failed, expect user home phone number: %v, actual: %v", pbUserPhoneNumber.HomePhoneNumber, userHomePhoneNumberInDB.PhoneNumber.String)
	case (studentInDB.ContactPreference.Status != pgtype.Null) && pbUserPhoneNumber.ContactPreference.String() != studentInDB.ContactPreference.String:
		return fmt.Errorf("validation student phone number failed, expect contact preference: %v, actual: %v", pbUserPhoneNumber.ContactPreference, studentInDB.ContactPreference.String)
	}
	return nil
}
func validateStudentDontHavePhoneNumber(ctx context.Context, db database.QueryExecer, studentID string) error {
	userPhoneNumberRepo := &repository.UserPhoneNumberRepo{}

	userPhoneNumbers, err := userPhoneNumberRepo.FindByUserID(ctx, db, database.Text(studentID))
	if err != nil {
		return err
	}

	if len(userPhoneNumbers) > 0 {
		return fmt.Errorf("validation student phone number in DB failed, expect user %s does not have any phone number, actual: %v", studentID, len(userPhoneNumbers))
	}

	return nil
}
