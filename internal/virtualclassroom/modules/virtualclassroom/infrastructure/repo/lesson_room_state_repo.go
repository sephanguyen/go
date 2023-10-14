package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type LessonRoomStateRepo struct{}

func (l *LessonRoomStateRepo) UpsertRecordingState(ctx context.Context, db database.QueryExecer, lessonID string, recording *domain.CompositeRecordingState) error {
	dto := &LessonRoomState{}
	if err := dto.Recording.Set(database.JSONB(recording)); err != nil {
		return err
	}

	return l.upsertLessonRoomState(ctx, db, lessonID, dto.Recording, "recording")
}

func (l *LessonRoomStateRepo) UpdateRecordingState(ctx context.Context, db database.QueryExecer, lessonID string, recording *domain.CompositeRecordingState) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.UpdateRecordingState")
	defer span.End()
	dto, err := toRecordingStateDto(lessonID, recording)
	if err != nil {
		return err
	}
	query := `UPDATE lesson_room_states SET recording = $2, updated_at = now(), deleted_at = NULL
		WHERE lesson_id = $1`
	_, err = db.Exec(ctx, query, &dto.LessonID, &dto.Recording)
	return err
}

func toRecordingStateDto(lessonID string, recording *domain.CompositeRecordingState) (*LessonRoomState, error) {
	if lessonID == "" {
		return nil, fmt.Errorf("lessonID should not empty")
	}
	w, err := json.Marshal(recording)
	if err != nil {
		return nil, fmt.Errorf("error when json.Marshal %v", recording)
	}

	dto := &LessonRoomState{}
	if err := multierr.Combine(
		dto.LessonID.Set(lessonID),
		dto.Recording.Set(database.JSONB(w)),
	); err != nil {
		return nil, err
	}
	return dto, nil
}

func (l *LessonRoomStateRepo) upsertLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID string, value interface{}, fieldName string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.UpsertLessonRoomState with field "+fieldName)
	defer span.End()
	dto := &LessonRoomState{}
	if err := dto.LessonID.Set(lessonID); err != nil {
		return err
	}
	dto.PreInsert()
	query := fmt.Sprintf(`INSERT INTO lesson_room_states (lesson_room_state_id, lesson_id, %s) VALUES ($1, $2, $3)
				ON CONFLICT ON CONSTRAINT unique__lesson_id 
					DO UPDATE SET %s = $3, updated_at = now(), deleted_at = NULL`, fieldName, fieldName)
	_, err := db.Exec(ctx, query, &dto.LessonRoomStateID, &dto.LessonID, &value)

	return err
}

func (l *LessonRoomStateRepo) UpsertCurrentPollingState(ctx context.Context, db database.QueryExecer, lessonID string, polling *domain.CurrentPolling) error {
	dto := &LessonRoomState{}
	if err := dto.CurrentPolling.Set(database.JSONB(polling)); err != nil {
		return err
	}

	return l.upsertLessonRoomState(ctx, db, lessonID, dto.CurrentPolling, "current_polling")
}

func (l *LessonRoomStateRepo) UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, wbZoomState *domain.WhiteboardZoomState) error {
	dto := &LessonRoomState{}
	if err := dto.WhiteboardZoomState.Set(database.JSONB(wbZoomState)); err != nil {
		return err
	}

	return l.upsertLessonRoomState(ctx, db, lessonID, dto.WhiteboardZoomState, "whiteboard_zoom_state")
}

func (l *LessonRoomStateRepo) UpsertSpotlightState(ctx context.Context, db database.QueryExecer, lessonID string, spotlightedUser string) error {
	dto := &LessonRoomState{}
	if err := dto.SpotlightedUser.Set(spotlightedUser); err != nil {
		return err
	}

	return l.upsertLessonRoomState(ctx, db, lessonID, dto.SpotlightedUser, "spotlighted_user")
}

func (l *LessonRoomStateRepo) UpsertLiveLessonSessionTime(ctx context.Context, db database.QueryExecer, lessonID string, sessionTime time.Time) error {
	dto := &LessonRoomState{}
	if err := dto.SessionTime.Set(sessionTime); err != nil {
		return err
	}

	return l.upsertLessonRoomState(ctx, db, lessonID, dto.SessionTime, "session_time")
}

func (l *LessonRoomStateRepo) UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.UnSpotlight")
	defer span.End()

	query := `UPDATE lesson_room_states SET spotlighted_user = NULL, updated_at = now() 
			  WHERE lesson_id = $1 AND spotlighted_user IS NOT NULL`
	_, err := db.Exec(ctx, query, &lessonID)

	return err
}

func (l *LessonRoomStateRepo) UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID string, currentMaterial *domain.CurrentMaterial) error {
	dto := &LessonRoomState{}
	if err := dto.CurrentMaterial.Set(database.JSONB(currentMaterial)); err != nil {
		return err
	}

	return l.upsertLessonRoomState(ctx, db, lessonID, dto.CurrentMaterial, "current_material")
}

func (l *LessonRoomStateRepo) GetLessonRoomStateByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (*domain.LessonRoomState, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.GetLessonRoomStateByLessonID")
	defer span.End()

	dto := &LessonRoomState{}
	fields, values := dto.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE lesson_id = $1 
		AND deleted_at IS NULL `,
		strings.Join(fields, ", "),
		dto.TableName(),
	)
	err := db.QueryRow(ctx, query, &lessonID).Scan(values...)
	if err == pgx.ErrNoRows {
		return &domain.LessonRoomState{}, domain.ErrLessonRoomStateNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "Scan: %w")
	}

	return dto.ToLessonRoomState()
}
