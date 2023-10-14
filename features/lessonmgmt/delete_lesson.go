package lessonmgmt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/proto"
)

func (s *Suite) createDeletedLessonSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonDeletedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *bpb.EvtLesson_DeletedLessons_:
			stepState.DeletedLessonIDs = r.GetDeletedLessons().LessonIds
			stepState.FoundChanForJetStream <- r.Message
			return true, nil
		}
		return false, errors.New("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonDeleted, opts, handlerLessonDeletedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userDeleteALesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := lpb.DeleteLessonRequest{
		LessonId: stepState.CurrentLessonID,
	}
	ctx, err := s.createDeletedLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createDeletedLessonSubscription: %w", err)
	}
	_, stepState.ResponseErr = lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).DeleteLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userNoLongerSeesTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonRepo{}
	if len(stepState.CurrentLessonID) > 0 {
		_, err := repo.FindByID(ctx, s.BobDB, database.Text(stepState.CurrentLessonID))
		if err == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s be deleted, but not", stepState.CurrentLessonID)
		} else if err.Error() != fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
			return StepStateToContext(ctx, stepState), err
		}
	} else {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current lessonId is empty")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userStillSeesTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonRepo{}
	if len(stepState.CurrentLessonID) > 0 {
		lesson, err := repo.FindByID(ctx, s.BobDB, database.Text(stepState.CurrentLessonID))
		if err == nil {
			if lesson.LessonID.String == stepState.CurrentLessonID {
				return StepStateToContext(ctx, stepState), nil
			}
		} else if err.Error() == fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s not be deleted, but not", stepState.CurrentLessonID)
		}
	} else {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current lessonId is empty")
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("error in lesson repo.FindByID")
}

func (s *Suite) userNoLongerSeesTheLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonReportRepo{}
	if len(stepState.CurrentLessonID) > 0 {
		_, err := repo.FindByLessonID(ctx, s.BobDB, database.Text(stepState.CurrentLessonID))
		if err == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected all lesson report of lesson %s be deleted, but not", stepState.CurrentLessonID)
		} else if err.Error() != fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
			return StepStateToContext(ctx, stepState), err
		}
	} else {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current lessonId is empty")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userStillSeesTheLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonReportRepo{}
	if len(stepState.CurrentLessonID) > 0 {
		lessonReport, err := repo.FindByLessonID(ctx, s.BobDB, database.Text(stepState.CurrentLessonID))
		if err == nil {
			if lessonReport.LessonID.String == stepState.CurrentLessonID {
				return StepStateToContext(ctx, stepState), err
			}
		} else if err.Error() == fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected all lesson report of lesson %s be deleted, but not", stepState.CurrentLessonID)
		}
	} else {
		return StepStateToContext(ctx, stepState), fmt.Errorf("current lessonId is empty")
	}

	return StepStateToContext(ctx, stepState), fmt.Errorf("error in lesson report repo.FindByLessonID")
}
