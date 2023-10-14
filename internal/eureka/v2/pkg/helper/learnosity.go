package helper

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

func NewLearnositySecurity(ctx context.Context, config configurations.LearnosityConfig, domain string, timestamp time.Time) learnosity.Security {
	return learnosity.Security{
		ConsumerKey:    config.ConsumerKey,
		Domain:         domain,
		Timestamp:      learnosity.FormatUTCTime(timestamp),
		UserID:         interceptors.UserIDFromContext(ctx),
		ConsumerSecret: config.ConsumerSecret,
	}
}
