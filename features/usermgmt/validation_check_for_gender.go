package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) onlyStudentInfoWith(ctx context.Context, gender string) (context.Context, error) {
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
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}

	switch gender {
	case "MALE":
		req.StudentProfile.Gender = pb.Gender_MALE
	case "FEMALE":
		req.StudentProfile.Gender = pb.Gender_FEMALE
	default:
		req.StudentProfile.Gender = pb.Gender(99)
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}
