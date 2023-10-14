package validation

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

func ValidateKafkaContext(ctx context.Context) error {
	// check if UserID exists
	if userID := interceptors.UserIDFromContext(ctx); userID == "" {
		claim := interceptors.JWTClaimsFromContext(ctx)
		if claim.Manabie.UserID == "" {
			return fmt.Errorf("missing UserID in context")
		}
	}

	// check if ResourcePath exists
	if _, err := interceptors.ResourcePathFromContext(ctx); err != nil {
		return err
	}
	return nil
}
