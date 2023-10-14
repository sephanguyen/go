package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
)

type LearningMaterial struct {
	ID           string
	TopicID      string
	DisplayOrder int
	Name         string
	Type         constants.LearningMaterialType
	VendorType   string
	Published    bool

	// learning objective
	// a child table in postgres
	Video       string
	StudyGuide  string
	VideoScript string

	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt *time.Time
}

func (t *LearningMaterial) String() string {
	return fmt.Sprintf("{ID: %s; Name: %s; Order: %d, TopicID: %s; Type: %s, Published: %v}",
		t.ID, t.Name, t.DisplayOrder, t.TopicID, t.Type, t.Published)
}
