package controller

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/media/application/commands"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/notification/modules/media/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMediaModifierService_UpsertMedia(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mediaRepo := new(mock_repo.MockMediaRepo)
	db := new(mock_database.Ext)

	mediaModifierService := &MediaModifierService{
		UpsertMediaCommandHandler: commands.UpsertMediaCommandHandler{
			DB:        db,
			MediaRepo: mediaRepo,
		},
	}

	comment := &npb.Comment{
		Comment:  "comment",
		Duration: durationpb.New(10 * time.Second),
	}
	userID := "user-id"

	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := []struct {
		Name  string
		Err   error
		Req   *npb.UpsertMediaRequest
		Res   *npb.UpsertMediaResponse
		Setup func(ctx context.Context)
	}{
		{
			Name: "error upsert media",
			Req: &npb.UpsertMediaRequest{
				Media: []*npb.Media{
					{
						MediaId: "media-id",
						Name:    "name",
						Comments: []*npb.Comment{
							comment,
						},
						Resource:  "resource",
						CreatedAt: timestamppb.Now(),
						UpdatedAt: timestamppb.Now(),
					},
				},
			},
			Res: nil,
			Err: status.Errorf(codes.Internal, "svc.UpsertMediaCommandHandler.UpsertMedia: %v", pgx.ErrNoRows),
			Setup: func(ctx context.Context) {
				mediaRepo.On("UpsertMediaBatch", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			Name: "success upsert",
			Req: &npb.UpsertMediaRequest{
				Media: []*npb.Media{
					{
						MediaId: "media-id-1",
						Name:    "name",
						Comments: []*npb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/ascii.pdf",
						Type:      npb.MediaType_MEDIA_TYPE_PDF,
						CreatedAt: timestamppb.Now(),
						UpdatedAt: timestamppb.Now(),
					},
					{
						MediaId: "media-id-2",
						Name:    "name",
						Comments: []*npb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/jp.特殊文字.pdf",
						Type:      npb.MediaType_MEDIA_TYPE_PDF,
						CreatedAt: timestamppb.Now(),
						UpdatedAt: timestamppb.Now(),
					},
					{
						MediaId: "media-id-3",
						Name:    "name",
						Comments: []*npb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/jp.特殊文字.mp4",
						Type:      npb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.Now(),
						UpdatedAt: timestamppb.Now(),
					},
					{
						MediaId: "media-id-4",
						Name:    "name",
						Comments: []*npb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/ascii.@lcha%20!$^&*()[]r.mp3",
						Type:      npb.MediaType_MEDIA_TYPE_AUDIO,
						CreatedAt: timestamppb.Now(),
						UpdatedAt: timestamppb.Now(),
					},
				},
			},
			Res: &npb.UpsertMediaResponse{MediaIds: []string{"media-id-1", "media-id-2", "media-id-3", "media-id-4"}},
			Err: nil,
			Setup: func(ctx context.Context) {
				mediaRepo.On("UpsertMediaBatch", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			Req := testCase.Req
			rsp, err := mediaModifierService.UpsertMedia(ctx, Req)
			assert.Equal(t, status.Code(testCase.Err), status.Code(err))
			if testCase.Err != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}
