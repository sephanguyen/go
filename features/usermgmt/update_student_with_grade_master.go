package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) studentInfoWithGradeMasterUpdateRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	domainGradeRepo := &repository.DomainGradeRepo{}
	grades, err := domainGradeRepo.GetByPartnerInternalIDs(ctx, s.BobDBTrace, stepState.PartnerInternalIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domainGradeRepo.GetByPartnerInternalIDs err: %v", err)
	}

	req := &pb.UpdateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                stepState.CurrentStudentID,
			Name:              fmt.Sprintf("updated-%s", stepState.CurrentStudentID),
			GradeId:           grades[0].GradeID().String(),
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			StudentExternalId: fmt.Sprintf("student-external-id-%v", stepState.CurrentStudentID),
			StudentNote:       fmt.Sprintf("some random student note edited %v", stepState.CurrentStudentID),
			Email:             fmt.Sprintf("student-email-edited-%s@example.com", stepState.CurrentStudentID),
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            pb.Gender_MALE,
			LocationIds:       []string{s.ExistingLocations[0].LocationID.String},
		},
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInfoWithInvalidGradeMasterUpdateRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	randomID := newID()

	req := &pb.UpdateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &pb.UpdateStudentRequest_StudentProfile{
			Id:                stepState.CurrentStudentID,
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Name:              fmt.Sprintf("user-%v", randomID),
			EnrollmentStatus:  pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
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

func (s *suite) newStudentAccountUpdatedSuccessWithGradeMaster(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.UpdateStudentRequest)
	resp := stepState.Response.(*pb.UpdateStudentResponse)

	if req.StudentProfile.GradeId != resp.StudentProfile.GradeId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation grade master failed, expect grade_id: %v, actual: %v", req.StudentProfile.GradeId, resp.StudentProfile.GradeId)
	}

	if err := validateGradeMaster(ctx, s.BobDBTrace, req.StudentProfile.GradeId, stepState.CurrentStudentID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validateGradeMaster: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
