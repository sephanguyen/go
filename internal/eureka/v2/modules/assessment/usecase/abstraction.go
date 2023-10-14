package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository"
	learnosity_repo "github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/learnosity"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres"
	book_repository "github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository"
	book_postgres "github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository/postgres"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

type AssessmentUsecase interface {
	GetAssessmentSignedRequest(ctx context.Context, session domain.Session, hostDomain, config string) (string, error)
	ListAssessmentAttemptHistory(ctx context.Context, userID, courseID, lmID string) ([]domain.Session, error)
	ListLearnositySessionStatuses(ctx context.Context, courseID, userID string, learningMaterialIDs []string) (map[string]bool, error)
	ListNonQuizLearningMaterialStatuses(ctx context.Context, courseID, userID string, learningMaterialIDs []string) (map[string]bool, error)
	CompleteAssessmentSession(ctx context.Context, sessionID string) error
	GetAssessmentSubmissionDetail(ctx context.Context, id string) (*domain.Submission, error)
	AllocateMarkerSubmissions(ctx context.Context, submissions []domain.Submission) error
}

type AssessmentUsecaseImpl struct {
	DB               database.Ext
	LearnosityConfig configurations.LearnosityConfig
	HTTP             learnosity.HTTP
	DataAPI          learnosity.DataAPI

	AssessmentRepo        repository.AssessmentRepo
	SubmissionRepo        repository.SubmissionRepo
	FeedbackSessionRepo   repository.FeedbackSessionRepo
	AssessmentSessionRepo repository.AssessmentSessionRepo
	StudentEventLogRepo   repository.StudentEventLogRepo
	LearningMaterialRepo  book_repository.LearningMaterialRepo
	LearnositySessionRepo repository.LearnositySessionRepo
}

func NewAssessmentUsecase(db database.Ext, learnosityConfig configurations.LearnosityConfig,
	http learnosity.HTTP, api learnosity.DataAPI) AssessmentUsecase {
	return &AssessmentUsecaseImpl{
		DB:                    db,
		LearnosityConfig:      learnosityConfig,
		HTTP:                  http,
		DataAPI:               api,
		AssessmentRepo:        &postgres.AssessmentRepo{},
		SubmissionRepo:        &postgres.SubmissionRepo{},
		StudentEventLogRepo:   &postgres.StudentEventLogRepo{},
		FeedbackSessionRepo:   &postgres.FeedbackSessionRepo{},
		AssessmentSessionRepo: &postgres.AssessmentSessionRepo{},
		LearningMaterialRepo:  &book_postgres.LearningMaterialRepo{DB: db},
		LearnositySessionRepo: learnosity_repo.NewSessionRepo(http, api),
	}
}
