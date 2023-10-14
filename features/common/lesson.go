package common

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	classdo_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/gogo/protobuf/types"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) AListOfLessonsAreExistedInDBOfWithStartTimeAndEndTime(ctx context.Context, lesson_opt, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	courseID := "course-live-teacher-1"
	courseID2 := "course-live-teacher-1"
	if lesson_opt == "above teacher and belong to multy course" {
		courseID = "course-live-teacher-5"
		courseID2 = "course-live-teacher-6"
	}
	if lesson_opt == "above teacher and belong to single course" {
		courseID = "course-live-teacher-4"
		courseID2 = "course-live-teacher-4"
	}
	courseID += stepState.Random
	courseID2 += stepState.Random
	classID := idutil.ULIDNow()

	if lesson_opt == "JPREP whitelist" {
		courseID = "JPREP_COURSE_000000162"
		courseID2 = "JPREP_COURSE_000000218"
	}
	// create lesson group
	lg := &bob_entities.LessonGroup{}
	database.AllNullEntity(lg)
	lg.MediaIDs = database.TextArray(stepState.MediaIDs)
	lg.CourseID = database.Text(courseID)
	err = (&repositories.LessonGroupRepo{}).Create(ctx, s.BobDB, lg)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
	}
	for i := 0; i < 20; i++ {
		status := cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()
		if i > 0 {
			status = cpb.LessonSchedulingStatus_name[int32(rand.Intn(4))]
		}
		lesson := &bob_entities.Lesson{}
		database.AllNullEntity(lesson)

		err = multierr.Combine(
			lesson.LessonID.Set(s.newID()),
			lesson.CourseID.Set(courseID),
			lesson.TeacherID.Set(stepState.CurrentTeacherID),
			lesson.CreatedAt.Set(timeutil.Now()),
			lesson.UpdatedAt.Set(timeutil.Now()),
			lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
			lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
			lesson.StreamLearnerCounter.Set(database.Int4(0)),
			lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
			lesson.StartTime.Set(startDate),
			lesson.EndTime.Set(endDate),
			lesson.LessonGroupID.Set(lg.LessonGroupID),
			lesson.ClassID.Set(classID),
			lesson.SchedulingStatus.Set(database.Text(status)),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if err := lesson.Normalize(); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
		}

		cmdTag, err := database.Insert(ctx, lesson, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
		}

		if lesson_opt == "above teacher and belong to multy course" {
			sql := `INSERT INTO lessons_courses
				(lesson_id, course_id, created_at)
				VALUES ($1, $2, $4), ($1, $3, $4)`
			_, err = s.BobDB.Exec(ctx, sql, lesson.LessonID, database.Text(courseID),
				database.Text(courseID2),
				database.Timestamptz(time.Now()))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_course, err = %v", err)
			}
		}

		stepState.LessonIDs = append(stepState.LessonIDs, lesson.LessonID.String)

		if err := (&repositories.LessonRepo{}).UpsertLessonMembers(ctx, s.BobDB, lesson.LessonID, database.TextArray(stepState.StudentIds)); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("UpsertLessonMembers")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserCreateALessonWithRequest(ctx context.Context, req *bpb.CreateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := bpb.NewLessonManagementServiceClient(s.BobConn).CreateLesson(contextWithToken(s, ctx), req)
	stepState.ResponseErr = err
	stepState.Response = res
	if err == nil {
		stepState.CurrentLessonID = res.Id
	}
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserCreateALessonWithRequestInLessonmgmt(ctx context.Context, req *lpb.CreateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).CreateLesson(contextWithToken(s, ctx), req)
	stepState.ResponseErr = err
	stepState.Response = res
	if err == nil {
		stepState.CurrentLessonID = res.Id
		stepState.LessonIDs = append(stepState.LessonIDs, res.Id)
	}
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserCreateALessonRequestWithMissingFields(ctx context.Context, teachingMedium cpb.LessonTeachingMedium, missingFields ...string) *bpb.CreateLessonRequest {
	stepState := StepStateFromContext(ctx)
	now := time.Now().Round(time.Second)
	locationID := stepState.CenterIDs[len(stepState.CenterIDs)-1]
	req := &bpb.CreateLessonRequest{
		StartTime:       timestamppb.New(now.Add(-2 * time.Hour)),
		EndTime:         timestamppb.New(now.Add(2 * time.Hour)),
		TeachingMedium:  teachingMedium,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		CenterId:        locationID,
		StudentInfoList: []*bpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*bpb.Material{},
		SavingOption: &bpb.CreateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
		SchedulingStatus: bpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	// For lesson group
	switch stepState.CurrentTeachingMethod {
	case "group":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP
			req.ClassId = stepState.CurrentClassId
			req.CourseId = stepState.CurrentCourseID
		}
	case "individual":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL
		}
	}

	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentID := stepState.StudentIDWithCourseID[i]
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		addedStudentIDs[studentID] = true
		req.StudentInfoList = append(req.StudentInfoList, &bpb.CreateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
			LocationId:       locationID,
		})
	}

	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &bpb.Material{
			Resource: &bpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}

	for _, missingField := range missingFields {
		switch strings.ToLower(missingField) {
		case "start time", "starttime", "start_time":
			req.StartTime = nil
		case "end time", "endtime", "end_time":
			req.EndTime = nil
		case "teaching medium", "teachingmedium", "teaching_medium":
			req.TeachingMedium = 0
		case "teaching method", "teachingmethod", "teaching_method":
			req.TeachingMethod = 0
		case "center id", "centerid", "center_id":
			req.CenterId = ""
		case "teacher ids", "teacherids", "teacher_ids":
			req.TeacherIds = nil
		case "student info list", "studentinfolist", "student_info_list":
			req.StudentInfoList = nil
		case "material info", "materials":
			req.Materials = nil
		}
	}
	return req
}

