package domain

import (
	"fmt"
	"time"
)

type Topic struct {
	ID            string
	ChapterID     string
	Name          string
	Status        string
	DisplayOrder  int
	IconURL       string
	CopiedTopicID string

	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt *time.Time

	LearningMaterials []LearningMaterial
}

func (t *Topic) RemoveUnpublishedMaterials() {
	publishedMaterials := make([]LearningMaterial, 0, len(t.LearningMaterials))

	for i := 0; i < len(t.LearningMaterials); i++ {
		if t.LearningMaterials[i].Published {
			publishedMaterials = append(publishedMaterials, t.LearningMaterials[i])
		}
	}

	t.LearningMaterials = publishedMaterials
}

func (t *Topic) String() string {
	return fmt.Sprintf("{ID: %s; Name: %s; ChapterID: %s; Status: %s; IconUrl: %s; Order: %d}",
		t.ID, t.Name, t.ChapterID, t.Status, t.IconURL, t.DisplayOrder)
}
