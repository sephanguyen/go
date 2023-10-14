package tom

import (
	"context"

	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) updateTeacherUserGroups(ctx context.Context, locationStatus string) (context.Context, error) {
	for _, teacherID := range s.teachersInConversation {
		teacherProfile := s.TeacherProfileMap[teacherID]

		_, err := upb.NewStaffServiceClient(s.CommonSuite.UserMgmtConn).UpdateStaff(contextWithToken(ctx, s.schoolAdminToken), &upb.UpdateStaffRequest{
			Staff: &upb.UpdateStaffRequest_StaffProfile{
				StaffId:      teacherID,
				UserGroupIds: s.userGroupIDs,
				Name:         teacherProfile.GetName(),
				Email:        teacherProfile.GetEmail(),
				LocationIds:  teacherProfile.GetLocationIds(),
			},
		})
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}
