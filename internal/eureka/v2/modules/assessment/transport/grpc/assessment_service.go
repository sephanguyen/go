package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AssessmentService struct {
	pb.UnimplementedAssessmentServiceServer
	AssessmentUsecase usecase.AssessmentUsecase
}

func NewAssessmentService(assessmentUsecase usecase.AssessmentUsecase) *AssessmentService {
	return &AssessmentService{
		AssessmentUsecase: assessmentUsecase,
	}
}

var _ pb.AssessmentServiceServer = (*AssessmentService)(nil)

func (a *AssessmentService) GetAssessmentSignedRequest(ctx context.Context, req *pb.GetAssessmentSignedRequestRequest) (*pb.GetAssessmentSignedRequestResponse, error) {
	if err := a.validateGetAssessmentSignedRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	session := a.transformSessionIdentityPb(req.SessionIdentity)

	signedRequest, err := a.AssessmentUsecase.GetAssessmentSignedRequest(ctx, session, req.Domain, req.Config)
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return &pb.GetAssessmentSignedRequestResponse{
		SignedRequest: signedRequest,
	}, nil
}

func (a *AssessmentService) validateGetAssessmentSignedRequest(req *pb.GetAssessmentSignedRequestRequest) error {
	if req.SessionIdentity == nil {
		return errors.New("req must have SessionIdentity", nil)
	}
	if req.SessionIdentity.CourseId == "" {
		return errors.New("req must have CourseId", nil)
	}
	if req.SessionIdentity.LearningMaterialId == "" {
		return errors.New("req must have LearningMaterialId", nil)
	}
	if req.SessionIdentity.UserId == "" {
		return errors.New("req must have UserId", nil)
	}
	if req.Domain == "" {
		return errors.New("req must have Domain", nil)
	}

	return nil
}

func (a *AssessmentService) transformSessionIdentityPb(identity *pb.SessionIdentity) domain.Session {
	return domain.Session{
		CourseID:           identity.CourseId,
		LearningMaterialID: identity.LearningMaterialId,
		UserID:             identity.UserId,
	}
}

