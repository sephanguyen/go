package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/constants"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vl_payloads "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type VirtualLessonRepo struct{}

func (v *VirtualLessonRepo) getVirtualLessonByID(ctx context.Context, db database.QueryExecer, id string) (*VirtualLesson, error) {
	lesson := &VirtualLesson{}
	fields, values := lesson.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM lessons
		WHERE lesson_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return lesson, nil
}

func (v *VirtualLessonRepo) GetVirtualLessonByID(ctx context.Context, db database.QueryExecer, id string) (*domain.VirtualLesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetVirtualLessonByID")
	defer span.End()
	lesson, err := v.getVirtualLessonByID(ctx, db, id)

	if err != nil {
		return nil, err
	}
	// LessonMembers
	lessonLearners := domain.LessonLearners{}

	lessonMembers, err := (&LessonMemberRepo{}).GetLessonMembersInLesson(ctx, db, lesson.LessonID.String)
	if err != nil {
		return nil, fmt.Errorf("LessonMemberRepo.GetLessonMembersInLesson: %w", err)
	}
	for _, lm := range lessonMembers {
		lessonLearners = append(lessonLearners, &domain.LessonLearner{
			LearnerID:    lm.UserID.String,
			CourseID:     lm.CourseID.String,
			AttendStatus: domain.StudentAttendStatus(lm.AttendanceStatus.String),
		})
	}
	// LessonGroup
	gr := &LessonGroupDTO{}
	if lesson.LessonGroupID.Status == pgtype.Present && lesson.CourseID.Status == pgtype.Present {
		gr, err = (&LessonGroupRepo{}).GetByIDAndCourseID(ctx, db, lesson.LessonGroupID.String, lesson.CourseID.String)
		if err != nil {
			return nil, fmt.Errorf("LessonGroupRepo.GetByIDAndCourseID: %w", err)
		}
	}
	// LessonTeacher
	lessonTeacher, err := (&LessonTeacherRepo{}).GetTeacherIDsByLessonID(ctx, db, lesson.LessonID.String)
	if err != nil {
		return nil, err
	}

	// RoomState
	roomState := &domain.OldLessonRoomState{}
	if len(lesson.RoomState.Bytes) > 0 {
		src := lesson.RoomState
		err := src.AssignTo(roomState)
		if err != nil {
			return nil, fmt.Errorf("could to unmarshal roomstate: %v", err)
		}
	}

	// Builder
	res := domain.NewVirtualLesson().
		WithLessonID(lesson.LessonID.String).
		WithName(lesson.Name.String).
		WithCenterID(lesson.CenterID.String).
		WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
		WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
		WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
		WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
		WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
		WithLearners(lessonLearners).
		WithLearnerIDs(lessonLearners.GetLearnerIDs()).
		WithTeacherIDs(lessonTeacher).
		WithMaterials(database.FromTextArray(gr.MediaIDs)).
		WithCourseID(lesson.CourseID.String).
		WithClassID(lesson.ClassID.String).
		WithSchedulerID(lesson.SchedulerID.String).
		WithLessonGroupID(lesson.LessonGroupID.String).
		WithRoomState(roomState).
		WithRoomID(lesson.RoomID.String).
		BuildDraft()

	return res, nil
}

func (v *VirtualLessonRepo) GetVirtualLessonOnlyByID(ctx context.Context, db database.QueryExecer, id string) (*domain.VirtualLesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetVirtualLessonOnlyByID")
	defer span.End()
	lesson, err := v.getVirtualLessonByID(ctx, db, id)
	if err != nil {
		return nil, err
	}

	res := domain.NewVirtualLesson().
		WithLessonID(lesson.LessonID.String).
		WithName(lesson.Name.String).
		WithCenterID(lesson.CenterID.String).
		WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
		WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
		WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
		WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
		WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
		WithCourseID(lesson.CourseID.String).
		WithClassID(lesson.ClassID.String).
		WithSchedulerID(lesson.SchedulerID.String).
		WithLessonGroupID(lesson.LessonGroupID.String).
		WithRoomID(lesson.RoomID.String).
		WithClassDoOwnerID(lesson.ClassDoOwnerID.String).
		WithClassDoLink(lesson.ClassDoLink.String).
		BuildDraft()

	return res, nil
}

