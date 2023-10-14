package staff

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StaffService) UpdateStaffSetting(ctx context.Context, req *pb.UpdateStaffSettingRequest) (*pb.UpdateStaffSettingResponse, error) {
	if err := validateUpdateStaffSetting(req); err != nil {
		return nil, err
	}
	staff, err := s.StaffRepo.FindByID(ctx, s.DB, database.Text(req.StaffId))
	if err != nil {
		return nil, errorx.ToStatusError(err)
	}
	staff.AutoCreateTimesheet = database.Bool(req.AutoCreateTimesheet)
	_, err = s.StaffRepo.Update(ctx, s.DB, staff)
	if err != nil {
		return nil, errorx.ToStatusError(err)
	}
	eventStaffConfig := toEventStaffUpsertTimesheetSetting(staff.ID.String, staff.AutoCreateTimesheet.Bool, staff.UpdatedAt.Time)
	err = s.publishStaffSettingEvent(ctx, constants.SubjectStaffUpsertTimesheetConfig, eventStaffConfig)
	if err != nil {
		return nil, errorx.ToStatusError(err)
	}
	return &pb.UpdateStaffSettingResponse{
		Successful: true,
	}, nil
}

func validateUpdateStaffSetting(req *pb.UpdateStaffSettingRequest) error {
	if req.StaffId == "" {
		return status.Error(codes.InvalidArgument, "staff id cannot be null or empty")
	}
	return nil
}

func toEventStaffUpsertTimesheetSetting(staffID string, autoCreateTimesheet bool, updatedAt time.Time) *pb.EvtStaffUpsertTimesheetConfig {
	return &pb.EvtStaffUpsertTimesheetConfig{
		StaffId:                   staffID,
		AutoCreateTimesheetConfig: autoCreateTimesheet,
		UpdatedAt:                 timestamppb.New(updatedAt),
	}
}

func toEventUpsertStaff(staffID string, userGroupIDs []string, locationIDs []string, evtType pb.EvtUpsertStaff_UpsertStaffType) *pb.EvtUpsertStaff {
	return &pb.EvtUpsertStaff{
		StaffId:      staffID,
		UserGroupIds: userGroupIDs,
		LocationIds:  locationIDs,
		Type:         evtType,
	}
}

func (s *StaffService) publishStaffSettingEvent(ctx context.Context, subject string, event *pb.EvtStaffUpsertTimesheetConfig) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event %s error, %w", subject, err)
	}
	_, err = s.JSM.PublishContext(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("publishStaffSettingEvent error: %w", err)
	}
	return nil
}

func (s *StaffService) publishUpsertStaffEvent(ctx context.Context, subject string, event *pb.EvtUpsertStaff) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event %s error, %w", subject, err)
	}
	_, err = s.JSM.PublishContext(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("publishUpsertStaffEvent error: %w", err)
	}
	return nil
}
