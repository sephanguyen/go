package bob

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	types "github.com/gogo/protobuf/types"
	"go.uber.org/multierr"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) bobMustStoreAllReturnedMedia(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.UpsertMediaResponse)
	query := `SELECT count(*) FROM media WHERE media_id = ANY($1)`
	var count int
	if err := s.DB.QueryRow(ctx, query, rsp.MediaIds).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(rsp.MediaIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %d stored media got %d", len(rsp.MediaIds), count)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobMustRecordAllMediaList(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.returnsStatusCode(ctx, "OK")
	ctx, err2 := s.bobMustStoreAllReturnedMedia(ctx)
	return ctx, multierr.Combine(err1, err2)
}

func (s *suite) generateMedia(randStr string) *pb.Media {
	return &pb.Media{
		MediaId:   "",
		Name:      fmt.Sprintf("random-name-%s", randStr),
		Resource:  s.newID(),
		CreatedAt: types.TimestampNow(),
		UpdatedAt: types.TimestampNow(),
		Comments: []*pb.Comment{
			{Comment: "Comment-1", Duration: types.DurationProto(10 * time.Second)},
			{Comment: "Comment-2", Duration: types.DurationProto(20 * time.Second)},
		},
		Type: pb.MEDIA_TYPE_VIDEO,
	}
}

func (s *suite) generateMediaWithType(randStr string, mediaType pb.MediaType) *pb.Media {
	return &pb.Media{
		MediaId:   "",
		Name:      fmt.Sprintf("random-name-%s", randStr),
		Resource:  s.newID(),
		CreatedAt: types.TimestampNow(),
		UpdatedAt: types.TimestampNow(),
		Comments: []*pb.Comment{
			{Comment: "Comment-1", Duration: types.DurationProto(10 * time.Second)},
			{Comment: "Comment-2", Duration: types.DurationProto(20 * time.Second)},
		},
		Type: mediaType,
	}
}

func (s *suite) userUpsertValidMediaList(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	n := rand.Intn(20) + 5
	mediaList := make([]*pb.Media, 0, n)
	for i := 0; i < n; i++ {
		var mediaType pb.MediaType
		if int32(i)%int32(pb.MEDIA_TYPE_AUDIO) == 0 {
			mediaType = pb.MEDIA_TYPE_AUDIO
		} else if int32(i)%int32(pb.MEDIA_TYPE_PDF) == 0 {
			mediaType = pb.MEDIA_TYPE_PDF
		} else if int32(i)%int32(pb.MEDIA_TYPE_IMAGE) == 0 {
			mediaType = pb.MEDIA_TYPE_IMAGE
		} else {
			mediaType = pb.MEDIA_TYPE_VIDEO
		}
		mediaList = append(mediaList, s.generateMediaWithType(stepState.Random, mediaType))
	}
	req := &pb.UpsertMediaRequest{
		Media: mediaList,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).UpsertMedia(s.signedCtx(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}
