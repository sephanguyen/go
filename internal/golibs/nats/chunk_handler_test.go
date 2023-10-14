package nats

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkHandler(t *testing.T) {
	t.Parallel()
	t.Run("err handler", func(t *testing.T) {
		t.Parallel()
		err := ChunkHandler(10, 7, func(start, end int) error {
			return fmt.Errorf("err handler")
		})

		assert.EqualError(t, err, "err handler; err handler")
	})

	t.Run("sucess", func(t *testing.T) {
		t.Parallel()
		array := make([]string, 10)
		err := ChunkHandler(10, 7, func(start, end int) error {
			for i := range array[start:end] {
				array[start+i] = strconv.Itoa(start + i)
			}

			return nil
		})

		assert.Nil(t, err)
		for i, v := range array {
			assert.Equal(t, strconv.Itoa(i), v)
		}
	})
}
