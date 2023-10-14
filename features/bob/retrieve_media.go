package bob

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/multierr"
)

func (s *suite) returnValidMedia(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.RetrieveMediaRequest)
	rsp := stepState.Response.(*pb.RetrieveMediaResponse)
	mediaIDs := make([]string, len(rsp.Media))
	for i, media := range rsp.Media {
		mediaIDs[i] = media.MediaId
	}

	query := `SELECT count(*) FROM media WHERE media_id = ANY($1)`
	var count int
	if err := s.DB.QueryRow(ctx, query, &mediaIDs).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(req.MediaIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d media got %d", len(req.MediaIds), count)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnAllMedia(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.returnsStatusCode(ctx, "OK")
	ctx, err2 := s.returnValidMedia(ctx)
	return ctx, multierr.Combine(err1, err2)
}
func (s *suite) studentHasMultipleMedia(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.aSignedInStudent(ctx)
	ctx, err2 := s.userUpsertValidMediaList(ctx)
	return ctx, multierr.Combine(err1, err2)
}
func (s *suite) studentRetrieveMediaByIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	mediaIDs := stepState.Response.(*pb.UpsertMediaResponse).MediaIds
	req := &pb.RetrieveMediaRequest{
		MediaIds: mediaIDs,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).RetrieveMedia(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
