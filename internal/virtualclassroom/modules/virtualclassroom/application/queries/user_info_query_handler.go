package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type UserInfoQuery struct {
	WrapperDBConnection *support.WrapperDBConnection
	UserBasicInfoRepo   infrastructure.UserBasicInfoRepo
}

func (u *UserInfoQuery) GetUserInfosByIDs(ctx context.Context, userIDs []string) ([]domain.UserBasicInfo, error) {
	conn, err := u.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	userInfos, err := u.UserBasicInfoRepo.GetUserInfosByIDs(ctx, conn, userIDs)
	if err != nil {
		return nil, fmt.Errorf("error in UserBasicInfoRepo.GetUserInfosByIDs, userIDs %s: %w", userIDs, err)
	}

	return userInfos, nil
}
