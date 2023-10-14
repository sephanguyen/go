package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LiveRoomStateRepo struct{}

func (l *LiveRoomStateRepo) GetLiveRoomStateByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (*domain.LiveRoomState, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomStateRepo.GetLiveRoomStateByChannelID")
	defer span.End()

	dto := &LiveRoomState{}
	fields, values := dto.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE channel_id = $1 
		AND deleted_at IS NULL `,
		strings.Join(fields, ", "),
		dto.TableName(),
	)
	err := db.QueryRow(ctx, query, &channelID).Scan(values...)
	if err == pgx.ErrNoRows {
		return &domain.LiveRoomState{}, domain.ErrChannelNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "Scan: %w")
	}

	return dto.ToLiveRoomStateDomain()
}

func (l *LiveRoomStateRepo) UpsertLiveRoomState(ctx context.Context, db database.QueryExecer, channelID string, value interface{}, fieldName string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomStateRepo.UpsertLiveRoomState with field "+fieldName)
	defer span.End()

	dto := &LiveRoomState{}
	if err := dto.ChannelID.Set(channelID); err != nil {
		return err
	}
	dto.PreInsert()

	query := fmt.Sprintf(`INSERT INTO %s (live_room_state_id, channel_id, %s) VALUES ($1, $2, $3)
			ON CONFLICT ON CONSTRAINT unique__channel_id 
			DO UPDATE SET %s = $3, updated_at = now(), deleted_at = NULL`,
		dto.TableName(),
		fieldName,
		fieldName,
	)
	_, err := db.Exec(ctx, query, &dto.LiveRoomStateID, &dto.ChannelID, &value)

	return err
}

func (l *LiveRoomStateRepo) UpsertLiveRoomCurrentPollingState(ctx context.Context, db database.QueryExecer, channelID string, polling *vc_domain.CurrentPolling) error {
	dto := &LiveRoomState{}
	if err := dto.CurrentPolling.Set(database.JSONB(polling)); err != nil {
		return err
	}

	return l.UpsertLiveRoomState(ctx, db, channelID, dto.CurrentPolling, "current_polling")
}

func (l *LiveRoomStateRepo) UpsertLiveRoomSpotlightState(ctx context.Context, db database.QueryExecer, channelID, spotlightedUser string) error {
	dto := &LiveRoomState{}
	if err := dto.SpotlightedUser.Set(spotlightedUser); err != nil {
		return err
	}

	return l.UpsertLiveRoomState(ctx, db, channelID, dto.SpotlightedUser, "spotlighted_user")
}

func (l *LiveRoomStateRepo) UpsertLiveRoomWhiteboardZoomState(ctx context.Context, db database.QueryExecer, channelID string, wbZoomState *vc_domain.WhiteboardZoomState) error {
	dto := &LiveRoomState{}
	if err := dto.WhiteboardZoomState.Set(database.JSONB(wbZoomState)); err != nil {
		return err
	}

	return l.UpsertLiveRoomState(ctx, db, channelID, dto.WhiteboardZoomState, "whiteboard_zoom_state")
}

func (l *LiveRoomStateRepo) UpsertLiveRoomCurrentMaterialState(ctx context.Context, db database.QueryExecer, channelID string, currentMaterial *vc_domain.CurrentMaterial) error {
	dto := &LiveRoomState{}
	if err := dto.CurrentMaterial.Set(database.JSONB(currentMaterial)); err != nil {
		return err
	}

	return l.UpsertLiveRoomState(ctx, db, channelID, dto.CurrentMaterial, "current_material")
}

func (l *LiveRoomStateRepo) UpsertLiveRoomSessionTime(ctx context.Context, db database.QueryExecer, channelID string, sessionTime time.Time) error {
	dto := &LiveRoomState{}
	if err := dto.SessionTime.Set(sessionTime); err != nil {
		return err
	}

	return l.UpsertLiveRoomState(ctx, db, channelID, dto.SessionTime, "session_time")
}

func (l *LiveRoomStateRepo) UnSpotlight(ctx context.Context, db database.QueryExecer, channelID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomStateRepo.UnSpotlight")
	defer span.End()

	query := `UPDATE live_room_state SET spotlighted_user = NULL, updated_at = now() 
			  WHERE channel_id = $1 AND spotlighted_user IS NOT NULL`
	_, err := db.Exec(ctx, query, &channelID)

	return err
}

func (l *LiveRoomStateRepo) UpsertRecordingState(ctx context.Context, db database.QueryExecer, channelID string, recording *vc_domain.CompositeRecordingState) error {
	dto := &LiveRoomState{}
	if err := dto.Recording.Set(database.JSONB(recording)); err != nil {
		return err
	}

	return l.UpsertLiveRoomState(ctx, db, channelID, dto.Recording, "recording")
}

func (l *LiveRoomStateRepo) GetStreamingLearners(ctx context.Context, db database.QueryExecer, channelID string, lockForUpdate bool) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomStateRepo.GetStreamingLearners")
	defer span.End()

	dto := &LiveRoomState{}
	query := fmt.Sprintf(`SELECT streaming_learners FROM %s 
			  WHERE channel_id = $1 
			  AND deleted_at IS NULL `,
		dto.TableName(),
	)
	if lockForUpdate {
		query += ` FOR UPDATE `
	}

	var streamingLearners pgtype.TextArray
	err := db.QueryRow(ctx, query, &channelID).Scan(&streamingLearners)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrChannelNotFound
	} else if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	ids := database.FromTextArray(streamingLearners)
	return ids, nil
}

func (l *LiveRoomStateRepo) IncreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, channelID, learnerID string, maximumLearnerStreamings int) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomStateRepo.IncreaseNumberOfStreaming")
	defer span.End()

	dto := &LiveRoomState{}
	if err := dto.ChannelID.Set(channelID); err != nil {
		return err
	}
	dto.PreInsert()
	learnerIDArray := []string{learnerID}

	query := `INSERT INTO live_room_state (live_room_state_id, channel_id, stream_learner_counter, streaming_learners) VALUES ($1, $2, 1, $3)
			ON CONFLICT ON CONSTRAINT unique__channel_id 
			DO UPDATE SET stream_learner_counter = live_room_state.stream_learner_counter+1, 
				streaming_learners = array_append(live_room_state.streaming_learners, $4), 
				updated_at = now(), 
				deleted_at = NULL
			WHERE live_room_state.channel_id = $2
			AND live_room_state.stream_learner_counter < $5 
			AND NOT($4 = ANY(live_room_state.streaming_learners))`

	cmdTag, err := db.Exec(ctx, query, &dto.LiveRoomStateID, &dto.ChannelID, &learnerIDArray, &learnerID, &maximumLearnerStreamings)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return domain.ErrNoChannelUpdated
	}

	return nil
}

func (l *LiveRoomStateRepo) DecreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, channelID, learnerID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomStateRepo.DecreaseNumberOfStreaming")
	defer span.End()

	query := `UPDATE live_room_state 
			  SET stream_learner_counter = stream_learner_counter-1, streaming_learners = array_remove(streaming_learners, $2) 
			  WHERE channel_id = $1 
			  AND stream_learner_counter > 0 
			  AND $2 = ANY(streaming_learners)`

	cmdTag, err := db.Exec(ctx, query, &channelID, &learnerID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return domain.ErrNoChannelUpdated
	}

	return nil
}