func (a *AssessmentService) GetLearningMaterialStatuses(ctx context.Context, req *pb.GetLearningMaterialStatusesRequest) (*pb.GetLearningMaterialStatusesResponse, error) {
	courseID := req.GetCourseId()
	userID := req.GetUserId()
	lmIDs := req.GetLearningMaterialIds()

	if strings.TrimSpace(userID) == "" || strings.TrimSpace(courseID) == "" || lmIDs == nil || len(lmIDs) == 0 {
		err := errors.NewValidationError("Input are missing: courseID, userID, LM IDs", nil)
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	nonQuizStatuses, err := a.AssessmentUsecase.ListNonQuizLearningMaterialStatuses(ctx, courseID, userID, lmIDs)
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	if len(nonQuizStatuses) == len(lmIDs) {
		return transformToStatuses(nonQuizStatuses), nil
	}

	learnosityStatuses, err := a.AssessmentUsecase.ListLearnositySessionStatuses(ctx, courseID, userID, lmIDs)
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	merged := mergeTwoMaps(nonQuizStatuses, learnosityStatuses)

	return transformToStatuses(merged), nil
}

func transformToStatuses(status map[string]bool) *pb.GetLearningMaterialStatusesResponse {
	statuses := make([]*pb.GetLearningMaterialStatusesResponse_LearningMaterialStatus, 0, len(status))
	for k, v := range status {
		statuses = append(statuses, &pb.GetLearningMaterialStatusesResponse_LearningMaterialStatus{
			LearningMaterialId: k,
			IsCompleted:        v,
		})
	}
	return &pb.GetLearningMaterialStatusesResponse{Statuses: statuses}
}

func mergeTwoMaps(m1, m2 map[string]bool) map[string]bool {
	mergedMap := make(map[string]bool)

	for k, v := range m1 {
		mergedMap[k] = v
	}
	for k, v := range m2 {
		v1 := mergedMap[k]
		if !v1 {
			mergedMap[k] = v
		}
	}

	return mergedMap
}

// ListAssessmentSubmissionResult actually get attempt histories
// TODO: should be renamed afterwards
func (a *AssessmentService) ListAssessmentSubmissionResult(ctx context.Context, req *pb.ListAssessmentSubmissionResultRequest) (*pb.ListAssessmentSubmissionResultResponse, error) {
	if err := a.validateListAssessmentSubmissionResult(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	idt := req.SessionIdentity
	subs, err := a.AssessmentUsecase.ListAssessmentAttemptHistory(ctx, idt.UserId, idt.CourseId, idt.LearningMaterialId)
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	resp := transformAttemptHistories(subs)

	return resp, nil
}

func transformAttemptHistories(sessions []domain.Session) *pb.ListAssessmentSubmissionResultResponse {
	histories := make([]*pb.AssessmentSubmission, len(sessions))

	for k, s := range sessions {
		h := &pb.AssessmentSubmission{
			SessionId:               s.ID,
			TotalPoint:              uint32(s.MaxScore),    // TODO: tobe deleted next ME prod version
			TotalGradedPoint:        uint32(s.GradedScore), // TODO: tobe deleted next ME prod version
			MaxScore:                uint32(s.MaxScore),
			GradedScore:             uint32(s.GradedScore),
			AssessmentSessionStatus: transformSessionStatusToPb(s.Status),
			CreatedAt:               timestamppb.New(s.CreatedAt),
			GradingStatus:           transformGradingStatusToPb(domain.GradingStatusNone),
		}
		if s.Submission != nil {
			sub := s.Submission
			h.GradingStatus = transformGradingStatusToPb(sub.GradingStatus)
			h.TotalPoint = uint32(sub.MaxScore)          // TODO: tobe deleted next ME prod version
			h.TotalGradedPoint = uint32(sub.GradedScore) // TODO: tobe deleted next ME prod version
			h.MaxScore = uint32(sub.MaxScore)
			h.GradedScore = uint32(sub.GradedScore)
			h.FeedbackBy = sub.FeedBackBy
			h.FeedbackSessionId = sub.FeedBackSessionID
			h.SubmissionId = sub.ID
			h.CompletedAt = timestamppb.New(sub.CompletedAt)
		}
		histories[k] = h
	}

	return &pb.ListAssessmentSubmissionResultResponse{
		AssessmentSubmissions: histories,
	}
}

func (a *AssessmentService) validateListAssessmentSubmissionResult(req *pb.ListAssessmentSubmissionResultRequest) error {
	if req.SessionIdentity == nil {
		return errors.New("req must have SessionIdentity", nil)
	}
	if req.SessionIdentity.CourseId == "" {
		return errors.New("req must have CourseId", nil)
	}
	if req.SessionIdentity.LearningMaterialId == "" {
		return errors.New("req must have LearningMaterialId", nil)
	}
	if req.SessionIdentity.UserId == "" {
		return errors.New("req must have UserId", nil)
	}

	return nil
}

func (a *AssessmentService) CompleteAssessmentSession(ctx context.Context, req *pb.CompleteAssessmentSessionRequest) (*pb.CompleteAssessmentSessionResponse, error) {
	if req.SessionId == "" {
		return nil, errors.NewGrpcError(errors.NewValidationError("req must have SessionId", nil), transport.GrpcErrorMap)
	}

	if err := a.AssessmentUsecase.CompleteAssessmentSession(ctx, req.SessionId); err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return &pb.CompleteAssessmentSessionResponse{}, nil
}

func (a *AssessmentService) GetAssessmentSubmissionDetail(ctx context.Context, req *pb.GetAssessmentSubmissionDetailRequest) (*pb.GetAssessmentSubmissionDetailResponse, error) {
	if strings.TrimSpace(req.GetSubmissionId()) == "" {
		return nil, errors.NewGrpcError(errors.NewValidationError("submission id is required", nil), transport.GrpcErrorMap)
	}
	sub, err := a.AssessmentUsecase.GetAssessmentSubmissionDetail(ctx, req.SubmissionId)
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}
	return transformSubmissionDetailToPb(sub), nil
}

func transformSubmissionDetailToPb(sub *domain.Submission) *pb.GetAssessmentSubmissionDetailResponse {
	resp := &pb.GetAssessmentSubmissionDetailResponse{
		SubmissionId:      sub.ID,
		CompletedAt:       timestamppb.New(sub.CompletedAt),
		GradingStatus:     transformGradingStatusToPb(sub.GradingStatus),
		AllocatedMarkerId: sub.AllocatedMarkerID,
		MaxScore:          uint32(sub.MaxScore),
		GradedScore:       uint32(sub.GradedScore),
		StudentSessionId:  sub.SessionID,
		FeedbackSessionId: sub.FeedBackSessionID,
		FeedbackBy:        sub.FeedBackBy,
		MarkedBy:          sub.MarkedBy,
		StudentId:         sub.StudentID,
	}
	if resp.MarkedAt != nil {
		resp.MarkedAt = timestamppb.New(*sub.MarkedAt)
	}
	return resp
}

func transformGradingStatusToPb(s domain.GradingStatus) pb.GradingStatus {
	stt := pb.GradingStatus_GRADING_STATUS_NONE
	switch s {
	case domain.GradingStatusMarked:
		stt = pb.GradingStatus_GRADING_STATUS_MARKED
	case domain.GradingStatusInProgress:
		stt = pb.GradingStatus_GRADING_STATUS_IN_PROGRESS
	case domain.GradingStatusReturned:
		stt = pb.GradingStatus_GRADING_STATUS_RETURNED
	case domain.GradingStatusNotMarked:
		stt = pb.GradingStatus_GRADING_STATUS_NOT_MARKED
	}
	return stt
}

func transformSessionStatusToPb(s domain.SessionStatus) pb.AssessmentSessionStatus {
	stt := pb.AssessmentSessionStatus_ASSESSMENT_SESSION_STATUS_NONE
	switch s {
	case domain.SessionStatusNone:
		stt = pb.AssessmentSessionStatus_ASSESSMENT_SESSION_STATUS_NONE
	case domain.SessionStatusCompleted:
		stt = pb.AssessmentSessionStatus_ASSESSMENT_SESSION_STATUS_COMPLETED
	case domain.SessionStatusIncomplete:
		stt = pb.AssessmentSessionStatus_ASSESSMENT_SESSION_STATUS_INCOMPLETE
	}
	return stt
}

func (a *AssessmentService) AllocateMarkerSubmissions(ctx context.Context, req *pb.AllocateMarkerSubmissionsRequest) (*pb.AllocateMarkerSubmissionsResponse, error) {
	if err := validateAllocateMarkerSubmissionsRequest(req); err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	submissions := make([]domain.Submission, 0, len(req.GetAllocateMarkerSubmissions()))

	for _, allocateMarker := range req.GetAllocateMarkerSubmissions() {
		submissions = append(submissions, domain.Submission{
			ID:                allocateMarker.SubmissionId,
			AllocatedMarkerID: allocateMarker.AllocatedUserId,
		})
	}

	err := a.AssessmentUsecase.AllocateMarkerSubmissions(ctx, submissions)

	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return &pb.AllocateMarkerSubmissionsResponse{}, nil
}

func validateAllocateMarkerSubmissionsRequest(req *pb.AllocateMarkerSubmissionsRequest) error {
	for _, allocateMarker := range req.GetAllocateMarkerSubmissions() {
		if allocateMarker.SubmissionId == "" {
			return errors.NewValidationError("req must have the SubmissionId", nil)
		}
	}

	return nil
}

func (a *AssessmentService) UpdateManualGradingSubmission(_ context.Context, _ *pb.UpdateManualGradingSubmissionRequest) (*pb.UpdateManualGradingSubmissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (a *AssessmentService) ListSubmissions(_ context.Context, _ *pb.ListSubmissionsRequest) (*pb.ListSubmissionsResponse, error) {
	// TODO implement me
	panic("implement me")
}
