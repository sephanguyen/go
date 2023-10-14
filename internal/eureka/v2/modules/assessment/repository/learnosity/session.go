package learnosity

import (
	"context"
	"encoding/json"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	learnosity_entity "github.com/manabie-com/backend/internal/golibs/learnosity/entity"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
)

type SessionRepo struct {
	http learnosity.HTTP
	api  learnosity.DataAPI
}

func NewSessionRepo(http learnosity.HTTP, api learnosity.DataAPI) repository.LearnositySessionRepo {
	return &SessionRepo{http: http, api: api}
}

func (s *SessionRepo) GetSessionStatuses(ctx context.Context, security learnosity.Security, request learnosity.Request) (rs []domain.Session, err error) {
	ctx, span := interceptors.StartSpan(ctx, "LearnositySessionRepo.GetSessionStatuses")
	defer span.End()
	endpoint := learnosity.EndpointDataAPISessionsStatuses

	results, err := s.api.RequestIterator(ctx, s.http, endpoint, security, request)
	if err != nil {
		return rs, errors.NewLearnosityError("LearnositySessionRepo.GetSessionStatuses", err)
	}

	for _, r := range results {
		records := r.Meta.Records()
		ssr := make([]learnosity_entity.SessionStatus, records)
		_ = json.Unmarshal(r.Data, &ssr)
		childSessions := sliceutils.Map(ssr, func(s learnosity_entity.SessionStatus) domain.Session {
			return domain.Session{
				ID:           s.SessionID,
				AssessmentID: s.ActivityID,
				UserID:       s.UserID,
				Status:       toSessionDomainStatus(s.Status),
				CompletedAt:  s.DtCompleted,
			}
		})
		rs = append(rs, childSessions...)
	}
	return rs, nil
}

func toSessionDomainStatus(s learnosity.SessionStatus) domain.SessionStatus {
	var ds domain.SessionStatus
	switch s {
	case learnosity.SessionStatusIncomplete:
		ds = domain.SessionStatusIncomplete
	case learnosity.SessionStatusCompleted:
		ds = domain.SessionStatusCompleted
	default:
		ds = domain.SessionStatusNone
	}
	return ds
}

func (s *SessionRepo) GetSessionResponses(ctx context.Context, security learnosity.Security, request learnosity.Request) (domain.Sessions, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearnositySessionRepo.GetSessionResponses")
	defer span.End()
	endpoint := learnosity.EndpointDataAPISessionsResponses

	resps, err := s.api.RequestIterator(ctx, s.http, endpoint, security, request)
	if err != nil {
		return nil, errors.NewLearnosityError("LearnositySessionRepo.GetSessionResponses", err)
	}

	results := make(domain.Sessions, 0)
	for _, r := range resps {
		childResps := make([]learnosity_entity.SessionResponse, 0, r.Meta.Records())
		if err = json.Unmarshal(r.Data, &childResps); err != nil {
			return nil, errors.NewLearnosityError("json.Unmarshal", err)
		}

		for _, cr := range childResps {
			sessionStatus := domain.SessionStatusNone
			switch cr.Status {
			case "Incomplete":
				sessionStatus = domain.SessionStatusIncomplete
			case "Completed":
				sessionStatus = domain.SessionStatusCompleted
			}

			results = append(results, domain.Session{
				ID:          cr.SessionID,
				MaxScore:    cr.MaxScore,
				GradedScore: cr.Score,
				Status:      sessionStatus,
				CreatedAt:   cr.DtStarted,
				CompletedAt: cr.DtCompleted,
			})
		}
	}

	return results, nil
}
