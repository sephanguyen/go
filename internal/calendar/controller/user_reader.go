package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/calendar/application"
	"github.com/manabie-com/backend/internal/calendar/application/queries"
	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserReaderService struct {
	UserPort         application.GetStaffPort
	db               database.Ext
	unleashClientIns unleashclient.ClientInstance
	env              string
}

func NewUserReaderService(
	userRepo infrastructure.UserPort,
	db database.Ext,
	unleashClient unleashclient.ClientInstance,
	env string,
) *UserReaderService {
	return &UserReaderService{
		UserPort: &queries.GetStaff{
			UserRepo: userRepo,
		},
		db:               db,
		unleashClientIns: unleashClient,
		env:              env,
	}
}

func (u *UserReaderService) GetStaffsByLocation(ctx context.Context, in *cpb.GetStaffsByLocationRequest) (*cpb.GetStaffsByLocationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	isUnleashToggled, err := u.unleashClientIns.IsFeatureEnabledOnOrganization("Lesson_LessonManagement_BackOffice_SwitchNewDBConnection", u.env, golibs.ResourcePathFromCtx(ctx))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to unleash: %w", err)
	}

	req := &payloads.GetStaffRequest{
		LocationID:                in.GetLocationId(),
		IsUsingUserBasicInfoTable: isUnleashToggled,
	}
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := u.UserPort.GetStaffsByLocation(ctx, u.db, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	staffInfo := make([]*cpb.GetStaffsByLocationResponse_StaffInfo, 0, len(resp.User))
	for _, t := range resp.User {
		staffInfo = append(staffInfo, &cpb.GetStaffsByLocationResponse_StaffInfo{
			Id:    t.UserID,
			Name:  t.Name,
			Email: t.Email,
		})
	}

	return &cpb.GetStaffsByLocationResponse{
		Staffs: staffInfo,
	}, nil
}

func (u *UserReaderService) GetStaffsByLocationIDsAndNameOrEmail(ctx context.Context, in *cpb.GetStaffsByLocationIDsAndNameOrEmailRequest) (*cpb.GetStaffsByLocationIDsAndNameOrEmailResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req := &payloads.GetStaffByLocationIDsAndNameOrEmailRequest{
		LocationIDs:        in.LocationIds,
		Keyword:            in.Keyword,
		FilteredTeacherIDs: in.FilteredTeacherIds,
		Limit:              int(in.Limit),
	}
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := u.UserPort.GetStaffsByLocationIDsAndNameOrEmail(ctx, u.db, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	staffInfo := make([]*cpb.StaffInfo, 0, len(resp.User))
	for _, t := range resp.User {
		staffInfo = append(staffInfo, &cpb.StaffInfo{
			Id:    t.UserID,
			Name:  t.Name,
			Email: t.Email,
		})
	}

	return &cpb.GetStaffsByLocationIDsAndNameOrEmailResponse{
		Staffs: staffInfo,
	}, nil
}
