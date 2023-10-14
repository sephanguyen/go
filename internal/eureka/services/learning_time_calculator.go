package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bob_legacy "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LearningTimeCalculator struct {
	DB database.Ext
	// Logger              *zap.Logger
	StudentEventLogRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, studentID, sessionID string, from, to *pgtype.Timestamptz) ([]*entities.StudentEventLog, error)
	}
	StudentLearningTimeDailyRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, s *entities.StudentLearningTimeDaily) error
		Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
	}
	UsermgmtUserReaderService interface {
		SearchBasicProfile(ctx context.Context, in *upb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*upb.SearchBasicProfileResponse, error)
	}
}

const (
	// LO event names
	LOEventStarted   = "started"
	LOEventPaused    = "paused"
	LOEventResumed   = "resumed"
	LOEventCompleted = "completed"
	LOEventExited    = "exited"
)

/*
because we face a lot of weird cases
+ case 1: started -> started - > completed: we will get the first started in this case, omit the started between the first start and completed
+ case 2: started -> completed - > started -> completed: we will device it to 2 batch and calculate each batch
+ case 3: started -> completed -> started: we will get batch 1, omit the second started bcz we don't have second completed or exited
+ case 4: started -> completed -> completed : we will get the last completed
*/
// about the completed/exited event: if we already have completed event we will omit exit and NOT vice versa
// Calculate learning time by get input array of logs from a student with the same session_id
//gocyclo:ignore
func (c *LearningTimeCalculator) Calculate(logs []*entities.StudentEventLog) (learningTime time.Duration, completedAt *time.Time, err error) {
	var (
		started, completed, exited *time.Time
		paused, resumed            []time.Time
	)
	var isStart bool
	var totalPausedTime, batchLearningTime time.Duration
	isNewFlow := false
	isCompletedAt := false
	for _, log := range logs {
		payload := make(map[string]interface{})
		if err := log.Payload.AssignTo(&payload); err != nil {
			return 0, nil, errors.Wrap(err, "log.Payload.AssignTo")
		}
		event, ok := payload["event"].(string)
		if !ok {
			continue
		}
		t := log.CreatedAt.Time.UTC()
		switch event {
		case LOEventStarted:
			if !isStart {
				isStart = true
				started = &t
				isNewFlow = true
			} else if completed != nil || exited != nil {
				started = &t
				isNewFlow = true
				// reset value when start new flow
				resumed = make([]time.Time, 0)
				paused = make([]time.Time, 0)
				totalPausedTime = 0
			}
		case LOEventCompleted:
			if isNewFlow && started != nil { // only excute when face started event before
				completed = &t
				for len(resumed) > 0 {
					if len(paused) == 0 {
						break
					}
					totalPausedTime += resumed[0].Sub(paused[0])
					resumed = resumed[1:]
					paused = paused[1:]
				}
				batchLearningTime = completed.Sub(*started)
				batchLearningTime -= totalPausedTime

				learningTime += batchLearningTime
				completedAt = completed
				isCompletedAt = true // avoid re-calculate in exit event
				// reset value, in case we face the second completed which have completed before (case 4)

				isNewFlow = false
				resumed = make([]time.Time, 0)
				paused = make([]time.Time, 0)
				totalPausedTime = 0
			} else if started != nil { // if we face another completed, we will get the second. case 4
				for len(resumed) > 0 {
					if len(paused) == 0 {
						break
					}
					totalPausedTime += resumed[0].Sub(paused[0])
					resumed = resumed[1:]
					paused = paused[1:]
				}
				batchLearningTime = t.Sub(*completedAt)
				completed = &t
				completedAt = completed
				batchLearningTime -= totalPausedTime
				learningTime += batchLearningTime
			}
		case LOEventExited:
			// the flow like completed but before go insight, we have make sure we don't have completed event
			// only excute when face started event before
			if isNewFlow && !isCompletedAt && started != nil {
				exited = &t
				for len(resumed) > 0 {
					if len(paused) == 0 {
						break
					}
					totalPausedTime += resumed[0].Sub(paused[0])
					resumed = resumed[1:]
					paused = paused[1:]
				}
				batchLearningTime = exited.Sub(*started)
				batchLearningTime -= totalPausedTime

				learningTime += batchLearningTime

				completedAt = exited
				isCompletedAt = true
				isNewFlow = false
			} else if !isNewFlow && !isCompletedAt && started != nil {
				for len(resumed) > 0 {
					if len(paused) == 0 {
						break
					}
					totalPausedTime += resumed[0].Sub(paused[0])
					resumed = resumed[1:]
					paused = paused[1:]
				}
				batchLearningTime = t.Sub(*completedAt)
				exited = &t
				completedAt = exited
				batchLearningTime -= totalPausedTime
				learningTime += batchLearningTime
			}
		case LOEventPaused:
			if len(resumed) == len(paused) {
				paused = append(paused, t)
			}
		case LOEventResumed:
			if len(paused) == len(resumed)+1 {
				resumed = append(resumed, t)
			}
		}
	}
	if started == nil || (completed == nil && exited == nil) {
		return 0, nil, nil
	}
	// avoid negative learning
	if learningTime < 0 {
		learningTime = 0
	}
	return
}
func (c *LearningTimeCalculator) CalculateLearningTimeByEventLogs(ctx context.Context, studentID string, logs []*epb.StudentEventLog) error {
	for _, log := range logs {
		if err := c.calculateLearningTime(ctx, studentID, log); err != nil {
			return errors.Wrap(err, "c.calculateLearningTime")
		}
	}
	return nil
}
func (c *LearningTimeCalculator) calculateLearningTime(ctx context.Context, studentID string, log *epb.StudentEventLog) error {
	if log.EventType != LOEventType {
		return nil
	}
	if log.Payload == nil {
		return nil
	}

	if log.Payload.Event == "" || (log.Payload.Event != "completed" && log.Payload.Event != "exited") {
		return nil
	}
	sessionID := log.Payload.SessionId
	if sessionID == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	logs, err := c.StudentEventLogRepo.Retrieve(ctx, c.DB, studentID, sessionID, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "s.StudentEventLogRepo.Retrieve: studentID: %q, sessionID: %q", studentID, sessionID)
	}
	learningTime, completedAt, err := c.Calculate(logs)
	if err != nil {
		return errors.Wrapf(err, "c.Calculate: studentID: %v", studentID)
	}
	if learningTime == time.Duration(0) {
		return nil
	}
	country, err := c.getStudentCountry(ctx, studentID)
	if err != nil {
		return errors.Wrap(err, "c.getStudentCountry")
	}
	day := timeutil.MidnightIn(bob_legacy.Country(epb.Country_value[country]), *completedAt)
	var pgDay pgtype.Timestamptz
	_ = pgDay.Set(day.UTC())

	if err := database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		dailies, err := c.StudentLearningTimeDailyRepo.Retrieve(ctx, tx, database.Text(studentID), &pgDay, &pgDay, repositories.WithUpdateLock())
		if err != nil {
			return errors.Wrap(err, "s.StudentLearningTimeDailyRepo.Retrieve")
		}
		var daily *entities.StudentLearningTimeDaily
		if len(dailies) > 0 {
			if len(dailies) > 1 {
				return errors.Errorf("expected only 1 results, got: %d", len(dailies))
			}
			daily = dailies[0]
			// check if learning time in this sessionID is calculated or not
			sessions := strings.Split(daily.Sessions.String, ",")
			if golibs.InArrayString(sessionID, sessions) {
				return nil
			}
		}
		var newSessions string
		if daily != nil {
			newSessions = fmt.Sprintf("%s,%s", daily.Sessions.String, sessionID)
		} else {
			newSessions = sessionID
		}

		s := new(entities.StudentLearningTimeDaily)
		s.StudentID.Set(studentID)
		s.LearningTime.Set(int64(learningTime.Seconds()))
		s.Day.Set(day.UTC()) // DB always store UTC time
		s.Sessions.Set(newSessions)
		if err := c.StudentLearningTimeDailyRepo.Upsert(ctx, tx, s); err != nil {
			return errors.Wrap(err, "c.StudentLearningTimeDailyRepo.Upsert")
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "database.ExecInTx")
	}
	return nil
}

func (c *LearningTimeCalculator) getStudentCountry(ctx context.Context, studentID string) (string, error) {
	upbReq := &upb.SearchBasicProfileRequest{
		UserIds: []string{studentID},
		Paging:  &cpb.Paging{Limit: uint32(1)},
	}

	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "GetOutgoingContext: %v", err)
	}
	resp, err := c.UsermgmtUserReaderService.SearchBasicProfile(mdCtx, upbReq)
	if err != nil {
		return "", status.Errorf(codes.Internal, "s.UsermgmtUserReaderService.SearchBasicProfile: %v", err)
	}

	if len(resp.Profiles) == 0 {
		return "", status.Errorf(codes.NotFound, "s.UsermgmtUserReaderService.SearchBasicProfile: user %s not found", studentID)
	}

	return resp.Profiles[0].Country.String(), nil
}
