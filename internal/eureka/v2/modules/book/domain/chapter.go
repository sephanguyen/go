package domain

import "time"

type Chapter struct {
	ID                       string
	BookID                   string
	Name                     string
	DisplayOrder             int
	CopiedFrom               string
	CurrentTopicDisplayOrder int

	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt *time.Time

	Topics []Topic
}

func (c *Chapter) RemoveUnpublishedTopics() {
	publishedTopics := make([]Topic, 0, len(c.Topics))

	for i := 0; i < len(c.Topics); i++ {
		topic := c.Topics[i]
		topic.RemoveUnpublishedMaterials()
		if len(topic.LearningMaterials) > 0 {
			publishedTopics = append(publishedTopics, topic)
		}
	}

	c.Topics = publishedTopics
}
