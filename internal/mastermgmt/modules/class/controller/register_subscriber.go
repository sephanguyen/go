package controller

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/consumers"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"

	"go.uber.org/zap"
)

type RegisterSubscriber struct {
	JSM             nats.JetStreamManagement
	Logger          *zap.Logger
	DB              database.Ext
	ClassMemberRepo infrastructure.ClassMemberRepo
}

func (r *RegisterSubscriber) Subscribe() error {
	subscriberHandler := &consumers.StudentPackageHandler{
		Logger:          r.Logger,
		DB:              r.DB,
		ClassMemberRepo: r.ClassMemberRepo,
		JSM:             r.JSM,
	}
	s := &StudentPackageSubscriber{
		JSM:               r.JSM,
		Logger:            r.Logger,
		SubscriberHandler: subscriberHandler,
	}

	return s.Subscribe()
}
