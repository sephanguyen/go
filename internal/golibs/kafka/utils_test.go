package kafka

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetTopicNameWithPrefix(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		topic := GetTopicNameWithPrefix("topic-A", "local.manabie.")
		assert.Equal(t, topic, "local.manabie.topic-A")
	})
}