func (v *VirtualLessonRepo) GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetLearnerIDsOfLesson")
	defer span.End()
	var learnerIDs []string

	query := `SELECT user_id 
			FROM lesson_members
			WHERE lesson_id = $1 AND deleted_at IS NULL`

	rows, err := db.Query(ctx, query, &lessonID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan :%v", err)
		}
		learnerIDs = append(learnerIDs, id.String)
	}
	return learnerIDs, nil
}

func (v *VirtualLessonRepo) GetTeacherIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetTeacherIDsOfLesson")
	defer span.End()

	var teacherIDs []string

	query := `SELECT teacher_id 
			FROM lessons_teachers
			WHERE lesson_id = $1 AND deleted_at IS NULL`

	rows, err := db.Query(ctx, query, &lessonID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan :%v", err)
		}
		teacherIDs = append(teacherIDs, id.String)
	}
	return teacherIDs, nil
}

func (v *VirtualLessonRepo) UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID string, state *domain.OldLessonRoomState) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.UpdateLessonRoomState")
	defer span.End()
	stateJSON := pgtype.JSONB{}
	if err := stateJSON.Set(state); err != nil {
		return fmt.Errorf("could not marshal room state to jsonb: %w", err)
	}
	query := `UPDATE lessons SET room_state = $2
			WHERE lesson_id = $1 AND deleted_at IS NULL`

	_, err := db.Exec(ctx, query, &lessonID, &stateJSON)
	if err != nil {
		return err
	}

	return nil
}

func (v *VirtualLessonRepo) GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID string, recordingState []byte) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GrantRecordingPermission")
	defer span.End()
	recordingsJSON := pgtype.JSONB{}
	if err := recordingsJSON.Set(recordingState); err != nil {
		return err
	}
	state, err := recordingsJSON.MarshalJSON()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`update lessons set room_state = coalesce(room_state || '%s', '%s')
	where lesson_id = $1 and (coalesce(room_state->'recording'->'is_recording', 'false') = 'false');`, string(state), string(state))
	_, err = db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (v *VirtualLessonRepo) StopRecording(ctx context.Context, db database.QueryExecer, lessonID string, creator string, recordingState []byte) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.StopRecording")
	defer span.End()
	recordingsJSON := pgtype.JSONB{}
	if err := recordingsJSON.Set(recordingState); err != nil {
		return err
	}
	state, err := recordingsJSON.MarshalJSON()
	if err != nil {
		return err
	}
	escapeCreator := fmt.Sprintf("\"%s\"", creator)
	query := fmt.Sprintf(`update lessons set room_state = coalesce(room_state || '%s', '%s')
	where lesson_id = $1 and (coalesce(room_state->'recording'->'creator', '%s') = '%s');`, string(state), string(state), escapeCreator, escapeCreator)
	_, err = db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (v *VirtualLessonRepo) GetVirtualLessonByLessonIDsAndCourseIDs(ctx context.Context, db database.QueryExecer, lessonIDs, courseIDs []string) ([]*domain.VirtualLesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetVirtualLessonByLessonIDsAndCourseIDs")
	defer span.End()

	lesson := &VirtualLesson{}
	fields, values := lesson.FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE deleted_at IS NULL
		AND lesson_id = ANY($1)
		AND course_id = ANY($2)`,
		strings.Join(fields, ","),
		lesson.TableName(),
	)

	rows, err := db.Query(ctx, query, &lessonIDs, &courseIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	virtualLessons := []*domain.VirtualLesson{}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		virtualLesson := domain.NewVirtualLesson().
			WithLessonID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithCenterID(lesson.CenterID.String).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithSchedulerID(lesson.SchedulerID.String).
			WithLessonGroupID(lesson.LessonGroupID.String).
			WithRoomID(lesson.RoomID.String).
			BuildDraft()
		virtualLessons = append(virtualLessons, virtualLesson)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return virtualLessons, nil
}

func (v *VirtualLessonRepo) GetVirtualLessonsByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*domain.VirtualLesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetVirtualLessonsByLessonIDs")
	defer span.End()

	lesson := &VirtualLesson{}
	fields, values := lesson.FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE deleted_at IS NULL
		AND lesson_id = ANY($1)`,
		strings.Join(fields, ","),
		lesson.TableName(),
	)

	rows, err := db.Query(ctx, query, &lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	virtualLessons := []*domain.VirtualLesson{}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		virtualLesson := domain.NewVirtualLesson().
			WithLessonID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithCenterID(lesson.CenterID.String).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithSchedulerID(lesson.SchedulerID.String).
			WithLessonGroupID(lesson.LessonGroupID.String).
			WithRoomID(lesson.RoomID.String).
			BuildDraft()
		virtualLessons = append(virtualLessons, virtualLesson)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return virtualLessons, nil
}

