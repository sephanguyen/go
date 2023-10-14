package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TimesheetAutoCreateFlagEventSubscription struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement

	AutoCreateFlagService interface {
		UpdateLessonHoursFlag(ctx context.Context, req *dto.AutoCreateTimesheetFlag) error
	}
}

func (t *TimesheetAutoCreateFlagEventSubscription) Subscribe() error {
	autoCreateLogSubOptions := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamTimesheetAutoCreateFlag, constants.DurableTimesheetAutoCreateFlag),
			nats.DeliverSubject(constants.DeliverTimesheetAutoCreateFlag),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := t.JSM.QueueSubscribe(constants.SubjectTimesheetAutoCreateFlag, constants.QueueTimesheetAutoCreateFlag, autoCreateLogSubOptions, t.HandleNatsMessage)
	if err != nil {
		t.Logger.Error("err Subscribe JSM.QueueSubscribe failed", zap.Error(err))
		return fmt.Errorf("JSM.QueueSubscribe SubjectTimesheetAutoCreateFlag: %w", err)
	}

	return nil
}

func (t *TimesheetAutoCreateFlagEventSubscription) HandleNatsMessage(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &tpb.NatsUpdateAutoCreateTimesheetFlagRequest{}
	err := proto.Unmarshal(data, req)

	if err != nil {
		t.Logger.Error(err.Error())
		return true, err
	}

	autoFlag := dto.NewAutoCreateTimeSheetFlagFromNATSUpdateRequest(req)

	err = t.AutoCreateFlagService.UpdateLessonHoursFlag(ctx, autoFlag)
	if err != nil {
		t.Logger.Error("err AutoCreateFlagService.UpsertFlag failed", zap.Error(err))
		return true, err
	}

	return false, nil
}
