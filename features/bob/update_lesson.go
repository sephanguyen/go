package bob

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) createCenterInDB(name string) (string, error) {
	repo := lesson_repo.MasterDataRepo{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id := idutil.ULIDNow()
	if _, err := repo.InsertCenter(ctx, s.DB, &domain.Location{
		LocationID: id,
		Name:       name,
	}); err != nil {
		return "", fmt.Errorf("could not CreateCenterInDB: %w", err)
	}

	return id, nil
}

func (s *suite) someCenters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.ManabieSchool
	ctx, err := s.signedAsAccountV2(ctx, "staff granted role school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aListOfLocationsInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.courseIds = []string{s.newID(), s.newID()}
	for _, id := range stepState.courseIds {
		if ctx, err := s.upsertLiveCourse(ctx, id, stepState.TeacherIDs[0]); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertStudentSubscription(ctx context.Context, studentIDWithCourseID ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	//HACK start at ,end_at
	startAt := time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)
	endAt := time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)

	queueFn := func(b *pgx.Batch, studentID, courseID string) {
		id := idutil.ULIDNow()
		query := `INSERT INTO lesson_student_subscriptions (student_subscription_id, subscription_id, student_id, course_id, start_at, end_at) VALUES ($1, $2, $3, $4, $5, $6)`
		b.Queue(query, id, id, studentID, courseID, startAt, endAt)
	}

	b := &pgx.Batch{}
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		queueFn(b, studentIDWithCourseID[i], studentIDWithCourseID[i+1])
	}
	result := s.DB.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return nil
}

func (s *suite) someStudentSubscriptions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := stepState.courseIds[len(stepState.courseIds)-1]
	studentIDWithCourseID := make([]string, 0, len(stepState.StudentIds)*2)
	for _, studentID := range stepState.StudentIds {
		studentIDWithCourseID = append(studentIDWithCourseID, studentID, courseID)
	}
	if err := s.insertStudentSubscription(ctx, studentIDWithCourseID...); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createsANewLessonWithPayload(
	ctx context.Context,
	req *bpb.CreateLessonRequest,
) (*bpb.CreateLessonResponse, error) {
	return bpb.NewLessonManagementServiceClient(s.Conn).CreateLesson(s.signedCtx(ctx), req)
}