func (v *VirtualLessonRepo) UpdateRoomID(ctx context.Context, db database.QueryExecer, lessonID, roomID string) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.UpdateRoomID")
	defer span.End()

	query := `UPDATE lessons 
		SET updated_at = now(), room_id = $1, status = 'LESSON_STATUS_NOT_STARTED' 
		WHERE lesson_id = $2`

	cmdTag, err := db.Exec(ctx, query, &roomID, &lessonID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("cannot update lesson %s room ID", lessonID)
	}

	return nil
}

func (v *VirtualLessonRepo) EndLiveLesson(ctx context.Context, db database.QueryExecer, lessonID string, endTime time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.EndLiveLesson")
	defer span.End()

	var endAt pgtype.Timestamptz
	if err := endAt.Set(endTime); err != nil {
		return fmt.Errorf("endAt.Set: %w", err)
	}

	query := `UPDATE lessons 
		SET end_at = $1
		WHERE lesson_id = $2`

	cmdTag, err := db.Exec(ctx, query, &endAt, &lessonID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("cannot update lesson %s end time", lessonID)
	}

	return nil
}

func (v *VirtualLessonRepo) GetStreamingLearners(ctx context.Context, db database.QueryExecer, lessonID string, lockForUpdate bool) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetStreamingLearners")
	defer span.End()

	query := `SELECT learner_ids FROM lessons 
			  WHERE lesson_id = $1`

	if lockForUpdate {
		query += ` FOR UPDATE`
	}

	var learnerIDs pgtype.TextArray
	err := db.QueryRow(ctx, query, &lessonID).Scan(&learnerIDs)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	ids := database.FromTextArray(learnerIDs)

	return ids, nil
}

func (v *VirtualLessonRepo) IncreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID string, maximumLearnerStreamings int) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.IncreaseNumberOfStreaming")
	defer span.End()

	query := `UPDATE lessons 
			  SET stream_learner_counter = stream_learner_counter+1, learner_ids = array_append(learner_ids, $1) 
			  WHERE lesson_id = $2 
			  AND stream_learner_counter < $3 
			  AND NOT($1 = ANY(learner_ids))`

	cmdTag, err := db.Exec(ctx, query, &learnerID, &lessonID, &maximumLearnerStreamings)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf(constants.NoRowsUpdatedError)
	}

	return nil
}

func (v *VirtualLessonRepo) DecreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID string) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.DecreaseNumberOfStreaming")
	defer span.End()

	query := `UPDATE lessons 
			  SET stream_learner_counter = stream_learner_counter-1, learner_ids = array_remove(learner_ids, $1) 
			  WHERE lesson_id = $2 
			  AND stream_learner_counter > 0 
			  AND $1 = ANY(learner_ids)`

	cmdTag, err := db.Exec(ctx, query, &learnerID, &lessonID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf(constants.NoRowsUpdatedError)
	}

	return nil
}

