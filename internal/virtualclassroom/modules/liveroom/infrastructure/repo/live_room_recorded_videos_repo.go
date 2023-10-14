package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type LiveRoomRecordedVideosRepo struct{}

func (l *LiveRoomRecordedVideosRepo) InsertRecordedVideos(ctx context.Context, db database.QueryExecer, videos []*vc_domain.RecordedVideo) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomRecordedVideosRepo.InsertRecordedVideos")
	defer span.End()

	refdto := &LiveRoomRecordedVideo{}
	fieldNames, _ := refdto.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		refdto.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	b := &pgx.Batch{}
	for _, v := range videos {
		if v == nil {
			return fmt.Errorf("could not insert a null recorded video")
		}
		dto, err := NewLiveRoomRecordedVideoFromEntity(v)
		if err != nil {
			return err
		}
		if err = dto.PreInsert(); err != nil {
			return fmt.Errorf("got error when PreInsert live room recorded video dto: %w", err)
		}

		scanFields := database.GetScanFields(dto, fieldNames)
		b.Queue(query, scanFields...)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("live room recorded video is not inserted")
		}
	}

	return nil
}

func (l *LiveRoomRecordedVideosRepo) GetLiveRoomRecordingsByChannelIDs(ctx context.Context, db database.QueryExecer, channelIDs []string) (vc_domain.RecordedVideos, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomRecordedVideosRepo.GetLiveRoomRecordingsByChannelIDs")
	defer span.End()

	if len(channelIDs) == 0 {
		return nil, fmt.Errorf("channel IDs must be not empty")
	}

	dto := &LiveRoomRecordedVideo{}
	fieldNames := database.GetFieldNames(dto)

	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE channel_id = ANY($1) 
		AND deleted_at IS NULL`,
		strings.Join(fieldNames, ","),
		dto.TableName(),
	)
	result := make(LiveRoomRecordedVideos, 0)
	err := database.Select(ctx, db, query, &channelIDs).ScanAll(&result)
	if err != nil {
		return nil, err
	}

	return result.ToRecordedVideosEntity(), nil
}
