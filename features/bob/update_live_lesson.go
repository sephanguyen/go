package bob

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	yasuorepo "github.com/manabie-com/backend/internal/yasuo/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) anExistingLiveLesson(ctx context.Context) (context.Context, error) {
	// we don't care about the status when using this function, simply choose one

	return s.CommonSuite.UserCreateALiveLessonWithMissingFields(ctx)
}
func (s *suite) anExistingLiveLessonWithStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, createRequest, err := s.initCreateLiveLessonRequest(ctx, status)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("fail to initialize CreateLiveLessonRequest: %s", err)
	}

	resp, err := bpb.NewLessonModifierServiceClient(s.Conn).CreateLiveLesson(s.signedCtx(ctx), createRequest)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create live lesson: %s", err)
	}

	stepState.lessonID = resp.Id
	stepState.CurrentLessonID = resp.Id
	stepState.Request = createRequest
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) initCreateLiveLessonRequest(ctx context.Context, lessonStatus string) (context.Context, *bpb.CreateLiveLessonRequest, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.CreateTeacherAccounts(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create some teachers: %s", err)
	}

	if len(stepState.TeacherIDs) < 2 {
		return ctx, nil, fmt.Errorf("required at least 2 teacher IDs, got %d", len(stepState.TeacherIDs))
	}

	ctx, err = s.CreateStudentAccounts(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("createStudentAccounts: %s", err)
	}

	if len(stepState.StudentIds) < 2 {
		return ctx, nil, fmt.Errorf("required at least 2 student IDs, got %d", len(stepState.StudentIds))
	}

	ctx, err = s.CreateLiveCourse(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create some courses: %s", err)
	}

	if len(stepState.courseIds) < 2 {
		return ctx, nil, fmt.Errorf("required at least 2 course IDs, got %d", len(stepState.courseIds))
	}

	ctx, err = s.CreateMedias(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("failed to create some media: %s", err)
	}

	if len(stepState.MediaIDs) < 1 {
		return ctx, nil, fmt.Errorf("required at least 1 media ID, got %d", len(stepState.MediaIDs))
	}
	material := make([]*bpb.Material, 0, len(stepState.MediaIDs)+1)
	material = append(material, &bpb.Material{Resource: &bpb.Material_BrightcoveVideo_{BrightcoveVideo: &bpb.Material_BrightcoveVideo{Name: "video 1", Url: "https://brightcove.com/account/2/video?videoId=abc123"}}})
	for _, id := range stepState.MediaIDs {
		material = append(material, &bpb.Material{Resource: &bpb.Material_MediaId{MediaId: id}})
	}
	var startTime, endTime *timestamppb.Timestamp
	const oneDayDuration = time.Second * 60 * 60 * 24
	now := time.Now().Round(time.Second).UTC()
	switch lessonStatus {
	case "not started":
		startTime = timestamppb.New(now.Add(oneDayDuration))
		endTime = timestamppb.New(now.Add(oneDayDuration * 2))
	case "in progress":
		startTime = timestamppb.New(now.Add(-oneDayDuration))
		endTime = timestamppb.New(now.Add(oneDayDuration))
	case "completed":
		startTime = timestamppb.New(now.Add(-oneDayDuration * 2))
		endTime = timestamppb.New(now.Add(-oneDayDuration))
	default:
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf(`invalid input live lesson status "%s"`, lessonStatus)
	}
	return StepStateToContext(ctx, stepState), &bpb.CreateLiveLessonRequest{Name: "lesson_name" + s.newID(), StartTime: startTime, EndTime: endTime, TeacherIds: stepState.TeacherIDs, CourseIds: stepState.courseIds, LearnerIds: stepState.StudentIds, Materials: material}, nil
}

