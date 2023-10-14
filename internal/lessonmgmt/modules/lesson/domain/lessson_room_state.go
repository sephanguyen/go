package domain

import (
	"time"

	virDomain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type LessonRoomState struct {
	ID                  string
	LessonID            string
	CurrentMaterial     *virDomain.CurrentMaterial
	SpotlightedUser     string
	WhiteboardZoomState *virDomain.WhiteboardZoomState
	Recording           *virDomain.CompositeRecordingState
	CurrentPolling      *virDomain.CurrentPolling
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

type WhiteboardZoomState struct {
	PdfScaleRatio float64 `json:"pdf_scale_ratio"`
	CenterX       float64 `json:"center_x"`
	CenterY       float64 `json:"center_y"`
	PdfWidth      float64 `json:"pdf_width"`
	PdfHeight     float64 `json:"pdf_height"`
}

func (w *WhiteboardZoomState) SetDefault() *WhiteboardZoomState {
	w.PdfScaleRatio = 100.0
	w.PdfWidth = 1920.0
	w.PdfHeight = 1080.0
	w.CenterX = 0.0
	w.CenterY = 0.0
	return w
}
