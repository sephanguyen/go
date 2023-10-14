package staff

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgx/v4"
)

func userToSchoolAdmin(user *entity.LegacyUser, schoolID int32) *entity.SchoolAdmin {
	return &entity.SchoolAdmin{
		SchoolAdminID: user.ID,
		ResourcePath:  user.ResourcePath,
		LegacyUser:    *user,
		SchoolID:      database.Int4(schoolID),
	}
}

func (s *StaffService) createSchoolAdmin(ctx context.Context, tx pgx.Tx, user *entity.LegacyUser, schoolIDs []int64) error {
	// school admin just only have one school id
	schoolAdmin := userToSchoolAdmin(user, int32(schoolIDs[0]))
	return s.UserModifierService.SchoolAdminRepo.CreateMultiple(ctx, tx, []*entity.SchoolAdmin{schoolAdmin})
}
