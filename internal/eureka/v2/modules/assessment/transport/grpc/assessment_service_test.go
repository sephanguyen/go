package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_usecase "github.com/manabie-com/backend/mock/eureka/v2/modules/assessment/usecase"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAssessmentService_GetAssessmentSignedRequest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	assessmenUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
	assessmentSvc := NewAssessmentService(assessmenUsecase)

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		Setup            func(ctx context.Context)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name: "SessionIdentity is empty",
			Ctx:  ctx,
			Request: &pb.GetAssessmentSignedRequestRequest{
				SessionIdentity: nil,
				Domain:          "domain",
				Config:          "config",
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have SessionIdentity", nil)),
		},
		{
			Name: "CourseId is empty",
			Ctx:  ctx,
			Request: &pb.GetAssessmentSignedRequestRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "",
					LearningMaterialId: "lm_id",
					UserId:             "user_id",
				},
				Domain: "domain",
				Config: "config",
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have CourseId", nil)),
		},
		{
			Name: "LearningMaterialId is empty",
			Ctx:  ctx,
			Request: &pb.GetAssessmentSignedRequestRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "",
					UserId:             "user_id",
				},
				Domain: "domain",
				Config: "config",
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have LearningMaterialId", nil)),
		},
		{
			Name: "UserId is empty",
			Ctx:  ctx,
			Request: &pb.GetAssessmentSignedRequestRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "lm_id",
					UserId:             "",
				},
				Domain: "domain",
				Config: "config",
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have UserId", nil)),
		},
		{
			Name: "Domain is empty",
			Ctx:  ctx,
			Request: &pb.GetAssessmentSignedRequestRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "lm_id",
					UserId:             "user_id",
				},
				Domain: "",
				Config: "config",
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have Domain", nil)),
		},
		{
			Name: "AssessmentUsecase.GetAssessmentSignedRequest return error",
			Ctx:  ctx,
			Request: &pb.GetAssessmentSignedRequestRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "lm_id",
					UserId:             "user_id",
				},
				Domain: "domain",
				Config: "config",
			},
			Setup: func(ctx context.Context) {
				assessmenUsecase.On("GetAssessmentSignedRequest", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return("", errors.New("error", nil))
			},
			ExpectedResponse: nil,
			ExpectedError:    status.Error(codes.Internal, errors.New("error", nil).Error()),
		},
		{
			Name: "happy case",
			Ctx:  ctx,
			Request: &pb.GetAssessmentSignedRequestRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "lm_id",
					UserId:             "user_id",
				},
				Domain: "domain",
				Config: "config",
			},
			Setup: func(ctx context.Context) {
				assessmenUsecase.On("GetAssessmentSignedRequest", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return("SignedRequest", nil)
			},
			ExpectedResponse: &pb.GetAssessmentSignedRequestResponse{
				SignedRequest: "SignedRequest",
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Setup != nil {
				tc.Setup(tc.Ctx)
			}
			res, err := assessmentSvc.GetAssessmentSignedRequest(tc.Ctx, tc.Request.(*pb.GetAssessmentSignedRequestRequest))
			if tc.ExpectedError != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResponse, res)
			}
		})
	}
}

