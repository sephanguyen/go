package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMediaGRPCService_RetrieveMediasByIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mediaRepo := &mock_repositories.MockMediaRepo{}
	mediaIDs := []string{"id-1", "id-2"}

	tcs := []struct {
		name     string
		req      *lpb.RetrieveMediasByIDsRequest
		res      *lpb.RetrieveMediasByIDsResponse
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "get successfully",
			req: &lpb.RetrieveMediasByIDsRequest{
				MediaIds: mediaIDs,
			},
			res: &lpb.RetrieveMediasByIDsResponse{
				Medias: []*lpb.Media{
					{MediaId: "id-1"},
					{MediaId: "id-2"},
				},
			},
			setup: func(ctx context.Context) {
				mediaRepo.On("RetrieveMediasByIDs", ctx, db, mediaIDs).
					Return(media_domain.Medias{
						&media_domain.Media{ID: "media-id-1"},
						&media_domain.Media{ID: "media-id-2"},
					}, nil).Once()
			},
			hasError: false,
		},
		{
			name: "get failed",
			req: &lpb.RetrieveMediasByIDsRequest{
				MediaIds: mediaIDs,
			},
			res: &lpb.RetrieveMediasByIDsResponse{},
			setup: func(ctx context.Context) {
				mediaRepo.On("RetrieveMediasByIDs", ctx, db, mediaIDs).
					Return(media_domain.Medias{}, fmt.Errorf("err")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewMediaGRPCService(db, mediaRepo)
			medias, err := srv.RetrieveMediasByIDs(ctx, tc.req)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, len(tc.res.GetMedias()), len(medias.GetMedias()))
			}
			mock.AssertExpectationsForObjects(t, db, mediaRepo)
		})
	}
}

func TestMediaGRPCService_CreateMedia(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mediaRepo := &mock_repositories.MockMediaRepo{}

	now := time.Now()
	mediaPb := &lpb.Media{
		MediaId:   "media-id1",
		Name:      "media name",
		Resource:  "resource-media",
		Type:      lpb.MediaType_MEDIA_TYPE_RECORDING_VIDEO,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
		Comments: []*lpb.Comment{
			{
				Comment: "hello",
			},
		},
		FileSizeBytes: int64(123456),
		Duration:      durationpb.New(time.Duration(5)),
	}

	tcs := []struct {
		name     string
		req      *lpb.CreateMediaRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "create media successfully",
			req: &lpb.CreateMediaRequest{
				Media: mediaPb,
			},
			setup: func(ctx context.Context) {
				mediaRepo.On("CreateMedia", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						media := args.Get(2).(*media_domain.Media)
						assert.Equal(t, media.ID, "media-id1")
						assert.Equal(t, media.Name, "media name")
						assert.Equal(t, media.Type, media_domain.MediaTypeRecordingVideo)
						assert.NotEmpty(t, media.CreatedAt, now)
						assert.NotEmpty(t, media.UpdatedAt, now)
						assert.Equal(t, media.FileSizeBytes, int64(123456))
						assert.Equal(t, media.Duration, time.Duration(5))
					}).
					Return(nil).Once()
			},
			hasError: false,
		},
		{
			name: "failed to create media",
			req: &lpb.CreateMediaRequest{
				Media: mediaPb,
			},
			setup: func(ctx context.Context) {
				mediaRepo.On("CreateMedia", ctx, db, mock.Anything).
					Run(func(args mock.Arguments) {
						media := args.Get(2).(*media_domain.Media)
						assert.Equal(t, media.ID, "media-id1")
						assert.Equal(t, media.Name, "media name")
						assert.Equal(t, media.Type, media_domain.MediaTypeRecordingVideo)
						assert.NotEmpty(t, media.CreatedAt, now)
						assert.NotEmpty(t, media.UpdatedAt, now)
						assert.Equal(t, media.FileSizeBytes, int64(123456))
						assert.Equal(t, media.Duration, time.Duration(5))
					}).
					Return(fmt.Errorf("error")).Once()
			},
			hasError: true,
		},
		{
			name: "invalid media missing ID",
			req: &lpb.CreateMediaRequest{
				Media: &lpb.Media{
					MediaId:   "",
					Name:      "media name",
					Resource:  "resource-media",
					Type:      lpb.MediaType_MEDIA_TYPE_RECORDING_VIDEO,
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
					Comments: []*lpb.Comment{
						{
							Comment: "hello",
						},
					},
					FileSizeBytes: int64(123456),
					Duration:      durationpb.New(time.Duration(5)),
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "invalid media missing resource",
			req: &lpb.CreateMediaRequest{
				Media: &lpb.Media{
					MediaId:   "media-id1",
					Name:      "media name",
					Type:      lpb.MediaType_MEDIA_TYPE_RECORDING_VIDEO,
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
					Comments: []*lpb.Comment{
						{
							Comment: "hello",
						},
					},
					FileSizeBytes: int64(123456),
					Duration:      durationpb.New(time.Duration(5)),
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "invalid media missing type",
			req: &lpb.CreateMediaRequest{
				Media: &lpb.Media{
					MediaId:   "media-id1",
					Name:      "media name",
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
					Comments: []*lpb.Comment{
						{
							Comment: "hello",
						},
					},
					FileSizeBytes: int64(123456),
					Duration:      durationpb.New(time.Duration(5)),
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewMediaGRPCService(db, mediaRepo)
			_, err := srv.CreateMedia(ctx, tc.req)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, mediaRepo)
		})
	}
}

func TestMediaGRPCService_DeleteMedias(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mediaRepo := &mock_repositories.MockMediaRepo{}

	mediaIDs := []string{
		"media-id1",
		"media-id2",
		"media-id3",
	}

	tcs := []struct {
		name     string
		req      *lpb.DeleteMediasRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "delete medias successfully",
			req: &lpb.DeleteMediasRequest{
				MediaIds: mediaIDs,
			},
			setup: func(ctx context.Context) {
				mediaRepo.On("DeleteMedias", ctx, db, mediaIDs).
					Return(nil).Once()
			},
			hasError: false,
		},
		{
			name: "failed to delete medias",
			req: &lpb.DeleteMediasRequest{
				MediaIds: mediaIDs,
			},
			setup: func(ctx context.Context) {
				mediaRepo.On("DeleteMedias", ctx, db, mediaIDs).
					Return(fmt.Errorf("error")).Once()
			},
			hasError: true,
		},
		{
			name: "invalid no media IDs",
			req: &lpb.DeleteMediasRequest{
				MediaIds: []string{},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			srv := NewMediaGRPCService(db, mediaRepo)
			_, err := srv.DeleteMedias(ctx, tc.req)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, mediaRepo)
		})
	}
}