func (s *suite) createLessonUpdatedSubscribe(ctx context.Context) (context.Context, error) {
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
			req := stepState.Request.(*bpb.UpdateLiveLessonRequest)
			if req.GetId() == r.GetUpdateLesson().LessonId && cmp.Equal(req.LearnerIds, r.GetUpdateLesson().LearnerIds) {
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

func (s *suite) userUpdatesFieldInTheLiveLesson(ctx context.Context, fieldName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, createRequest, err := s.extractCreateLessonRequestFromPreviousStep(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get CreateLiveLessonRequest: %s", err)
	}
	updateRequest := s.initUpdateLiveLessonRequest(createRequest, stepState.lessonID)

	// Change the field to something else for the update request
	// For TeacherIds, CourseIds, and LearnerIds: remove one, keep one, and add some new elements
	updateAllFields := fieldName == "all fields"
	if updateAllFields || fieldName == lessonName {
		updateRequest.Name = "changed_lesson_name" + s.newID()
	}
	if updateAllFields || fieldName == startTimeString {
		updateRequest.StartTime = timestamppb.New(updateRequest.StartTime.AsTime().Add(time.Hour))
	}
	if updateAllFields || fieldName == endTimeString {
		updateRequest.EndTime = timestamppb.New(updateRequest.EndTime.AsTime().Add(time.Hour))
	}
	if updateAllFields || fieldName == teacherIds {
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
	if updateAllFields || fieldName == courseIds {
		stepState.RemovedCourseIDs = []string{updateRequest.CourseIds[len(updateRequest.CourseIds)-1]}
		updateRequest.CourseIds = updateRequest.CourseIds[:len(updateRequest.CourseIds)-1]
		stepState.courseIds = []string{}
		ctx, err := s.CreateLiveCourse(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create courses: %s", err)
		}
		if len(stepState.courseIds) < 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("required at least 1 course, got %d", len(stepState.courseIds))
		}
		updateRequest.CourseIds = append(updateRequest.CourseIds, stepState.courseIds...)
	}
	if updateAllFields || fieldName == "learner ids" {
		updateRequest.LearnerIds = updateRequest.LearnerIds[:len(updateRequest.LearnerIds)-1]
		stepState.StudentIds = []string{}
		ctx, err := s.CreateStudentAccounts(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create students: %s", err)
		}
		if len(stepState.StudentIds) < 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("required at least 1 student, got %d", len(stepState.StudentIds))
		}
		updateRequest.LearnerIds = append(updateRequest.LearnerIds, stepState.StudentIds...)
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
	ctx, err = s.createLessonUpdatedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createLessonUpdatedSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).UpdateLiveLesson(s.signedCtx(ctx), updateRequest)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) extractCreateLessonRequestFromPreviousStep(ctx context.Context) (context.Context, *bpb.CreateLiveLessonRequest, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Request == nil {
		return ctx, nil, errors.New("missing CreateLiveLesson request, please call anExistingLiveLesson first")
	}
	createRequest, ok := stepState.Request.(*bpb.CreateLiveLessonRequest)
	if !ok {
		return ctx, nil, fmt.Errorf("expect request of type *bpb.CreateLiveLessonRequest, got type %T", stepState.Request)
	}
	return ctx, createRequest, nil
}
func (s *suite) initUpdateLiveLessonRequest(createRequest *bpb.CreateLiveLessonRequest, lessonID string) *bpb.UpdateLiveLessonRequest {
	return &bpb.UpdateLiveLessonRequest{Id: lessonID, Name: createRequest.Name, StartTime: createRequest.StartTime, EndTime: createRequest.EndTime, TeacherIds: createRequest.TeacherIds, CourseIds: createRequest.CourseIds, LearnerIds: createRequest.LearnerIds, Materials: createRequest.Materials}
}
func (s *suite) theLiveLessonIsUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonRepo := repositories.LessonRepo{}
	lesson, err := lessonRepo.FindByID(ctx, s.DB, database.Text(stepState.lessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}

	updateRequest, ok := stepState.Request.(*bpb.UpdateLiveLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", updateRequest)
	}
	if ctx, err := s.validateLessonForUpdateRequest(ctx, lesson, updateRequest); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for UpdateLiveLesson: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) validateLessonForUpdateRequest(ctx context.Context, e *entities.Lesson, req *bpb.UpdateLiveLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if e.LessonID.String != req.Id {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LessonID mismatched")
	}
	if e.Name.String != req.Name {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.Name mismatched")
	}
	if e.TeacherID.String != req.TeacherIds[0] {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher ID, got %s", req.TeacherIds[0], e.TeacherID.String)
	}
	if e.CourseID.String != req.CourseIds[0] {
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
	if e.LessonType.String != string(entities.LessonTypeOnline) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LessonType mismatched")
	}

	ctx, err1 := s.validateLessonGroupForLesson(ctx, e.LessonGroupID, e.CourseID, req.Materials)
	ctx, err2 := s.validateTeachersForLesson(ctx, e.LessonID, req.TeacherIds)
	ctx, err3 := s.validateLearnersForLesson(ctx, e.LessonID, req.LearnerIds)
	ctx, err4 := s.validateCoursesAndPresetStudyPlansForLesson(ctx, e.LessonID, req.CourseIds, e.StartTime.Time, e.EndTime.Time)
	ctx, err5 := s.validatePresetStudyPlansWeeklyAndTopicsForLesson(ctx, e)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2, err3, err4, err5)

}
func (s *suite) validateLessonGroupForLesson(ctx context.Context, lessonGroupID, courseID pgtype.Text, expectedMaterials []*bpb.Material) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lgRepo := repositories.LessonGroupRepo{}
	medias, err := lgRepo.GetMedias(ctx, s.DB, lessonGroupID, courseID, database.Int4(100), pgtype.Text{Status: pgtype.Null})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query media for lesson group: %s", err)
	}

	for _, material := range expectedMaterials {
		switch resource := material.Resource.(type) {
		case *bpb.Material_BrightcoveVideo_:
			if ctx, err := s.validateBrightcoveVideoInMedias(ctx, resource.BrightcoveVideo, medias); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		case *bpb.Material_MediaId:
			if ctx, err := s.validateMediaIDInMedias(ctx, resource, medias); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("unhandled material resource type: %T", resource)
		}
	}

	// media can be removed from lesson group but must not be deleted from database
	if len(stepState.RemovedMediaIDs) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	query := `
		SELECT count(1)
		FROM media
		WHERE media_id = ANY($1::text[])
			AND deleted_at IS NULL`
	var count int
	err = s.DB.QueryRow(ctx, query, database.TextArray(stepState.RemovedMediaIDs)).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query media: %s", err)
	}

	if count != len(stepState.RemovedMediaIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected to find %d media in database, got %d", len(stepState.RemovedMediaIDs), count)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) validateBrightcoveVideoInMedias(ctx context.Context, bcVideo *bpb.Material_BrightcoveVideo, medias entities.Medias) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	videoID, err := golibs.GetBrightcoveVideoIDFromURL(bcVideo.Url)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get Brightcove video ID: %s", err)
	}
	for _, media := range medias {
		if media.Type.String == string(entities.MediaTypeVideo) &&
			media.Resource.String == videoID &&
			media.Name.String == bcVideo.Name {
			return StepStateToContext(ctx, stepState), nil // found
		}
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("medias does not contain Brightcovde video with videoID %s", videoID)
}
func (s *suite) validateMediaIDInMedias(ctx context.Context, m *bpb.Material_MediaId, medias entities.Medias) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, media := range medias {
		if media.MediaID.String == m.MediaId {
			return StepStateToContext(ctx, stepState), nil // found
		}
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("medias does not contain media with ID %s", m.MediaId)
}
func (s *suite) validateTeachersForLesson(ctx context.Context, lessonID pgtype.Text, expectedTeacherIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonRepo := repositories.LessonRepo{}
	teacherIDs, err := lessonRepo.GetTeacherIDsOfLesson(ctx, s.DB, lessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query teacher IDs for lesson: %s", err)
	}
	if !stringutil.SliceElementsMatch(database.FromTextArray(teacherIDs), expectedTeacherIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher IDs, got %s", expectedTeacherIDs, database.FromTextArray(teacherIDs))
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) validateLearnersForLesson(ctx context.Context, lessonID pgtype.Text, expectedLearnerIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonRepo := repositories.LessonRepo{}
	studentIDs, err := lessonRepo.GetLearnerIDsOfLesson(ctx, s.DB, lessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query student IDs for lesson: %s", err)
	}
	if !stringutil.SliceElementsMatch(database.FromTextArray(studentIDs), expectedLearnerIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LearnerIDs mismatched")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) validateCoursesAndPresetStudyPlansForLesson(ctx context.Context, lessonID pgtype.Text, expectedCourseIDs []string, expectedStartTime, expectedEndTime time.Time) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonRepo := repositories.LessonRepo{}
	courseIDs, err := lessonRepo.GetCourseIDsOfLesson(ctx, s.DB, lessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query course IDs for lesson: %s", err)
	}
	if !stringutil.SliceElementsMatch(database.FromTextArray(courseIDs), expectedCourseIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.CourseIDs mismatched")
	}

	// Also query preset_study_plans for courses
	pspYasuoRepo := yasuorepo.PresetStudyPlanRepo{}
	pspByCourseID, err := pspYasuoRepo.FindByCourseIDs(ctx, s.DB, courseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query preset study plans for courses: %s", err)
	}

	// Check each course has a valid preset_study_plan and valid start/end date.
	// Since each course only belongs to one lesson in this test, its start/end date
	// are equal to the lesson's start/end time.
	courseRepo := repositories.CourseRepo{}
	courseByID, err := courseRepo.FindByIDs(ctx, s.DB, courseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query courses: %s", err)
	}
	for courseID, course := range courseByID {
		if len(course.PresetStudyPlanID.String) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("course %s does not have a preset_study_plan_id", courseID.String)
		}
		psp, ok := pspByCourseID[courseID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to find preset study plan %s for course %s", course.PresetStudyPlanID.String, courseID.String)
		}
		if ctx, err := s.validatePresetStudyPlanForCourses(ctx, psp, course); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid preset study plan for course %s: %s", courseID.String, err)
		}

		if !course.StartDate.Time.Equal(expectedStartTime) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecte start time %v, got %v", expectedStartTime, course.StartDate.Time)
		}
		if !course.EndDate.Time.Equal(expectedEndTime) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecte end time %v, got %v", expectedEndTime, course.EndDate.Time)
		}
	}

	stepState.CourseByID = courseByID // reuse course entities in later steps

	// Check that removed courses should still have its preset study plan
	if len(stepState.RemovedCourseIDs) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	removedCourseIDs := database.TextArray(stepState.RemovedCourseIDs)
	pspIDs, err := courseRepo.GetPresetStudyPlanIDsByCourseIDs(ctx, s.DB, removedCourseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("courseRepo.GetPresetStudyPlanIDsByCourseIDs: %s", err)
	}
	if len(pspIDs) != len(stepState.RemovedCourseIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected to have %d preset study plans, found %d", len(stepState.RemovedCourseIDs), len(pspIDs))
	}

	// Since we removed the courses from the only lesson they belong to
	// Their start/end date should be null
	removedCourses, err := courseRepo.FindByIDs(ctx, s.DB, removedCourseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("courseRepo.FindByIDs: %s", err)
	}
	for _, course := range removedCourses {
		if course.StartDate.Status != pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected NULL start_date for removed course, got %+v", course.StartDate)
		}
		if course.EndDate.Status != pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected NULL end_date for removed course, got %+v", course.EndDate)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) validatePresetStudyPlanForCourses(ctx context.Context, psp *entities.PresetStudyPlan, course *entities.Course) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if psp.Name != course.Name {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected name %+v, got %+v", course.Name, psp.Name)
	}
	if psp.Country != course.Country {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected course %+v, got %+v", course.Country, psp.Country)
	}
	if psp.Grade != course.Grade {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected grade %+v, got %+v", course.Grade, psp.Grade)
	}
	if psp.Country != course.Country {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected course %+v, got %+v", course.Country, psp.Country)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) validatePresetStudyPlansWeeklyAndTopicsForLesson(ctx context.Context, lesson *entities.Lesson) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	pspwByPspID, err := s.queryPresetStudyPlansWeeklyForLesson(ctx, s.DB, lesson.LessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query preset study plan weekly for lesson: %s", err)
	}
	topicByPspwID, err := s.queryTopicByPresetStudyPlanWeekly(ctx, s.DB, extractIDsForPSPWFromMap(pspwByPspID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query topics for preset study plans weekly: %s", err)
	}

	// Reuse stepState.CourseByID from validateCoursesForLesson()
	if len(stepState.CourseByID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stepState.CourseByID is missing")
	}
	for courseID, course := range stepState.CourseByID {
		pspID := course.PresetStudyPlanID
		pspw, ok := pspwByPspID[pspID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("could not find preset study plan weekly for course %s", courseID.String)
		}
		if ctx, err := s.validatePresetStudyPlanWeeklyForLesson(ctx, pspw, lesson); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid preset study plan %s: %s", pspw.ID.String, err)
		}
		topic, ok := topicByPspwID[pspw.ID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("could not find topic for preset study plan %s", pspw.ID.String)
		}
		ctx, err := s.validateTopic(ctx,
			topic,
			lesson.Name, // name of all topics must be updated to match lesson
			course.Country,
			course.Grade,
			course.Subject,
			course.SchoolID,
			database.Text(string(entities.TopicTypeLiveLesson)),
			database.Text(string(entities.TopicStatusPublished)),
			database.Int2(1))

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid topic %s: %s", topic.ID.String, err)
		}
	}

	// Removing courses from lesson should delete its related preset study plan weekly and topic
	removedCourseIDs := database.TextArray(stepState.RemovedCourseIDs)
	deletedPspws, err := s.findDeletedPresetStudyPlansWeekly(ctx, s.DB, lesson.LessonID, removedCourseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get deleted preset study plan weekly: %s", err)
	}
	if len(deletedPspws) != len(stepState.RemovedCourseIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d deleted preset study plans weekly, got %d", len(stepState.RemovedCourseIDs), len(deletedPspws))
	}
	deletedTopics, err := s.findDeletedTopicByPresetStudyPlanWeeklyID(ctx, s.DB, extractIDsForPSPW(deletedPspws))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get deleted topics: %s", err)
	}
	if len(deletedTopics) != len(stepState.RemovedCourseIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d deleted topics, got %d", len(stepState.RemovedCourseIDs), len(deletedTopics))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) queryPresetStudyPlansWeeklyForLesson(ctx context.Context, db database.Ext, lessonIDs pgtype.Text) (map[pgtype.Text]*entities.PresetStudyPlanWeekly, error) {
	fields := database.GetFieldNames(&entities.PresetStudyPlanWeekly{})
	query := fmt.Sprintf(`
		SELECT %s
		FROM preset_study_plans_weekly
		WHERE lesson_id = $1
			AND deleted_at IS NULL`, strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %s", err)
	}
	defer rows.Close()
	pspwByPSPID := make(map[pgtype.Text]*entities.PresetStudyPlanWeekly)
	for rows.Next() {
		pspw := new(entities.PresetStudyPlanWeekly)
		if err := rows.Scan(database.GetScanFields(pspw, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %s", err)
		}
		pspwByPSPID[pspw.PresetStudyPlanID] = pspw
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %s", err)
	}
	return pspwByPSPID, nil
}
func (s *suite) validatePresetStudyPlanWeeklyForLesson(ctx context.Context, pspw *entities.PresetStudyPlanWeekly, lesson *entities.Lesson) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if pspw.LessonID != lesson.LessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for start_time, got %+v", pspw.LessonID, lesson.LessonID)
	}
	if !pspw.StartDate.Time.Equal(lesson.StartTime.Time) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for start_date, got %s", lesson.StartTime.Time, pspw.StartDate.Time)
	}
	if !pspw.EndDate.Time.Equal(lesson.EndTime.Time) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for end_date, got %s", lesson.EndTime.Time, pspw.EndDate.Time)
	}
	if len(pspw.TopicID.String) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing topic ID")
	}
	return StepStateToContext(ctx, stepState), nil
}
func extractIDsForPSPWFromMap(pspwInMap map[pgtype.Text]*entities.PresetStudyPlanWeekly) pgtype.TextArray {
	ids := make([]string, 0, len(pspwInMap))
	for _, pspw := range pspwInMap {
		ids = append(ids, pspw.ID.String)
	}
	return database.TextArray(ids)
}
func extractIDsForPSPW(pspws []*entities.PresetStudyPlanWeekly) pgtype.TextArray {
	ids := make([]string, 0, len(pspws))
	for _, pspw := range pspws {
		ids = append(ids, pspw.ID.String)
	}
	return database.TextArray(ids)
}
func (s *suite) queryTopicByPresetStudyPlanWeekly(ctx context.Context, db database.Ext, pspwIDs pgtype.TextArray) (map[pgtype.Text]*entities.Topic, error) {
	fields := database.GetFieldNames(&entities.Topic{})
	query := fmt.Sprintf(`
		SELECT pspw.preset_study_plan_weekly_id, t.%s
		FROM topics t
		JOIN (
			SELECT preset_study_plan_weekly_id, topic_id
			FROM preset_study_plans_weekly
			WHERE preset_study_plan_weekly_id = ANY($1::text[])
				AND deleted_at IS NULL
		) pspw USING(topic_id)
		WHERE t.deleted_at IS NULL`, strings.Join(fields, ", t."))
	rows, err := db.Query(ctx, query, pspwIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %s", err)
	}
	defer rows.Close()
	topicByPspwID := make(map[pgtype.Text]*entities.Topic)
	for rows.Next() {
		pspwID := pgtype.Text{}
		topic := new(entities.Topic)
		scanFields := []interface{}{&pspwID}
		scanFields = append(scanFields, database.GetScanFields(topic, fields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %s", err)
		}
		topicByPspwID[pspwID] = topic
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %s", err)
	}
	return topicByPspwID, nil
}
func (s *suite) validateTopic(ctx context.Context,
	t *entities.Topic,
	name pgtype.Text,
	country pgtype.Text,
	grade pgtype.Int2,
	subject pgtype.Text,
	schoolID pgtype.Int4,
	topicType pgtype.Text,
	status pgtype.Text,
	displayOrder pgtype.Int2) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if t.Name != name {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for name, got %+v", name, t.Name)
	}
	if t.Country != country {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for country, got %+v", country, t.Country)
	}
	if t.Grade != grade {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for grade, got %+v", grade, t.Grade)
	}
	if t.Subject != subject {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for subject, got %+v", subject, t.Subject)
	}
	if t.SchoolID != schoolID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for school id, got %+v", schoolID, t.SchoolID)
	}
	if t.TopicType != topicType {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for topic type, got %+v", topicType, t.TopicType)
	}
	if t.Status != status {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for status, got %+v", status, t.Status)
	}
	if t.DisplayOrder != displayOrder {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %+v for display order, got %+v", displayOrder, t.DisplayOrder)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) findDeletedPresetStudyPlansWeekly(ctx context.Context, db database.Ext, lessonID pgtype.Text, removedCourseIDs pgtype.TextArray) ([]*entities.PresetStudyPlanWeekly, error) {
	fields := database.GetFieldNames(&entities.PresetStudyPlanWeekly{})
	query := fmt.Sprintf(`
		SELECT %s
		FROM preset_study_plans_weekly
		WHERE deleted_at IS NOT NULL
			AND lesson_id = $1
			AND preset_study_plan_id IN (
				SELECT preset_study_plan_id
				FROM courses
				WHERE course_id = ANY($2::text[])
			)`, strings.Join(fields, ", "))
	pspws := entities.PresetStudyPlansWeekly{}
	err := database.Select(ctx, db, query, lessonID, removedCourseIDs).ScanAll(&pspws)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %s", err)
	}
	return pspws, nil
}
func (s *suite) findDeletedTopicByPresetStudyPlanWeeklyID(ctx context.Context, db database.Ext, pspwIDs pgtype.TextArray) ([]*entities.Topic, error) {
	fields := database.GetFieldNames(&entities.Topic{})
	query := fmt.Sprintf(`
		SELECT %s
		FROM topics
		WHERE deleted_at IS NOT NULL
			AND topic_id IN (
				SELECT topic_id
				FROM preset_study_plans_weekly
				WHERE preset_study_plan_weekly_id = ANY($1)
			)`, strings.Join(fields, ", "))
	topics := entities.Topics{}
	err := database.Select(ctx, db, query, pspwIDs).ScanAll(&topics)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %s", err)
	}
	return topics, nil
}
func (s *suite) userUpdatesTheLiveLessonWithStartTimeLaterThanEndTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, createRequest, err := s.extractCreateLessonRequestFromPreviousStep(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get CreateLiveLessonRequest: %s", err)
	}

	updateRequest := s.initUpdateLiveLessonRequest(createRequest, stepState.lessonID)
	updateRequest.StartTime = timestamppb.New(updateRequest.EndTime.AsTime().Add(time.Second * 3600))

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).UpdateLiveLesson(s.signedCtx(ctx), updateRequest)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userUpdatesTheLiveLessonWithMissingField(ctx context.Context, fieldName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, createRequest, err := s.extractCreateLessonRequestFromPreviousStep(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get CreateLiveLessonRequest: %s", err)
	}

	updateRequest := s.initUpdateLiveLessonRequest(createRequest, stepState.lessonID)
	switch fieldName {
	case lessonName:
		updateRequest.Name = ""
	case startTimeString:
		updateRequest.StartTime = nil
	case endTimeString:
		updateRequest.EndTime = nil
	case teacherIds:
		updateRequest.TeacherIds = nil
	case courseIds:
		updateRequest.CourseIds = nil
	case "learner ids":
		updateRequest.LearnerIds = nil
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`invalid field name "%s"`, fieldName)
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).UpdateLiveLesson(s.signedCtx(ctx), updateRequest)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLiveLessonIsNotUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// Query the lesson from database and check it against the create request
	ctx, createRequest, err := s.extractCreateLessonRequestFromPreviousStep(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get CreateLiveLessonRequest: %s", err)
	}

	lessonID := stepState.lessonID
	lesson, err := (&repositories.LessonRepo{}).FindByID(ctx, s.DB, database.Text(lessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not find lesson: %s", err)
	}
	if lesson.LessonID.String != lessonID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LessonID mismatched")
	}
	if lesson.Name.String != createRequest.Name {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.Name mismatched")
	}
	if lesson.TeacherID.String != createRequest.TeacherIds[0] {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.TeacherID mismatched")
	}
	if lesson.CourseID.String != createRequest.CourseIds[0] {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.CourseID mismatched")
	}
	if lesson.DeletedAt.Status != pgtype.Null {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.DeletedAt is not null")
	}
	if !lesson.StartTime.Time.Equal(createRequest.StartTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for start time, got %s", createRequest.StartTime.AsTime(), lesson.StartTime.Time)
	}
	if !lesson.EndTime.Time.Equal(createRequest.EndTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for end time, got %s", createRequest.EndTime.AsTime(), lesson.EndTime.Time)
	}
	if lesson.LessonType.String != string(entities.LessonTypeOnline) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for Lesson.LessonType, got %s", entities.LessonTypeOnline, lesson.LessonType.String)
	}
	return StepStateToContext(ctx, stepState), nil
}
