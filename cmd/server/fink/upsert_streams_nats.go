package fink

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/fink/streams"
	"github.com/manabie-com/backend/internal/fink/configurations"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"

	nats_org "github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func RunUpsertStreams(ctx context.Context, c *configurations.Config) {
	zapLogger := logger.NewZapLogger("debug", c.Common.Environment == "local")

	// init nats-jetstream
	var jsm nats.JetStreamManagement
	jsm, err := nats.NewJetStreamManagement(c.NatsJS.Address, c.NatsJS.User, c.NatsJS.Password, c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)
	if err != nil {
		zapLogger.Fatal("failed to connect to nats jetstream", zap.Error(err))
	}
	jsm.ConnectToJS()
	defer jsm.Close()

	err = upsertStreamsNats(jsm)
	if err != nil {
		zapLogger.Fatal("error when upsert stream", zap.Error(err))
	}

	zapLogger.Info("Upsert all streams have succeed")
}

func upsertStreamsNats(jsm nats.JetStreamManagement) error {
	arrStreamConfig := []*nats_org.StreamConfig{}

	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigBob()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigEureka()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigHephaestus()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigLessonmgmt()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigMastermgmt()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigNotificationmgmt()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigPayment()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigTimesheet()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigTom()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigUsermgmt()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigYasuo()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigZeus()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigFatima()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigDraft()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigVirtualClassroom()...)
	arrStreamConfig = append(arrStreamConfig, streams.GetStreamConfigDiscount()...)

	var err error
	for i := range arrStreamConfig {
		err = jsm.UpsertStream(arrStreamConfig[i])
		if err != nil {
			return fmt.Errorf("failed to upsertStream: %s. Detail: %v", arrStreamConfig[i].Name, err)
		}
	}
	return nil
}
