package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRetrieveLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lessonRepo := new(mock_repositories.MockLessonRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)
	configRepo := new(mock_repositories.MockConfigRepo)
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	courseService := &CourseService{
		UserRepo:        userRepo,
		LessonRepo:      lessonRepo,
		TopicRepo:       topicRepo,
		ClassMemberRepo: classMemberRepo,
		CourseClassRepo: courseClassRepo,
		ConfigRepo:      configRepo,
		Cfg: &configurations.Config{
			Common: configs.CommonConfig{
				Environment: "prod",
			},
		},
		UnleashClientIns: mockUnleashClient,
		Env:              "local",
	}
	var pgtypeTime pgtype.Timestamptz
	pgtypeTime.Set(time.Now())
	lessons := []*repositories.LessonWithTime{
		{
			Lesson: entities_bob.Lesson{
				LessonID:  database.Text("lessonID"),
				TeacherID: database.Text("teacherID"),
				CourseID:  database.Text("courseID"),
				EndAt:     pgtypeTime,
			},
			PresetStudyPlanWeekly: entities_bob.PresetStudyPlanWeekly{
				PresetStudyPlanID: database.Text("presetStudyPlanID"),
				TopicID:           database.Text("topicID"),
				StartDate:         pgtypeTime,
				EndDate:           pgtypeTime,
			},
		},
	}
	teachers := []*entities_bob.User{
		{
			ID: database.Text("teacherID"),
		},
	}

	mapClassIDByCourseID := make(map[pgtype.Text]pgtype.Int4Array)
	mapClassIDByCourseID[lessons[0].Lesson.CourseID] = database.Int4Array([]int32{0, 1, 2})
	emptymapClassIDByCourseID := make(map[pgtype.Text]pgtype.Int4Array)
	var userClassIDs []int32
	var total pgtype.Int8
	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "1",
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	// expectedCode := codes.Unknown

	pgSchedulingStatus := pgtype.Text{String: string(entities.LessonSchedulingStatusPublished), Status: pgtype.Present}
	pgSchedulingStatusNull := pgtype.Text{String: "", Status: pgtype.Null}

	testCases := []TestCase{
		{
			name:         "err find lesson",
			ctx:          ctx,
			req:          &pb.RetrieveLiveLessonRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(mapClassIDByCourseID, nil)
				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, pgx.ErrNoRows)
			},
		},
		{
			name:         "err find teacher",
			ctx:          ctx,
			req:          &pb.RetrieveLiveLessonRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(mapClassIDByCourseID, nil)
				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, pgx.ErrNoRows)
			},
		},
		{
			name:         "teacher retrieve live lesson happy case",
			ctx:          ctx,
			req:          &pb.RetrieveLiveLessonRequest{},
			expectedResp: &pb.RetrieveLiveLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(mapClassIDByCourseID, nil)
				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "student retrieve live lesson happy case",
			ctx:          ctx,
			req:          &pb.RetrieveLiveLessonRequest{},
			expectedResp: &pb.RetrieveLiveLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_STUDENT.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(mapClassIDByCourseID, nil)
				configRepo.On("RetrieveWithResourcePath",
					ctx, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"specificCourseIDsForLesson"}),
					database.Text("1")).
					Return([]*entities.Config{
						{
							Key:     database.Text("specificCourseIDsForLesson"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("JPREP_COURSE_000000162,JPREP_COURSE_000000218,JPREP_COURSE_000000163"),
						},
					}, nil)
				lessonRepo.On("FindLessonJoined", ctx, mock.Anything, database.Text(userID), mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "happy case with empty class map",
			ctx:          ctx,
			req:          &pb.RetrieveLiveLessonRequest{},
			expectedResp: &pb.RetrieveLiveLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(emptymapClassIDByCourseID, nil)
				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "teacher retrieve live lesson happy case with unleash scheduling status is false",
			ctx:          ctx,
			req:          &pb.RetrieveLiveLessonRequest{},
			expectedResp: &pb.RetrieveLiveLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(mapClassIDByCourseID, nil)
				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "student retrieve live lesson happy case  with unleash scheduling status is false",
			ctx:          ctx,
			req:          &pb.RetrieveLiveLessonRequest{},
			expectedResp: &pb.RetrieveLiveLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_STUDENT.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(mapClassIDByCourseID, nil)
				configRepo.On("RetrieveWithResourcePath",
					ctx, mock.Anything,
					database.Text("COUNTRY_MASTER"),
					database.Text("lesson"),
					database.TextArray([]string{"specificCourseIDsForLesson"}),
					database.Text("1")).
					Return([]*entities.Config{
						{
							Key:     database.Text("specificCourseIDsForLesson"),
							Group:   database.Text("lesson"),
							Country: database.Text("COUNTRY_MASTER"),
							Value:   database.Text("JPREP_COURSE_000000162,JPREP_COURSE_000000218,JPREP_COURSE_000000163"),
						},
					}, nil)
				lessonRepo.On("FindLessonJoined", ctx, mock.Anything, database.Text(userID), mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			rsp, err := courseService.RetrieveLiveLesson(ctx, testCase.req.(*pb.RetrieveLiveLessonRequest))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestToPbLesson(t *testing.T) {
	t.Parallel()

	t.Run("empty class id", func(t *testing.T) {
		t.Parallel()
		var pgtypeTime pgtype.Timestamptz
		pgtypeTime.Set(time.Now())
		teacherMap := map[string]*pb.BasicProfile{
			"teacherID": {
				UserId: "teacherID",
			},
		}

		lessons := []*repositories.LessonWithTime{
			{
				Lesson: entities_bob.Lesson{
					LessonID:  database.Text("lessonID"),
					TeacherID: database.Text("teacherID"),
					CourseID:  database.Text("courseID"),
					EndAt:     pgtypeTime,
				},
				PresetStudyPlanWeekly: entities_bob.PresetStudyPlanWeekly{
					PresetStudyPlanID: database.Text("presetStudyPlanID"),
					TopicID:           database.Text("topicID"),
					StartDate:         pgtypeTime,
					EndDate:           pgtypeTime,
				},
			},
		}
		mapClassIDByCourseID := make(map[pgtype.Text]pgtype.Int4Array)
		mapClassIDByCourseID[lessons[0].Lesson.CourseID] = database.Int4Array([]int32{0, 1, 2})
		results := toPbLessons(lessons, teacherMap)
		assert.Equal(t, len(results[0].UserClassIds), 0)
	})
}

func TestCourseService_RetrieveCoursesByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	courseRepo := new(mock_repositories.MockCourseRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	chapterRepo := new(mock_repositories.MockChapterRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)
	bookRepo := new(mock_repositories.MockBookRepo)
	courseBookRepo := new(mock_repositories.MockCourseBookRepo)

	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)

	s := &CourseService{
		CourseRepo:      courseRepo,
		UserRepo:        userRepo,
		CourseClassRepo: courseClassRepo,
		ChapterRepo:     chapterRepo,
		LessonRepo:      lessonRepo,
		BookRepo:        bookRepo,
		CourseBookRepo:  courseBookRepo,
		TopicRepo:       topicRepo,

		BookChapterRepo: bookChapterRepo,
	}

	users := []*entities_bob.User{}

	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	_ = course.TeacherIDs.Set([]string{"teacherIDs"})
	_ = course.ID.Set("1")
	_ = course.Country.Set(pb.COUNTRY_VN.String())
	_ = course.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = course.Grade.Set(8)
	_ = course.SchoolID.Set(1)

	chapter := &entities_bob.Chapter{}
	database.AllNullEntity(chapter)
	_ = chapter.ID.Set("1")
	_ = chapter.Country.Set(pb.COUNTRY_VN.String())
	_ = chapter.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = chapter.Grade.Set(8)
	_ = chapter.SchoolID.Set(1)

	lesson := &entities_bob.Lesson{}
	database.AllNullEntity(lesson)
	_ = lesson.CourseID.Set("1")
	_ = lesson.LessonID.Set("1")

	courses := entities_bob.Courses{course}
	courseClass := make(map[pgtype.Text]pgtype.Int4Array)
	validRequest := &pb.RetrieveCoursesByIDsRequest{}

	enChapters := []*entities_bob.Chapter{}
	enChapters = append(enChapters, chapter)

	enLessons := []*entities_bob.Lesson{}
	enLessons = append(enLessons, lesson)

	book := &entities_bob.Book{}
	database.AllNullEntity(book)
	_ = book.ID.Set("1")
	_ = book.Country.Set(pb.COUNTRY_VN.String())
	_ = book.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = book.Grade.Set(8)
	_ = book.SchoolID.Set(1)

	enCourseBook := map[string][]string{}
	enCourseBook[course.ID.String] = []string{book.ID.String}

	enBooks := map[string]*entities_bob.Book{}
	enBooks[book.ID.String] = book

	enTopics := []*entities_bob.Topic{}
	topic := &entities_bob.Topic{}
	database.AllNullEntity(topic)
	_ = topic.ID.Set("1")
	_ = topic.ChapterID.Set(chapter.ID.String)
	_ = topic.Country.Set(pb.COUNTRY_VN.String())
	_ = topic.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = topic.Grade.Set(8)
	_ = topic.SchoolID.Set(1)
	_ = topic.DisplayOrder.Set(1)
	_ = topic.CreatedAt.Set(time.Now())
	_ = topic.UpdatedAt.Set(time.Now())

	enTopics = append(enTopics, topic)

	enBookChapters := map[string][]*entities_bob.BookChapter{}
	bookChapter := &entities_bob.BookChapter{}
	database.AllNullEntity(bookChapter)
	bookChapter.BookID.Set(book.ID.String)
	bookChapter.ChapterID.Set(chapter.ID.String)

	enBookChapters[book.ID.String] = []*entities_bob.BookChapter{bookChapter}

	testcases := map[string]TestCase{
		"err query retrieve course": {
			ctx:         interceptors.ContextWithUserID(ctx, "err query retrieve course"),
			req:         validRequest,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.Courses{}, pgx.ErrTxClosed)
			},
		},
		"err retrieve users": {
			ctx:         interceptors.ContextWithUserID(ctx, "err retrieve users"),
			req:         validRequest,
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, pgx.ErrNoRows)
			},
		},
		"err find course class": {
			ctx:         interceptors.ContextWithUserID(ctx, "err find course class"),
			req:         validRequest,
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)
				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, pgx.ErrNoRows)
			},
		},
		"happy case retrieve all": {
			ctx:         interceptors.ContextWithUserID(ctx, "happy case retrieve all"),
			req:         validRequest,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(enCourseBook, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{course.ID.String}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)
				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
	}
	for name, testCase := range testcases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.RetrieveCoursesByIDsRequest)
			_, err := s.RetrieveCoursesByIDs(testCase.ctx, req)

			if testCase.expectedErr != nil {
				assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestGetTeacherProfileEmpty(t *testing.T) {
	t.Parallel()
	c := &CourseService{}
	tm, err := c.getTeacherProfile(context.Background(), []string{})
	assert.NoError(t, err, "expectign no error when passing empty ID slice")
	assert.Empty(t, tm, "expecting empty teacher's profile map")

	tm, err = c.getTeacherProfile(context.Background(), nil)
	assert.NoError(t, err, "expectign no error when passing nil ID slice")
	assert.Empty(t, tm, "expecting empty teacher's profile map")
}

func TestGetTeacherProfileTransform(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	teacherIDs := []string{"abc", "123"}
	userRepo := new(mock_repositories.MockUserRepo)
	teachers := []*entities_bob.User{
		{
			ID:       database.Text("abc"),
			LastName: database.Text("def"),
		},
		{
			ID:       database.Text("123"),
			LastName: database.Text("456"),
		},
	}

	userRepo.On("Retrieve", ctx, nil, database.TextArray(teacherIDs), []string(nil)).Once().Return(teachers, nil)
	c := &CourseService{
		DB:       nil,
		UserRepo: userRepo,
	}
	teachersMap, err := c.getTeacherProfile(ctx, teacherIDs)
	assert.NoError(t, err, "expecting no error")
	assert.Equal(t, len(teachers), len(teachersMap), "expecting same number of item")
	for _, i := range teacherIDs {
		_, ok := teachersMap[i]
		assert.True(t, ok, "unexpected teacher item")
	}
}