func (s *suite) UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx context.Context, teachingMedium cpb.LessonTeachingMedium, missingFields ...string) *lpb.CreateLessonRequest {
	stepState := StepStateFromContext(ctx)
	now := time.Now().Round(time.Second)
	locationID := stepState.CenterIDs[len(stepState.CenterIDs)-1]
	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(now.Add(9 * time.Hour)),
		EndTime:         timestamppb.New(now.Add(10 * time.Hour)),
		TeachingMedium:  teachingMedium,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      locationID,
		StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*lpb.Material{},
		SavingOption: &lpb.CreateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
		SchedulingStatus: lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	if teachingMedium == cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ZOOM {
		req.ZoomInfo = &lpb.ZoomInfo{
			ZoomLink:         stepState.ZoomLink,
			ZoomAccountOwner: stepState.ZoomAccount.ID,
		}
	}
	if teachingMedium == cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_CLASS_DO {
		req.ClassDoInfo = &lpb.ClassDoInfo{
			ClassDoOwnerId: stepState.ClassDoAccount.ClassDoID,
			ClassDoLink:    stepState.ClassDoLink,
			ClassDoRoomId:  stepState.ClassDoRoomID,
		}
	}
	// For lesson group
	switch stepState.CurrentTeachingMethod {
	case "group":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP
			req.ClassId = stepState.CurrentClassId
			req.CourseId = stepState.CurrentCourseID
		}
	case "individual":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL
		}
	}

	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentID := stepState.StudentIDWithCourseID[i]
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		addedStudentIDs[studentID] = true
		req.StudentInfoList = append(req.StudentInfoList, &lpb.CreateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			LocationId:       locationID,
			AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
			AttendanceNotice: lpb.StudentAttendanceNotice_ON_THE_DAY,
			AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
		})
	}
	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &lpb.Material{
			Resource: &lpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}

	for _, missingField := range missingFields {
		switch strings.ToLower(missingField) {
		case "start time", "starttime", "start_time":
			req.StartTime = nil
		case "end time", "endtime", "end_time":
			req.EndTime = nil
		case "teaching medium", "teachingmedium", "teaching_medium":
			req.TeachingMedium = 0
		case "teaching method", "teachingmethod", "teaching_method":
			req.TeachingMethod = 0
		case "center id", "centerid", "center_id":
			req.LocationId = ""
		case "teacher ids", "teacherids", "teacher_ids", "teachers":
			req.TeacherIds = nil
		case "student info list", "studentinfolist", "student_info_list":
			req.StudentInfoList = nil
		case "material info", "materials":
			req.Materials = nil
		case "students":
			req.StudentInfoList = nil
			req.Materials = nil
		}
	}
	return req
}

