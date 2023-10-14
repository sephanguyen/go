package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentSubmissionWriterService struct {
	DB database.Ext

	StudentSubmissionRepo interface {
		Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.StudentSubmission, error)
		DeleteByStudyPlanItemIDs(
			ctx context.Context, db database.QueryExecer,
			studyPlanItemIDs pgtype.TextArray, deletedBy pgtype.Text,
		) error
	}

	StudentLatestSubmissionRepo interface {
		DeleteByStudyPlanItemID(
			ctx context.Context, db database.QueryExecer,
			studyPlanItemID, deletedBy pgtype.Text,
		) error
	}

	StudyPlanItemRepo interface {
		UnMarkItemCompleted(ctx context.Context, db database.QueryExecer,
			itemID pgtype.Text) error
	}
}

func NewStudentSubmissionWriterService(db database.Ext) *StudentSubmissionWriterService {
	return &StudentSubmissionWriterService{
		DB:                          db,
		StudentSubmissionRepo:       &repositories.StudentSubmissionRepo{},
		StudentLatestSubmissionRepo: &repositories.StudentLatestSubmissionRepo{},
		StudyPlanItemRepo:           &repositories.StudyPlanItemRepo{},
	}
}

func (s *StudentSubmissionWriterService) DeleteStudentSubmission(ctx context.Context, req *pb.DeleteStudentSubmissionRequest) (*pb.DeleteStudentSubmissionResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	if req.StudentSubmissionId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "StudentSubmissionService.DeleteStudentSubmission: No StudentSubmissionID")
	}

	studentSubmission, err := s.StudentSubmissionRepo.Get(ctx, s.DB, database.Text(req.StudentSubmissionId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "StudentSubmissionService.DeleteStudentSubmission1: %v", err)
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		studyPlanItemIDs := database.TextArray([]string{studentSubmission.StudyPlanItemID.String})
		if err := s.StudentSubmissionRepo.
			DeleteByStudyPlanItemIDs(ctx, tx, studyPlanItemIDs, database.Text(userID)); err != nil {
			return fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission2: %w", err)
		}

		if err := s.StudentLatestSubmissionRepo.
			DeleteByStudyPlanItemID(ctx, tx, studentSubmission.StudyPlanItemID, database.Text(userID)); err != nil {
			return fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission3: %w", err)
		}

		if err := s.StudyPlanItemRepo.UnMarkItemCompleted(ctx, tx, database.Text(studentSubmission.StudyPlanItemID.String)); err != nil {
			return fmt.Errorf("StudentSubmissionService.DeleteStudentSubmission4: %w", err)
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteStudentSubmissionResponse{}, nil
}
