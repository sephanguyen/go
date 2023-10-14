package helpers

import (
	"context"

	"github.com/manabie-com/backend/features/communication/common/entities"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func (helper *CommunicationHelper) CreateTags(admin *entities.Staff, tags []*entities.Tag) error {
	ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), admin.Token)
	defer cancel()
	for _, tag := range tags {
		req := &npb.UpsertTagRequest{
			TagId: tag.ID,
			Name:  tag.Name,
		}
		_, err := npb.NewTagMgmtModifierServiceClient(helper.NotificationMgmtGRPCConn).UpsertTag(ctx, req)
		if err != nil {
			return err
		}
	}

	return nil
}
