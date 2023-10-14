package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	media_infra "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type RecordedVideo struct {
	RecordedVideoID  pgtype.Text
	LessonID         pgtype.Text
	Description      pgtype.Text
	DateTimeRecorded pgtype.Timestamptz
	Creator          pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	MediaID          pgtype.Text
}

func (r *RecordedVideo) FieldMap() ([]string, []interface{}) {
	return []string{
			"recorded_video_id",
			"lesson_id",
			"description",
			"date_time_recorded",
			"creator",
			"created_at",
			"updated_at",
			"deleted_at",
			"media_id",
		}, []interface{}{
			&r.RecordedVideoID,
			&r.LessonID,
			&r.Description,
			&r.DateTimeRecorded,
			&r.Creator,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
			&r.MediaID,
		}
}

func (r *RecordedVideo) TableName() string {
	return "lesson_recorded_videos"
}

func (r *RecordedVideo) PreInsert() error {
	now := time.Now()
	if err := multierr.Combine(
		r.CreatedAt.Set(now),
		r.UpdatedAt.Set(now),
	); err != nil {
		return err
	}
	return nil
}

func (r *RecordedVideo) ToRecordedVideoEntity() *domain.RecordedVideo {
	return &domain.RecordedVideo{
		ID:                 r.RecordedVideoID.String,
		RecordingChannelID: r.LessonID.String,
		Description:        r.Description.String,
		DateTimeRecorded:   r.DateTimeRecorded.Time,
		Creator:            r.Creator.String,
		CreatedAt:          r.CreatedAt.Time,
		UpdatedAt:          r.UpdatedAt.Time,
		Media:              &media_domain.Media{ID: r.MediaID.String},
	}
}

type RecordedVideos []*RecordedVideo

func (r *RecordedVideos) Add() database.Entity {
	e := &RecordedVideo{}
	*r = append(*r, e)

	return e
}

func (r RecordedVideos) ToRecordedVideosEntity() domain.RecordedVideos {
	res := make(domain.RecordedVideos, 0, len(r))
	for _, video := range r {
		res = append(res, &domain.RecordedVideo{
			ID:                 video.RecordedVideoID.String,
			RecordingChannelID: video.LessonID.String,
			Description:        video.Description.String,
			DateTimeRecorded:   video.DateTimeRecorded.Time,
			Creator:            video.Creator.String,
			CreatedAt:          video.CreatedAt.Time,
			UpdatedAt:          video.UpdatedAt.Time,
			Media: &media_domain.Media{
				ID: video.MediaID.String,
			},
		})
	}

	return res
}

func NewRecordedVideoFromEntity(r *domain.RecordedVideo) (*RecordedVideo, error) {
	dto := &RecordedVideo{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.RecordedVideoID.Set(r.ID),
		dto.LessonID.Set(r.RecordingChannelID),
		dto.Description.Set(r.Description),
		dto.DateTimeRecorded.Set(r.DateTimeRecorded),
		dto.Creator.Set(r.Creator),
		dto.CreatedAt.Set(r.CreatedAt),
		dto.UpdatedAt.Set(r.UpdatedAt),
		dto.MediaID.Set(r.GetMediaID()),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from recorded video entity to recorded video dto: %w", err)
	}

	return dto, nil
}

type RecordedVideoRepo struct{}

func (r *RecordedVideoRepo) InsertRecordedVideos(ctx context.Context, db database.QueryExecer, videos []*domain.RecordedVideo) error {
	ctx, span := interceptors.StartSpan(ctx, "RecordedVideoRepo.InsertRecordedVideos")
	defer span.End()

	b := &pgx.Batch{}
	for _, v := range videos {
		if v == nil {
			return fmt.Errorf("could not insert a null recorded video")
		}
		dto, err := NewRecordedVideoFromEntity(v)
		if err != nil {
			return err
		}
		if err = dto.PreInsert(); err != nil {
			return fmt.Errorf("got error when PreInsert recorded video dto: %w", err)
		}

		fieldNames, _ := new(RecordedVideo).FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := fmt.Sprintf("INSERT INTO lesson_recorded_videos (%s) VALUES (%s)",
			strings.Join(fieldNames, ","),
			placeHolders,
		)
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
			return fmt.Errorf("lesson recorded video is not inserted")
		}
	}

	return nil
}

func (r *RecordedVideoRepo) ListRecordingByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (domain.RecordedVideos, error) {
	ctx, span := interceptors.StartSpan(ctx, "RecordedVideoRepo.ListRecordingByLessonIDs")
	defer span.End()

	if len(lessonIDs) == 0 {
		return nil, fmt.Errorf("lessonID must be not empty")
	}

	e := &RecordedVideo{}
	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE lesson_id = ANY($1) AND deleted_at IS NULL`, strings.Join(fieldNames, ","), e.TableName())
	result := make(RecordedVideos, 0)
	err := database.Select(ctx, db, query, &lessonIDs).ScanAll(&result)
	if err != nil {
		return nil, err
	}

	return result.ToRecordedVideosEntity(), nil
}

func (r *RecordedVideoRepo) GetRecordingByID(ctx context.Context, db database.QueryExecer, payload *payloads.GetRecordingByIDPayload) (*domain.RecordedVideo, error) {
	ctx, span := interceptors.StartSpan(ctx, "RecordedVideoRepo.GetRecordingByID")
	defer span.End()

	if len(payload.RecordedVideoID) == 0 {
		return nil, fmt.Errorf("lessonID must be not empty")
	}

	e := &RecordedVideo{}
	fieldNames := database.GetFieldNames(e)
	_, values := e.FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE recorded_video_id = $1 AND deleted_at IS NULL`, strings.Join(fieldNames, ","), e.TableName())
	err := db.QueryRow(ctx, query, &payload.RecordedVideoID).Scan(values...)
	if err != nil {
		return nil, err
	}
	result := e.ToRecordedVideoEntity()

	return result, nil
}

