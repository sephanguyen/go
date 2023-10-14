package idutil

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/oklog/ulid/v2"
)

// ULID generates a string representation of a ULID.
// rand.New is not threadsafe, so we create a pool of rand to speed up the id generation.
var randPool = sync.Pool{
	New: func() interface{} {
		// Note that this implementation to create the entropy:
		//	t := time.Now()
		//	return ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
		// is not correct if there have *extremely* multiple concurrent calls.
		return ulid.Monotonic(rand.Reader, 0)
	},
}

func ULID(t time.Time) string {
	entropy := randPool.Get().(*ulid.MonotonicEntropy)
	defer randPool.Put(entropy)

	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// ULIDNow returns a new ULID.
func ULIDNow() string {
	return ULID(timeutil.Now())
}