func (v *VirtualLessonRepo) GetVirtualLessons(ctx context.Context, db database.QueryExecer, params *vl_payloads.GetVirtualLessonsArgs) ([]*domain.VirtualLesson, int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetVirtualLessons")
	defer span.End()

	paramsDTO := ToListVirtualLessonParamsDTO(params)

	lesson := &VirtualLesson{}
	fields, values := lesson.FieldMap()

	baseQuery := fmt.Sprintf(`SELECT DISTINCT ls.%s
		FROM %s ls `,
		strings.Join(fields, ", ls."),
		lesson.TableName(),
	)
	joinQuery := ""

	whereQuery := ` WHERE ls.lesson_type = 'LESSON_TYPE_ONLINE'
				    AND ls.deleted_at IS NULL `

	orderQuery := ` ORDER BY ls.start_time, ls.end_time ASC `

	queryArgs := []interface{}{}
	paramsNum := len(queryArgs)

	if paramsDTO.LocationIDs.Status == pgtype.Present {
		paramsNum++
		whereQuery += fmt.Sprintf(` AND ls.center_id = ANY($%d)`, paramsNum)
		queryArgs = append(queryArgs, &paramsDTO.LocationIDs)
	}

	if paramsDTO.StudentIDs.Status == pgtype.Present {
		paramsNum++
		joinQuery += ` JOIN lesson_members lm ON ls.lesson_id = lm.lesson_id `
		whereQuery += fmt.Sprintf(` AND lm.user_id = ANY($%d) AND lm.deleted_at IS NULL `, paramsNum)
		queryArgs = append(queryArgs, &paramsDTO.StudentIDs)
	}

	if paramsDTO.CourseIDs.Status == pgtype.Present {
		paramsNum++
		joinQuery += ` LEFT JOIN lessons_courses lc ON ls.lesson_id = lc.lesson_id AND lc.deleted_at IS NULL `
		whereQuery += fmt.Sprintf(` AND (lc.course_id = ANY($%d) OR ls.course_id = ANY($%d)) `, paramsNum, paramsNum)

		if params.ReplaceCourseIDColumn {
			baseQuery = strings.ReplaceAll(baseQuery, "ls.course_id",
				fmt.Sprintf("CASE WHEN ls.course_id = ANY($%d) THEN ls.course_id ELSE lc.course_id END as course_id", paramsNum),
			)
		}

		queryArgs = append(queryArgs, &paramsDTO.CourseIDs)
	}

	if paramsDTO.LessonSchedulingStatus.Status == pgtype.Present {
		paramsNum++
		whereQuery += fmt.Sprintf(` AND ls.scheduling_status = ANY($%d) `, paramsNum)
		queryArgs = append(queryArgs, &paramsDTO.LessonSchedulingStatus)
	}

	if paramsDTO.StartDate.Status == pgtype.Present && paramsDTO.EndDate.Status == pgtype.Present {
		paramsNum += 2
		whereQuery += fmt.Sprintf(` AND (ls.start_time <= $%d AND ls.end_time >= $%d) `, paramsNum-1, paramsNum)
		queryArgs = append(queryArgs, &paramsDTO.EndDate, &paramsDTO.StartDate)
	}

	totalQuery := fmt.Sprintf(`SELECT COUNT(DISTINCT ls.lesson_id)
			FROM %s ls `,
		lesson.TableName(),
	) + joinQuery + whereQuery

	var total pgtype.Int8
	if err := db.QueryRow(ctx, totalQuery, queryArgs...).Scan(&total); err != nil {
		return nil, int32(0), fmt.Errorf("get total error: %w", err)
	}

	query := baseQuery + joinQuery + whereQuery + orderQuery
	query, queryArgs = database.AddPagingQuery(query, params.Limit, params.Page, queryArgs...)
	rows, err := db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, int32(0), fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var virtualLessons []*domain.VirtualLesson
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, int32(0), fmt.Errorf("rows.Scan: %w", err)
		}

		virtualLessonBuilder := domain.NewVirtualLesson().
			WithLessonID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithCenterID(lesson.CenterID.String).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithSchedulerID(lesson.SchedulerID.String).
			WithLessonGroupID(lesson.LessonGroupID.String).
			WithRoomID(lesson.RoomID.String).
			WithTeacherID(lesson.TeacherID.String).
			WithZoomLink(lesson.ZoomLink.String)

		if lesson.EndAt.Status == pgtype.Present {
			virtualLessonBuilder = virtualLessonBuilder.WithEndAt(&lesson.EndAt.Time)
		}

		virtualLesson := virtualLessonBuilder.BuildDraft()
		virtualLessons = append(virtualLessons, virtualLesson)
	}
	if err := rows.Err(); err != nil {
		return nil, int32(0), fmt.Errorf("rows.Err: %w", err)
	}

	return virtualLessons, int32(total.Int), nil
}