func (s *suite) UserCreateALessonRequestInDateWithMissingFieldsInLessonmgmt(
	ctx context.Context,
	teachingMedium cpb.LessonTeachingMedium,
	date time.Time,
	missingFields ...string,
) *lpb.CreateLessonRequest {
	stepState := StepStateFromContext(ctx)
	date = date.Round(time.Second)

	minStart := 7  // 7h
	maxStart := 21 // 21h
	randTimeStart := (time.Duration)(rand.Intn(maxStart-minStart) + minStart)
	startTime := date.Add(randTimeStart * time.Hour)

	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(startTime),
		EndTime:         timestamppb.New(startTime.Add(2 * time.Hour)),
		TeachingMedium:  teachingMedium,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      stepState.CenterIDs[len(stepState.CenterIDs)-1],
		StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*lpb.Material{},
		SavingOption: &lpb.CreateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
		SchedulingStatus: lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	// For lesson group
	switch stepState.CurrentTeachingMethod {
	case "group":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP
			req.ClassId = stepState.CurrentClassId
			req.CourseId = stepState.CurrentCourseID
		}
	case "individual":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL
		}
	}

	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentID := stepState.StudentIDWithCourseID[i]
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		addedStudentIDs[studentID] = true
		req.StudentInfoList = append(req.StudentInfoList, &lpb.CreateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
			AttendanceNotice: lpb.StudentAttendanceNotice_ON_THE_DAY,
			AttendanceReason: lpb.StudentAttendanceReason_PHYSICAL_CONDITION,
		})
	}
	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &lpb.Material{
			Resource: &lpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}

	for _, missingField := range missingFields {
		switch strings.ToLower(missingField) {
		case "start time", "starttime", "start_time":
			req.StartTime = nil
		case "end time", "endtime", "end_time":
			req.EndTime = nil
		case "teaching medium", "teachingmedium", "teaching_medium":
			req.TeachingMedium = 0
		case "teaching method", "teachingmethod", "teaching_method":
			req.TeachingMethod = 0
		case "center id", "centerid", "center_id":
			req.LocationId = ""
		case "teacher ids", "teacherids", "teacher_ids", "teachers":
			req.TeacherIds = nil
		case "student info list", "studentinfolist", "student_info_list":
			req.StudentInfoList = nil
		case "material info", "materials":
			req.Materials = nil
		case "students":
			req.StudentInfoList = nil
			req.Materials = nil
		}
	}
	return req
}

func (s *suite) UserCreateALessonWithMissingFields(ctx context.Context, missingFields ...string) (context.Context, error) {
	req := s.UserCreateALessonRequestWithMissingFields(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE, missingFields...)

	return s.UserCreateALessonWithRequest(ctx, req)
}
func (s *suite) UserCreateSomeLessonsWithMissingFieldsAndSchedulingStatusInLessonmgmt(ctx context.Context, status string, missingFields ...string) (context.Context, error) {
	var err error
	for i := 0; i < 5; i++ {
		req := s.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE, missingFields...)
		switch status {
		case "published":
			req.SchedulingStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
		case "completed":
			req.SchedulingStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_COMPLETED
		case "canceled":
			req.SchedulingStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_CANCELED
		case "draft":
			req.SchedulingStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
		}
		ctx, err = s.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, err
}

