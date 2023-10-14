package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/calendar/domain/entities"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"go.uber.org/multierr"
)

type CreateSchedulerCommand struct {
	SchedulerRepo infrastructure.SchedulerPort
}

type CreateSchedulerRequest struct {
	StartDate time.Time
	EndDate   time.Time
	Frequency string
}

type CreateSchedulerResponse struct {
	SchedulerID string
}

func (c *CreateSchedulerCommand) CreateScheduler(ctx context.Context, db database.QueryExecer, req *CreateSchedulerRequest) (*CreateSchedulerResponse, error) {
	freq := strings.ToLower(req.Frequency)
	scheduler := entities.NewScheduler(
		req.StartDate,
		req.EndDate,
		constants.Frequency(freq),
		c.SchedulerRepo,
	)
	schedulerID, err := scheduler.Create(ctx, db)
	if err != nil {
		return nil, err
	}
	return &CreateSchedulerResponse{
		SchedulerID: schedulerID,
	}, nil
}

func (c *CreateSchedulerCommand) CreateManySchedulers(ctx context.Context, db database.QueryExecer, req *cpb.CreateManySchedulersRequest) (*cpb.CreateManySchedulersResponse, error) {
	var err error
	params := sliceutils.Map(req.Schedulers, func(s *cpb.CreateSchedulerWithIdentityRequest) *dto.CreateSchedulerParamWithIdentity {
		if s.Identity == "" {
			err = multierr.Append(err, fmt.Errorf("missing Identity"))
			return nil
		}
		param := s.Request
		if param == nil {
			err = multierr.Append(err, fmt.Errorf("%s missing param to create scheduler", s.Identity))
			return nil
		}
		freq := strings.ToLower(param.Frequency.String())
		startDate := param.StartDate.AsTime()
		endDate := param.EndDate.AsTime()

		if endDate.Before(startDate) {
			err = multierr.Append(err, fmt.Errorf("end date of %s is earlier than start date", s.Identity))
			return nil
		}

		return &dto.CreateSchedulerParamWithIdentity{
			ID: s.Identity,
			CreateSchedulerParam: dto.CreateSchedulerParams{
				SchedulerID: idutil.ULIDNow(),
				StartDate:   startDate,
				EndDate:     endDate,
				Frequency:   freq,
			},
		}
	})

	if err != nil {
		return nil, fmt.Errorf("build params to create many schedulers fail: %w", err)
	}

	mapSchedulers, err := c.SchedulerRepo.CreateMany(ctx, db, params)

	if err != nil {
		return nil, fmt.Errorf("create many schedulers fail: %w", err)
	}

	return &cpb.CreateManySchedulersResponse{
		MapSchedulers: mapSchedulers,
	}, nil
}
