package domain

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type ImportLog struct {
	ID         string
	UserID     string
	ImportType string
	Payload    string
	CreatedAt  time.Time
	DeletedAt  *time.Time
}

func NewImportLog(userID string, importType string, payload []byte) *ImportLog {
	return &ImportLog{
		ID:         idutil.ULIDNow(),
		UserID:     userID,
		ImportType: importType,
		Payload:    string(payload),
		CreatedAt:  time.Now(),
	}
}
