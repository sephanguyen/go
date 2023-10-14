package repo

import (
	"encoding/json"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type LessonRoomState struct {
	LessonRoomStateID   pgtype.Text
	LessonID            pgtype.Text
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

func (l *LessonRoomState) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_room_state_id",
			"lesson_id",
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
			&l.LessonRoomStateID,
			&l.LessonID,
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

func (l *LessonRoomState) TableName() string {
	return "lesson_room_states"
}

func (l *LessonRoomState) PreInsert() {
	if l.LessonRoomStateID.Status != pgtype.Present {
		l.LessonRoomStateID = database.Text(idutil.ULIDNow())
	}
}

func (l *LessonRoomState) ToLessonRoomState() (*domain.LessonRoomState, error) {
	lessonRoomState := &domain.LessonRoomState{
		ID:                  l.LessonRoomStateID.String,
		LessonID:            l.LessonID.String,
		CurrentMaterial:     nil,
		SpotlightedUser:     "",
		WhiteboardZoomState: new(domain.WhiteboardZoomState).SetDefault(),
		Recording:           &domain.CompositeRecordingState{},
		CurrentPolling:      nil,
		SessionTime:         nil,
		CreatedAt:           l.CreatedAt.Time,
		UpdatedAt:           l.UpdatedAt.Time,
		DeletedAt:           nil,
	}

	if l.CurrentMaterial.Status == pgtype.Present {
		var v domain.CurrentMaterial
		if err := json.Unmarshal(l.CurrentMaterial.Bytes, &v); err != nil {
			return nil, err
		}
		lessonRoomState.CurrentMaterial = &v
	}

	if l.SpotlightedUser.Status == pgtype.Present {
		lessonRoomState.SpotlightedUser = l.SpotlightedUser.String
	}

	if l.WhiteboardZoomState.Status == pgtype.Present {
		var v domain.WhiteboardZoomState
		if err := json.Unmarshal(l.WhiteboardZoomState.Bytes, &v); err != nil {
			return nil, err
		}
		lessonRoomState.WhiteboardZoomState = &v
	}

	if l.Recording.Status == pgtype.Present {
		var v domain.CompositeRecordingState
		if err := json.Unmarshal(l.Recording.Bytes, &v); err != nil {
			return nil, err
		}
		lessonRoomState.Recording = &v
	}

	if l.CurrentPolling.Status == pgtype.Present {
		var v domain.CurrentPolling
		if err := json.Unmarshal(l.CurrentPolling.Bytes, &v); err != nil {
			return nil, err
		}
		if !v.CreatedAt.IsZero() {
			lessonRoomState.CurrentPolling = &v
		}
	}

	if l.DeletedAt.Status == pgtype.Present {
		lessonRoomState.DeletedAt = &l.DeletedAt.Time
	}

	if l.SessionTime.Status == pgtype.Present {
		lessonRoomState.SessionTime = &l.SessionTime.Time
	}

	return lessonRoomState, nil
}