func (s *suite) aLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.CurrentSchoolID = constants.ManabieSchool

	ctx, err = s.signedAsAccountV2(ctx, "staff granted role school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now().Round(time.Second)
	req := &bpb.CreateLessonRequest{
		StartTime:       timestamppb.New(now.Add(-2 * time.Hour)),
		EndTime:         timestamppb.New(now.Add(2 * time.Hour)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		CenterId:        stepState.CenterIDs[len(stepState.CenterIDs)-1],
		StudentInfoList: []*bpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*bpb.Material{},
		SavingOption: &bpb.CreateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
		SchedulingStatus: bpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}

	courseID := stepState.courseIds[len(stepState.courseIds)-1]
	for _, studentID := range stepState.StudentIds {
		req.StudentInfoList = append(req.StudentInfoList, &bpb.CreateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
			LocationId:       stepState.CenterIDs[len(stepState.CenterIDs)-1],
		})
	}

	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &bpb.Material{
			Resource: &bpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}
	res, err := s.createsANewLessonWithPayload(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.Response = res
	stepState.lessonID = res.Id
	stepState.CurrentLessonID = res.Id
	stepState.CreateLessonRequest = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdatesFieldInTheLesson(ctx context.Context, fieldName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createRequest := stepState.CreateLessonRequest
	updateRequest := &bpb.UpdateLessonRequest{
		LessonId:        stepState.lessonID,
		StartTime:       createRequest.StartTime,
		EndTime:         createRequest.EndTime,
		TeachingMedium:  createRequest.TeachingMedium,
		TeachingMethod:  createRequest.TeachingMethod,
		TeacherIds:      createRequest.TeacherIds,
		CenterId:        createRequest.CenterId,
		StudentInfoList: []*bpb.UpdateLessonRequest_StudentInfo{},
		Materials:       createRequest.Materials,
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: createRequest.SavingOption.Method,
		},
	}
	for _, studentInf := range createRequest.StudentInfoList {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentInf.StudentId,
			CourseId:         studentInf.CourseId,
			AttendanceStatus: studentInf.AttendanceStatus,
			LocationId:       createRequest.CenterId,
		})
	}

	// Change the field to something else for the update request
	var err error
	updateAllFields := fieldName == "all fields"
	if updateAllFields || fieldName == startTimeString {
		updateRequest.StartTime = timestamppb.New(updateRequest.StartTime.AsTime().Add(time.Hour))
	}
	if updateAllFields || fieldName == endTimeString {
		updateRequest.EndTime = timestamppb.New(updateRequest.EndTime.AsTime().Add(time.Hour))
	}
	if updateAllFields || fieldName == "center id" {
		ctx, err = s.someCenters(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create center: %s", err)
		}
		stepState = StepStateFromContext(ctx)
		updateRequest.CenterId = stepState.CenterIDs[len(stepState.CenterIDs)-1]
	}
	if updateAllFields || fieldName == "teacher ids" {
		updateRequest.TeacherIds = updateRequest.TeacherIds[:len(updateRequest.TeacherIds)-1]
		stepState.TeacherIDs = []string{}
		ctx, err := s.CreateTeacherAccounts(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create teachers: %s", err)
		}
		if len(stepState.TeacherIDs) < 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("required at least 1 teacher, got %d", len(stepState.TeacherIDs))
		}
		updateRequest.TeacherIds = append(updateRequest.TeacherIds, stepState.TeacherIDs...)
	}
	if updateAllFields || fieldName == "student info list" {
		updateRequest.StudentInfoList = updateRequest.StudentInfoList[:len(updateRequest.StudentInfoList)-1]
		stepState.StudentIds = []string{}
		ctx, err = s.CreateStudentAccounts(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create students: %s", err)
		}
		if len(stepState.StudentIds) < 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("required at least 1 student, got %d", len(stepState.StudentIds))
		}
		ctx, err = s.someStudentSubscriptions(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create student subscription: %s", err)
		}
		stepState = StepStateFromContext(ctx)
		for _, studentID := range stepState.StudentIds {
			updateRequest.StudentInfoList = append(updateRequest.StudentInfoList,
				&bpb.UpdateLessonRequest_StudentInfo{
					StudentId:        studentID,
					CourseId:         stepState.courseIds[len(stepState.courseIds)-1],
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_INFORMED_ABSENT,
					LocationId:       createRequest.CenterId,
				},
			)
		}
	}
	if updateAllFields || fieldName == "teaching medium" {
		if updateRequest.TeachingMedium == cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE {
			updateRequest.TeachingMedium = cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE
		} else {
			updateRequest.TeachingMedium = cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE
		}
	}
	if updateAllFields || fieldName == "teaching method" {
		if updateRequest.TeachingMethod == cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP {
			updateRequest.CourseId = "bdd-test-update-course-id-" + idutil.ULIDNow()
			updateRequest.ClassId = "bdd-test-update-class-id-" + idutil.ULIDNow()
		} else {
			updateRequest.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL
		}
	}
	if updateAllFields || fieldName == "material info" {
		stepState.RemovedMediaIDs = []string{updateRequest.Materials[len(updateRequest.Materials)-1].Resource.(*bpb.Material_MediaId).MediaId}
		updateRequest.Materials = updateRequest.Materials[:len(updateRequest.Materials)-1]
		stepState.MediaIDs = []string{}
		ctx, err := s.CreateMedias(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create media: %s", err)
		}
		if len(stepState.MediaIDs) < 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("required at least 1 media, got %d", len(stepState.MediaIDs))
		}
		for _, id := range stepState.MediaIDs {
			updateRequest.Materials = append(updateRequest.Materials,
				&bpb.Material{
					Resource: &bpb.Material_MediaId{
						MediaId: id,
					},
				},
			)
		}
	}
	stepState.Request = updateRequest
	ctx, err = s.createEditLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditLessonSubscription: %w", err)
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.Conn).UpdateLesson(s.signedCtx(ctx), updateRequest)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createEditLessonSubscription(ctx context.Context) (context.Context, error) {
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
		r := &pb.EvtLesson{}
		err := r.Unmarshal(data)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *pb.EvtLesson_UpdateLesson_:
			req := stepState.Request.(*bpb.UpdateLessonRequest)
			learnerIDs := make([]string, 0, len(req.StudentInfoList))
			for _, studentInfo := range req.StudentInfoList {
				learnerIDs = append(learnerIDs, studentInfo.StudentId)
			}
			if req.GetLessonId() == r.GetUpdateLesson().LessonId && cmp.Equal(learnerIDs, r.GetUpdateLesson().LearnerIds) {
				stepState.FoundChanForJetStream <- r.Message
				return false, nil
			}
		}
		return false, errors.New("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonUpdated, opts, handlerLessonUpdatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theLessonIsUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repositories.LessonRepo{}
	lesson, err := lessonRepo.FindByID(ctx, s.DB, database.Text(stepState.lessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	updateRequest, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", updateRequest)
	}
	if ctx, err = s.validateLessonForUpdateRequestMGMT(ctx, lesson, updateRequest); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for UpdateLiveLesson: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateLessonForUpdateRequestMGMT(ctx context.Context, e *entities.Lesson, req *bpb.UpdateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if e.LessonID.String != req.LessonId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LessonID mismatched")
	}
	if e.TeacherID.String != req.TeacherIds[0] {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher ID, got %s", req.TeacherIds[0], e.TeacherID.String)
	}
	if e.CourseID.String != req.StudentInfoList[0].CourseId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.CourseID mismatched")
	}
	if e.DeletedAt.Status != pgtype.Null {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.DeletedAt is not null")
	}
	if !e.StartTime.Time.Equal(req.StartTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for start time, got %s", req.StartTime.AsTime(), e.StartTime.Time)
	}
	if !e.EndTime.Time.Equal(req.EndTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for end time, got %s", req.EndTime.AsTime(), e.EndTime.Time)
	}
	if e.LessonGroupID.Status == pgtype.Null || len(e.LessonGroupID.String) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LessonGroupID is not null")
	}
	if e.CenterID.String != req.CenterId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected CenterId %s but got %s", req.CenterId, e.CenterID.String)
	}
	if req.TeachingMedium.String() != e.TeachingMedium.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMedium %s but got %s", req.TeachingMedium.String(), e.TeachingMedium.String)
	}
	if req.TeachingMethod.String() != e.TeachingMethod.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMethod %s but got %s", req.TeachingMethod.String(), e.TeachingMethod.String)
	}
	if e.SchedulingStatus.String != string(domain.LessonSchedulingStatusPublished) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected SchedulingStatus %s but got %s", domain.LessonSchedulingStatusPublished, e.SchedulingStatus.String)
	}

	ctx, err1 := s.validateLessonGroupForLesson(ctx, e.LessonGroupID, e.CourseID, req.Materials)
	ctx, err2 := s.validateTeachersForLesson(ctx, e.LessonID, req.TeacherIds)
	learnerIds := make([]string, 0, len(req.StudentInfoList))
	for _, studentInfo := range req.StudentInfoList {
		learnerIds = append(learnerIds, studentInfo.StudentId)
	}
	ctx, err3 := s.validateLearnersForLesson(ctx, e.LessonID, learnerIds)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2, err3)
}
