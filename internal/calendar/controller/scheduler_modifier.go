package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/calendar/application"
	"github.com/manabie-com/backend/internal/calendar/application/command"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SchedulerModifierService struct {
	db                 database.Ext
	createSchedulerCmd application.CreateSchedulerPort
	updateSchedulerCmd application.UpdateSchedulerPort
}

func NewSchedulerModifierService(schedulerRepo infrastructure.SchedulerPort, db database.Ext) *SchedulerModifierService {
	return &SchedulerModifierService{
		db: db,
		createSchedulerCmd: &command.CreateSchedulerCommand{
			SchedulerRepo: schedulerRepo,
		},
		updateSchedulerCmd: &command.UpdateSchedulerCommand{
			SchedulerRepo: schedulerRepo,
		},
	}
}

func (s *SchedulerModifierService) CreateScheduler(ctx context.Context, in *cpb.CreateSchedulerRequest) (*cpb.CreateSchedulerResponse, error) {
	req := &command.CreateSchedulerRequest{
		StartDate: in.StartDate.AsTime(),
		EndDate:   in.EndDate.AsTime(),
		Frequency: in.Frequency.String(),
	}
	response, err := s.createSchedulerCmd.CreateScheduler(ctx, s.db, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.createSchedulerCmd.CreateScheduler: %w", err).Error())
	}
	return &cpb.CreateSchedulerResponse{
		SchedulerId: response.SchedulerID,
	}, nil
}

func (s *SchedulerModifierService) UpdateScheduler(ctx context.Context, in *cpb.UpdateSchedulerRequest) (*cpb.UpdateSchedulerResponse, error) {
	req := &command.UpdateSchedulerRequest{
		SchedulerID: in.SchedulerId,
		EndDate:     in.EndDate.AsTime(),
	}
	err := s.updateSchedulerCmd.UpdateScheduler(ctx, s.db, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.updateSchedulerCmd.UpdateScheduler: %w", err).Error())
	}
	return &cpb.UpdateSchedulerResponse{}, nil
}

func (s *SchedulerModifierService) CreateManySchedulers(ctx context.Context, in *cpb.CreateManySchedulersRequest) (*cpb.CreateManySchedulersResponse, error) {
	response, err := s.createSchedulerCmd.CreateManySchedulers(ctx, s.db, in)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("s.createSchedulerCmd.CreateManySchedulers: %w", err).Error())
	}
	return response, nil
}
