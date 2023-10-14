package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) studentAccountUpdatedSuccessWithHomeAddresses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpdateStudentRequest)
	if err := validateHomeAddressesInDB(ctx, s.BobDBTrace, req.UserAddresses, stepState.CurrentStudentID, fmt.Sprint(OrgIDFromCtx(ctx))); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateSchoolHistoriesInDB: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInfoWithHomeAddressesUpdateValidRequest(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.UpdateStudentRequest{
		SchoolId: int32(OrgIDFromCtx(ctx)),
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                stepState.CurrentStudentID,
			Name:              fmt.Sprintf("updated-%s", stepState.CurrentStudentID),
			Grade:             int32(1),
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			StudentExternalId: fmt.Sprintf("student-external-id-%v", stepState.CurrentStudentID),
			StudentNote:       fmt.Sprintf("some random student note edited %v", stepState.CurrentStudentID),
			Email:             fmt.Sprintf("student-email-edited-%s@example.com", stepState.CurrentStudentID),
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	userAddresses, err := generateUpdateHomeAddressesPbWithCondition(ctx, s.BobDBTrace, condition)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateHomeAddressesPbWithCondition: %v", err)
	}
	req.UserAddresses = userAddresses
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInfoWithHomeAddressesUpdateInvalidRequest(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.UpdateStudentRequest{
		SchoolId: int32(OrgIDFromCtx(ctx)),
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                stepState.CurrentStudentID,
			Name:              fmt.Sprintf("updated-%s", stepState.CurrentStudentID),
			Grade:             int32(1),
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			StudentExternalId: fmt.Sprintf("student-external-id-%v", stepState.CurrentStudentID),
			StudentNote:       fmt.Sprintf("some random student note edited %v", stepState.CurrentStudentID),
			Email:             fmt.Sprintf("student-email-edited-%s@example.com", stepState.CurrentStudentID),
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	userAddresses, err := generateHomeAddressPbWithInvalidCondition(condition)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateSchoolHistoryPbWithInvalidCondition: %v", err)
	}
	req.UserAddresses = userAddresses
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func validateHomeAddressesInDB(ctx context.Context, db database.Ext, reqHomeAddresses []*pb.UserAddress, studentID string, resourcePath string) error {
	userAddressRepo := &repository.UserAddressRepo{}
	prefectureRepo := &repository.PrefectureRepo{}

	userAddresses, err := userAddressRepo.GetByUserID(ctx, db, database.Text(studentID))
	if err != nil {
		return fmt.Errorf("userAddressRepo.GetByUserID: %v", err)
	}

	count := 0
	for _, userAddress := range userAddresses {
		for _, userAddressPb := range reqHomeAddresses {
			if userAddressPb.Prefecture != "" {
				_, err := prefectureRepo.GetByPrefectureID(ctx, db, database.Text(userAddressPb.Prefecture))
				if err != nil {
					return fmt.Errorf("prefectureRepo.GetByPrefectureCode: %v", err)
				}
			}

			switch {
			case pb.AddressType_name[int32(userAddressPb.AddressType)] != userAddress.AddressType.String:
				return fmt.Errorf("validation user_address failed, expect address_type: %v, actual: %v", pb.AddressType_name[int32(userAddressPb.AddressType)], userAddress.AddressType.String)
			case userAddress.ResourcePath.String != resourcePath:
				return fmt.Errorf("validation user_address failed, expect resource_path: %v, actual: %v", userAddress.ResourcePath.String, resourcePath)
			}
			count++
		}
	}

	if len(userAddresses)*len(reqHomeAddresses) != count {
		return fmt.Errorf("cannot find any user address match with request")
	}

	return nil
}
