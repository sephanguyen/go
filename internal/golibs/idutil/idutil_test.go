package idutil

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs"
)

func TestULID(t *testing.T) {
	t.Parallel()
	const times = 10000
	ch := make(chan string, times)

	for i := 1; i <= times; i++ {
		go func() {
			ch <- ULIDNow()
		}()
	}

	ids := make([]string, 0, times)
	for id := range ch {
		ids = append(ids, id)
		if len(ids) == times {
			break
		}
	}

	if u := golibs.Uniq(ids); len(u) != len(ids) {
		t.Errorf("duplicated ids: orig: %d, uniq: %d", len(ids), len(u))
	}
}
