package lessonmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const ManabieOrgLocationType = "01FR4M51XJY9E77GSN4QZ1Q9M1"
const LocalTimezone = "Asia/Ho_Chi_Minh"
const ImportLessonLocationID = "5-19"
const ImportLessonPartnerInternalID = "partner-internal-id-5-19"

func (s *Suite) avalidLessonRequestPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request := fmt.Sprintf(`partner_internal_id,start_date_time,end_date_time,teaching_method
		%s,2023-01-02 05:40:00,2023-01-02 06:45:00,1
		%s,2023-01-03 06:40:00,2023-01-03 07:45:00,1
		%s,2023-01-06 07:00:00,2023-01-06 08:00:00,1`,
		ImportLessonPartnerInternalID,
		ImportLessonPartnerInternalID,
		ImportLessonPartnerInternalID)

	stepState.Request = &lpb.ImportLessonRequest{
		Payload: []byte(request),
	}
	stepState.ImportLessonPartnerInternalIDs = []string{ImportLessonPartnerInternalID}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) avalidLessonRequestPayloadV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentIDWithCourseIDs := []string{}
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentCourseIDStr := fmt.Sprintf("%s/%s", stepState.StudentIDWithCourseID[i], stepState.StudentIDWithCourseID[i+1])
		studentIDWithCourseIDs = append(studentIDWithCourseIDs, studentCourseIDStr)
	}

	request := fmt.Sprintf(`partner_internal_id,start_date_time,end_date_time,teaching_method,teaching_medium,teacher_ids,student_course_ids
		%s,2023-01-02 05:40:00,2023-01-02 06:45:00,1,1,%s,
		%s,2023-01-03 06:40:00,2023-01-03 07:45:00,1,1,,
		%s,2023-01-04 07:00:00,2023-01-04 08:00:00,2,,%s,%s
		%s,2023-01-06 07:00:00,2023-01-06 08:00:00,1,2,,%s`,
		ImportLessonPartnerInternalID, strings.Join(stepState.TeacherIDs, "_"),
		ImportLessonPartnerInternalID,
		ImportLessonPartnerInternalID, strings.Join(stepState.TeacherIDs, "_"), strings.Join(studentIDWithCourseIDs, "_"),
		ImportLessonPartnerInternalID, strings.Join(studentIDWithCourseIDs, "_"))

	stepState.Request = &lpb.ImportLessonRequest{
		Payload: []byte(request),
	}
	// grant this student with course to the location using to import lesson
	stmt := `INSERT INTO lesson_student_subscription_access_path (student_subscription_id,location_id)
			SELECT student_subscription_id, $1 FROM lesson_student_subscriptions
			WHERE (student_id || '/' || course_id) = ANY($2)
			ON CONFLICT DO NOTHING`
	_, err := s.BobDB.Exec(ctx, stmt, ImportLessonLocationID, studentIDWithCourseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_student_subscription_access_path with student_course_ids:%s, location_id:%s, err:%v", studentIDWithCourseIDs, ImportLessonLocationID, err)
	}

	stepState.ImportLessonPartnerInternalIDs = []string{ImportLessonPartnerInternalID}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) anInvalidLessonRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentIDWithCourseIDs := []string{}
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentCourseIDStr := fmt.Sprintf("%s/%s", stepState.StudentIDWithCourseID[i], stepState.StudentIDWithCourseID[i+1])
		studentIDWithCourseIDs = append(studentIDWithCourseIDs, studentCourseIDStr)
	}

	switch invalidFormat {
	case "no data":
		stepState.Request = &lpb.ImportLessonRequest{}
	case "header only":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_internal_id,start_date_time,end_date_time,teaching_method`),
		}
	case "number of column is not equal 4":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_internal_id,teaching_method
			pid_1,1
			pid_2,2
			pid_3,1`),
		}
	case "mismatched number of fields in header and content":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_internal_id,start_date_time,end_date_time,teaching_method
			pid_1,1
			pid_2,2
			pid_3,1`),
		}
	case "wrong id column name in header":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_internal_id,teaching_method,start_date_time,end_date_time
			partner-internal-id-5-19,1,2023-01-02 05:40:00,2023-01-02 06:45:00
			partner-internal-id-5-19,1,2023-01-02 05:40:00,2023-01-02 06:45:00
			partner-internal-id-5-19,2,2023-01-02 05:40:00,2023-01-02 06:45:00`),
		}
	case "wrong name column name in header":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,1
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,1
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,2`),
		}
	case "mismatched valid and invalid rows":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,1
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,1
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,3`),
		}
	case "invalid partner_internal_id":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			pid_1,2023-01-02 05:40:00,2023-01-02 06:45:00,1
			pid_2,2023-01-02 05:40:00,2023-01-02 06:45:00,2
			pid_7,2023-01-02 05:40:00,2023-01-02 06:45:00,1`),
		}
	case "invalid date time format":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			partner-internal-id-5-19,2023-01-02T05:40:00Z,2023-01-0206:45:00,1
			partner-internal-id-5-19,2023-01-02T05:40:00Z,2023-01-0206:45:00,1`),
		}
	case "start time > end time":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			partner-internal-id-5-19,2023-01-02 09:40:00,2023-01-02 06:45:00,1
			partner-internal-id-5-19,2023-01-02 10:40:00,2023-01-02 06:45:00,2`),
		}
	case "start date <> end date":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-05 06:45:00,1
			partner-internal-id-5-19,2023-01-02 02:40:00,2023-01-10 06:45:00,2`),
		}
	case "invalid teaching_method":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,3
			partner-internal-id-5-19,2023-01-02 02:40:00,2023-01-02 06:45:00,4`),
		}
	case "missing value in madatory column":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_id,starting_date_time,end_date_time,teaching_method
			,2023-01-02 05:40:00,2023-01-02 06:45:00,1
			partner-internal-id-5-19,,2023-01-02 06:45:00,1
			partner-internal-id-5-19,2023-01-02 05:40:00,,1
			partner-internal-id-5-19,2023-01-02 02:40:00,2023-01-02 06:45:00,`),
		}
	case "number of column is not equal 6":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_internal_id,teaching_method
			pid_1,1
			pid_2,2
			pid_3,1`),
		}
	case "invalid teaching medium":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_internal_id,start_date_time,end_date_time,teaching_method,teaching_medium,teacher_ids,student_course_ids
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,1,1,%s,
			partner-internal-id-5-19,2023-01-02 02:40:00,2023-01-02 06:45:00,1,3,,`),
		}
	case "invalid teacher ids":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(`partner_internal_id,start_date_time,end_date_time,teaching_method,teaching_medium,teacher_ids,student_course_ids
			partner-internal-id-5-19,2023-01-02 05:40:00,2023-01-02 06:45:00,1,1,teacher-id1_teacher-id2,
			partner-internal-id-5-19,2023-01-02 02:40:00,2023-01-02 06:45:00,1,2,,`),
		}
	case "invalid student course ids":
		stepState.Request = &lpb.ImportLessonRequest{
			Payload: []byte(fmt.Sprintf(`partner_internal_id,start_date_time,end_date_time,teaching_method,teaching_medium,teacher_ids,student_course_ids
			%s,2023-01-02 05:40:00,2023-01-02 06:45:00,1,1,%s,
			%s,2023-01-03 06:40:00,2023-01-03 07:45:00,1,1,,
			%s,2023-01-04 07:00:00,2023-01-04 08:00:00,2,,%s,%s
			%s,2023-01-06 07:00:00,2023-01-06 08:00:00,1,2,%s,`,
				ImportLessonPartnerInternalID, strings.Join(stepState.TeacherIDs, "_"),
				ImportLessonPartnerInternalID,
				ImportLessonPartnerInternalID, strings.Join(stepState.TeacherIDs, "_"), strings.Join(studentIDWithCourseIDs, "_"),
				ImportLessonPartnerInternalID, strings.Join(stepState.TeacherIDs, "_"))),
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) importingLessons(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	ctx, err := s.createLessonSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createLessonSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = lpb.NewLessonExecutorServiceClient(s.LessonMgmtConn).
		ImportLesson(contextWithToken(s, ctx), stepState.Request.(*lpb.ImportLessonRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theValidLessonsLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*lpb.ImportLessonRequest)
	res := stepState.Response.(*lpb.ImportLessonResponse)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	if len(res.Errors) > 0 {
		return ctx, fmt.Errorf("response errors: %s", res.Errors)
	}

	lessons, err := s.selectNewLessons(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	masterDataRepo := repo.MasterDataRepo{}
	centerByPartnerID, err := masterDataRepo.GetLowestLocationsByPartnerInternalIDs(ctx, s.BobDBTrace, stepState.ImportLessonPartnerInternalIDs)
	if err != nil {
		return ctx, fmt.Errorf("can not get centers by partner_internal_id: %s", err)
	}

	timezone := LoadLocalLocation()
	for _, row := range stepState.ValidCsvRows {
		rowValues := strings.Split(row, ",")
		pID := strings.ToLower(rowValues[0])
		startDateTime, err1 := timeutil.ParsingTimeFromYYYYMMDDStr(rowValues[1], req.TimeZone)
		endDateTime, err2 := timeutil.ParsingTimeFromYYYYMMDDStr(rowValues[2], req.TimeZone)
		err = multierr.Combine(err1, err2)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf(`could not parsing time: %T`, err.Error()))
		}
		endTime := time.Date(startDateTime.Year(), startDateTime.Month(), startDateTime.Day(),
			endDateTime.Hour(), endDateTime.Minute(), endDateTime.Second(), endDateTime.Nanosecond(),
			endDateTime.Location())

		teachingMethod := rowValues[3]
		centerID := centerByPartnerID[pID].LocationID
		found := false
		for _, ls := range lessons {
			lessonStartTime := ls.StartTime.In(timezone)
			lessonEndTime := ls.EndTime.In(timezone)
			if ls.LocationID == centerID && teachingMethod == domain.MapKeyLessonTeachingMethod[domain.LessonTeachingMethod(string(ls.TeachingMethod))] && lessonStartTime.Equal(startDateTime) && lessonEndTime.Equal(endTime) {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to import valid csv row: %s", rowValues)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createLessonSubscribe(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	eventHandler := func(ctx context.Context, data []byte) (bool, error) {
		lessonEvent := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, lessonEvent)
		if err != nil {
			return false, err
		}
		switch msg := lessonEvent.Message.(type) {
		case *bpb.EvtLesson_CreateLessons_:
			stepState.FoundChanForJetStream <- msg
			return false, nil
		default:
			return true, fmt.Errorf("wrong message type")
		}
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonCreated, opts, eventHandler)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) selectNewLessons(ctx context.Context) ([]*domain.Lesson, error) {
	var allEntities []*domain.Lesson
	stmt := `SELECT lesson_id, center_id, teaching_method, start_time, end_time
		FROM lessons
		where deleted_at is null
		order by updated_at desc limit 50`
	rows, err := s.BobDBTrace.Query(ctx, stmt)
	if err != nil {
		return nil, errors.Wrap(err, "query new lessons")
	}
	defer rows.Close()
	for rows.Next() {
		e := &domain.Lesson{}
		err := rows.Scan(
			&e.LessonID,
			&e.LocationID,
			&e.TeachingMethod,
			&e.StartTime,
			&e.EndTime,
		)
		if err != nil {
			return nil, errors.WithMessage(err, "rows.Scan new lessons")
		}
		allEntities = append(allEntities, e)
	}
	return allEntities, nil
}

func (s *Suite) theInvalidLessonMustReturnedError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*lpb.ImportLessonResponse)
	if len(resp.Errors) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid file is not returned error list in response")
	}
	return StepStateToContext(ctx, stepState), nil
}
