package domain

import (
	"fmt"
	"time"
)

type (
	PlayerState string
)

const (
	PlayerStatePause   PlayerState = "PLAYER_STATE_PAUSE"
	PlayerStatePlaying PlayerState = "PLAYER_STATE_PLAYING"
	PlayerStateEnded   PlayerState = "PLAYER_STATE_ENDED"
)

type PresentMaterialState struct {
	VideoPresentMaterialState
	PDFPresentMaterialState
	MediaID string
}
type presentMaterialState interface {
	isPresentMaterialType()
	IsValid() error
}

var (
	_ presentMaterialState = new(VideoPresentMaterialState)
	_ presentMaterialState = new(PDFPresentMaterialState)
)

type VideoPresentMaterialState struct {
	Material   *VideoMaterial
	UpdatedAt  time.Time   `json:"updated_at"`
	VideoState *VideoState `json:"video_state,omitempty"`
}

func (v *VideoPresentMaterialState) isPresentMaterialType() {}

func (v *VideoPresentMaterialState) IsValid() error {
	if v.Material == nil {
		return fmt.Errorf("material cannot be empty")
	}

	if err := v.Material.isValid(); err != nil {
		return fmt.Errorf("invalid material: %w", err)
	}

	if v.VideoState == nil {
		return fmt.Errorf("video state cannot be empty")
	}

	if len(v.VideoState.PlayerState) == 0 {
		return fmt.Errorf("video's player state cannot be empty")
	}

	return nil
}

type VideoState struct {
	CurrentTime Duration    `json:"current_time"`
	PlayerState PlayerState `json:"player_state"`
}

type PDFPresentMaterialState struct {
	Material  *PDFMaterial
	UpdatedAt time.Time
}

func (p *PDFPresentMaterialState) isPresentMaterialType() {}

func (p *PDFPresentMaterialState) IsValid() error {
	if p.Material == nil {
		return fmt.Errorf("material cannot be empty")
	}

	if err := p.Material.isValid(); err != nil {
		return fmt.Errorf("invalid material: %w", err)
	}

	return nil
}

type AudioState struct {
	CurrentTime Duration    `json:"current_time"`
	PlayerState PlayerState `json:"player_state"`
}
