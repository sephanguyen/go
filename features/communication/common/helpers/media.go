package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (helper *CommunicationHelper) generateMedia() *npb.Media {
	return &npb.Media{
		MediaId:  "",
		Name:     fmt.Sprintf("test-file-%s.pdf", idutil.ULIDNow()),
		Resource: idutil.ULIDNow(), CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Comments: []*npb.Comment{
			{
				Comment:  "Comment-1",
				Duration: durationpb.New(10 * time.Second),
			},
			{
				Comment:  "Comment-2",
				Duration: durationpb.New(10 * time.Second),
			},
		},
		Type:     npb.MediaType_MEDIA_TYPE_PDF,
		FileSize: 315,
	}
}

func (helper *CommunicationHelper) CreateMediaViaGRPC(authToken string, numberMedia int) ([]string, error) {
	mediaList := []*npb.Media{}
	for i := 0; i < numberMedia; i++ {
		mediaList = append(mediaList, helper.generateMedia())
	}
	req := &npb.UpsertMediaRequest{
		Media: mediaList,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = contextWithToken(ctx, authToken)
	res, err := npb.NewMediaModifierServiceClient(helper.NotificationMgmtGRPCConn).UpsertMedia(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("notificationmgmt.UpsertMedia: %v", err)
	}

	return res.MediaIds, nil
}