func (s *suite) UserCreateALessonWithMissingFieldsInLessonmgmt(ctx context.Context, missingFields ...string) (context.Context, error) {
	req := s.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE, missingFields...)

	return s.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *suite) UserCreateALessonZoomWithMissingFieldsInLessonmgmt(ctx context.Context, missingFields ...string) (context.Context, error) {
	req := s.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ZOOM, missingFields...)

	return s.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *suite) UserCreateALessonClassDoWithMissingFieldsInLessonmgmt(ctx context.Context, missingFields ...string) (context.Context, error) {
	req := s.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_CLASS_DO, missingFields...)

	return s.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *suite) UserCreateALessonInDateWithMissingFieldsInLessonmgmt(ctx context.Context, date time.Time, missingFields ...string) (context.Context, error) {
	req := s.UserCreateALessonRequestInDateWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE, date, missingFields...)

	return s.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *suite) UserCreateALiveLessonWithMissingFields(ctx context.Context, missingFields ...string) (context.Context, error) {
	req := s.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE, missingFields...)

	return s.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *suite) UserUpdateALessonWithRequest(ctx context.Context, req *bpb.UpdateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := bpb.NewLessonManagementServiceClient(s.BobConn).UpdateLesson(contextWithToken(s, ctx), req)
	stepState.ResponseErr = err
	stepState.Response = res
	stepState.CurrentLessonID = req.LessonId
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserUpdateALessonWithRequestInLessonmgmt(ctx context.Context, req *lpb.UpdateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).UpdateLesson(contextWithToken(s, ctx), req)
	stepState.ResponseErr = err
	stepState.Response = res
	stepState.CurrentLessonID = req.LessonId
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserUpdateALessonSchedulingStatusWithRequest(ctx context.Context, req *lpb.UpdateLessonSchedulingStatusRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).UpdateLessonSchedulingStatus(contextWithToken(s, ctx), req)
	stepState.ResponseErr = err
	stepState.Response = res
	stepState.CurrentLessonID = req.LessonId
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateEditLessonSubscription(ctx context.Context) (context.Context, error) {
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
		switch r.Message.(type) {
		case *bpb.EvtLesson_UpdateLesson_:
			req := stepState.Request.(*bpb.UpdateLessonRequest)
			learnerIDs := make([]string, 0, len(req.StudentInfoList))
			for _, studentInfo := range req.StudentInfoList {
				learnerIDs = append(learnerIDs, studentInfo.StudentId)
			}
			if req.GetLessonId() == r.GetUpdateLesson().LessonId && cmp.Equal(learnerIDs, r.GetUpdateLesson().LearnerIds) {
				stepState.FoundChanForJetStream <- r.Message
				return true, nil
			}
		}
		return false, fmt.Errorf("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonUpdated, opts, handlerLessonUpdatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)

	return s.getLessonByID(ctx)
}

func (s *suite) CreateEditStatusLessonSubscription(ctx context.Context) (context.Context, error) {
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
		switch r.Message.(type) {
		case *bpb.EvtLesson_UpdateLesson_:
			req := stepState.Request.(*lpb.UpdateLessonSchedulingStatusRequest)
			if req.GetLessonId() == r.GetUpdateLesson().LessonId {
				stepState.FoundChanForJetStream <- r.Message
				return true, nil
			}
		}
		return false, fmt.Errorf("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonUpdated, opts, handlerLessonUpdatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)

	return s.getLessonByID(ctx)
}

func (s *suite) CreateEditLessonSubscriptionLessonmgmt(ctx context.Context) (context.Context, error) {
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
		switch r.Message.(type) {
		case *bpb.EvtLesson_UpdateLesson_:
			req := stepState.Request.(*lpb.UpdateLessonRequest)
			learnerIDs := make([]string, 0, len(req.StudentInfoList))
			for _, studentInfo := range req.StudentInfoList {
				learnerIDs = append(learnerIDs, studentInfo.StudentId)
			}
			if req.GetLessonId() == r.GetUpdateLesson().LessonId && cmp.Equal(learnerIDs, r.GetUpdateLesson().LearnerIds) {
				stepState.FoundChanForJetStream <- r.Message
				return false, nil
			}
		}
		return false, fmt.Errorf("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonUpdated, opts, handlerLessonUpdatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)

	return s.getLessonByID(ctx)
}

func (s *suite) getLessonByID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	s.Lesson = lesson

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getCountDeletedLessonByLessonID(ctx context.Context, lessonIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count pgtype.Int8
	query := `SELECT count(*) FROM lessons
				WHERE lesson_id = ANY($1) 
				AND deleted_at IS NOT NULL `

	if err := s.LessonmgmtDB.QueryRow(ctx, query, &lessonIDs).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query count deleted lesson: %s", err)
	}

	stepState.DeletedLessonCount = int(count.Int)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserUpdatesCurrentLessonWithMissingFields(ctx context.Context, missingFields ...string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createdRequest, ok := stepState.Request.(*bpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	updateRequest := &bpb.UpdateLessonRequest{
		LessonId:        stepState.CurrentLessonID,
		StartTime:       createdRequest.StartTime,
		EndTime:         createdRequest.EndTime,
		TeachingMedium:  createdRequest.TeachingMedium,
		TeachingMethod:  createdRequest.TeachingMethod,
		TeacherIds:      createdRequest.TeacherIds,
		CenterId:        createdRequest.CenterId,
		StudentInfoList: make([]*bpb.UpdateLessonRequest_StudentInfo, 0, len(createdRequest.StudentInfoList)),
		Materials:       createdRequest.Materials,
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: createdRequest.SavingOption.Method,
		},
	}
	for _, studentInf := range createdRequest.StudentInfoList {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentInf.StudentId,
			CourseId:         studentInf.CourseId,
			AttendanceStatus: studentInf.AttendanceStatus,
		})
	}

	for _, missingField := range missingFields {
		switch strings.ToLower(missingField) {
		case "start time", "starttime", "start_time":
			updateRequest.StartTime = nil
		case "end time", "endtime", "end_time":
			updateRequest.EndTime = nil
		case "teaching medium", "teachingmedium", "teaching_medium":
			updateRequest.TeachingMedium = 0
		case "teaching method", "teachingmethod", "teaching_method":
			updateRequest.TeachingMethod = 0
		case "center id", "centerid", "center_id":
			updateRequest.CenterId = ""
		case "teacher ids", "teacherids", "teacher_ids":
			updateRequest.TeacherIds = nil
		case "student info list", "studentinfolist", "student_info_list":
			updateRequest.StudentInfoList = nil
		case "material info", "materials":
			updateRequest.Materials = nil
		}
	}

	stepState.Request = updateRequest
	ctx, err := s.CreateEditLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditLessonSubscription: %w", err)
	}

	return s.UserUpdateALessonWithRequest(ctx, updateRequest)
}

func (s *suite) UserUpdatesCurrentLessonWithMissingFieldsInLessonmgmt(ctx context.Context, missingFields ...string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.CreateLessonRequest, got %T", createdRequest)
	}
	updateRequest := &lpb.UpdateLessonRequest{
		LessonId:        stepState.CurrentLessonID,
		StartTime:       createdRequest.StartTime,
		EndTime:         createdRequest.EndTime,
		TeachingMedium:  createdRequest.TeachingMedium,
		TeachingMethod:  createdRequest.TeachingMethod,
		TeacherIds:      createdRequest.TeacherIds,
		LocationId:      createdRequest.LocationId,
		StudentInfoList: make([]*lpb.UpdateLessonRequest_StudentInfo, 0, len(createdRequest.StudentInfoList)),
		Materials:       createdRequest.Materials,
		SavingOption: &lpb.UpdateLessonRequest_SavingOption{
			Method: createdRequest.SavingOption.Method,
		},
	}
	for _, studentInf := range createdRequest.StudentInfoList {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &lpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentInf.StudentId,
			CourseId:         studentInf.CourseId,
			AttendanceStatus: studentInf.AttendanceStatus,
			AttendanceNote:   studentInf.AttendanceNote,
			AttendanceNotice: studentInf.AttendanceNotice,
			AttendanceReason: studentInf.AttendanceReason,
		})
	}

	for _, missingField := range missingFields {
		switch strings.ToLower(missingField) {
		case "start time", "starttime", "start_time":
			updateRequest.StartTime = nil
		case "end time", "endtime", "end_time":
			updateRequest.EndTime = nil
		case "teaching medium", "teachingmedium", "teaching_medium":
			updateRequest.TeachingMedium = 0
		case "teaching method", "teachingmethod", "teaching_method":
			updateRequest.TeachingMethod = 0
		case "center id", "centerid", "center_id":
			updateRequest.LocationId = ""
		case "teacher ids", "teacherids", "teacher_ids":
			updateRequest.TeacherIds = nil
		case "student info list", "studentinfolist", "student_info_list":
			updateRequest.StudentInfoList = nil
		case "material info", "materials":
			updateRequest.Materials = nil
		}
	}

	stepState.Request = updateRequest
	ctx, err := s.CreateEditLessonSubscriptionLessonmgmt(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditLessonSubscription: %w", err)
	}

	return s.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
}