func (v *VirtualLessonRepo) GetLessons(ctx context.Context, db database.QueryExecer, payload vl_payloads.GetLessonsArgs) (lessons []domain.VirtualLesson, total uint32, offsetID string, preTotal uint32, err error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonRepo.GetLessons")
	defer span.End()

	var (
		listQuery, offsetQuery, totalQuery string
		totalLessons, preTotalLessons      pgtype.Int8
		preOffset                          pgtype.Text
	)

	queryBuild := NewGetLessonsQueryBuild(payload)
	baseQuery := ` WITH filter_lesson AS (
		SELECT %s l.lesson_id, l."name", l.start_time , l.end_time, l.teaching_method, 
		l.teaching_medium, l.center_id, l.course_id, l.class_id, l.scheduling_status, l.lesson_capacity, 
		l.end_at, l.zoom_link
		FROM lessons l `
	baseTotalQuery := ` SELECT count(%s l.lesson_id) FROM lessons l `

	// add distinct to queries if needed
	baseQuery = fmt.Sprintf(baseQuery, queryBuild.Distinct)
	baseTotalQuery = fmt.Sprintf(baseTotalQuery, queryBuild.Distinct)

	// get total
	totalQuery = baseTotalQuery + queryBuild.JoinQuery + queryBuild.WhereQuery
	if queryRowErr := db.QueryRow(ctx, totalQuery, queryBuild.QueryArgs...).Scan(&totalLessons); queryRowErr != nil {
		err = fmt.Errorf("failed to get total lessons, db.QueryRow: %w", queryRowErr)
		return
	}
	total = uint32(totalLessons.Int)

	// get lesson list
	queryBuild.QueryArgs = append(queryBuild.QueryArgs, &queryBuild.OffsetLessonID, &payload.Limit)
	offsetParamsNum := queryBuild.ParamsCount + 1
	limitParamsNum := queryBuild.ParamsCount + 2

	listQuery = baseQuery + queryBuild.JoinQuery + queryBuild.WhereQuery + ")" + fmt.Sprintf(queryBuild.MainSelectQuery, offsetParamsNum, offsetParamsNum, offsetParamsNum, offsetParamsNum, limitParamsNum)

	rows, queryErr := db.Query(ctx, listQuery, queryBuild.QueryArgs...)
	if queryErr != nil {
		err = fmt.Errorf("failed to get lessons, db.Query: %w", queryErr)
		return
	}
	defer rows.Close()

	lesson := &VirtualLesson{}
	fields := []string{"lesson_id", "name", "start_time", "end_time",
		"teaching_method", "teaching_medium", "center_id", "course_id",
		"class_id", "scheduling_status", "lesson_capacity", "end_at", "zoom_link",
	}
	scanFields := database.GetScanFields(lesson, fields)

	for rows.Next() {
		if scanErr := rows.Scan(scanFields...); scanErr != nil {
			err = fmt.Errorf("rows.Scan: %w", scanErr)
			return
		}
		lessonEntity := domain.NewVirtualLesson().
			WithLessonID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithCenterID(lesson.CenterID.String).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithLessonCapacity(lesson.LessonCapacity.Int).
			WithEndAt(database.FromTimestamptz(lesson.EndAt)).
			WithZoomLink(lesson.ZoomLink.String).
			BuildDraft()

		lessons = append(lessons, *lessonEntity)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		err = fmt.Errorf("rows.Err: %w", rowsErr)
		return
	}

	// get offsets
	if queryBuild.OffsetLessonID.Status == pgtype.Present {
		offsetQuery = baseQuery + queryBuild.JoinQuery + queryBuild.WhereQuery + ")" + fmt.Sprintf(queryBuild.PreviousQuery, offsetParamsNum, offsetParamsNum, offsetParamsNum, offsetParamsNum, limitParamsNum)

		if offsetErr := db.QueryRow(ctx, offsetQuery, queryBuild.QueryArgs...).Scan(&preOffset, &preTotalLessons); offsetErr != nil && offsetErr != pgx.ErrNoRows {
			err = fmt.Errorf("failed to get offset lesson ID and count, db.QueryRow: %w", offsetErr)
			return
		}
		offsetID = preOffset.String
		preTotal = uint32(preTotalLessons.Int)
	}

	return
}
