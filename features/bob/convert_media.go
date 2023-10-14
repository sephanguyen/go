package bob

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
)

func (s *suite) aListOfMedia(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	medias := []*pb.Media{
		s.generateMedia(stepState.Random),
		s.generateMedia(stepState.Random),
		s.generateMedia(stepState.Random),
		s.generateMedia(stepState.Random),
		s.generateMedia(stepState.Random),
	}
	medias[1].Type = pb.MEDIA_TYPE_PDF
	medias[3].Type = pb.MEDIA_TYPE_PDF
	medias[4].Type = pb.MEDIA_TYPE_PDF

	stepState.Request = medias
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userConvertsMediaToImage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	medias := stepState.Request.([]*pb.Media)
	if _, err := pb.NewClassClient(s.Conn).UpsertMedia(s.signedCtx(ctx), &pb.UpsertMediaRequest{
		Media: medias,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	mediasV1 := make([]*bpb.Media, 0, len(medias))
	for _, media := range medias {
		mediasV1 = append(mediasV1, &bpb.Media{
			Name:     media.Name,
			Resource: media.Resource,
			Type:     bpb.MediaType(media.Type),
		})
	}

	stepState.Request = mediasV1
	if _, err := bpb.NewClassModifierServiceClient(s.Conn).ConvertMedia(s.signedCtx(ctx), &bpb.ConvertMediaRequest{
		Media: mediasV1,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) mediaConversionTasksMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var resourceURLs []string
	for _, media := range stepState.Request.([]*bpb.Media) {
		if media.Type == bpb.MediaType_MEDIA_TYPE_PDF {
			resourceURLs = append(resourceURLs, media.Resource)
		}
	}
	query := "SELECT COUNT(*) FROM conversion_tasks WHERE resource_url = ANY($1)"

	var count int
	if err := s.DB.QueryRow(ctx, query, database.TextArray(resourceURLs)).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(resourceURLs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("conversion tasks are missing, got: %d, want: %d", count, len(resourceURLs))
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aListOfMediaConversionTasks(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.aListOfMedia(ctx)
	ctx, err2 := s.aSignedInTeacher(ctx)
	ctx, err3 := s.userConvertsMediaToImage(ctx)
	return ctx, multierr.Combine(err1, err2, err3)

}
func (s *suite) ourSystemReceivesAFinishedConversionEvent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	medias := stepState.Request.([]*bpb.Media)
	var media *bpb.Media
	for _, m := range medias {
		if m.Type == bpb.MediaType_MEDIA_TYPE_PDF {
			media = m
			break
		}
	}

	var jobID string
	query := "SELECT task_uuid FROM conversion_tasks WHERE resource_url = $1"

	if err := database.Select(ctx, s.DB, query, media.Resource).ScanFields(&jobID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	rnd := s.newID()
	now := time.Now().UTC().Format("2006-01-02")
	data := &npb.CloudConvertJobData{
		JobId:      jobID,
		JobStatus:  "job.finished",
		Signature:  "sig",
		RawPayload: nil,
		ExportName: fmt.Sprintf("export-%s", rnd),
		ConvertedFiles: []string{
			fmt.Sprintf("assignment/%s/images/%s/f1", now, rnd),
			fmt.Sprintf("assignment/%s/images/%s/f2", now, rnd),
		},
	}
	msg, err := proto.Marshal(data)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.Request = media
	if _, err := s.JSM.PublishContext(ctx, constants.SubjectCloudConvertJobEventNatsJS, msg); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) finishedConversionTasksMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// waiting a bit to let subscription has a chance to execute
	time.Sleep(time.Second)

	media := stepState.Request.(*bpb.Media)

	query := "SELECT status FROM conversion_tasks WHERE resource_url = $1"
	rows, err := s.DB.Query(ctx, query, media.Resource)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	defer rows.Close()

	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if status != bpb.ConversionTaskStatus_CONVERSION_TASK_STATUS_FINISHED.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected conversion task status: %q, expected: %q", status, bpb.ConversionTaskStatus_CONVERSION_TASK_STATUS_FINISHED.String())
		}
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	resourcePath := golibs.ResourcePathFromCtx(ctx)

	query = "SELECT media_id FROM media WHERE resource = $1 AND resource_path = $2"
	rows, err = s.DB.Query(ctx, query, media.Resource, resourcePath)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, err := pb.NewClassClient(s.Conn).RetrieveMedia(s.signedCtx(ctx), &pb.RetrieveMediaRequest{
		MediaIds: ids,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	for _, media := range resp.Media {
		if len(media.Images) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("converted images in media must exist")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
