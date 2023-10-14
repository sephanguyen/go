package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type LessonRoomStateRepo struct{}

func (l *LessonRoomStateRepo) UpsertCurrentMaterial(ctx context.Context, db database.QueryExecer, material *domain.CurrentMaterial) (*domain.CurrentMaterial, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.UpsertCurrentMaterial")
	defer span.End()

	dto := &LessonRoomState{}
	if err := multierr.Combine(
		dto.SetCurrentMaterial(material),
		dto.LessonID.Set(material.LessonID),
	); err != nil {
		return nil, err
	}
	dto.PreInsert()
	query := `
			INSERT INTO lesson_room_states (lesson_room_state_id, lesson_id, current_material) VALUES ($1, $2, $3)
			ON CONFLICT ON CONSTRAINT unique__lesson_id 
				DO UPDATE SET current_material = $3, updated_at = now(), deleted_at = NULL`
	_, err := db.Exec(ctx, query, &dto.LessonRoomStateID, &dto.LessonID, &dto.CurrentMaterial)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (l *LessonRoomStateRepo) Spotlight(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.Spotlight")
	defer span.End()

	dto := &LessonRoomState{}
	if err := multierr.Combine(
		dto.LessonID.Set(lessonID),
		dto.SpotlightedUser.Set(userID),
	); err != nil {
		return err
	}
	dto.PreInsert()
	query := `
			INSERT INTO lesson_room_states (lesson_room_state_id, lesson_id, spotlighted_user) VALUES ($1, $2, $3)
			ON CONFLICT ON CONSTRAINT unique__lesson_id 
				DO UPDATE SET spotlighted_user = $3, updated_at = now(), deleted_at = NULL`
	_, err := db.Exec(ctx, query, &dto.LessonRoomStateID, &dto.LessonID, &dto.SpotlightedUser)
	return err
}

func (l *LessonRoomStateRepo) UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.UnSpotlight")
	defer span.End()
	query := "UPDATE lesson_room_states SET spotlighted_user = NULL, updated_at = now() WHERE lesson_id = $1 AND spotlighted_user is not null"
	_, err := db.Exec(ctx, query, &lessonID)
	return err
}

func (l *LessonRoomStateRepo) UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, whiteboardZoomState *domain.WhiteboardZoomState) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.UpsertWhiteboardZoomState")
	defer span.End()
	w, err := json.Marshal(whiteboardZoomState)
	if err != nil {
		return fmt.Errorf("error when json.Marshal %v", whiteboardZoomState)
	}

	dto := &LessonRoomState{}
	if err := multierr.Combine(
		dto.LessonID.Set(lessonID),
		dto.WhiteboardZoomState.Set(database.JSONB(w)),
	); err != nil {
		return err
	}
	dto.PreInsert()
	query := `
			INSERT INTO lesson_room_states (lesson_room_state_id, lesson_id, whiteboard_zoom_state) VALUES ($1, $2, $3)
			ON CONFLICT ON CONSTRAINT unique__lesson_id 
				DO UPDATE SET whiteboard_zoom_state = $3, updated_at = now(), deleted_at = NULL`
	_, err = db.Exec(ctx, query, &dto.LessonRoomStateID, &dto.LessonID, &dto.WhiteboardZoomState)
	return err
}

func (l *LessonRoomStateRepo) GetLessonRoomStateByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*domain.LessonRoomState, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.GetLessonRoomStateByLessonID")
	defer span.End()
	dto := &LessonRoomState{}
	fields, values := dto.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = $1 and deleted_at is null", strings.Join(fields, ", "), dto.TableName())
	err := db.QueryRow(ctx, query, &lessonID).Scan(values...)
	if err == pgx.ErrNoRows {
		return &domain.LessonRoomState{}, domain.ErrNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "Scan: %w")
	}
	return dto.ToLessonRoomState()
}

func (l *LessonRoomStateRepo) UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, currentMaterial pgtype.JSONB) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRoomState.UpsertCurrentMaterialState")
	defer span.End()

	dto := &LessonRoomState{}
	dto.PreInsert()

	query := `INSERT INTO lesson_room_states (lesson_room_state_id, lesson_id, current_material) VALUES ($1, $2, $3)
			ON CONFLICT ON CONSTRAINT unique__lesson_id 
				DO UPDATE SET current_material = $3, updated_at = now(), deleted_at = NULL`
	_, err := db.Exec(ctx, query, &dto.LessonRoomStateID, &lessonID, &currentMaterial)

	return err
}
