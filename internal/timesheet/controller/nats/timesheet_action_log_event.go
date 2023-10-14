package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	timesheet_constants "github.com/manabie-com/backend/internal/timesheet/domain/constant"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TimesheetActionLogEventSubscription struct {
	Logger        *zap.Logger
	JSM           nats.JetStreamManagement
	UnleashClient unleashclient.ClientInstance
	Env           string

	ActionLogService interface {
		Create(ctx context.Context, req *tpb.TimesheetActionLogRequest) error
	}
}

func (t *TimesheetActionLogEventSubscription) Subscribe() error {
	t.UnleashClient.WaitForUnleashReady()
	isFeatureEnabled, err := t.UnleashClient.IsFeatureEnabled(timesheet_constants.FeatureToggleActionLog, t.Env)
	if err != nil {
		t.Logger.Error("err Subscribe UnleashClient.IsFeatureEnabled failed", zap.Error(err))
		return fmt.Errorf("%s unleashClient.IsFeatureEnabled: %w", timesheet_constants.FeatureToggleActionLog, err)
	}

	// if feature flag disabled, then do nothing
	if !isFeatureEnabled {
		return nil
	}

	actionLogSubOptions := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamTimesheetActionLog, constants.DurableTimesheetActionLog),
			nats.DeliverSubject(constants.DeliverTimesheetActionLogEvent),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err = t.JSM.QueueSubscribe(constants.SubjectTimesheetActionLog, constants.QueueTimesheetActionLog, actionLogSubOptions, t.HandleNatsMessage)
	if err != nil {
		t.Logger.Error("err Subscribe JSM.QueueSubscribe failed", zap.Error(err))
		return fmt.Errorf("JSM.QueueSubscribe SubjectTimesheetActionLog: %w", err)
	}

	return nil
}

func (t *TimesheetActionLogEventSubscription) HandleNatsMessage(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &tpb.TimesheetActionLogRequest{}
	err := proto.Unmarshal(data, req)

	if err != nil {
		t.Logger.Error(err.Error())
		return true, err
	}

	err = t.ActionLogService.Create(ctx, req)
	if err != nil {
		t.Logger.Error("err ActionLogService.Create failed", zap.Error(err))
		return true, err
	}

	return false, nil
}
