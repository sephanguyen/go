package subscription

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"go.uber.org/zap"
)

type OrganizationModifier struct {
	DB               database.Ext
	Logger           *zap.Logger
	JSM              nats.JetStreamManagement
	LocationTypeRepo domain.LocationTypeRepo
	LocationRepo     domain.LocationRepo
}

// publisher has been locked
func (o *OrganizationModifier) HandleOrganizationCreated(ctx context.Context, data []byte) (bool, error) {
	return true, nil
}