func TestAssessmentService_ListAssessmentSubmissionResult(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	assessmentUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
	assessmentSvc := NewAssessmentService(assessmentUsecase)

	createdAt := time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local)
	completedAt := time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local)

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		Setup            func(ctx context.Context)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name: "SessionIdentity is empty",
			Ctx:  ctx,
			Request: &pb.ListAssessmentSubmissionResultRequest{
				SessionIdentity: nil,
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have SessionIdentity", nil)),
		},
		{
			Name: "CourseId is empty",
			Ctx:  ctx,
			Request: &pb.ListAssessmentSubmissionResultRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "",
					LearningMaterialId: "lm_id",
					UserId:             "user_id",
				},
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have CourseId", nil)),
		},
		{
			Name: "LearningMaterialId is empty",
			Ctx:  ctx,
			Request: &pb.ListAssessmentSubmissionResultRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "",
					UserId:             "user_id",
				},
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have LearningMaterialId", nil)),
		},
		{
			Name: "UserId is empty",
			Ctx:  ctx,
			Request: &pb.ListAssessmentSubmissionResultRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "lm_id",
					UserId:             "",
				},
			},
			ExpectedResponse: nil,
			ExpectedError:    fmt.Errorf("invalid request: %w", errors.New("req must have UserId", nil)),
		},
		{
			Name: "AssessmentUsecase.ListAssessmentSubmissionResult return error",
			Ctx:  ctx,
			Request: &pb.ListAssessmentSubmissionResultRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "lm_id",
					UserId:             "user_id",
				},
			},
			Setup: func(ctx context.Context) {
				assessmentUsecase.On("ListAssessmentAttemptHistory", ctx, "user_id", "course_id", "lm_id").
					Once().
					Return(nil, errors.New("error", nil))
			},
			ExpectedResponse: nil,
			ExpectedError:    status.Error(codes.Internal, errors.New("error", nil).Error()),
		},
		{
			Name: "happy case",
			Ctx:  ctx,
			Request: &pb.ListAssessmentSubmissionResultRequest{
				SessionIdentity: &pb.SessionIdentity{
					CourseId:           "course_id",
					LearningMaterialId: "lm_id",
					UserId:             "user_id",
				},
			},
			Setup: func(ctx context.Context) {
				assessmentUsecase.On("ListAssessmentAttemptHistory", ctx, "user_id", "course_id", "lm_id").
					Once().Return([]domain.Session{
					{
						ID:          "session_id",
						MaxScore:    8,
						GradedScore: 4,
						CreatedAt:   createdAt,
						Status:      domain.SessionStatusCompleted,
						Submission: &domain.Submission{
							ID:                "sub_id",
							SessionID:         "session_id",
							MaxScore:          8,
							GradedScore:       6,
							CreatedAt:         completedAt,
							CompletedAt:       completedAt,
							GradingStatus:     domain.GradingStatusMarked,
							FeedBackBy:        "ID F",
							FeedBackSessionID: "ID F2",
						},
					},
				}, nil)
			},
			ExpectedResponse: &pb.ListAssessmentSubmissionResultResponse{
				AssessmentSubmissions: []*pb.AssessmentSubmission{
					{
						SessionId:               "session_id",
						SubmissionId:            "sub_id",
						TotalPoint:              8,
						TotalGradedPoint:        6,
						MaxScore:                8,
						GradedScore:             6,
						AssessmentSessionStatus: pb.AssessmentSessionStatus_ASSESSMENT_SESSION_STATUS_COMPLETED,
						CreatedAt:               timestamppb.New(createdAt),
						CompletedAt:             timestamppb.New(completedAt),
						GradingStatus:           pb.GradingStatus_GRADING_STATUS_MARKED,
						FeedbackBy:              "ID F",
						FeedbackSessionId:       "ID F2",
					},
				},
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Setup != nil {
				tc.Setup(tc.Ctx)
			}
			res, err := assessmentSvc.ListAssessmentSubmissionResult(tc.Ctx, tc.Request.(*pb.ListAssessmentSubmissionResultRequest))
			if tc.ExpectedError != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResponse, res)
			}
		})
	}
}

