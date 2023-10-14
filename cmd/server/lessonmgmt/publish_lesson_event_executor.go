package lessonmgmt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	domain_lesson "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_userID           string
	_organizationID   string
	_lessonIDs        string
	_schedulingStatus string
	_startTime        string
	_duration         int
	_sleepAfter       int
	_batchLength      int
	_limit            int
	_offset           int
)

func init() {
	bootstrap.RegisterJob("publish_lesson_event_executor", PublishLessonEventExecutor).
		StringVar(&_userID, "userID", "", "userID for RLS").
		StringVar(&_organizationID, "organizationID", "", "specific organization").
		StringVar(&_lessonIDs, "lessonIDs", "", "specific lessons").
		StringVar(&_schedulingStatus, "schedulingStatus", "", "specific lesson's scheduling status ").
		StringVar(&_startTime, "startTime", "", "filter by start time ").
		IntVar(&_sleepAfter, "sleepAfter", 1, "sleep after publish n messages").
		IntVar(&_duration, "duration", 1000, "duration between messages in millisecond").
		IntVar(&_batchLength, "batchLength", 100, "max lessons length per event").
		IntVar(&_limit, "limit", 10000, "limit records").
		IntVar(&_offset, "offset", 0, "offset for paging")
}

func PublishLessonEventExecutor(ctx context.Context, cfg configurations.Config, rsc *bootstrap.Resources) error {
	sugaredLogger := rsc.Logger().Sugar()
	sugaredLogger.Infof("Running on env: %s", cfg.Common.Environment)

	if strings.TrimSpace(_userID) == "" {
		sugaredLogger.Error("userID cannot be empty")
		return fmt.Errorf("userID cannot be empty")
	}
	if strings.TrimSpace(_organizationID) == "" {
		sugaredLogger.Error("organizationID cannot be empty")
		return fmt.Errorf("organizationID cannot be empty")
	}

	var lessonIDs []string
	if strings.TrimSpace(_lessonIDs) != "" {
		lessonIDs = strings.Split(_lessonIDs, "; ")
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			ResourcePath: _organizationID,
			UserID:       _userID,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	jsm := rsc.NATS()
	lessonDB := rsc.DBWith("lessonmgmt")

	baseArgs := []interface{}{&_organizationID}
	paramsNum := len(baseArgs)
	baseQuery := "select l.lesson_id, l.center_id, l.scheduling_status, l.start_time, l.end_time from lessons l "
	whereQuery := "where l.resource_path = $1 and l.deleted_at is null "
	if strings.TrimSpace(_startTime) != "" {
		whereQuery += fmt.Sprintf("and l.start_time::timestamptz at time zone 'Asia/Ho_Chi_Minh' >= '%s'::timestamptz at time zone 'Asia/Ho_Chi_Minh' ", _startTime)
	}
	if strings.TrimSpace(_schedulingStatus) != "" {
		whereQuery += fmt.Sprintf("and l.scheduling_status = '%s' ", _schedulingStatus)
	}
	if len(lessonIDs) > 0 {
		baseArgs = append(baseArgs, &lessonIDs)
		paramsNum++
		whereQuery += fmt.Sprintf("and l.lesson_id = ANY($%d) ", paramsNum)
	}
	baseQuery += whereQuery + fmt.Sprintf("order by l.start_time asc limit $%d offset $%d ", paramsNum+1, paramsNum+2)
	messagesPublished := 0
	sugaredLogger.Info("==========================================")
	sugaredLogger.Infof("Prepare data for offset from: %d, %d", _offset, _offset+_limit)
	args := baseArgs
	args = append(args, _limit, _offset)
	rows, err := lessonDB.Query(ctx, baseQuery, args...)
	if err != nil {
		return fmt.Errorf("query lesson error with offset %d, limit %d, status %s, organization %s: %v", _offset, _limit, _schedulingStatus, _organizationID, err)
	}
	var lessons []domain_lesson.Lesson
	for rows.Next() {
		var (
			lessonID         pgtype.Text
			centerID         pgtype.Text
			schedulingStatus pgtype.Text
			startTime        pgtype.Timestamptz
			endTime          pgtype.Timestamptz
		)
		value := []interface{}{&lessonID, &centerID, &schedulingStatus, &startTime, &endTime}
		if err = rows.Scan(value...); err != nil {
			return fmt.Errorf("rows.Scan: %v", err)
		}
		lessons = append(lessons, domain_lesson.Lesson{
			LessonID:         lessonID.String,
			LocationID:       centerID.String,
			StartTime:        startTime.Time,
			EndTime:          endTime.Time,
			SchedulingStatus: domain_lesson.LessonSchedulingStatusPublished,
		})
	}
	sugaredLogger.Infof("total lessons: %d", len(lessons))
	if len(lessons) == 0 {
		sugaredLogger.Infof("Lesson batch offset from: %d, %d is empty", _offset, _offset+_limit)
		sugaredLogger.Info("==========================================")
		return nil
	}
	tempOffset := _offset
	index := 0
	for {
		lessonEvt := []*bpb.EvtLesson_Lesson{}
		var tempLessonIDs []string
		sugaredLogger.Infof("Prepare data for batch offset from: %d to %d", tempOffset, tempOffset+_batchLength)
		for i := 0; i < _batchLength; i++ {
			if index == len(lessons) {
				break
			}
			tempLessonIDs = append(tempLessonIDs, lessons[index].LessonID)
			lessonEvt = append(lessonEvt, &bpb.EvtLesson_Lesson{
				LessonId:         lessons[index].LessonID,
				LocationId:       lessons[index].LocationID,
				StartAt:          timestamppb.New(lessons[index].StartTime),
				EndAt:            timestamppb.New(lessons[index].EndTime),
				SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(lessons[index].SchedulingStatus)]),
			})
			index++
		}

		mapLearnerIDs, err := getLearnerIDsByLessonIDs(ctx, lessonDB, tempLessonIDs)
		if err != nil {
			sugaredLogger.Errorf(fmt.Sprintf("get learner ids of lessonIDs %s failed: %s", tempLessonIDs, err))
			return fmt.Errorf("get learner ids of lesson %s failed: %w", lessons[index].LessonID, err)
		}
		mapTeacherIDs, err := getTeacherIDsByLessonIDs(ctx, lessonDB, tempLessonIDs)
		if err != nil {
			sugaredLogger.Errorf(fmt.Sprintf("get teacher ids of lessonIDs %s failed: %s", tempLessonIDs, err))
			return fmt.Errorf("get teacher ids of lesson %s failed: %w", lessons[index].LessonID, err)
		}
		for i := 0; i < len(lessonEvt); i++ {
			lessonID := lessonEvt[i].LessonId
			lessonEvt[i].TeacherIds = mapTeacherIDs[lessonID]
			lessonEvt[i].LearnerIds = mapLearnerIDs[lessonID]
		}

		msg := &bpb.EvtLesson{
			Message: &bpb.EvtLesson_CreateLessons_{
				CreateLessons: &bpb.EvtLesson_CreateLessons{
					Lessons: lessonEvt,
				},
			},
		}
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}
		sugaredLogger.Infof("Start PublishAsyncContext for subject Lesson.Created")
		msgID, err := jsm.PublishAsyncContext(ctx, "Lesson.Created", data)
		if err != nil {
			sugaredLogger.Errorf(fmt.Sprintf("Job publish lesson event failed with offset from %d to %d: msgID, err: %s, %s", tempOffset, tempOffset+len(lessonEvt), msgID, err))
			return nats.HandlePushMsgFail(ctx, fmt.Errorf("job publish lesson event failed with offset from %d to %d: msgID, err: %s, %w", tempOffset, tempOffset+len(lessonEvt), msgID, err))
		}
		sugaredLogger.Infof("Publish lesson event success for batch offset from: %d, %d", tempOffset, tempOffset+len(lessonEvt))
		sugaredLogger.Info("==========================================")
		tempOffset += len(lessonEvt)
		if tempOffset >= _offset+_limit || len(lessonEvt) < _batchLength || index == len(lessons) {
			break
		}
		messagesPublished++
		if messagesPublished == _sleepAfter {
			sugaredLogger.Infof("Sleep after: %d message", messagesPublished)
			messagesPublished = 0
			time.Sleep(time.Duration(_duration) * time.Millisecond)
		}
	}
	return nil
}

