package services

import (
	"context"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repositories "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewCourseModifierService create CourseModifierService
func NewCourseModifierService(eurekaDBTrace database.Ext, lessonDBTrace database.Ext, db database.Ext, oldCourseService *CourseService, unleashClientIns unleashclient.ClientInstance, env string) *CourseModifierService {
	return &CourseModifierService{
		EurekaDBTrace:    eurekaDBTrace,
		DB:               db,
		LessonDBTrace:    lessonDBTrace,
		OldCourseService: oldCourseService,
		UnleashClientIns: unleashClientIns,
		QuizRepo:         &bob_repositories.QuizRepo{},
		QuizSetRepo:      &bob_repositories.QuizSetRepo{},
		LessonGroupRepo:  &bob_repositories.LessonGroupRepo{},
		CourseRepo:       &bob_repositories.CourseRepo{},
		Env:              env,
	}
}

// CourseModifierService implements bob proto CourseModifierServiceServer
type CourseModifierService struct {
	ypb.UnimplementedCourseModifierServiceServer
	Env              string
	EurekaDBTrace    database.Ext
	LessonDBTrace    database.Ext
	DB               database.Ext
	OldCourseService *CourseService
	UnleashClientIns unleashclient.ClientInstance

	QuizRepo interface {
		DeleteByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) error
	}

	QuizSetRepo interface {
		GetQuizSetsContainQuiz(ctx context.Context, db database.QueryExecer, quizID pgtype.Text) (entities_bob.QuizSets, error)
		GetQuizSetsOfLOContainQuiz(ctx context.Context, db database.QueryExecer, loID pgtype.Text, quizID pgtype.Text) (entities_bob.QuizSets, error)
		Create(ctx context.Context, db database.QueryExecer, quizset *entities_bob.QuizSet) error
		Delete(ctx context.Context, db database.QueryExecer, id pgtype.Text) error
		GetQuizSetByLoID(ctx context.Context, db database.QueryExecer, loID pgtype.Text) (*entities_bob.QuizSet, error)
	}

	CourseRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*entities_bob.Course, error)
	}

	LessonGroupRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities_bob.LessonGroup) error
	}
}

func (s *CourseModifierService) AttachMaterialsToCourse(ctx context.Context, req *ypb.AttachMaterialsToCourseRequest) (*ypb.AttachMaterialsToCourseResponse, error) {
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, "Connect unleash server failed: %s"+err.Error())
	}
	course, err := s.CourseRepo.FindByID(ctx, s.DB, database.Text(req.CourseId))
	if err != nil {
		return nil, status.Error(codes.Internal, "s.CourseRepo.FindByID: "+err.Error())
	}

	if course == nil {
		return nil, status.Error(codes.NotFound, "not found course")
	}

	entity := &entities_bob.LessonGroup{}
	database.AllNullEntity(entity)
	if err := multierr.Combine(
		entity.LessonGroupID.Set(req.LessonGroupId),
		entity.CourseID.Set(req.CourseId),
		entity.MediaIDs.Set(req.MaterialIds),
	); err != nil {
		return nil, err
	}

	db := s.DB
	if isUnleashToggled {
		db = s.LessonDBTrace
	}

	if err := s.LessonGroupRepo.BulkUpsert(ctx, db, []*entities_bob.LessonGroup{entity}); err != nil {
		return nil, err
	}

	return &ypb.AttachMaterialsToCourseResponse{}, nil
}
