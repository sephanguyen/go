package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RecordedVideoQuery struct {
	WrapperDBConnection *support.WrapperDBConnection

	RecordedVideoRepo infrastructure.RecordedVideoRepo
	MediaModulePort   infrastructure.MediaModulePort
}

type RetrieveRecordedVideosByLessonIDQueryResponse struct {
	Recs      domain.RecordedVideos
	Total     uint32
	PrePageID string
	Err       error
}

func (r *RecordedVideoQuery) RetrieveRecordedVideosByLessonID(ctx context.Context, args *payloads.RetrieveRecordedVideosByLessonIDPayload) *RetrieveRecordedVideosByLessonIDQueryResponse {
	conn, err := r.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return &RetrieveRecordedVideosByLessonIDQueryResponse{
			Recs: domain.RecordedVideos{},
			Err:  err}
	}
	records, total, prePageID, preTotal, err := r.RecordedVideoRepo.ListRecordingByLessonIDWithPaging(ctx, conn, args)
	if err != nil {
		return &RetrieveRecordedVideosByLessonIDQueryResponse{
			Recs: domain.RecordedVideos{},
			Err:  err}
	}
	if preTotal <= args.Limit {
		prePageID = ""
	}
	return &RetrieveRecordedVideosByLessonIDQueryResponse{
		Recs:      records,
		Total:     total,
		PrePageID: prePageID,
		Err:       nil,
	}
}

func (r *RecordedVideoQuery) GetRecordingByID(ctx context.Context, args *payloads.GetRecordingByIDPayload) (*domain.RecordedVideo, error) {
	conn, err := r.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if len(args.RecordedVideoID) == 0 {
		return nil, status.Error(codes.Internal, "missing paging info")
	}
	record, err := r.RecordedVideoRepo.GetRecordingByID(ctx, conn, args)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when call RecordedVideoRepo.GetRecordingByID: %s", err))
	}
	medias, err := r.MediaModulePort.RetrieveMediasByIDs(ctx, []string{record.Media.ID})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when call MediaModulePort.RetrieveMediasByIDs: %v", err))
	}
	if len(medias) == 0 {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when call MediaModulePort.RetrieveMediasByIDs %s don't have any media with id: ", args.RecordedVideoID))
	}
	record.Media = medias[0]
	return record, nil
}
