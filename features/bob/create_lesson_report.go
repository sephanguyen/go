package bob

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) UserCreateIndividualLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.CreateIndividualLessonReportRequest{
		StartTime:  timestamppb.New(time.Now()),
		EndTime:    timestamppb.New(time.Now()),
		TeacherIds: []string{"t_1", "t_2"},
		ReportDetail: []*bpb.IndividualLessonReportDetail{
			{
				StudentId: "s_1",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "attitude",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING,
						Value:          &bpb.DynamicFieldValue_StringValue{StringValue: "good"},
					},
					{
						DynamicFieldId: "home_work_score",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT,
						Value:          &bpb.DynamicFieldValue_IntValue{IntValue: 100},
					},
				},
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
				AttendanceRemark: "abc",
			},
			{
				StudentId: "s_2",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "attitude",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING,
						Value:          &bpb.DynamicFieldValue_StringValue{StringValue: "ok"},
					},
					{
						DynamicFieldId: "home_work_score",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT,
						Value:          &bpb.DynamicFieldValue_IntValue{IntValue: 80},
					},
				},
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				AttendanceRemark: "abc",
			},
		},
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonReportModifierServiceClient(s.Conn).CreateIndividualLessonReport(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserSubmitANewLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.WriteLessonReportRequest{
		LessonId: stepState.CurrentLessonID,
		Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				AttendanceRemark: "very good",
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &bpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &bpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &bpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &bpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &bpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	ctx, err := s.subscribeLessonUpdatedForUpdateSchedulingStatus(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = req
	res, err := bpb.NewLessonReportModifierServiceClient(s.Conn).SubmitLessonReport(contextWithToken(s, ctx), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
	}

	stepState.DynamicFieldValuesByUser = map[string][]*entities.PartnerDynamicFormFieldValue{
		stepState.StudentIds[0]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(5),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("title"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
				StringValue:      database.Text("monitor"),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("scores"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("comments"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
			},
			{
				FieldID:        database.Text("buddy"),
				ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
			},
			{
				FieldID:     database.Text("finished-exams"),
				ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
			},
		},
		stepState.StudentIds[1]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(15),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
		},
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserSubmitANewLessonReportWithMultiVersionFeatureNameIs(ctx context.Context, featureName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.WriteLessonReportRequest{
		LessonId:    stepState.CurrentLessonID,
		FeatureName: featureName,
		Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				AttendanceRemark: "very good",
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &bpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &bpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &bpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &bpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &bpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	stepState.Request = req
	res, err := bpb.NewLessonReportModifierServiceClient(s.Conn).SubmitLessonReport(contextWithToken(s, ctx), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
	}

	stepState.DynamicFieldValuesByUser = map[string][]*entities.PartnerDynamicFormFieldValue{
		stepState.StudentIds[0]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(5),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("title"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
				StringValue:      database.Text("monitor"),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("scores"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("comments"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
			},
			{
				FieldID:        database.Text("buddy"),
				ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
			},
			{
				FieldID:     database.Text("finished-exams"),
				ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
			},
		},
		stepState.StudentIds[1]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(15),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
		},
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) subscribeLessonUpdatedForUpdateSchedulingStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonUpdatedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		stepState.FoundChanForJetStream <- r.Message
		return true, nil
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonUpdated, opts, handlerLessonUpdatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustHaveEventFromStatusToAfterStatus(ctx context.Context, mustHaveEvent, statusString, newStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if mustHaveEvent == "yes" {
		timer := time.NewTimer(time.Minute * 1)
		allowedSchedulingStatuses := make(map[domain.LessonSchedulingStatus]bool)
		allowedSchedulingStatuses[domain.LessonSchedulingStatus(newStatus)] = false
		if len(statusString) > 0 {
			statuses := strings.Split(statusString, ",")
			for _, s := range statuses {
				allowedSchedulingStatuses[domain.LessonSchedulingStatus(s)] = false
			}
		}
		i := 0
		for {
			select {
			case data := <-stepState.FoundChanForJetStream:
				switch v := data.(type) {
				case *bpb.EvtLesson_UpdateLesson_:
					if v.UpdateLesson.LessonId == stepState.CurrentLessonID {
						status := domain.LessonSchedulingStatus(v.UpdateLesson.GetSchedulingStatusAfter().String())
						if !allowedSchedulingStatuses[status] {
							allowedSchedulingStatuses[status] = true
							i++
							if i == len(allowedSchedulingStatuses) {
								return StepStateToContext(ctx, stepState), nil
							}
						}
					}
				}
			case <-ctx.Done():
				return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) LessonSchedulingStatusUpdatesTo(ctx context.Context, schedulingStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.DB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}

	if lesson.SchedulingStatus != domain.LessonSchedulingStatus(schedulingStatus) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected SchedulingStatus %s but got %s", domain.LessonSchedulingStatus(schedulingStatus), lesson.SchedulingStatus)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdatesStatusInTheLessonIsValue(ctx context.Context, value string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	status := entities.SchedulingStatus(value)

	if status != entities.LessonSchedulingStatusCanceled && status != entities.LessonSchedulingStatusCompleted && status != entities.LessonSchedulingStatusDraft && status != entities.LessonSchedulingStatusPublished {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%s is not type of SchedulingStatus", value)
	}
	// update lesson
	lesson := &entities.Lesson{
		LessonID:         database.Text(stepState.CurrentLessonID),
		SchedulingStatus: database.Text(string(status)),
		UpdatedAt:        database.Timestamptz(time.Now()),
	}
	_, err := database.UpdateFields(ctx, lesson, db.Exec, "lesson_id", []string{
		"scheduling_status",
		"updated_at",
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("got error when update lesson status: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserSubmitToUpdateALessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.WriteLessonReportRequest{
		LessonReportId: stepState.LessonReportID,
		LessonId:       stepState.CurrentLessonID,
		Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
				AttendanceRemark: "very good",
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &bpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &bpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &bpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &bpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &bpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LATE,
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	stepState.Request = req
	res, err := bpb.NewLessonReportModifierServiceClient(s.Conn).SubmitLessonReport(contextWithToken(s, ctx), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
	}

	stepState.DynamicFieldValuesByUser = map[string][]*entities.PartnerDynamicFormFieldValue{
		stepState.StudentIds[0]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(5),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("title"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
				StringValue:      database.Text("monitor"),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("scores"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("comments"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
			},
			{
				FieldID:        database.Text("buddy"),
				ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
			},
			{
				FieldID:     database.Text("finished-exams"),
				ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
			},
		},
		stepState.StudentIds[1]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(15),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
		},
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserSaveANewDraftLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.WriteLessonReportRequest{
		LessonId: stepState.CurrentLessonID,
		Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceRemark: "very good",
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &bpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &bpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &bpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &bpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &bpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	req.Details[0].AttendanceStatus = bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY
	stepState.Request = req
	res, err := bpb.NewLessonReportModifierServiceClient(s.Conn).SaveDraftLessonReport(contextWithToken(s, ctx), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
	}

	stepState.DynamicFieldValuesByUser = map[string][]*entities.PartnerDynamicFormFieldValue{
		stepState.StudentIds[0]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(5),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("title"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
				StringValue:      database.Text("monitor"),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("scores"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("comments"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
			},
			{
				FieldID:        database.Text("buddy"),
				ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
			},
			{
				FieldID:     database.Text("finished-exams"),
				ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
			},
		},
		stepState.StudentIds[1]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(15),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
		},
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserSaveANewDraftLessonReportWithMultiVersionFeatureNameIs(ctx context.Context, featureName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.WriteLessonReportRequest{
		LessonId:    stepState.CurrentLessonID,
		FeatureName: featureName,
		Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceRemark: "very good",
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &bpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &bpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &bpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &bpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &bpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LEAVE_EARLY,
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	req.Details[0].AttendanceStatus = bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_EMPTY
	stepState.Request = req
	res, err := bpb.NewLessonReportModifierServiceClient(s.Conn).SaveDraftLessonReport(contextWithToken(s, ctx), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
	}

	stepState.DynamicFieldValuesByUser = map[string][]*entities.PartnerDynamicFormFieldValue{
		stepState.StudentIds[0]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(5),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("title"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
				StringValue:      database.Text("monitor"),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("scores"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("comments"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
			},
			{
				FieldID:        database.Text("buddy"),
				ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
			},
			{
				FieldID:     database.Text("finished-exams"),
				ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
			},
		},
		stepState.StudentIds[1]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(15),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
		},
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserSaveToUpdateADraftLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.WriteLessonReportRequest{
		LessonReportId: stepState.LessonReportID,
		LessonId:       stepState.CurrentLessonID,
		Details: []*bpb.WriteLessonReportRequest_LessonReportDetail{
			{
				StudentId:        stepState.StudentIds[0],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				AttendanceRemark: "very good",
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(5),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "title",
						Value: &bpb.DynamicFieldValue_StringValue{
							StringValue: "monitor",
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_STRING,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "scores",
						Value: &bpb.DynamicFieldValue_IntArrayValue_{
							IntArrayValue: &bpb.DynamicFieldValue_IntArrayValue{
								ArrayValue: []int32{9, 10, 8, 10},
							},
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT_ARRAY,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "comments",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_ARRAY,
						Value: &bpb.DynamicFieldValue_StringArrayValue_{
							StringArrayValue: &bpb.DynamicFieldValue_StringArrayValue{
								ArrayValue: []string{"excellent", "creative", "diligent"},
							},
						},
					},
					{
						DynamicFieldId: "buddy",
						ValueType:      bpb.ValueType_VALUE_TYPE_STRING_SET,
						Value: &bpb.DynamicFieldValue_StringSetValue_{
							StringSetValue: &bpb.DynamicFieldValue_StringSetValue{
								ArrayValue: []string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz", "Gabriel"},
							},
						},
					},
					{
						DynamicFieldId: "finished-exams",
						ValueType:      bpb.ValueType_VALUE_TYPE_INT_SET,
						Value: &bpb.DynamicFieldValue_IntSetValue_{
							IntSetValue: &bpb.DynamicFieldValue_IntSetValue{
								ArrayValue: []int32{1, 2, 3, 5, 6, 1},
							},
						},
					},
				},
			},
			{
				StudentId:        stepState.StudentIds[1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_LATE,
				AttendanceNotice: bpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: bpb.StudentAttendanceReason_PHYSICAL_CONDITION,
				AttendanceNote:   "lazy",
				FieldValues: []*bpb.DynamicFieldValue{
					{
						DynamicFieldId: "ordinal-number",
						Value: &bpb.DynamicFieldValue_IntValue{
							IntValue: int32(15),
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_INT,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
					{
						DynamicFieldId: "is-pass-lesson",
						Value: &bpb.DynamicFieldValue_BoolValue{
							BoolValue: true,
						},
						ValueType:        bpb.ValueType_VALUE_TYPE_BOOL,
						FieldRenderGuide: []byte(`{"filed": "fake guide to render this field"}`),
					},
				},
			},
		},
	}
	stepState.Request = req
	res, err := bpb.NewLessonReportModifierServiceClient(s.Conn).SaveDraftLessonReport(contextWithToken(s, ctx), req)
	stepState.Response = res
	stepState.ResponseErr = err
	if err == nil {
		stepState.LessonReportID = res.LessonReportId
	}

	stepState.DynamicFieldValuesByUser = map[string][]*entities.PartnerDynamicFormFieldValue{
		stepState.StudentIds[0]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(5),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("title"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING.String()),
				StringValue:      database.Text("monitor"),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("scores"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntArrayValue:    database.Int4Array([]int32{9, 10, 8, 10}),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("comments"),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringArrayValue: database.TextArray([]string{"excellent", "creative", "diligent"}),
			},
			{
				FieldID:        database.Text("buddy"),
				ValueType:      database.Text(bpb.ValueType_VALUE_TYPE_STRING_ARRAY.String()),
				StringSetValue: database.TextArray([]string{"Charles", "Eric", "Gabriel", "Hanna", "Beatriz"}),
			},
			{
				FieldID:     database.Text("finished-exams"),
				ValueType:   database.Text(bpb.ValueType_VALUE_TYPE_INT_ARRAY.String()),
				IntSetValue: database.Int4Array([]int32{1, 2, 3, 5, 6}),
			},
		},
		stepState.StudentIds[1]: {
			{
				FieldID:          database.Text("ordinal-number"),
				IntValue:         database.Int4(15),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_INT.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
			{
				FieldID:          database.Text("is-pass-lesson"),
				BoolValue:        database.Bool(true),
				ValueType:        database.Text(bpb.ValueType_VALUE_TYPE_BOOL.String()),
				FieldRenderGuide: database.JSONB([]byte(`{"filed": "fake guide to render this field"}`)),
			},
		},
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AFormConfigForFeature(ctx context.Context, feature string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	var featureName pgtype.Text
	switch feature {
	case "individual lesson report":
		featureName = database.Text(string(entities.FeatureNameIndividualLessonReport))
	case "group lesson report":
		featureName = database.Text(string(entities.FeatureNameGroupLessonReport))
	default:
		featureName = database.Text(feature)
	}

	_, err := (&repositories.PartnerFormConfigRepo{}).FindByPartnerAndFeatureName(ctx, db, database.Int4(stepState.CurrentSchoolID), featureName)
	if err == nil {
		return StepStateToContext(ctx, stepState), nil
	}
	if err.Error() != fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
		return ctx, fmt.Errorf("PartnerFormConfigRepo.FindByPartnerAndFeatureName: %w", err)
	}

	now := time.Now()
	config := &entities.PartnerFormConfig{
		FormConfigID: database.Text(idutil.ULIDNow()),
		PartnerID:    database.Int4(stepState.CurrentSchoolID),
		CreatedAt:    database.Timestamptz(now),
		UpdatedAt:    database.Timestamptz(now),
		DeletedAt: pgtype.Timestamptz{
			Status: pgtype.Null,
		},
		FeatureName: featureName,
		FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-0",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "attendance_status",
									"label": "attendance status",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_remark",
									"label": "attendance remark",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_notice",
									"label": "attendance notice",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_reason",
									"label": "attendance reason",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_note",
									"label": "attendance note",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "ordinal-number",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_INT",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "title",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "is-pass-lesson",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_BOOL",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "buddy",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "scores",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								},
								{
									"field_id": "comments",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_STRING_ARRAY",
									"is_required": false
								},
								{
									"field_id": "finished-exams",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_SET",
									"is_required": false
								}
							]
						}
					]
				}
			`),
	}

	fieldNames, args := config.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		config.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return ctx, fmt.Errorf("insert a form config: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) BobHaveANewLessonReport(ctx context.Context) (context.Context, error) {
	return s.checkLessonReport(ctx, entities.ReportSubmittingStatusSubmitted, false)
}

func (s *suite) BobHaveANewDraftLessonReport(ctx context.Context) (context.Context, error) {
	return s.checkLessonReport(ctx, entities.ReportSubmittingStatusSaved, false)
}

func (s *suite) BobHaveANewLessonReportWithLessonIsLocked(ctx context.Context) (context.Context, error) {
	return s.checkLessonReport(ctx, entities.ReportSubmittingStatusSubmitted, true)
}

func (s *suite) BobHaveANewDraftLessonReportWithLessonIsLocked(ctx context.Context) (context.Context, error) {
	return s.checkLessonReport(ctx, entities.ReportSubmittingStatusSaved, true)
}

func (s *suite) checkLessonReport(ctx context.Context, status entities.ReportSubmittingStatus, isLocked bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*bpb.WriteLessonReportRequest)
	// get lesson report record
	fields, _ := (&entities.LessonReport{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_reports
		WHERE lesson_report_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	report := entities.LessonReport{}
	err := database.Select(ctx, db, query, &stepState.LessonReportID).ScanOne(&report)
	if err != nil {
		return ctx, fmt.Errorf("database.Select: %w", err)
	}

	if report.ReportSubmittingStatus.String != string(status) {
		return ctx, fmt.Errorf("expected report_submitting_status %s but got %s", status, report.ReportSubmittingStatus.String)
	}
	if report.LessonID.String != req.LessonId {
		return ctx, fmt.Errorf("expected lesson_id %s but got %s", req.LessonId, report.LessonID.String)
	}
	if len(report.FormConfigID.String) == 0 {
		return ctx, fmt.Errorf("expected form_config_id but got empty")
	}
	if report.CreatedAt.Status != pgtype.Present || report.CreatedAt.Time.IsZero() {
		return ctx, fmt.Errorf("expected created_at field but got empty")
	}
	if report.UpdatedAt.Status != pgtype.Present || report.UpdatedAt.Time.IsZero() {
		return ctx, fmt.Errorf("expected updated_at field but got empty")
	}

	// get lesson details
	details, err := (&repositories.LessonReportDetailRepo{}).GetByLessonReportID(ctx, db, database.Text(stepState.LessonReportID))
	if err != nil {
		return ctx, err
	}
	detailIDs := make([]string, 0, len(details))
	for i := range req.Details {
		detailIDs = append(detailIDs, details[i].LessonReportDetailID.String)
	}
	if err = s.checkLessonReportDetails(req.Details, details); err != nil {
		return ctx, err
	}

	// get detail's field values
	values, err := s.getPartnerFormDynamicFieldValues(ctx, detailIDs)
	if err != nil {
		return ctx, err
	}
	valuesByDetailID := make(map[string]entities.PartnerDynamicFormFieldValues)
	for _, value := range values {
		v := valuesByDetailID[value.LessonReportDetailID.String]
		v = append(v, value)
		valuesByDetailID[value.LessonReportDetailID.String] = v
	}
	valuesByUserID := make(map[string]entities.PartnerDynamicFormFieldValues)
	for _, detail := range details {
		valuesByUserID[detail.StudentID.String] = valuesByDetailID[detail.LessonReportDetailID.String]
	}
	if err = s.checkLessonReportDetailFieldValues(valuesByUserID, stepState.DynamicFieldValuesByUser); err != nil {
		return ctx, err
	}

	// get lesson member
	membersByID, err := s.getLessonMembersByLesson(ctx, report.LessonID.String)
	if err != nil {
		return ctx, err
	}

	err = s.checkLessonReportLessonMember(req.Details, membersByID)

	if isLocked {
		if err == nil {
			return ctx, fmt.Errorf("AttendanceRemark and AttendanceStatus of lesson member don't should to change when lesson is locked")
		}
	} else {
		if err != nil {
			return ctx, err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkLessonReportDetails(req []*bpb.WriteLessonReportRequest_LessonReportDetail, actual entities.LessonReportDetails) error {
	for i, expected := range req {
		if actual[i].StudentID.String != expected.StudentId {
			return fmt.Errorf("expected student_id %s but got %s", expected.StudentId, actual[i].StudentID.String)
		}
		if actual[i].CreatedAt.Status != pgtype.Present || actual[i].CreatedAt.Time.IsZero() {
			return fmt.Errorf("expected created_at field but got empty")
		}
		if actual[i].UpdatedAt.Status != pgtype.Present || actual[i].UpdatedAt.Time.IsZero() {
			return fmt.Errorf("expected updated_at field but got empty")
		}
	}
	return nil
}

func (s *suite) checkLessonReportDetailFieldValues(valuesByUserID map[string]entities.PartnerDynamicFormFieldValues, dynamicFieldValuesByUser map[string][]*entities.PartnerDynamicFormFieldValue) error {
	for userID, actualValues := range valuesByUserID {
		expected, ok := dynamicFieldValuesByUser[userID]
		if !ok {
			return fmt.Errorf("no fields values of lesson report details which have user id %s", userID)
		}

		expectedByFieldID := make(map[string]*entities.PartnerDynamicFormFieldValue)
		for i := range expected {
			expectedByFieldID[expected[i].FieldID.String] = expected[i]
		}

		if len(expectedByFieldID) != len(actualValues) {
			return fmt.Errorf("expected %d values but got %d", len(expectedByFieldID), len(actualValues))
		}

		for _, actualValue := range actualValues {
			v, ok := expectedByFieldID[actualValue.FieldID.String]
			if !ok {
				return fmt.Errorf("expected field %s's values but got null", actualValue.FieldID.String)
			}

			if string(v.FieldRenderGuide.Bytes) != string(actualValue.FieldRenderGuide.Bytes) {
				return fmt.Errorf("expected field render %s guide but got %s", string(v.FieldRenderGuide.Bytes), string(actualValue.FieldRenderGuide.Bytes))
			}
			if v.IntValue.Status == pgtype.Present && (v.IntValue.Int != actualValue.IntValue.Int) {
				return fmt.Errorf("expected int value %d but got %d", v.IntValue.Int, actualValue.IntValue.Int)
			}
			if v.StringValue.Status == pgtype.Present && (v.StringValue.String != actualValue.StringValue.String) {
				return fmt.Errorf("expected string value %s but got %s", v.StringValue.String, actualValue.StringValue.String)
			}
			if v.BoolValue.Status == pgtype.Present && (v.BoolValue.Bool != actualValue.BoolValue.Bool) {
				return fmt.Errorf("expected boolean value %v but got %v", v.BoolValue.Bool, actualValue.BoolValue.Bool)
			}
			if v.StringArrayValue.Status == pgtype.Present {
				expectedArr := database.FromTextArray(v.StringArrayValue)
				actualArr := database.FromTextArray(actualValue.StringArrayValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected string array %v but got %v", expectedArr, actualArr)
					}
				}
			}
			if v.IntArrayValue.Status == pgtype.Present {
				expectedArr := database.FromInt4Array(v.IntArrayValue)
				actualArr := database.FromInt4Array(actualValue.IntArrayValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected int array %v but got %v", expectedArr, actualArr)
					}
				}
			}
			if v.StringSetValue.Status == pgtype.Present {
				expectedArr := database.FromTextArray(v.StringSetValue)
				actualArr := database.FromTextArray(actualValue.StringSetValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected string set %v but got %v", expectedArr, actualArr)
					}
				}
			}
			if v.IntSetValue.Status == pgtype.Present {
				expectedArr := database.FromInt4Array(v.IntSetValue)
				actualArr := database.FromInt4Array(actualValue.IntSetValue)
				for i := range expectedArr {
					if expectedArr[i] != actualArr[i] {
						return fmt.Errorf("expected int set %v but got %v", expectedArr, actualArr)
					}
				}
			}
		}
	}
	return nil
}

func (s *suite) checkLessonReportLessonMember(details []*bpb.WriteLessonReportRequest_LessonReportDetail, actualList map[string]*entities.LessonMember) error {
	for _, member := range details {
		actual := actualList[member.StudentId]
		if actual.AttendanceStatus.String != member.AttendanceStatus.String() {
			return fmt.Errorf("expected attendance status %s but got %s", member.AttendanceStatus.String(), actual.AttendanceStatus.String)
		}
		if actual.AttendanceRemark.String != member.AttendanceRemark {
			return fmt.Errorf("expected attendance remark %s but got %s", member.AttendanceRemark, actual.AttendanceRemark.String)
		}
		if actual.AttendanceNotice.String != member.AttendanceNotice.String() {
			return fmt.Errorf("expected attendance notice %s but got %s", member.AttendanceNotice.String(), actual.AttendanceNotice.String)
		}
		if actual.AttendanceReason.String != member.AttendanceReason.String() {
			return fmt.Errorf("expected attendance reason %s but got %s", member.AttendanceReason.String(), actual.AttendanceReason.String)
		}
		if actual.AttendanceNote.String != member.AttendanceNote {
			return fmt.Errorf("expected attendance note %s but got %s", member.AttendanceNote, actual.AttendanceNote.String)
		}
	}

	return nil
}

func (s *suite) getPartnerFormDynamicFieldValues(ctx context.Context, lessonReportDetailIDs []string) (entities.PartnerDynamicFormFieldValues, error) {
	fields, _ := (&entities.PartnerDynamicFormFieldValue{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM partner_dynamic_form_field_values
		WHERE lesson_report_detail_id = ANY($1) AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	values := entities.PartnerDynamicFormFieldValues{}
	err := database.Select(ctx, db, query, &lessonReportDetailIDs).ScanAll(&values)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return values, nil
}

func (s *suite) getLessonMembersByLesson(ctx context.Context, lessonID string) (map[string]*entities.LessonMember, error) {
	fields, _ := (&entities.LessonMember{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_members
		WHERE lesson_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	members := entities.LessonMembers{}
	err := database.Select(ctx, db, query, &lessonID).ScanAll(&members)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	membersByID := make(map[string]*entities.LessonMember)
	for i, member := range members {
		membersByID[member.UserID.String] = members[i]
	}

	return membersByID, nil
}

func (s *suite) lessonIsLockedByTimesheet(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	_, err := s.DB.Exec(ctx, `UPDATE lessons SET is_locked = $1, updated_at = $2 WHERE lesson_id = $3`, true, database.Timestamptz(time.Now()), stepState.CurrentLessonID)

	return StepStateToContext(ctx, stepState), err
}