func TestAssessmentService_GetLearningMaterialStatuses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	courseID := idutil.ULIDNow()
	userID := idutil.ULIDNow()
	lmIDs := []string{"LM1", "LM2"}
	req := &pb.GetLearningMaterialStatusesRequest{
		CourseId:            courseID,
		LearningMaterialIds: lmIDs,
		UserId:              userID,
	}

	t.Run("Return invalid arguments when input are missing", func(t *testing.T) {
		t.Parallel()
		// arrange
		assessmentUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		sut := NewAssessmentService(assessmentUsecase)
		req := &pb.GetLearningMaterialStatusesRequest{
			CourseId:            "",
			LearningMaterialIds: lmIDs,
			UserId:              userID,
		}
		req2 := &pb.GetLearningMaterialStatusesRequest{
			CourseId:            courseID,
			LearningMaterialIds: lmIDs,
			UserId:              "",
		}
		req3 := &pb.GetLearningMaterialStatusesRequest{
			CourseId:            courseID,
			LearningMaterialIds: []string{},
			UserId:              userID,
		}
		rootErr := errors.NewValidationError("Input are missing: courseID, userID, LM IDs", nil)
		expectedErr := status.Error(codes.InvalidArgument, fmt.Errorf("%w", rootErr).Error())

		// act
		resp, err := sut.GetLearningMaterialStatuses(ctx, req)
		resp2, err2 := sut.GetLearningMaterialStatuses(ctx, req2)
		resp3, err3 := sut.GetLearningMaterialStatuses(ctx, req3)

		// assert
		assert.Equal(t, err, expectedErr)
		assert.Equal(t, err2, expectedErr)
		assert.Equal(t, err3, expectedErr)
		assert.Nil(t, resp)
		assert.Nil(t, resp2)
		assert.Nil(t, resp3)
	})

	t.Run("Return correct status when there are only non-quiz LMs", func(t *testing.T) {
		t.Parallel()
		// arrange
		assessmentUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		sut := NewAssessmentService(assessmentUsecase)
		nonQuizStatuses := map[string]bool{
			"LM1": true,
			"LM2": false,
		}
		assessmentUsecase.On("ListNonQuizLearningMaterialStatuses", ctx, courseID, userID, lmIDs).
			Once().Return(nonQuizStatuses, nil)

		// act
		resp, err := sut.GetLearningMaterialStatuses(ctx, req)

		// assert
		assert.Nil(t, err)
		for _, v := range resp.GetStatuses() {
			assert.Equal(t, nonQuizStatuses[v.LearningMaterialId], v.IsCompleted)
		}
		mock.AssertExpectationsForObjects(t, assessmentUsecase)
	})

	t.Run("Return correct status when there are only learnosity quiz LMs", func(t *testing.T) {
		t.Parallel()
		// arrange
		assessmentUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		sut := NewAssessmentService(assessmentUsecase)
		nonQuizStatuses := map[string]bool{}
		learnosityStatuses := map[string]bool{
			"LM1": false,
			"LM2": true,
		}
		assessmentUsecase.On("ListNonQuizLearningMaterialStatuses", ctx, courseID, userID, lmIDs).
			Once().Return(nonQuizStatuses, nil)
		assessmentUsecase.On("ListLearnositySessionStatuses", ctx, courseID, userID, lmIDs).
			Once().Return(learnosityStatuses, nil)

		// act
		resp, err := sut.GetLearningMaterialStatuses(ctx, req)

		// assert
		assert.Nil(t, err)
		for _, v := range resp.GetStatuses() {
			assert.Equal(t, learnosityStatuses[v.LearningMaterialId], v.IsCompleted)
		}
		mock.AssertExpectationsForObjects(t, assessmentUsecase)
	})

	t.Run("Return correct status when there are both quiz type LMs", func(t *testing.T) {
		t.Parallel()
		// arrange
		assessmentUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		sut := NewAssessmentService(assessmentUsecase)
		nonQuizStatuses := map[string]bool{
			"LM1": true,
		}
		learnosityStatuses := map[string]bool{
			"LM1": false,
			"LM2": true,
		}
		assessmentUsecase.On("ListNonQuizLearningMaterialStatuses", ctx, courseID, userID, lmIDs).
			Once().Return(nonQuizStatuses, nil)
		assessmentUsecase.On("ListLearnositySessionStatuses", ctx, courseID, userID, lmIDs).
			Once().Return(learnosityStatuses, nil)
		expectedStatuses := map[string]bool{
			"LM1": true,
			"LM2": true,
		}
		// act
		resp, err := sut.GetLearningMaterialStatuses(ctx, req)

		// assert
		assert.Nil(t, err)
		for _, v := range resp.GetStatuses() {
			assert.Equal(t, expectedStatuses[v.LearningMaterialId], v.IsCompleted)
		}
		mock.AssertExpectationsForObjects(t, assessmentUsecase)
	})

	t.Run("return internal error when something gone wrong", func(t *testing.T) {
		t.Parallel()
		// arrange
		assessmentUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		var nonQuizStatus map[string]bool
		sut := NewAssessmentService(assessmentUsecase)
		rootErr := fmt.Errorf("%s", "root")
		expectedErr := errors.NewGrpcError(rootErr, transport.GrpcErrorMap)
		assessmentUsecase.On("ListNonQuizLearningMaterialStatuses", ctx, courseID, userID, lmIDs).
			Once().Return(nonQuizStatus, rootErr)

		// act
		resp, err := sut.GetLearningMaterialStatuses(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, assessmentUsecase)
	})
}