func getLearnerIDsByLessonIDs(ctx context.Context, db database.Ext, lessonIDs []string) (map[string][]string, error) {
	result := map[string][]string{}
	for _, id := range lessonIDs {
		result[id] = []string{}
	}

	baseQuery := `select lm.user_id, lm.lesson_id from lesson_members lm
				  where lm.deleted_at is null and lm.lesson_id= any($1)`

	rows, err := db.Query(ctx, baseQuery, &lessonIDs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("query lesson member err: %v", err)
	}
	for rows.Next() {
		var userID pgtype.Text
		var lessonID pgtype.Text
		if err = rows.Scan(&userID, &lessonID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		learnerIDs := result[lessonID.String]
		result[lessonID.String] = append(learnerIDs, userID.String)
	}

	return result, nil
}

func getTeacherIDsByLessonIDs(ctx context.Context, db database.Ext, lessonIDs []string) (map[string][]string, error) {
	result := map[string][]string{}
	for _, id := range lessonIDs {
		result[id] = []string{}
	}

	baseQuery := `select lt.teacher_id, lt.lesson_id from lessons_teachers lt
				  where lt.deleted_at is null and lt.lesson_id = any($1)`

	rows, err := db.Query(ctx, baseQuery, &lessonIDs)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, nil
		}
		return nil, fmt.Errorf("query lesson member err: %v", err)
	}
	for rows.Next() {
		var userID pgtype.Text
		var lessonID pgtype.Text
		if err = rows.Scan(&userID, &lessonID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		teacherIDs := result[lessonID.String]
		result[lessonID.String] = append(teacherIDs, userID.String)
	}

	return result, nil
}
