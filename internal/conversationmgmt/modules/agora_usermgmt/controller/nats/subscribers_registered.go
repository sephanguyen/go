package nats

import (
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/application/consumers"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/golibs/chatvendor"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"go.uber.org/zap"
)

type SubscribersRegistered struct {
	db         database.Ext
	nats       nats.JetStreamManagement
	logger     *zap.Logger
	chatVendor chatvendor.ChatVendorClient
}

func NewSubscribersRegistered(
	db database.Ext,
	nats nats.JetStreamManagement,
	chatVendor chatvendor.ChatVendorClient,
	logger *zap.Logger,
) *SubscribersRegistered {
	return &SubscribersRegistered{
		db:         db,
		chatVendor: chatVendor,
		nats:       nats,
		logger:     logger,
	}
}

func (r *SubscribersRegistered) StartSubscribeForAllSubscribers() error {
	// Init user created handler
	userCreatedHandler := &consumers.UserCreatedHandler{
		DB:               r.db,
		ChatVendorClient: r.chatVendor,
		Logger:           *r.logger,

		AgoraUserRepo:     &repositories.AgoraUserRepo{},
		UserBasicInfoRepo: &repositories.UserBasicInfoRepo{},
	}
	userCreatedSubscriber := &UserCreatedSubscriber{
		nats:            r.nats,
		logger:          r.logger,
		ConsumerHandler: userCreatedHandler,
	}

	err := userCreatedSubscriber.StartSubscribe()
	if err != nil {
		return fmt.Errorf("error when userCreatedSubscriber.StartSubscribe(): [%v]", err)
	}

	// Init staff upserted handler
	staffCreatedHandler := &consumers.StaffUpsertedHandler{
		DB:               r.db,
		ChatVendorClient: r.chatVendor,
		Logger:           *r.logger,

		AgoraUserRepo:     &repositories.AgoraUserRepo{},
		UserBasicInfoRepo: &repositories.UserBasicInfoRepo{},
	}
	staffUpsertedSubscriber := &StaffUpsertedSubscriber{
		nats:            r.nats,
		logger:          r.logger,
		ConsumerHandler: staffCreatedHandler,
	}

	err = staffUpsertedSubscriber.StartSubscribe()
	if err != nil {
		return fmt.Errorf("error when staffUpsertedSubscriber.StartSubscribe(): [%v]", err)
	}

	return nil
}
