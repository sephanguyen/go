package repo

import (
	"encoding/json"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type LiveRoomState struct {
	LiveRoomStateID     pgtype.Text
	ChannelID           pgtype.Text
	CurrentMaterial     pgtype.JSONB
	SpotlightedUser     pgtype.Text
	WhiteboardZoomState pgtype.JSONB
	Recording           pgtype.JSONB
	CurrentPolling      pgtype.JSONB
	SessionTime         pgtype.Timestamptz
	CreatedAt           pgtype.Timestamptz
	UpdatedAt           pgtype.Timestamptz
	DeletedAt           pgtype.Timestamptz
}

func (l *LiveRoomState) FieldMap() ([]string, []interface{}) {
	return []string{
			"live_room_state_id",
			"channel_id",
			"current_material",
			"spotlighted_user",
			"whiteboard_zoom_state",
			"recording",
			"current_polling",
			"session_time",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&l.LiveRoomStateID,
			&l.ChannelID,
			&l.CurrentMaterial,
			&l.SpotlightedUser,
			&l.WhiteboardZoomState,
			&l.Recording,
			&l.CurrentPolling,
			&l.SessionTime,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		}
}

func (l *LiveRoomState) TableName() string {
	return "live_room_state"
}

func (l *LiveRoomState) PreInsert() {
	if l.LiveRoomStateID.Status != pgtype.Present {
		l.LiveRoomStateID = database.Text(idutil.ULIDNow())
	}
}

func (l *LiveRoomState) ToLiveRoomStateDomain() (*domain.LiveRoomState, error) {
	liveRoomState := &domain.LiveRoomState{
		LiveRoomStateID:     l.LiveRoomStateID.String,
		ChannelID:           l.ChannelID.String,
		CurrentMaterial:     nil,
		SpotlightedUser:     "",
		WhiteboardZoomState: new(vc_domain.WhiteboardZoomState).SetDefault(),
		Recording:           &vc_domain.CompositeRecordingState{},
		CurrentPolling:      nil,
		SessionTime:         nil,
		CreatedAt:           l.CreatedAt.Time,
		UpdatedAt:           l.UpdatedAt.Time,
		DeletedAt:           nil,
	}

	if l.CurrentMaterial.Status == pgtype.Present {
		var v vc_domain.CurrentMaterial
		if err := json.Unmarshal(l.CurrentMaterial.Bytes, &v); err != nil {
			return nil, err
		}
		liveRoomState.CurrentMaterial = &v
	}

	if l.SpotlightedUser.Status == pgtype.Present {
		liveRoomState.SpotlightedUser = l.SpotlightedUser.String
	}

	if l.WhiteboardZoomState.Status == pgtype.Present {
		var v vc_domain.WhiteboardZoomState
		if err := json.Unmarshal(l.WhiteboardZoomState.Bytes, &v); err != nil {
			return nil, err
		}
		liveRoomState.WhiteboardZoomState = &v
	}

	if l.Recording.Status == pgtype.Present {
		var v vc_domain.CompositeRecordingState
		if err := json.Unmarshal(l.Recording.Bytes, &v); err != nil {
			return nil, err
		}
		liveRoomState.Recording = &v
	}

	if l.CurrentPolling.Status == pgtype.Present {
		var v vc_domain.CurrentPolling
		if err := json.Unmarshal(l.CurrentPolling.Bytes, &v); err != nil {
			return nil, err
		}
		if !v.CreatedAt.IsZero() {
			liveRoomState.CurrentPolling = &v
		}
	}

	if l.DeletedAt.Status == pgtype.Present {
		liveRoomState.DeletedAt = &l.DeletedAt.Time
	}

	if l.SessionTime.Status == pgtype.Present {
		liveRoomState.SessionTime = &l.SessionTime.Time
	}

	return liveRoomState, nil
}
