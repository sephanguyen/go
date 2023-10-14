package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func generateHomeAddressesPbWithCondition(ctx context.Context, db database.Ext, condition string) ([]*pb.UserAddress, error) {
	var prefectureID string

	rows, err := db.Query(ctx, "SELECT prefecture_id FROM prefecture LIMIT 1")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err := rows.Scan(&prefectureID)
		if err != nil {
			return nil, err
		}
	}

	var userAddresses []*pb.UserAddress
	switch condition {
	case "one row":
		userAddresses = []*pb.UserAddress{
			{
				AddressType: pb.AddressType_HOME_ADDRESS,
				PostalCode:  "postal-code-1",
				City:        "city-1-create-onerows",
				FirstStreet: "FirstStreet-01",
			},
		}
	case "many rows":
		userAddresses = []*pb.UserAddress{
			{
				AddressType: pb.AddressType_HOME_ADDRESS,
				Prefecture:  prefectureID,
				City:        "city-1-create",
				FirstStreet: "FirstStreet-01",
			},
			{
				AddressType: pb.AddressType_HOME_ADDRESS,
				Prefecture:  prefectureID,
				City:        "city-2-create",
				FirstStreet: "FirstStreet-01",
			},
			{
				AddressType: pb.AddressType_HOME_ADDRESS,
				Prefecture:  prefectureID,
				City:        "city-3-create",
				FirstStreet: "FirstStreet-01",
			},
		}
	case "mandatory only":
		userAddresses = []*pb.UserAddress{
			{
				AddressType: pb.AddressType_HOME_ADDRESS,
				Prefecture:  prefectureID,
			},
		}
	}

	return userAddresses, nil
}

func generateUpdateHomeAddressesPbWithCondition(ctx context.Context, db database.Ext, condition string) ([]*pb.UserAddress, error) {
	var prefectureID string

	userAddressRows, err := db.Query(ctx, "SELECT user_address_id FROM user_address ORDER BY user_address_id DESC LIMIT 3")
	if err != nil {
		return nil, err
	}
	defer userAddressRows.Close()

	userAddressIDs := make([]string, 0)
	for userAddressRows.Next() {
		var userAddressID string
		err := userAddressRows.Scan(&userAddressID)
		if err != nil {
			return nil, err
		}
		userAddressIDs = append(userAddressIDs, userAddressID)
	}

	prefectureRows, err := db.Query(ctx, "SELECT prefecture_id FROM prefecture ORDER BY prefecture_id DESC LIMIT 1")
	if err != nil {
		return nil, err
	}

	for prefectureRows.Next() {
		err := prefectureRows.Scan(&prefectureID)
		if err != nil {
			return nil, err
		}
	}

	var userAddresses []*pb.UserAddress
	switch condition {
	case "one row":
		userAddresses = []*pb.UserAddress{
			{
				AddressId:   userAddressIDs[2],
				AddressType: pb.AddressType_HOME_ADDRESS,
				PostalCode:  "postal-code-1",
				Prefecture:  prefectureID,
				City:        "city-1-update-onerow",
				FirstStreet: "FirstStreet-01",
			},
		}
	case "many rows":
		userAddresses = []*pb.UserAddress{
			{
				AddressId:   userAddressIDs[2],
				AddressType: pb.AddressType_HOME_ADDRESS,
				PostalCode:  "postal-code-2",
				Prefecture:  prefectureID,
				City:        "city-1-update",
				FirstStreet: "FirstStreet-02",
			},
			{
				AddressId:   userAddressIDs[1],
				AddressType: pb.AddressType_HOME_ADDRESS,
				PostalCode:  "postal-code-3",
				City:        "city-2-update",
				FirstStreet: "FirstStreet-03",
			},
			{
				AddressId:   userAddressIDs[0],
				AddressType: pb.AddressType_HOME_ADDRESS,
				PostalCode:  "postal-code-4",
				City:        "city-3-update",
				FirstStreet: "FirstStreet-04",
			},
		}
	case "mandatory only":
		userAddresses = []*pb.UserAddress{
			{
				AddressId:   userAddressIDs[2],
				AddressType: pb.AddressType_HOME_ADDRESS,
				Prefecture:  prefectureID,
			},
		}
	}

	return userAddresses, nil
}

func (s *suite) studentInfoWithHomeAddressesRequestValid(ctx context.Context, condition string) (context.Context, error) {
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

	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	userAddresses, err := generateHomeAddressesPbWithCondition(ctx, s.BobDBTrace, condition)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateSchoolHistoryPbWithCondition: %v", err)
	}

	req.UserAddresses = userAddresses
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInfoWithHomeAddressesInvalidRequest(ctx context.Context, condition string) (context.Context, error) {
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
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	userAddresses, err := generateHomeAddressPbWithInvalidCondition(condition)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateHomeAddressPbWithInvalidCondition: %v", err)
	}
	req.UserAddresses = userAddresses
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newStudentAccountCreatedSuccessWithHomeAddresses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	if err := validateHomeAddressesInDB(ctx, s.BobDBTrace, req.UserAddresses, resp.StudentProfile.Student.UserProfile.UserId, fmt.Sprint(OrgIDFromCtx(ctx))); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateHomeAddressesInDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func generateHomeAddressPbWithInvalidCondition(condition string) ([]*pb.UserAddress, error) {
	var userAddresses []*pb.UserAddress
	switch condition {
	case "incorrect address type":
		userAddresses = []*pb.UserAddress{
			{
				AddressType: pb.AddressType_BILLING_ADDRESS,
			},
		}
	}
	return userAddresses, nil
}
