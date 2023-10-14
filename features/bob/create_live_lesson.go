package bob

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	yasuo_repo "github.com/manabie-com/backend/internal/yasuo/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/gogo/protobuf/types"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const lessonName = "lesson name"
const courseIds = "course ids"
const teacherIds = "teacher ids"
const startTimeString = "start time"
const endTimeString = "end time"
const studentIds = "student ids"

func (s *suite) CreateTeacherAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDs = []string{}
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	teacherOne, err := s.createUserWithRole(ctx, constant.RoleTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, teacherOne.UserID)

	teacherTwo, err := s.createUserWithRole(ctx, constant.RoleTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, teacherTwo.UserID)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) CreateStudentAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentOne, err := s.createUserWithRole(ctx, constant.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentIds = append(stepState.StudentIds, studentOne.UserID)

	studentTwo, err := s.createUserWithRole(ctx, constant.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentIds = append(stepState.StudentIds, studentTwo.UserID)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) upsertLiveCourse(ctx context.Context, id string, teacherID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var course entities_bob.Course
	database.AllNullEntity(&course)
	err := multierr.Combine(
		course.ID.Set(id),
		course.Name.Set("live-course "+stepState.Random),
		course.CreatedAt.Set(time.Now()),
		course.UpdatedAt.Set(time.Now()),
		course.DeletedAt.Set(nil),
		course.Grade.Set(3),
		course.Subject.Set(pb.SUBJECT_BIOLOGY.String()),
		course.TeacherIDs.Set([]string{teacherID}),
		course.Country.Set(pb.COUNTRY_VN.String()),
		course.StartDate.Set(time.Now()),
		course.StartDate.Set(time.Now().Add(2*time.Hour)),
		course.SchoolID.Set(fmt.Sprint(stepState.CurrentSchoolID)),
		course.Status.Set("COURSE_STATUS_NONE"),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	_, err = database.Insert(ctx, &course, s.DBPostgres.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) CreateLiveCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.courseIds = []string{s.newID(), s.newID()}
	stepState.CourseIDs = stepState.courseIds
	for _, id := range stepState.courseIds {
		if ctx, err := s.upsertLiveCourse(ctx, id, stepState.TeacherIDs[0]); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) CreateMedias(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	savedAuthToken := stepState.AuthToken

	var err error
	stepState.CurrentSchoolID = constants.ManabieSchool

	ctx, err = s.signedAsAccountV2(ctx, "staff granted role school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	ctx, err = s.userUpsertValidMediaList(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}

	resp := stepState.Response.(*pb.UpsertMediaResponse)
	stepState.MediaIDs = resp.MediaIds

	// Recover the original token
	stepState.AuthToken = savedAuthToken
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createLessonCreatedSubscribe(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonCreatedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &pb.EvtLesson{}
		err := r.Unmarshal(data)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *pb.EvtLesson_CreateLessons_:
			if cmp.Equal(stepState.StudentIds, r.GetCreateLessons().Lessons[0].GetLearnerIds()) {
				stepState.FoundChanForJetStream <- r.Message
				return false, nil
			}
		}
		return false, nil
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonCreated, opts, handlerLessonCreatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserCreateLiveLesson(ctx context.Context, name, startTimeStr, endTimeStr, brightcoveVideoURL string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	material := make([]*bpb.Material, 0, len(stepState.MediaIDs))
	for _, id := range stepState.MediaIDs {
		material = append(material,
			&bpb.Material{
				Resource: &bpb.Material_MediaId{
					MediaId: id,
				},
			},
		)
	}

	if len(brightcoveVideoURL) != 0 {
		material = append(material, &bpb.Material{
			Resource: &bpb.Material_BrightcoveVideo_{
				BrightcoveVideo: &bpb.Material_BrightcoveVideo{
					Url: brightcoveVideoURL,
				},
			},
		})
	}
	req := &bpb.CreateLiveLessonRequest{
		Name:       name,
		StartTime:  timestamppb.New(startTime),
		EndTime:    timestamppb.New(endTime),
		TeacherIds: stepState.TeacherIDs,
		CourseIds:  stepState.courseIds,
		LearnerIds: stepState.StudentIds,
		Materials:  material,
	}

	// for nats

	stepState.RequestSentAt = time.Now()
	// create subscribe

	ctx, err = s.createLessonCreatedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createLessonCreatedSubscribe: %w", err)
	}
	res, err := bpb.NewLessonModifierServiceClient(s.Conn).CreateLiveLesson(s.signedCtx(ctx), req)
	stepState.CurrentLessonID = res.GetId()
	stepState.Response, stepState.ResponseErr = res, err

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkLessonTeacher(ctx context.Context, resp *bpb.CreateLiveLessonResponse) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// check teacher ids
	actualTeacherIDs, err := s.getTeacherIDsLiveLessonInDB(ctx, resp.Id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	expectedTeacherIDs := make(map[string]bool, len(stepState.TeacherIDs))
	for _, id := range stepState.TeacherIDs {
		expectedTeacherIDs[id] = true
	}

	for _, actualTeacherID := range actualTeacherIDs.Elements {
		if ok := expectedTeacherIDs[actualTeacherID.String]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("teacher id %s not be expected", actualTeacherID.String)
		}
		delete(expectedTeacherIDs, actualTeacherID.String)
	}

	if len(expectedTeacherIDs) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could find teacher ids %v", expectedTeacherIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkCourses(ctx context.Context, resp *bpb.CreateLiveLessonResponse, startTime, endTime time.Time) (entities_bob.Courses, pgtype.TextArray, error) {
	actualCourseIDs, err := s.getCourseIDsLiveLessonInDB(ctx, resp.Id)
	if err != nil {
		return nil, pgtype.TextArray{}, err
	}

	stepState := StepStateFromContext(ctx)
	expectedCourseIDs := make(map[string]bool, len(stepState.courseIds))
	for _, id := range stepState.courseIds {
		expectedCourseIDs[id] = true
	}
	for _, actualCourseID := range actualCourseIDs.Elements {
		if ok := expectedCourseIDs[actualCourseID.String]; !ok {
			return nil, pgtype.TextArray{}, fmt.Errorf("course id %s not be expected", actualCourseID.String)
		} else {
			delete(expectedCourseIDs, actualCourseID.String)
		}
	}
	if len(expectedCourseIDs) != 0 {
		return nil, pgtype.TextArray{}, fmt.Errorf("could not find course ids %v", expectedCourseIDs)
	}
	coursesByID, err := s.getCoursesInDB(ctx, actualCourseIDs)
	if err != nil {
		return nil, pgtype.TextArray{}, err
	}
	courses := make(entities_bob.Courses, 0, len(coursesByID))
	for id := range coursesByID {
		courses = append(courses, coursesByID[id])
		if len(coursesByID[id].PresetStudyPlanID.String) == 0 {
			return nil, pgtype.TextArray{}, fmt.Errorf("expected preset study plan of course %s will be create but did not", coursesByID[id].ID.String)
		}
		if !coursesByID[id].StartDate.Time.Equal(startTime) {
			return nil, pgtype.TextArray{}, fmt.Errorf("expected course's start date %v but got %v", startTime, coursesByID[id].StartDate.Time)
		}
		if !coursesByID[id].EndDate.Time.Equal(endTime) {
			return nil, pgtype.TextArray{}, fmt.Errorf("expected course's end date %v but got %v", endTime, coursesByID[id].EndDate.Time)
		}
	}
	return courses, actualCourseIDs, nil
}
func (s *suite) checkLearners(ctx context.Context, resp *bpb.CreateLiveLessonResponse) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	actualLearnerIDs, err := s.getLearnerIDsOfLessonInDB(ctx, resp.Id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	expectedLearnerIDs := make(map[string]bool, len(stepState.StudentIds))
	for _, id := range stepState.StudentIds {
		expectedLearnerIDs[id] = true
	}

	for _, actualLearnerID := range actualLearnerIDs.Elements {
		if ok := expectedLearnerIDs[actualLearnerID.String]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("leanrer id %s not be expected", actualLearnerID.String)
		} else {
			delete(expectedLearnerIDs, actualLearnerID.String)
		}
	}

	if len(expectedLearnerIDs) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not find learner ids %v", expectedLearnerIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkPresetStudyPlans(ctx context.Context, courses entities_bob.Courses, coursesByID map[pgtype.Text]*entities_bob.Course) ([]string, error) {
	pSPByID, err := s.getPresetStudyPlanByCoursesInDB(ctx, courses)
	if err != nil {
		return nil, err
	}
	presetStudyPlanIDs := make([]string, 0, len(coursesByID))
	for _, course := range coursesByID {
		if v, ok := pSPByID[course.PresetStudyPlanID]; ok {
			if v.Name.String != course.Name.String {
				return nil, fmt.Errorf("expected preset study plan name %s but got %s", v.Name.String, course.Name.String)
			}
			if v.Country.String != course.Country.String {
				return nil, fmt.Errorf("expected preset study plan country %s but got %s", v.Country.String, course.Country.String)
			}
			if v.Grade.Int != course.Grade.Int {
				return nil, fmt.Errorf("expected preset study plan grade %d but got %d", v.Grade.Int, course.Grade.Int)
			}
			if v.Subject.String != course.Subject.String {
				return nil, fmt.Errorf("expected preset study plan subject %s but got %s", v.Subject.String, course.Subject.String)
			}
			presetStudyPlanIDs = append(presetStudyPlanIDs, course.PresetStudyPlanID.String)
		} else {
			return nil, fmt.Errorf("could not found preset study plan %s of course %s", course.PresetStudyPlanID.String, course.ID.String)
		}
	}
	return presetStudyPlanIDs, nil
}
func (s *suite) checkPresetStudyPlanWeeklies(ctx context.Context,
	presetStudyPlanIDs []string, coursesByID map[pgtype.Text]*entities_bob.Course, lesson *entities_bob.Lesson) (
	context.Context, []string, []*entities_bob.Topic, error) {
	pSPWByPresetStudyPlanID, err := s.getPresetStudyPlanWeekliesInDB(ctx, database.TextArray(presetStudyPlanIDs))
	if err != nil {
		return ctx, nil, nil, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.CurrentPresetStudyPlanWeeklyIDs = make([]string, 0, len(pSPWByPresetStudyPlanID))
	for _, item := range pSPWByPresetStudyPlanID {
		stepState.CurrentPresetStudyPlanWeeklyIDs = append(stepState.CurrentPresetStudyPlanWeeklyIDs, item.ID.String)
	}
	topicIDs := make([]string, 0, len(pSPWByPresetStudyPlanID))
	expectedTopics := make([]*entities_bob.Topic, 0, len(pSPWByPresetStudyPlanID))
	stepState.CurrentTopicIDs = make([]string, 0, len(coursesByID))
	for _, course := range coursesByID {
		if presetStudyPlanWeekly, ok := pSPWByPresetStudyPlanID[course.PresetStudyPlanID]; ok {
			if presetStudyPlanWeekly.LessonID.String != lesson.LessonID.String {
				return ctx, nil, nil, fmt.Errorf("expected lesson ID of preset study plan weekly %s but got %s", lesson.LessonID.String, presetStudyPlanWeekly.LessonID.String)
			}
			if !presetStudyPlanWeekly.StartDate.Time.Equal(lesson.StartTime.Time) {
				return ctx, nil, nil, fmt.Errorf("expected start time of preset study plan weekly %v but got %s", lesson.StartTime.Time, presetStudyPlanWeekly.StartDate.Time)
			}
			if !presetStudyPlanWeekly.EndDate.Time.Equal(lesson.EndTime.Time) {
				return ctx, nil, nil, fmt.Errorf("expected end time of preset study plan weekly %v but got %s", lesson.EndTime.Time, presetStudyPlanWeekly.EndDate.Time)
			}
			topicIDs = append(topicIDs, presetStudyPlanWeekly.TopicID.String)
			e := &entities_bob.Topic{}
			database.AllNullEntity(e)
			err = multierr.Combine(e.ID.Set(presetStudyPlanWeekly.TopicID), e.Name.Set(lesson.Name), e.Country.Set(course.Country), e.Grade.Set(course.Grade), e.Subject.Set(course.Subject), e.SchoolID.Set(course.SchoolID), e.TopicType.Set(entities_bob.TopicTypeLiveLesson), e.Status.Set(entities_bob.TopicStatusPublished), e.DisplayOrder.Set(1))
			if err != nil {
				return ctx, nil, nil, err
			}

			expectedTopics = append(expectedTopics, e)
			stepState.CurrentTopicIDs = append(stepState.CurrentTopicIDs, presetStudyPlanWeekly.TopicID.String)
		} else {
			return ctx, nil, nil, fmt.Errorf("could not found preset study plan weekly of preset study plan %s", course.PresetStudyPlanID.String)
		}
	}
	return StepStateToContext(ctx, stepState), topicIDs, expectedTopics, nil
}
func (s *suite) checkTopics(ctx context.Context, topicIDs []string, expectedTopics []*entities_bob.Topic) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topicsByID, err := s.getTopicsInDB(ctx, database.TextArray(topicIDs))
	if err != nil {
		return ctx, err

	}
	// check topic
	for _, expectedTopic := range expectedTopics {
		if v, ok := topicsByID[expectedTopic.ID]; ok {
			if expectedTopic.Name.String != v.Name.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic name %s but got %s", expectedTopic.Name.String, v.Name.String)
			}
			if expectedTopic.Country.String != v.Country.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic country %s but got %s", expectedTopic.Country.String, v.Country.String)
			}
			if expectedTopic.Grade.Int != v.Grade.Int {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic grade %d but got %d", expectedTopic.Grade.Int, v.Grade.Int)
			}
			if expectedTopic.Subject.String != v.Subject.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic subject %s but got %s", expectedTopic.Subject.String, v.Subject.String)
			}
			if expectedTopic.SchoolID.Int != v.SchoolID.Int {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic school id %d but got %d", expectedTopic.SchoolID.Int, v.SchoolID.Int)
			}
			if expectedTopic.TopicType.String != v.TopicType.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic type %s but got %s", expectedTopic.TopicType.String, v.TopicType.String)
			}
			if expectedTopic.Status.String != v.Status.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic status %s but got %s", expectedTopic.Status.String, v.Status.String)
			}
			if expectedTopic.DisplayOrder.Int != v.DisplayOrder.Int {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected topic display order %d but got %d", expectedTopic.DisplayOrder.Int, v.DisplayOrder.Int)
			}
		} else {
			return StepStateToContext(ctx, stepState), fmt.Errorf("could not found topic %s", expectedTopic.ID.String)
		}
	}

	return ctx, nil
}
func (s *suite) ThereIsLiveLessonInDB(ctx context.Context, name, startTimeStr, endTimeStr, brightcoveVideoURL string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	resp := stepState.Response.(*bpb.CreateLiveLessonResponse)
	if len(resp.Id) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected response id not empty but got empty id")
	}
	stepState.CurrentLessonID = resp.Id

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	lesson, err := s.getLiveLessonInDB(ctx, resp.Id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	// check lesson's data in db
	if len(lesson.LessonID.String) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected id not empty but got empty id")
	}

	if lesson.Name.String != name {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected name is %s but got %s", name, lesson.Name.String)
	}

	if !lesson.StartTime.Time.Equal(startTime) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected start time is %v but got %v", startTime, lesson.StartTime.Time)
	}

	if !lesson.EndTime.Time.Equal(endTime) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected end time is %v but got %v", endTime, lesson.EndTime.Time)
	}

	// check teacher ids
	if ctx, err = s.checkLessonTeacher(ctx, resp); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// check course ids
	courses, actualCourseIDs, err := s.checkCourses(ctx, resp, startTime, endTime)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	// check lesson learner ids
	if ctx, err = s.checkLearners(ctx, resp); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// check media
	actualMedias, err := s.getMediasByLessonGroupInDB(ctx, lesson.LessonGroupID, lesson.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	// check brightcove video URL
	if len(brightcoveVideoURL) != 0 {
		if len(actualMedias) != len(stepState.MediaIDs)+1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected 1 media but got %d ", len(actualMedias))
		}

		videoID, err := golibs.GetBrightcoveVideoIDFromURL(brightcoveVideoURL)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get Brightcove video ID: %v", err)
		}

		if actualMedias[0].Resource.String != videoID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected Brightcove video id is %s media but got %s ", videoID, actualMedias[0].Resource.String)
		}

		if actualMedias[0].Type.String != string(entities_bob.MediaTypeVideo) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected media type is %s media but got %s ", entities_bob.MediaTypeVideo, actualMedias[0].Type.String)
		}
	} else if len(actualMedias) != len(stepState.MediaIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d media but got %d ", len(stepState.MediaIDs), len(actualMedias))
	}

	// get list courses
	coursesByID, err := s.getCoursesInDB(ctx, actualCourseIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// check data of preset study plan
	presetStudyPlanIDs, err := s.checkPresetStudyPlans(ctx, courses, coursesByID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// check preset study plan weeklies
	ctx, topicIDs, expectedTopics, err := s.checkPresetStudyPlanWeeklies(ctx, presetStudyPlanIDs, coursesByID, lesson)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// check topic
	if ctx, err = s.checkTopics(ctx, topicIDs, expectedTopics); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) getTopicsInDB(ctx context.Context, topicIDs pgtype.TextArray) (map[pgtype.Text]*entities_bob.Topic, error) {
	repo := repo.TopicRepo{}
	topics, err := repo.RetrieveByIDs(ctx, s.DB, topicIDs)
	if err != nil {
		return nil, err
	}
	res := make(map[pgtype.Text]*entities_bob.Topic)
	for i := range topics {
		res[topics[i].ID] = topics[i]
	}
	return res, nil
}

func (s *suite) getPresetStudyPlanWeekliesInDB(ctx context.Context, presetStudyPlanIDs pgtype.TextArray) (map[pgtype.Text]*entities_bob.PresetStudyPlanWeekly, error) {
	fields := database.GetFieldNames(&entities_bob.PresetStudyPlanWeekly{})
	query := fmt.Sprintf("SELECT %s FROM preset_study_plans_weekly WHERE preset_study_plan_id = ANY($1) ORDER BY week ASC", strings.Join(fields, ","))
	rows, err := s.DB.Query(ctx, query, &presetStudyPlanIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pp []*entities_bob.PresetStudyPlanWeekly
	for rows.Next() {
		p := new(entities_bob.PresetStudyPlanWeekly)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	res := make(map[pgtype.Text]*entities_bob.PresetStudyPlanWeekly)
	for i := range pp {
		res[pp[i].PresetStudyPlanID] = pp[i]
	}
	return res, err
}
func (s *suite) getPresetStudyPlanByCoursesInDB(ctx context.Context, courses entities_bob.Courses) (map[pgtype.Text]*entities_bob.PresetStudyPlan, error) {
	ids := make([]string, 0, len(courses))
	for _, course := range courses {
		if len(course.PresetStudyPlanID.String) != 0 {
			ids = append(ids, course.PresetStudyPlanID.String)
		}
	}

	repo := yasuo_repo.PresetStudyPlanRepo{}
	pSPByID, err := repo.FindByIDs(ctx, s.DB, database.TextArray(ids))
	if err != nil {
		return nil, err
	}
	return pSPByID, nil
}
func (s *suite) getCoursesInDB(ctx context.Context, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities_bob.Course, error) {
	repo := repo.CourseRepo{}
	courses, err := repo.FindByIDs(ctx, s.DB, courseIDs)
	if err != nil {
		return nil, err
	}
	return courses, nil
}
func (s *suite) getLiveLessonInDB(ctx context.Context, id string) (*entities_bob.Lesson, error) {
	lessonRepo := repo.LessonRepo{}
	return lessonRepo.FindByID(ctx, s.DB, database.Text(id))
}
func (s *suite) getTeacherIDsLiveLessonInDB(ctx context.Context, id string) (pgtype.TextArray, error) {
	repo := repo.LessonRepo{}
	ids, err := repo.GetTeacherIDsOfLesson(ctx, s.DB, database.Text(id))
	if err != nil {
		return pgtype.TextArray{}, err
	}
	return ids, nil
}
func (s *suite) getCourseIDsLiveLessonInDB(ctx context.Context, id string) (pgtype.TextArray, error) {
	repo := repo.LessonRepo{}
	ids, err := repo.GetCourseIDsOfLesson(ctx, s.DB, database.Text(id))
	if err != nil {
		return pgtype.TextArray{}, err
	}
	return ids, nil
}
func (s *suite) getLearnerIDsOfLessonInDB(ctx context.Context, id string) (pgtype.TextArray, error) {
	repo := repo.LessonRepo{}
	ids, err := repo.GetLearnerIDsOfLesson(ctx, s.DB, database.Text(id))
	if err != nil {
		return pgtype.TextArray{}, err
	}
	return ids, nil
}
func (s *suite) getMediasByLessonGroupInDB(ctx context.Context, lessonGrID, courseID pgtype.Text) (entities_bob.Medias, error) {
	lgRepo := repo.LessonGroupRepo{}
	return lgRepo.GetMedias(ctx, s.DB, lessonGrID, courseID, database.Int4(100), pgtype.Text{Status: pgtype.Null})
}
func (s *suite) UserRetrieveListLessonsByAboveCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.RetrieveLiveLessonRequest{
		CourseIds: stepState.courseIds,
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  0,
		},
		From: &types.Timestamp{Seconds: time.Time{}.Unix()},
		To:   &types.Timestamp{Seconds: time.Now().Add(time.Hour).Unix()},
	}
	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveLiveLesson(s.signedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) TeacherGetALiveLesson(ctx context.Context, name, startTimeStr, endTimeStr, brightcoveVideoURL string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	resp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	lessonByID := make(map[string]*pb.Lesson)
	for i := range resp.Lessons {
		lessonByID[resp.Lessons[i].LessonId] = resp.Lessons[i]
	}
	if len(lessonByID) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d lessons but got %d", 1, len(lessonByID))
	}

	lesson, ok := lessonByID[stepState.CurrentLessonID]
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could found lesson %s", stepState.CurrentLessonID)
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	// check lesson
	if lesson.Topic.Name != name {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected name is %s but got %s", name, lesson.Topic.Name)
	}

	if !lesson.StartTime.Equal(&types.Timestamp{Seconds: startTime.Unix()}) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected start time is %v but got %v", startTime, lesson.StartTime)
	}

	if !lesson.EndTime.Equal(&types.Timestamp{Seconds: endTime.Unix()}) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected end time is %v but got %v", endTime, lesson.EndTime)
	}

	// check teacher ids
	// Hack: Current api just return 1 teacher id
	ok = false
	for _, id := range stepState.TeacherIDs {
		if lesson.Teacher[0].UserId == id {
			ok = true
		}
	}
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not found teacher id %s", lesson.Teacher[0].UserId)
	}

	// check course ids
	ok = false
	for _, id := range stepState.courseIds {
		if lesson.CourseId == id {
			ok = true
			break
		}
	}
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not found course id %s", lesson.CourseId)
	}

	// check total learner
	// current api not return
	//if int(lesson.TotalLearner) != len(stepState.StudentIds) {
	//	return fmt.Errorf("expected %d students but got %d", len(stepState.StudentIds), lesson.TotalLearner)
	//}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserCreateLiveLessonWithMissing(ctx context.Context, missingField string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now().Round(time.Second)
	req := &bpb.CreateLessonRequest{
		StartTime:       timestamppb.New(now.Add(9 * time.Hour)),
		EndTime:         timestamppb.New(now.Add(10 * time.Hour)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		CenterId:        constants.ManabieOrgLocation,
		StudentInfoList: []*bpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*bpb.Material{},
		SavingOption: &bpb.CreateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
		SchedulingStatus: bpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}

	switch missingField {
	case startTimeString:
		req.StartTime = nil
	case endTimeString:
		req.EndTime = nil
	case teacherIds:
		req.TeacherIds = nil
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
			LocationId:       constants.ManabieOrgLocation,
		})
	}

	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &bpb.Material{
			Resource: &bpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}

	res, err := bpb.NewLessonManagementServiceClient(s.Conn).CreateLesson(s.signedCtx(ctx), req)
	if err == nil {
		stepState.CurrentLessonID = res.Id
	}
	stepState.Response = res
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserCannotCreateAnyLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected error but not got")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) liveLessonHasRoomId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.FindByID(ctx, s.DB, database.Text(stepState.CurrentLessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(lesson.RoomID.String) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected has room_id but got nil")
	}
	return StepStateToContext(ctx, stepState), nil
}
