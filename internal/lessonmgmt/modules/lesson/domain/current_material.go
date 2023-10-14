package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type (
	PlayerState string
)

const (
	PlayerStatePause   PlayerState = "PLAYER_STATE_PAUSE"
	PlayerStatePlaying PlayerState = "PLAYER_STATE_PLAYING"
	PlayerStateEnded   PlayerState = "PLAYER_STATE_ENDED"
)

type CurrentMaterialBuilder struct {
	currentMaterial *CurrentMaterial
}

func NewCurrentMaterial() *CurrentMaterialBuilder {
	return &CurrentMaterialBuilder{
		currentMaterial: &CurrentMaterial{},
	}
}

func (c *CurrentMaterialBuilder) Build(ctx context.Context, db database.Ext) (*CurrentMaterial, error) {
	if err := c.currentMaterial.IsValid(ctx, db); err != nil {
		return nil, fmt.Errorf("invalid current material : %w", err)
	}
	return c.currentMaterial, nil
}

func (c *CurrentMaterialBuilder) WithLessonID(lessonID string) *CurrentMaterialBuilder {
	c.currentMaterial.LessonID = lessonID
	return c
}

func (c *CurrentMaterialBuilder) WithMediaID(mediaID string) *CurrentMaterialBuilder {
	c.currentMaterial.MediaID = &mediaID
	return c
}

func (c *CurrentMaterialBuilder) WithUpdatedAt(updatedAt time.Time) *CurrentMaterialBuilder {
	c.currentMaterial.UpdatedAt = updatedAt
	return c
}

func (c *CurrentMaterialBuilder) WithVideoCurrentTime(currentTime Duration) *CurrentMaterialBuilder {
	if c.currentMaterial.VideoState == nil {
		c.currentMaterial.VideoState = &VideoState{}
	}
	c.currentMaterial.VideoState.CurrentTime = currentTime
	return c
}

func (c *CurrentMaterialBuilder) WithVideoPlayerState(playerState PlayerState) *CurrentMaterialBuilder {
	if c.currentMaterial.VideoState == nil {
		c.currentMaterial.VideoState = &VideoState{}
	}
	c.currentMaterial.VideoState.PlayerState = playerState
	return c
}

func (c *CurrentMaterialBuilder) WithLessonRepo(repo LessonRepo) *CurrentMaterialBuilder {
	c.currentMaterial.LessonRepo = repo
	return c
}

func (c *CurrentMaterialBuilder) WithMediaModulePort(port MediaModulePort) *CurrentMaterialBuilder {
	c.currentMaterial.MediaModulePort = port
	return c
}

type CurrentMaterial struct {
	LessonID   string
	MediaID    *string     `json:"media_id,omitempty"`
	UpdatedAt  time.Time   `json:"updated_at"`
	VideoState *VideoState `json:"video_state,omitempty"`

	// repos
	LessonRepo      LessonRepo
	MediaModulePort MediaModulePort
}

func (c *CurrentMaterial) IsValid(ctx context.Context, db database.Ext) error {
	if len(c.LessonID) == 0 {
		return fmt.Errorf("lesson id could not be empty")
	}

	if c.MediaID != nil {
		if len(*c.MediaID) == 0 {
			return fmt.Errorf("invalid media id")
		}
	} else {
		if c.VideoState != nil {
			return fmt.Errorf("media is empty so could not has video state")
		}
	}

	if c.VideoState != nil {
		if err := c.VideoState.IsValid(); err != nil {
			return fmt.Errorf("invalid video_state: %v", err)
		}
	}

	// check data in db
	lesson, err := c.LessonRepo.GetLessonByID(ctx, db, c.LessonID)
	if err != nil {
		return fmt.Errorf("could not get lesson id %s: lessonRepo.GetLessonByID: %w", c.LessonID, err)
	}
	if c.MediaID != nil {
		// check media belong to lesson or not
		if lesson.Material == nil {
			return fmt.Errorf("media %s not belong to lesson", *c.MediaID)
		}
		isValid := false
		for _, mediaID := range lesson.Material.MediaIDs {
			if mediaID == *c.MediaID {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("media %s not belong to lesson %s", *c.MediaID, c.LessonID)
		}

		// check media data
		medias, err := c.MediaModulePort.RetrieveMediasByIDs(ctx, []string{*c.MediaID})
		if err != nil {
			return fmt.Errorf("MediaModulePort.RetrieveMediasByIDs(%s): %w", *c.MediaID, err)
		}
		if len(medias) == 0 {
			return fmt.Errorf("could not get media id %s: MediaModulePort.RetrieveMediasByIDs", *c.MediaID)
		}

		if medias[0].Type == media_domain.MediaTypeVideo && c.VideoState == nil {
			return fmt.Errorf("video state of media %s (video) could not be empty", *c.MediaID)
		}
		if medias[0].Type != media_domain.MediaTypeVideo && c.VideoState != nil {
			return fmt.Errorf("media id %s is not the video so could not has video state", *c.MediaID)
		}
	}

	return nil
}

func (c *CurrentMaterial) PreInsert() {
	c.UpdatedAt = time.Now()
}

type VideoState struct {
	CurrentTime Duration    `json:"current_time"`
	PlayerState PlayerState `json:"player_state"`
}

func (v *VideoState) IsValid() error {
	if len(v.PlayerState) == 0 {
		return fmt.Errorf("invalid player_state %s", v.PlayerState)
	}

	if v.PlayerState != PlayerStateEnded {
		if v.CurrentTime < 0 {
			return fmt.Errorf("invalid current_time %v", v.CurrentTime)
		}
	}

	return nil
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return fmt.Errorf("invalid duration")
	}
}

func (d *Duration) Duration() time.Duration {
	return time.Duration(*d)
}
