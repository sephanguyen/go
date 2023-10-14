package services

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentSubmissionReaderService struct {
	DB database.Ext

	StudentEventLogRepo interface {
		LogsQuestionSubmitionByLO(ctx context.Context, db database.QueryExecer, studentID string, loIDs pgtype.TextArray) (map[string][]*repositories.QuestionSubmissionResult, error)
	}

	QuizSetRepo interface {
		CountQuizOnLO(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error)
	}
}

func NewStudentSubmissionReaderService(db database.Ext) *StudentSubmissionReaderService {
	return &StudentSubmissionReaderService{
		DB:                  db,
		StudentEventLogRepo: &repositories.StudentEventLogRepo{},
		QuizSetRepo:         &repositories.QuizSetRepo{},
	}
}

func (s *StudentSubmissionReaderService) RetrieveStudentSubmissionHistoryByLoIDs(ctx context.Context, req *pb.RetrieveStudentSubmissionHistoryByLoIDsRequest) (*pb.RetrieveStudentSubmissionHistoryByLoIDsResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	history := make([]*pb.RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory, len(req.LoIds))

	submissionResults, err := s.StudentEventLogRepo.LogsQuestionSubmitionByLO(ctx, s.DB, userID, database.TextArray(req.LoIds))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "StudentEventLogRepo.LogsQuestionSubmitionByLO %v", err)
	}

	totalQuestions, err := s.QuizSetRepo.CountQuizOnLO(ctx, s.DB, database.TextArray(req.LoIds))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "QuizSetRepo.CountQuizOnLO %v", err)
	}

	for i, loID := range req.LoIds {
		results := make([]*pb.RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult, 0, len(submissionResults[loID]))
		for _, submissionResult := range submissionResults[loID] {
			results = append(results, &pb.RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult{
				QuestionId: submissionResult.QuestionID,
				Correct:    submissionResult.Correct,
			})
		}

		history[i] = &pb.RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory{
			LoId:          loID,
			Results:       results,
			TotalQuestion: totalQuestions[loID],
		}
	}

	return &pb.RetrieveStudentSubmissionHistoryByLoIDsResponse{
		Submissions: history,
	}, nil
}
