package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	LOCompletenessQuizEventType       = "quiz_finished"
	LOCompletenessVideoEventType      = "video_finished"
	LOCompletenessStudyGuideEventType = "study_guide_finished"
)

type (
	IStudentEventLogRepository interface {
		Create(ctx context.Context, db database.QueryExecer, ss []*entities.StudentEventLog) error
	}
)

type StudentEventLogModifierService struct {
	DB                  database.Ext
	JSM                 nats.JetStreamManagement
	StudentEventLogRepo IStudentEventLogRepository

	StudyPlanItemRepo interface {
		UpdateCompletedAtByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, completedAt pgtype.Timestamptz) error
	}
	LearningTimeCalculator    *LearningTimeCalculator
	StudentLOCompletenessRepo interface {
		UpsertLOCompleteness(context.Context, database.QueryExecer, []*entities.StudentsLearningObjectivesCompleteness) error
	}
}

func NewStudentEventLogModifierService(db database.Ext, jsm nats.JetStreamManagement, usermgmtUserReaderServiceClient upb.UserReaderServiceClient) epb.StudentEventLogModifierServiceServer {
	return &StudentEventLogModifierService{
		DB:                  db,
		JSM:                 jsm,
		StudentEventLogRepo: &repositories.StudentEventLogRepo{},
		StudyPlanItemRepo:   &repositories.StudyPlanItemRepo{},
		LearningTimeCalculator: &LearningTimeCalculator{
			DB:                           db,
			StudentEventLogRepo:          &repositories.StudentEventLogRepo{},
			StudentLearningTimeDailyRepo: &repositories.StudentLearningTimeDailyRepo{},
			UsermgmtUserReaderService:    usermgmtUserReaderServiceClient,
		},
		StudentLOCompletenessRepo: &repositories.StudentsLearningObjectivesCompletenessRepo{},
	}
}

func toStudentEventLogEntity(ctx context.Context, p *epb.StudentEventLog) (*entities.StudentEventLog, error) {
	ep := new(entities.StudentEventLog)
	ep.ID.Set(nil)
	ep.StudentID.Set(interceptors.UserIDFromContext(ctx))
	ep.StudyPlanID.Set(nil)
	ep.LearningMaterialID.Set(nil)
	ep.EventID.Set(p.EventId)
	ep.EventType.Set(p.EventType)

	if p.CreatedAt == nil {
		ep.CreatedAt.Set(nil)
	} else {
		ep.CreatedAt.Set(p.CreatedAt.AsTime())
	}

	if p.Payload == nil {
		ep.Payload.Set(nil)
	} else {
		ep.Payload.Set(p.Payload)
	}

	if p.ExtraPayload != nil && len(p.ExtraPayload) > 0 {
		convertedPayload := make(map[string]any, len(p.ExtraPayload))
		for k, v := range p.ExtraPayload {
			convertedPayload[k] = v
		}
		newPayload, err := database.AppendJSONBProps(ep.Payload, convertedPayload)
		if err != nil {
			return nil, err
		}
		ep.Payload = newPayload
	}

	return ep, nil
}

func (s *StudentEventLogModifierService) CreateStudentEventLogs(ctx context.Context, req *epb.CreateStudentEventLogsRequest) (*epb.CreateStudentEventLogsResponse, error) {
	currUserID := interceptors.UserIDFromContext(ctx)

	eventLogs := make([]*entities.StudentEventLog, 0, len(req.StudentEventLogs))
	for _, log := range req.StudentEventLogs {
		entity, err := toStudentEventLogEntity(ctx, log)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("toStudentEventLogEntity: %w", err).Error())
		}
		eventLogs = append(eventLogs, entity)
	}
	if err := s.StudentEventLogRepo.Create(ctx, s.DB, eventLogs); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("StudentEventLogRepo.Create: %w", err).Error())
	}

	req.StudentId = currUserID
	if err := s.handleStudentEvent(ctx, req, s.DB); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("handleStudentEvent: %w", err).Error())
	}

	errChan := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := s.LearningTimeCalculator.CalculateLearningTimeByEventLogs(ctx, currUserID, req.GetStudentEventLogs()); err != nil {
			errChan <- fmt.Errorf("LearningTimeCalculator.CalculateLearningTimeByEventLogs: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.upsertLOCompleteness(ctx, currUserID, req.GetStudentEventLogs()); err != nil {
			errChan <- fmt.Errorf("upsertLOCompleteness: %w", err)
		}
	}()
	go func() {
		wg.Wait()
		close(errChan)
	}()
	var err error
	for er := range errChan {
		if er == nil {
			continue
		}
		err = multierr.Append(err, er)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &epb.CreateStudentEventLogsResponse{Successful: true}, nil
}

