package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) studentInfoWithGradeMasterRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	domainGradeRepo := &repository.DomainGradeRepo{}
	grades, err := domainGradeRepo.GetByPartnerInternalIDs(ctx, s.BobDBTrace, stepState.PartnerInternalIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domainGradeRepo.GetByPartnerInternalIDs err: %v", err)
	}
	grade := grades[0]
	stepState.GradeName = grade.Name().String()
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
			GradeId:           grade.GradeID().String(),
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInfoWithInvalidGradeMasterRequest(ctx context.Context) (context.Context, error) {
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
			GradeId:           "invalid-grade-id",
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newStudentAccountCreatedSuccessWithGradeMaster(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.CreateStudentRequest)
	resp := stepState.Response.(*pb.CreateStudentResponse)

	if req.StudentProfile.GradeId != resp.StudentProfile.Student.GradeId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation grade master failed, expect grade_id: %v, actual: %v", req.StudentProfile.GradeId, resp.StudentProfile.Student.GradeId)
	}

	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	if err := validateGradeMaster(ctx, s.BobDBTrace, req.StudentProfile.GradeId, resp.StudentProfile.Student.UserProfile.UserId); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateGradeMaster: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func validateGradeMaster(ctx context.Context, db database.Ext, gradeIDreq, studentID string) error {
	studentRepo := &repository.StudentRepo{}
	students, err := studentRepo.FindStudentProfilesByIDs(ctx, db, database.TextArray([]string{studentID}))
	if err != nil {
		return fmt.Errorf("studentRepo.FindStudentProfilesByIDs: %v", err)
	}

	if len(students) == 0 {
		return fmt.Errorf("student with id %v does not exist", studentID)
	}

	student := students[0]

	if student.GradeID.String != gradeIDreq {
		return fmt.Errorf("validation grade master failed, expect grade_id: %v, actual: %v", gradeIDreq, student.GradeID.String)
	}

	return nil
}