func (r *RecordedVideoRepo) DeleteRecording(ctx context.Context, db database.QueryExecer, recordIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "RecordedVideoRepo.DeleteRecording")
	defer span.End()

	query := "UPDATE lesson_recorded_videos SET deleted_at = now(), updated_at = now() WHERE recorded_video_id = ANY($1) AND deleted_at IS NULL"
	command, err := db.Exec(ctx, query, &recordIDs)
	if command.RowsAffected() == 0 {
		return fmt.Errorf("not found any recorded video to update")
	}

	return err
}

func (r *RecordedVideoRepo) ListRecordingByLessonIDWithPaging(ctx context.Context, db database.QueryExecer, payload *payloads.RetrieveRecordedVideosByLessonIDPayload) (domain.RecordedVideos, uint32, string, uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "RecordedVideoRepo.ListRecordingByLessonIDWithPaging")
	defer span.End()

	if len(payload.LessonID) == 0 {
		return nil, 0, "", 0, fmt.Errorf("lessonID must be not empty")
	}

	// query for offset
	temporaryTable := `WITH filter_lesson_recorded_videos AS(select lrv.recorded_video_id, lrv.date_time_recorded, m.resource, m.file_size_bytes, m.duration_seconds :baseTable :condition ) `
	// query for selection
	creationQuerySelection := `:temporaryTable select recorded_video_id, date_time_recorded, resource, file_size_bytes, duration_seconds from filter_lesson_recorded_videos lrv :whereGetList`

	baseTable := `FROM lesson_recorded_videos lrv
				left join media m on m.media_id = lrv.media_id`
	condition := `WHERE lrv.deleted_at IS null 
			and m.type = 'MEDIA_TYPE_RECORDING_VIDEO'
			and m.deleted_at IS NULL and lrv.lesson_id = $1`
	args := []interface{}{payload.LessonID}
	paramsNum := len(args)
	countArgs := args
	paramNumLimit := paramsNum + 1
	paramNumSchoolID := paramsNum + 2
	args = append(args, &payload.Limit)
	whereGetList := ""
	if payload.RecordedVideoID != "" {
		whereGetList += fmt.Sprintf(` where (lrv.date_time_recorded, lrv.recorded_video_id) > ((SELECT date_time_recorded FROM lesson_recorded_videos WHERE recorded_video_id = $%d LIMIT 1), $%d)`, paramNumSchoolID, paramNumSchoolID)
		args = append(args, &payload.RecordedVideoID)
	}

	whereGetList += fmt.Sprintf(`order by lrv.date_time_recorded ASC, lrv.recorded_video_id ASC 
	LIMIT $%d`, paramNumLimit)

	// build condition and baseTable for placeholder temporaryTable
	temporaryTable = strings.Replace(temporaryTable, ":condition", condition, 1)
	temporaryTable = strings.Replace(temporaryTable, ":baseTable", baseTable, 1)

	// build temporaryTable and whereGetList for placeholder creationQuerySelection
	creationQuerySelection = strings.Replace(creationQuerySelection, ":temporaryTable", temporaryTable, 1)
	creationQuerySelection = strings.Replace(creationQuerySelection, ":whereGetList", whereGetList, 1)

	var rs domain.RecordedVideos
	var err error
	rows, err := db.Query(ctx, creationQuerySelection, args...)
	if err != nil {
		return nil, 0, "", 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var r RecordedVideo
		var m media_infra.Media
		if err := rows.Scan(&r.RecordedVideoID, &r.DateTimeRecorded, &m.Resource, &m.FileSizeBytes, &m.DurationSeconds); err != nil {
			return nil, 0, "", 0, errors.Wrap(err, "rows.Scan")
		}
		dRecord := r.ToRecordedVideoEntity()
		dRecord.Media = m.ToMediaEntity()
		rs = append(rs, dRecord)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "rows.Err")
	}

	var preOffset pgtype.Text
	var preTotal pgtype.Int8
	if payload.RecordedVideoID != "" {
		temporaryTableValue := strings.Replace(strings.Replace(temporaryTable, ":condition", condition, 1), ":baseTable", baseTable, 1)
		query := temporaryTableValue + fmt.Sprintf(`
		, previous_sort as(
			select lrv.recorded_video_id, lrv.date_time_recorded,
					COUNT(*) OVER() AS total
			from filter_lesson_recorded_videos lrv
			where $%d::text is not NULL
					and (lrv.date_time_recorded, lrv.recorded_video_id) < ((SELECT date_time_recorded FROM lesson_recorded_videos WHERE recorded_video_id = $%d LIMIT 1), $%d)
			order by lrv.date_time_recorded DESC, lrv.recorded_video_id DESC
			LIMIT $%d
		)
			select ps.recorded_video_id AS pre_offset, ps.total AS pre_total
			from previous_sort ps
			order by ps.date_time_recorded ASC 
			limit 1
			`, paramNumSchoolID, paramNumSchoolID, paramNumSchoolID, paramNumLimit)
		if err := db.QueryRow(ctx, query, args...).Scan(&preOffset, &preTotal); err != nil {
			if err.Error() != pgx.ErrNoRows.Error() {
				return nil, 0, "", 0, errors.Wrap(err, "get previous err")
			}
		}
	}

	// get total
	var total pgtype.Int8
	queryTotal := "select count(*) " + baseTable + " " + condition
	if err := db.QueryRow(ctx, queryTotal, countArgs...).Scan(&total); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "get total err")
	}

	return rs, uint32(total.Int), preOffset.String, uint32(preTotal.Int), nil
}
