package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	timesheet_nats "github.com/manabie-com/backend/internal/timesheet/controller/nats"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"go.uber.org/zap"
)

func RegisterLessonEventHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	unleashClient unleashclient.ClientInstance,
	env string,
	lessonNatsService interface {
		HandleEventLessonUpdate(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error
		HandleEventLessonCreate(ctx context.Context, msg *bpb.EvtLesson_CreateLessons) error
		HandleEventLessonDelete(ctx context.Context, msg *bpb.EvtLesson_DeletedLessons) error
	},
	mastermgmtConfigurationService interface {
		CheckPartnerTimesheetServiceIsOnWithoutToken(ctx context.Context) (bool, error)
	},
) error {
	s := &timesheet_nats.LessonEventSubscription{
		Logger:                         logger,
		LessonNatsService:              lessonNatsService,
		JSM:                            jsm,
		UnleashClient:                  unleashClient,
		Env:                            env,
		MastermgmtConfigurationService: mastermgmtConfigurationService,
	}

	return s.Subscribe()
}

func RegisterTimesheetActionLogEventHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	unleashClient unleashclient.ClientInstance,
	env string,
	timesheetActionLogNatsService interface {
		Create(ctx context.Context, req *tpb.TimesheetActionLogRequest) error
	},
) error {
	s := &timesheet_nats.TimesheetActionLogEventSubscription{
		Logger:           logger,
		ActionLogService: timesheetActionLogNatsService,
		JSM:              jsm,
		UnleashClient:    unleashClient,
		Env:              env,
	}

	return s.Subscribe()
}

func RegisterTimesheetAutoCreateFlagEventHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	autoCreateFlagService interface {
		UpdateLessonHoursFlag(ctx context.Context, req *dto.AutoCreateTimesheetFlag) error
	},
) error {
	s := &timesheet_nats.TimesheetAutoCreateFlagEventSubscription{
		Logger:                logger,
		AutoCreateFlagService: autoCreateFlagService,
		JSM:                   jsm,
	}

	return s.Subscribe()
}
