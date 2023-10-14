package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type ZoomAction string

const (
	ActionUpsert ZoomAction = "Upsert"
	ActionDelete ZoomAction = "Delete"
)
const ZoomIDLabel = "zoom_id"
const ZoomUsernameLabel = "zoom_username"
const ZoomAccountActionLabel = "Action"

type ZoomAccount struct {
	ID        string
	Email     string
	UserName  string
	Action    ZoomAction
	DeletedAt *time.Time
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type ZoomAccounts []*ZoomAccount

type ZoomAccountBuilder struct {
	zoomAccount *ZoomAccount
}

func NewZoomAccountBuilder() *ZoomAccountBuilder {
	return &ZoomAccountBuilder{
		zoomAccount: &ZoomAccount{},
	}
}

func (b *ZoomAccountBuilder) WithID(id string) *ZoomAccountBuilder {
	b.zoomAccount.ID = id
	if id == "" {
		b.zoomAccount.ID = idutil.ULIDNow()
	}
	return b
}

func (b *ZoomAccountBuilder) WithEmail(email string) *ZoomAccountBuilder {
	b.zoomAccount.Email = email
	return b
}

func (b *ZoomAccountBuilder) WithUsername(username string) *ZoomAccountBuilder {
	b.zoomAccount.UserName = username
	return b
}

func (b *ZoomAccountBuilder) WithAction(action string) *ZoomAccountBuilder {
	b.zoomAccount.Action = ZoomAction(action)
	now := time.Now()

	if ZoomAction(action) == ActionDelete {
		b.zoomAccount.DeletedAt = &now
	}
	b.zoomAccount.CreatedAt = &now
	return b
}

func (b *ZoomAccountBuilder) Build() (*ZoomAccount, error) {
	if err := b.zoomAccount.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid zoom account detail: %w", err)
	}
	now := time.Now()
	b.zoomAccount.UpdatedAt = &now
	return b.zoomAccount, nil
}

func (z *ZoomAccount) IsValid() error {
	if z.Email == "" {
		return fmt.Errorf("email could not be empty")
	}
	if z.Action == "" {
		return fmt.Errorf("action could not be empty")
	}
	return nil
}
