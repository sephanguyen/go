package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func ToAssignmentPb(src *entities.GeneralAssignment) (*sspb.AssignmentBase, error) {
	var attachment []string
	err := src.Attachments.AssignTo(&attachment)
	if err != nil {
		return nil, err
	}
	return &sspb.AssignmentBase{
		Base: &sspb.LearningMaterialBase{
			LearningMaterialId: src.LearningMaterial.ID.String,
			TopicId:            src.LearningMaterial.TopicID.String,
			Name:               src.LearningMaterial.Name.String,
			Type:               src.LearningMaterial.Type.String,
			DisplayOrder: &wrapperspb.Int32Value{
				Value: int32(src.DisplayOrder.Int),
			},
		},
		Attachments:            attachment,
		Instruction:            src.Instruction.String,
		MaxGrade:               uint32(src.MaxGrade.Int),
		IsRequiredGrade:        src.IsRequiredGrade.Bool,
		AllowResubmission:      src.AllowResubmission.Bool,
		RequireAssignmentNote:  src.RequireAssignmentNote.Bool,
		RequireVideoSubmission: src.RequireVideoSubmission.Bool,
	}, nil
}

func (s *AssignmentService) ListAssignment(ctx context.Context, req *sspb.ListAssignmentRequest) (*sspb.ListAssignmentResponse, error) {
	ids := req.LearningMaterialIds
	if len(ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "LearningMaterialIds must not be empty")
	}

	assignments, err := s.GeneralAssignmentRepo.List(ctx, s.DB, database.TextArray(req.LearningMaterialIds))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, fmt.Errorf("assignment not found: %w", err).Error())
		}
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.GeneralAssignmentRepo.List: %w", err).Error())
	}

	assignmentsPb := make([]*sspb.AssignmentBase, 0, len(assignments))
	for _, assignment := range assignments {
		assignmentBase, err := ToAssignmentPb(assignment)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error convert assignment to pb: %v", err)
		}
		assignmentsPb = append(assignmentsPb, assignmentBase)
	}
	return &sspb.ListAssignmentResponse{
		Assignments: assignmentsPb,
	}, nil
}

func (s *AssignmentService) SubmitAssignment(ctx context.Context, req *sspb.SubmitAssignmentRequest) (*sspb.SubmitAssignmentResponse, error) {
	if err := s.validateSubmitAssignmentRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validateSubmitAssignmentRequest: %s", err.Error())
	}
	if err := s.validateSubmitAssignmentPermission(ctx, req); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "validateSubmitAssignmentPermission: %s", err.Error())
	}

	submission, err := s.submitAssignmentRequestToEnt(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "submitAssignmentV2RequestToEnt: %s", err.Error())
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.calculateAssignmentLearningTime(ctx, tx, &entities.StudentAssignmentLearningTime{
			StudentID:    submission.StudentID.String,
			AssignmentID: req.GetSubmission().GetStudyPlanItemIdentity().GetLearningMaterialId(),
			CompleteDate: req.GetSubmission().GetCompleteDate(),
			Duration:     req.GetSubmission().GetDuration(),
		}); err != nil {
			return err
		}

		if err := s.SubmissionRepo.Create(ctx, tx, submission); err != nil {
			return fmt.Errorf("s.SubmissionRepo.Create: %w", err)
		}
		if err := s.StudentLatestSubmissionRepo.UpsertV2(ctx, tx, &entities.StudentLatestSubmission{StudentSubmission: *submission}); err != nil {
			return fmt.Errorf("s.StudentLatestSubmissionRepo.Upsert: %w", err)
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}
	return &sspb.SubmitAssignmentResponse{
		SubmissionId: submission.ID.String,
	}, nil
}

func (s *AssignmentService) validateSubmitAssignmentRequest(req *sspb.SubmitAssignmentRequest) error {
	if req.Submission == nil {
		return fmt.Errorf("submission can not be empty")
	}
	if req.Submission.StudyPlanItemIdentity == nil {
		return fmt.Errorf("studyPlanItemIdentity can not be empty")
	}
	if req.Submission.StudyPlanItemIdentity.StudyPlanId == "" {
		return fmt.Errorf("studyPlanId can not be empty")
	}
	if req.Submission.StudyPlanItemIdentity.LearningMaterialId == "" {
		return fmt.Errorf("learningMaterialId can not be empty")
	}
	if req.Submission.StudyPlanItemIdentity.StudentId == nil || req.Submission.StudyPlanItemIdentity.StudentId.Value == "" {
		return fmt.Errorf("studentId can not be empty")
	}
	return nil
}

func (s *AssignmentService) validateSubmitAssignmentPermission(ctx context.Context, req *sspb.SubmitAssignmentRequest) error {
	groupID := interceptors.UserGroupFromContext(ctx)
	switch groupID {
	case cpb.UserGroup_USER_GROUP_STUDENT.String(), constant.RoleStudent:
		req.Submission.StudyPlanItemIdentity.StudentId = wrapperspb.String(interceptors.UserIDFromContext(ctx))
	}

	ok, err := s.AssignmentRepo.IsStudentAssignedV2(
		ctx,
		s.DB,
		database.Text(req.Submission.StudyPlanItemIdentity.StudyPlanId),
		database.Text(req.Submission.StudyPlanItemIdentity.StudentId.Value))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("non-assigned assignment")
	}

	return nil
}