func (s *StudentEventLogModifierService) handleStudentEvent(ctx context.Context, req *epb.CreateStudentEventLogsRequest, db database.Ext) error {
	for _, log := range req.StudentEventLogs {
		if log.EventType == LOEventType {
			if log.Payload == nil {
				continue
			}

			if log.Payload.Event == CompletedEvent {
				loID := log.Payload.LoId
				if loID == "" {
					return errors.New("missing lo_id in req")
				}
				studyPlanItemID := log.Payload.StudyPlanItemId
				if studyPlanItemID == "" {
					return errors.New("missing study_plan_item_id in req")
				}
				completedAt := log.CreatedAt
				err := s.StudyPlanItemRepo.UpdateCompletedAtByID(ctx, db, database.Text(studyPlanItemID), database.Timestamptz(completedAt.AsTime()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *StudentEventLogModifierService) upsertLOCompleteness(ctx context.Context, studentID string, logs []*epb.StudentEventLog) error {
	var (
		los []*entities.StudentsLearningObjectivesCompleteness

		highestScoreLOs []*entities.StudentsLearningObjectivesCompleteness
	)

	for _, log := range logs {
		if lo := toStudentLOCompletenessEntity(studentID, log); lo != nil {
			los = append(los, lo)

			if lo.HighestQuizScore.Status == pgtype.Present {
				highestScoreLOs = append(highestScoreLOs, &entities.StudentsLearningObjectivesCompleteness{
					StudentID:        lo.StudentID,
					LoID:             lo.LoID,
					HighestQuizScore: lo.HighestQuizScore,
				})
			}
		}
	}
	if len(los) == 0 {
		return nil
	}

	if err := s.StudentLOCompletenessRepo.UpsertLOCompleteness(ctx, s.DB, los); err != nil {
		return fmt.Errorf("s.StudentLOCompletenessRepo.UpsertLOCompleteness: %w", err)
	}

	if len(highestScoreLOs) > 0 {
		if err := s.StudentLOCompletenessRepo.UpsertLOCompleteness(ctx, s.DB, highestScoreLOs); err != nil {
			return fmt.Errorf("s.StudentLOCompletenessRepo.UpsertLOCompleteness: %w", err)
		}
	}

	return nil
}

func toStudentLOCompletenessEntity(studentID string, p *epb.StudentEventLog) *entities.StudentsLearningObjectivesCompleteness {
	var (
		e = new(entities.StudentsLearningObjectivesCompleteness)
	)
	e.StudentID.Set(studentID)

	switch p.EventType {
	//nolint:errcheck
	case LOCompletenessQuizEventType:
		e.IsFinishedQuiz.Set(true)
		_ = e.FinishedQuizAt.Set(p.CreatedAt.AsTime())
		if p.Payload.TotalQuestions != 0 {
			score := (float64(p.Payload.Correct) / float64(p.Payload.TotalQuestions)) * 100
			e.FirstQuizCorrectness.Set(score)
			e.HighestQuizScore.Set(score)
		}
	case LOCompletenessVideoEventType:
		e.IsFinishedVideo.Set(true)
	case LOCompletenessStudyGuideEventType:
		e.IsFinishedStudyGuide.Set(true)
	default:
		return nil
	}

	if p.Payload != nil {
		e.LoID.Set(p.Payload.LoId)
	}
	return e
}
