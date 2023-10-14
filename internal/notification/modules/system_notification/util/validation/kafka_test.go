package validation

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/stretchr/testify/assert"
)

func Test_ValidateKafkaContext(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		userID := "userid"
		resourcePath := "123"
		claim := &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				UserID:       userID,
				ResourcePath: resourcePath,
			},
		}

		ctx := interceptors.ContextWithJWTClaims(context.Background(), claim)

		err := ValidateKafkaContext(ctx)

		assert.Equal(t, nil, err)
	})
}
