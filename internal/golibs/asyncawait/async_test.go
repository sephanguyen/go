package async

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func DoneAsync() (int, error) {
	time.Sleep(1 * time.Second)
	return 1, nil
}

func TestFuture_Exec(t *testing.T) {
	t.Run("should success when call function", func(t *testing.T) {
		start := time.Now()

		future := Exec(func() (interface{}, error) {
			return DoneAsync()
		})
		future2 := Exec(func() (interface{}, error) {
			return DoneAsync()
		})
		val, _ := future.Await()
		val2, _ := future2.Await()
		dur := time.Since(start)
		assert.Equal(t, dur.Round(time.Second), 1*time.Second)
		assert.Equal(t, 1, val.(int))
		assert.Equal(t, 1, val2.(int))
	})
}
