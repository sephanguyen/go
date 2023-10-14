package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) userUpdateStaffConfig(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpdateStaffSettingRequest)
	stepState.Response, stepState.ResponseErr = pb.NewStaffServiceClient(s.UserMgmtConn).UpdateStaffSetting(ctx, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStaffConfigWith(ctx context.Context, staffID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.UpdateStaffSettingRequest{}
	orgID := OrgIDFromCtx(ctx)

	switch staffID {
	case "exist":
		temCtx := s.signedIn(context.Background(), orgID, StaffRoleSchoolAdmin)
		roleWithLocationTeacher := RoleWithLocation{
			RoleName:    constant.RoleTeacher,
			LocationIDs: []string{constants.ManabieOrgLocation},
		}
		resp, err := CreateStaff(temCtx, s.BobDBTrace, s.UserMgmtConn, nil, []RoleWithLocation{roleWithLocationTeacher}, getChildrenLocation(orgID))
		if err != nil {
			return nil, fmt.Errorf("s.aStaffConfigWith: %w", err)
		}
		req.StaffId = resp.Staff.StaffId
	case "empty":
		req.StaffId = ""
	}

	req.AutoCreateTimesheet = randomBool()

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
