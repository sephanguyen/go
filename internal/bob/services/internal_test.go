package services

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/caching"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"
)

func TestNewInternalServiceCacher(t *testing.T) {
	//init vars
	var cacher caching.LocalCacher
	cacher = nil
	r := &mock_repositories.MockStudentOrderRepo{}
	svc := &InternalService{
		StudentOrderRepo: r,
	}
	iscHappy := InternalServiceCacher{
		group:          singleflight.Group{},
		Cacher:         cacher,
		InternalServer: svc,
	}
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		r := &mock_repositories.MockStudentOrderRepo{}
		svc := InternalService{
			StudentOrderRepo: r,
		}
		resp := NewInternalServiceCacher(cacher, &svc)
		assert.Equal(t, resp, &iscHappy)
	})
	t.Run("service without repo", func(t *testing.T) {
		t.Parallel()
		svc := InternalService{}
		resp := NewInternalServiceCacher(cacher, &svc)
		assert.NotEqual(t, resp, &iscHappy)
	})
}