func TestMergeTwoMaps(t *testing.T) {
	t.Parallel()

	t.Run("get only the first when the second contains nothing", func(t *testing.T) {
		// arrange
		m1 := map[string]bool{
			"apple":  true,
			"banana": false,
		}
		m2 := map[string]bool{}
		expected := map[string]bool{
			"apple":  true,
			"banana": false,
		}

		// act
		merged := mergeTwoMaps(m1, m2)

		// assert
		assert.Equal(t, expected, merged)
	})

	t.Run("get only the second when the first contains nothing", func(t *testing.T) {
		// arrange
		var m1 map[string]bool
		m2 := map[string]bool{
			"orange": true,
			"banana": true,
		}
		expected := map[string]bool{
			"orange": true,
			"banana": true,
		}

		// act
		merged := mergeTwoMaps(m1, m2)

		// assert
		assert.Equal(t, expected, merged)
	})

	t.Run("get merged map with overwritten values taking precedence", func(t *testing.T) {
		// arrange
		m1 := map[string]bool{
			"apple":  true,
			"banana": false,
		}
		m2 := map[string]bool{
			"orange": true,
			"banana": true,
		}
		expected := map[string]bool{
			"apple":  true,
			"banana": true,
			"orange": true,
		}

		// act
		merged := mergeTwoMaps(m1, m2)

		// assert
		assert.Equal(t, expected, merged)
	})
}

func TestAssessmentService_CompleteAssessmentSession(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.CompleteAssessmentSessionRequest{
			SessionId: "session_id",
		}
		mockUsecase.On("CompleteAssessmentSession", mock.Anything, req.SessionId).Once().Return(nil)
		expectedResp := &pb.CompleteAssessmentSessionResponse{}

		// actual
		resp, err := svc.CompleteAssessmentSession(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedResp, resp)
		mock.AssertExpectationsForObjects(t, mockUsecase)
	})

	t.Run("missing sessionID", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.CompleteAssessmentSessionRequest{
			SessionId: "",
		}
		expectedErr := errors.NewGrpcError(errors.NewValidationError("req must have SessionId", nil), transport.GrpcErrorMap)

		// actual
		resp, err := svc.CompleteAssessmentSession(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockUsecase)
	})

	t.Run("unexpected error", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.CompleteAssessmentSessionRequest{
			SessionId: "session_id",
		}
		mockUsecase.On("CompleteAssessmentSession", mock.Anything, req.SessionId).
			Once().Return(fmt.Errorf("unexpected error"))
		expectedErr := errors.NewGrpcError(fmt.Errorf("unexpected error"), transport.GrpcErrorMap)

		// actual
		resp, err := svc.CompleteAssessmentSession(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockUsecase)
	})
}

func TestAssessmentService_GetAssessmentSubmissionDetail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("return invalid argument error code when submission id is missing", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.GetAssessmentSubmissionDetailRequest{
			SubmissionId: "",
		}
		expectedErr := errors.NewGrpcError(errors.NewValidationError("submission id is required", nil), transport.GrpcErrorMap)

		// actual
		resp, err := svc.GetAssessmentSubmissionDetail(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockUsecase)
	})

	t.Run("return 404 error code when not found submission", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.GetAssessmentSubmissionDetailRequest{
			SubmissionId: "X",
		}
		usecaseErr := errors.NewEntityNotFoundError("submission is not found", nil)
		expectedErr := errors.NewGrpcError(usecaseErr, transport.GrpcErrorMap)
		mockUsecase.On("GetAssessmentSubmissionDetail", ctx, "X").Once().Return(nil, usecaseErr)

		// actual
		resp, err := svc.GetAssessmentSubmissionDetail(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockUsecase)
	})

	t.Run("return general error code when usecase encountered an error", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.GetAssessmentSubmissionDetailRequest{
			SubmissionId: "X",
		}
		usecaseErr := errors.NewDBError("submission is not found", nil)
		expectedErr := errors.NewGrpcError(usecaseErr, transport.GrpcErrorMap)
		mockUsecase.On("GetAssessmentSubmissionDetail", ctx, "X").Once().Return(nil, usecaseErr)

		// actual
		resp, err := svc.GetAssessmentSubmissionDetail(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockUsecase)
	})

	t.Run("return submission details when there is no error", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.GetAssessmentSubmissionDetailRequest{
			SubmissionId: "Some ID",
		}
		sub := &domain.Submission{
			ID:                "Some ID",
			SessionID:         idutil.ULIDNow(),
			AssessmentID:      idutil.ULIDNow(),
			StudentID:         idutil.ULIDNow(),
			AllocatedMarkerID: idutil.ULIDNow(),
			GradingStatus:     domain.GradingStatusInProgress,
			MaxScore:          20,
			GradedScore:       10,
			MarkedBy:          idutil.ULIDNow(),
			MarkedAt:          nil,
			FeedBackSessionID: "",
			FeedBackBy:        "",
			CreatedAt:         time.Time{},
			CompletedAt:       time.Time{},
		}
		expectedDetail := &pb.GetAssessmentSubmissionDetailResponse{
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
		mockUsecase.On("GetAssessmentSubmissionDetail", ctx, "Some ID").Once().Return(sub, nil)

		// actual
		resp, err := svc.GetAssessmentSubmissionDetail(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedDetail, resp)
		mock.AssertExpectationsForObjects(t, mockUsecase)
	})
}

