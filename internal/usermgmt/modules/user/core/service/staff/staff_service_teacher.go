package staff

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func userToTeacher(user *entity.LegacyUser, schoolIDs []int32) *entity.Teacher {
	return &entity.Teacher{
		ID:           user.ID,
		ResourcePath: user.ResourcePath,
		LegacyUser:   *user,
		SchoolIDs:    database.Int4Array(schoolIDs),
		DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
	}
}

func (s *StaffService) createTeacher(ctx context.Context, tx pgx.Tx, user *entity.LegacyUser, schoolIDs []int64) error {
	schoolIDsInt32 := make([]int32, 0)
	for _, schoolIDInt32 := range schoolIDs {
		schoolIDsInt32 = append(schoolIDsInt32, int32(schoolIDInt32))
	}

	teacher := userToTeacher(user, schoolIDsInt32)
	return s.UserModifierService.TeacherRepo.CreateMultiple(ctx, tx, []*entity.Teacher{teacher})
}
