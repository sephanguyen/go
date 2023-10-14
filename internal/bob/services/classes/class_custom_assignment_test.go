package classes

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)

	lessonModSrvMock *LessonModifierServicesMock
}

func TestClassCreateCustomAssignment_InvalidArgument(t *testing.T) {

}

func TestClassCreateCustomAssignment(t *testing.T) {
}

func TestToProtoEn(t *testing.T) {
	t.Parallel()
	protoTimeNow := types.TimestampNow()
	comment := &pb.Comment{
		Comment:  "comment",
		Duration: types.DurationProto(10 * time.Second),
	}
	media := &pb.Media{
		MediaId: "media-id",
		Name:    "name",
		Comments: []*pb.Comment{
			comment,
		},
		Resource:  "resource",
		CreatedAt: protoTimeNow,
		UpdatedAt: protoTimeNow,
		Type:      pb.MEDIA_TYPE_IMAGE,
		Images:    nil,
	}

	enComments := []*entities.Comment{
		{
			Comment:  "comment",
			Duration: 10,
		},
	}
	var json pgtype.JSONB
	json.Set(enComments)
	expectMedia := &entities.Media{
		MediaID:         database.Text("media-id"),
		Name:            database.Text("name"),
		Resource:        database.Text("resource"),
		Type:            database.Text(pb.MEDIA_TYPE_IMAGE.String()),
		Comments:        json,
		ConvertedImages: pgtype.JSONB{Status: pgtype.Null},
	}
	expectMedia.DeletedAt.Set(nil)
	expectMedia.CreatedAt.Set(time.Unix(media.CreatedAt.Seconds, int64(media.CreatedAt.Nanos)))
	expectMedia.UpdatedAt.Set(time.Unix(media.UpdatedAt.Seconds, int64(media.UpdatedAt.Nanos)))
	result, err := toMediaEn(media)
	assert.NoError(t, err)
	assert.Equal(t, result, expectMedia)
}

func TestUpsertMedia(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mediaRepo := new(mock_repositories.MockMediaRepo)

	classService := &ClassService{
		MediaRepo: mediaRepo,
	}

	comment := &pb.Comment{
		Comment:  "comment",
		Duration: types.DurationProto(10 * time.Second),
	}
	userID := "user-id"

	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := map[string]TestCase{
		"error upsert media": {
			ctx: ctx,
			req: &pb.UpsertMediaRequest{
				Media: []*pb.Media{
					{
						MediaId: "media-id",
						Name:    "name",
						Comments: []*pb.Comment{
							comment,
						},
						Resource:  "resource",
						CreatedAt: types.TimestampNow(),
						UpdatedAt: types.TimestampNow(),
					},
				},
			},
			expectedResp: nil,
			expectedErr:  services.ToStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mediaRepo.On("UpsertMediaBatch", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		"success upsert": {
			ctx: ctx,
			req: &pb.UpsertMediaRequest{
				Media: []*pb.Media{
					{
						MediaId: "media-id-1",
						Name:    "name",
						Comments: []*pb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/ascii.pdf",
						Type:      pb.MEDIA_TYPE_PDF,
						CreatedAt: types.TimestampNow(),
						UpdatedAt: types.TimestampNow(),
					},
					{
						MediaId: "media-id-2",
						Name:    "name",
						Comments: []*pb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/jp.特殊文字.pdf",
						Type:      pb.MEDIA_TYPE_PDF,
						CreatedAt: types.TimestampNow(),
						UpdatedAt: types.TimestampNow(),
					},
					{
						MediaId: "media-id-3",
						Name:    "name",
						Comments: []*pb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/jp.特殊文字.mp4",
						Type:      pb.MEDIA_TYPE_VIDEO,
						CreatedAt: types.TimestampNow(),
						UpdatedAt: types.TimestampNow(),
					},
					{
						MediaId: "media-id-4",
						Name:    "name",
						Comments: []*pb.Comment{
							comment,
						},
						Resource:  "https://storage/bucket/ascii.@lcha%20!$^&*()[]r.mp3",
						Type:      pb.MEDIA_TYPE_AUDIO,
						CreatedAt: types.TimestampNow(),
						UpdatedAt: types.TimestampNow(),
					},
				},
			},
			expectedResp: &pb.UpsertMediaResponse{MediaIds: []string{"media-id-1", "media-id-2", "media-id-3", "media-id-4"}},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mediaRepo.On("UpsertMediaBatch", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.UpsertMediaRequest)
			rsp, err := classService.UpsertMedia(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestRetrieveMedia(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mediaRepo := new(mock_repositories.MockMediaRepo)

	classService := &ClassService{
		MediaRepo: mediaRepo,
	}

	comment := &pb.Comment{
		Comment:  "comment",
		Duration: types.DurationProto(10 * time.Second),
	}
	eComment := []*entities.Comment{
		{
			Comment:  "comment",
			Duration: 10,
		},
	}
	var json pgtype.JSONB
	_ = json.Set(eComment)

	var convertedImages pgtype.JSONB
	convertedImages.Set([]*entities.ConvertedImage{
		{
			Width:    1920,
			Height:   920,
			ImageURL: "url1",
		},
		{
			Width:    1024,
			Height:   2048,
			ImageURL: "url2",
		},
	})
	enMedia := []*entities.Media{
		{
			MediaID:         database.Text("media-id"),
			Name:            database.Text("name"),
			CreatedAt:       pgtype.Timestamptz{},
			UpdatedAt:       pgtype.Timestamptz{},
			Resource:        database.Text("video-id"),
			Comments:        json,
			ConvertedImages: pgtype.JSONB{Status: pgtype.Null},
		},
		{
			MediaID:         database.Text("media-id-2"),
			Name:            database.Text("name"),
			CreatedAt:       pgtype.Timestamptz{},
			UpdatedAt:       pgtype.Timestamptz{},
			Resource:        database.Text("video-id"),
			Comments:        json,
			ConvertedImages: convertedImages,
		},
	}
	media := []*pb.Media{
		{
			MediaId: "media-id",
			Name:    "name",
			Comments: []*pb.Comment{
				comment,
			},
			Resource:  "video-id",
			CreatedAt: types.TimestampNow(),
			UpdatedAt: types.TimestampNow(),
		},
		{
			MediaId: "media-id-2",
			Name:    "name",
			Comments: []*pb.Comment{
				comment,
			},
			Resource:  "video-id",
			CreatedAt: types.TimestampNow(),
			UpdatedAt: types.TimestampNow(),
			Images: []*pb.ConvertedImage{
				{
					Width:    1920,
					Height:   920,
					ImageUrl: "url1",
				},
				{
					Width:    1024,
					Height:   2048,
					ImageUrl: "url2",
				},
			},
		},
	}
	validReq := &pb.RetrieveMediaRequest{
		MediaIds: []string{"media-id"},
	}
	userID := "user-id"
	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := map[string]TestCase{
		"error upsert media": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			expectedErr:  services.ToStatusError(pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mediaRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"success upsert": {
			ctx: ctx,
			req: validReq,
			expectedResp: &pb.RetrieveMediaResponse{
				Media: media,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mediaRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(enMedia, nil)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.RetrieveMediaRequest)
			rsp, err := classService.RetrieveMedia(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}