func TestAssessmentService_AllocateMarkerSubmissions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// For assign allocate marker
	submissionIdOne := "SubmissionIdOne"
	allocatedUserIdForOne := "allocatedUserIdForOne"

	// For unassign allocate marker
	submissionIdTwo := "submissionIdTwo"
	allocatedUserIdForTwo := ""

	t.Run("Happy case", func(t *testing.T) {
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}
		req := &pb.AllocateMarkerSubmissionsRequest{
			AllocateMarkerSubmissions: []*pb.AllocateMarkerSubmissionsRequest_AllocateMarkerSubmission{
				{
					SubmissionId:    submissionIdOne,
					AllocatedUserId: allocatedUserIdForOne,
				},
				{
					SubmissionId:    submissionIdTwo,
					AllocatedUserId: allocatedUserIdForTwo,
				},
			},
		}

		expectedSubmissions := []domain.Submission{{
			ID:                submissionIdOne,
			AllocatedMarkerID: allocatedUserIdForOne,
		}, {
			ID:                submissionIdTwo,
			AllocatedMarkerID: allocatedUserIdForTwo,
		}}

		mockUsecase.On("AllocateMarkerSubmissions", mock.Anything, expectedSubmissions).Once().Return(nil)

		// actual
		resp, err := svc.AllocateMarkerSubmissions(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, &pb.AllocateMarkerSubmissionsResponse{}, resp)
	})

	t.Run("usecase AllocateMarkerSubmissions error", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.AllocateMarkerSubmissionsRequest{
			AllocateMarkerSubmissions: []*pb.AllocateMarkerSubmissionsRequest_AllocateMarkerSubmission{
				{
					SubmissionId:    submissionIdOne,
					AllocatedUserId: allocatedUserIdForOne,
				},
				{
					SubmissionId:    submissionIdTwo,
					AllocatedUserId: allocatedUserIdForTwo,
				},
			},
		}

		usecaseErr := fmt.Errorf("AllocateMarkerSubmissions error")
		mockUsecase.On("AllocateMarkerSubmissions", mock.Anything, mock.Anything).Once().Return(usecaseErr)

		// actual
		resp, err := svc.AllocateMarkerSubmissions(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, errors.NewGrpcError(usecaseErr, transport.GrpcErrorMap), err)
	})

	t.Run("Missing SubmissionId", func(t *testing.T) {
		// arrange
		mockUsecase := &mock_usecase.MockAssessmentUsecaseImpl{}
		svc := &AssessmentService{
			AssessmentUsecase: mockUsecase,
		}

		req := &pb.AllocateMarkerSubmissionsRequest{
			AllocateMarkerSubmissions: []*pb.AllocateMarkerSubmissionsRequest_AllocateMarkerSubmission{
				{
					SubmissionId:    "",
					AllocatedUserId: allocatedUserIdForOne,
				},
			},
		}

		// actual
		resp, err := svc.AllocateMarkerSubmissions(ctx, req)

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, errors.NewGrpcError(errors.NewValidationError("req must have the SubmissionId", nil), transport.GrpcErrorMap), err)
	})
}
