package repo

import (
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	virDomain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

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

func (l *LessonRoomState) SetCurrentMaterial(material *domain.CurrentMaterial) error {
	src := pgtype.JSONB{}
	if err := src.Set(material); err != nil {
		return fmt.Errorf("could not marshal current material to jsonb: %w", err)
	}
	l.CurrentMaterial = src
	return nil
}

func (l *LessonRoomState) ToLessonRoomState() (*domain.LessonRoomState, error) {
	lessonRoomState := &domain.LessonRoomState{
		ID:                  l.LessonRoomStateID.String,
		LessonID:            l.LessonID.String,
		CurrentMaterial:     nil,
		CreatedAt:           l.CreatedAt.Time,
		UpdatedAt:           l.UpdatedAt.Time,
		WhiteboardZoomState: new(virDomain.WhiteboardZoomState).SetDefault(),
		Recording:           &virDomain.CompositeRecordingState{},
		CurrentPolling:      nil,
	}

	if l.DeletedAt.Status == pgtype.Present {
		lessonRoomState.DeletedAt = &l.DeletedAt.Time
	}
	if l.SpotlightedUser.Status == pgtype.Present {
		lessonRoomState.SpotlightedUser = l.SpotlightedUser.String
	}

	if l.WhiteboardZoomState.Status == pgtype.Present {
		var v virDomain.WhiteboardZoomState
		if err := json.Unmarshal(l.WhiteboardZoomState.Bytes, &v); err != nil {
			return nil, err
		}
		lessonRoomState.WhiteboardZoomState = &v
	}

	if l.Recording.Status == pgtype.Present {
		var v virDomain.CompositeRecordingState
		if err := json.Unmarshal(l.Recording.Bytes, &v); err != nil {
			return nil, err
		}
		lessonRoomState.Recording = &v
	}

	if l.CurrentPolling.Status == pgtype.Present {
		var v virDomain.CurrentPolling
		if err := json.Unmarshal(l.CurrentPolling.Bytes, &v); err != nil {
			return nil, err
		}
		if !v.CreatedAt.IsZero() {
			lessonRoomState.CurrentPolling = &v
		}
	}

	if l.CurrentMaterial.Status == pgtype.Present {
		var v virDomain.CurrentMaterial
		if err := json.Unmarshal(l.CurrentMaterial.Bytes, &v); err != nil {
			return nil, err
		}
		lessonRoomState.CurrentMaterial = &v
	}

	return lessonRoomState, nil
}