func (s *AssignmentService) submitAssignmentRequestToEnt(ctx context.Context, req *sspb.SubmitAssignmentRequest) (*entities.StudentSubmission, error) {
	e := &entities.StudentSubmission{}
	database.AllNullEntity(e)

	errs := multierr.Combine(
		e.StudyPlanID.Set(req.Submission.StudyPlanItemIdentity.StudyPlanId),
		e.LearningMaterialID.Set(req.Submission.StudyPlanItemIdentity.LearningMaterialId),
		e.Note.Set(req.Submission.Note),
		e.Status.Set(pb.SubmissionStatus_SUBMISSION_STATUS_NOT_MARKED.String()),
		e.Duration.Set(req.Submission.Duration),
		e.UnderstandingLevel.Set(req.Submission.UnderstandingLevel.String()),
	)
	if req.Submission.SubmissionContent != nil {
		errs = multierr.Append(errs, e.SubmissionContent.Set(req.Submission.SubmissionContent))
	}

	groupID := interceptors.UserGroupFromContext(ctx)
	switch groupID {
	case constant.RoleStudent:
		errs = multierr.Append(errs, e.StudentID.Set(interceptors.UserIDFromContext(ctx)))
	case constant.RoleTeacher:
		errs = multierr.Append(errs, e.StudentID.Set(req.Submission.StudyPlanItemIdentity.StudentId.Value))
	}

	if req.Submission.CompleteDate != nil && req.Submission.CompleteDate.IsValid() {
		errs = multierr.Append(errs, e.CompleteDate.Set(req.Submission.CompleteDate.AsTime()))
	} else {
		errs = multierr.Append(errs, e.CompleteDate.Set(time.Now()))
	}

	if req.Submission.CorrectScore != nil {
		errs = multierr.Append(errs, e.CorrectScore.Set(req.Submission.CorrectScore.Value))
	}
	if req.Submission.TotalScore != nil {
		errs = multierr.Append(errs, e.TotalScore.Set(req.Submission.TotalScore.Value))
	}

	if errs != nil {
		return nil, fmt.Errorf("set StudentSubmission data: %w", errs)
	}
	return e, nil
}

func (s *AssignmentService) calculateAssignmentLearningTime(ctx context.Context, tx database.QueryExecer, req *entities.StudentAssignmentLearningTime) error {
	country, err := s.getStudentCountry(ctx, req.StudentID)
	if err != nil {
		return err
	}

	if req.Duration != 0 && req.CompleteDate != nil {
		day := timeutil.MidnightIn(bob_pb.Country(bob_pb.Country_value[country]), req.CompleteDate.AsTime())

		pgDay := database.Timestamptz(day.UTC())
		var assignmentSubmissionIDs []string

		dailies, err := s.StudentLearningTimeDailyRepo.Retrieve(ctx, tx, database.Text(req.StudentID), &pgDay, &pgDay, repositories.WithUpdateLock())
		if err != nil {
			return fmt.Errorf("Retrieve: %w", err)
		}

		if len(dailies) > 0 {
			assignmentSubmissionIDs = database.FromTextArray(dailies[0].AssignmentSubmissionIDs)
		}

		if req.AssignmentID != "" && !golibs.InArrayString(req.AssignmentID, assignmentSubmissionIDs) {
			assignmentSubmissionIDs = append(assignmentSubmissionIDs, req.AssignmentID)
		}

		studentLearningTimeDaily := &entities.StudentLearningTimeDaily{
			StudentID:               database.Text(req.StudentID),
			LearningTime:            database.Int4(req.Duration),
			AssignmentLearningTime:  database.Int4(req.Duration),
			AssignmentSubmissionIDs: database.TextArray(assignmentSubmissionIDs),
			Day:                     pgDay, // DB always store UTC time
		}
		if err := s.StudentLearningTimeDailyRepo.UpsertTaskAssignment(ctx, tx, studentLearningTimeDaily); err != nil {
			return fmt.Errorf("Upsert: %w", err)
		}
	}
	return nil
}

func (s *AssignmentService) getStudentCountry(ctx context.Context, studentID string) (string, error) {
	upbReq := &upb.SearchBasicProfileRequest{
		UserIds: []string{studentID},
		Paging:  &cpb.Paging{Limit: uint32(1)},
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "GetOutgoingContext: %v", err)
	}
	resp, err := s.UsermgmtUserReaderService.SearchBasicProfile(mdCtx, upbReq)
	if err != nil {
		return "", status.Errorf(codes.Internal, "s.UsermgmtUserReaderService.SearchBasicProfile: %v", err)
	}

	if len(resp.Profiles) == 0 {
		return "", status.Errorf(codes.NotFound, "s.UsermgmtUserReaderService.SearchBasicProfile: user %s not found", studentID)
	}

	return resp.Profiles[0].Country.String(), nil
}
