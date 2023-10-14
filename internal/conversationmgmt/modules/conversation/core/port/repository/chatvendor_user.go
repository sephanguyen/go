package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ChatVendorUserRepo interface {
	GetByUserIDs(ctx context.Context, db database.Ext, userIDs []string) ([]*domain.ChatVendorUser, error)
	GetByUserID(ctx context.Context, db database.Ext, userID string) (*domain.ChatVendorUser, error)
	GetByVendorUserIDs(ctx context.Context, db database.Ext, vendorUserIDs []string) ([]*domain.ChatVendorUser, error)
	GetByVendorUserID(ctx context.Context, db database.Ext, vendorUserID string) (*domain.ChatVendorUser, error)
}