func (s *Suite) RetrieveLiveLessonByCourseWithStartTimeAndEndTime(ctx context.Context, courseID, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var from, to *types.Timestamp
	if len(startTimeString) != 0 {
		data, err := time.Parse(time.RFC3339, startTimeString)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		from, err = types.TimestampProto(data)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	if len(endTimeString) != 0 {
		data, err := time.Parse(time.RFC3339, startTimeString)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		to, err = types.TimestampProto(data)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	req := &pb.RetrieveLiveLessonRequest{
		CourseIds: []string{courseID},
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From: from,
		To:   to,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Connections.BobConn).
		RetrieveLiveLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ListStudentsInLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	paging := &cpb.Paging{
		Limit: uint32(100),
	}
	var students []*cpb.BasicProfile
	idx := 0
	for {
		if idx > 100 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected paging: infinite paging")
		}
		idx++

		resp, err := bpb.NewClassReaderServiceClient(s.BobConn).
			ListStudentsByLesson(contextWithToken(s, ctx), &bpb.ListStudentsByLessonRequest{
				LessonId: stepState.CurrentLessonID,
				Paging:   paging,
			})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(resp.Students) < int(paging.Limit) {
			break
		}
		if len(resp.Students) > int(paging.Limit) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total students: got: %d, want: %d", len(resp.Students), paging.Limit)
		}

		students = append(students, resp.Students...)

		paging = resp.NextPage
	}

	stepState.Response = students
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GetLessonMedias(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := bpb.NewCourseReaderServiceClient(s.BobConn).
		ListLessonMedias(contextWithToken(s, ctx), &bpb.ListLessonMediasRequest{
			LessonId: stepState.CurrentLessonID,
			Paging: &cpb.Paging{
				Limit: 1,
			},
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Medias = s.BobMediaFromV1(resp.Items)

	nextPage := resp.NextPage
	for len(resp.Items) != 0 {
		resp, err = bpb.NewCourseReaderServiceClient(s.BobConn).
			ListLessonMedias(contextWithToken(s, ctx), &bpb.ListLessonMediasRequest{
				LessonId: stepState.CurrentLessonID,
				Paging:   nextPage,
			})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		nextPage = resp.NextPage
		// stepState.MediaItems = append(stepState.MediaItems, resp.Items...)
		stepState.Medias = append(stepState.Medias, s.BobMediaFromV1(resp.Items)...)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) BobMediaFromV1(medias []*bpb.Media) []*pb.Media {
	res := make([]*pb.Media, 0, len(medias))
	for _, media := range medias {
		res = append(res, &pb.Media{
			MediaId:  media.MediaId,
			Name:     media.Name,
			Resource: media.Resource,
			Type:     pb.MediaType(media.Type),
		})
	}

	return res
}

func (s *suite) JoinLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.BobConn).
		JoinLesson(contextWithToken(s, ctx), &bpb.JoinLessonRequest{
			LessonId: stepState.CurrentLessonID,
		})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateClassrooms(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	locationIDs := stepState.LocationIDs
	classroomIDs := make([]string, 0, len(locationIDs))

	for _, locationID := range locationIDs {
		newID := idutil.ULIDNow()
		classroom := &repo.Classroom{}
		database.AllNullEntity(classroom)

		err := multierr.Combine(
			classroom.ClassroomID.Set(newID),
			classroom.Name.Set(fmt.Sprintf("classroom-%s", newID)),
			classroom.RoomArea.Set(fmt.Sprintf("room-area-%s", newID)),
			classroom.IsArchived.Set(false),
			classroom.LocationID.Set(locationID),
			classroom.CreatedAt.Set(timeutil.Now()),
			classroom.UpdatedAt.Set(timeutil.Now()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to set classroom values: %w", err)
		}

		cmdTag, err := database.Insert(ctx, classroom, s.LessonmgmtDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert classroom")
		}

		classroomIDs = append(classroomIDs, classroom.ClassroomID.String)
	}
	stepState.ClassroomIDs = classroomIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) HasAClassDoAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	classDoAccount := &classdo_domain.ClassDoAccount{
		ClassDoID:     id,
		ClassDoEmail:  fmt.Sprintf("email-%s@email.com", id),
		ClassDoAPIKey: fmt.Sprintf("APIKEY%s", id),
	}

	cmdTag, err := s.LessonmgmtDB.Exec(ctx, `INSERT INTO public."classdo_account"
					(classdo_id, classdo_email, classdo_api_key, created_at, updated_at)
					VALUES ($1, $2, $3, now(), now())
					ON CONFLICT ON CONSTRAINT pk__classdo_account 
					DO UPDATE SET classdo_email = $2, classdo_api_key = $3`,
		database.Text(classDoAccount.ClassDoID),
		database.Text(classDoAccount.ClassDoEmail),
		database.Text(classDoAccount.ClassDoAPIKey),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in upsert classdo account: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no ClassDo account was created")
	}

	stepState.ClassDoAccount = classDoAccount

	return StepStateToContext(ctx, stepState), nil
}
