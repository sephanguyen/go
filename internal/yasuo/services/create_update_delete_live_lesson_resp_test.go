package services

import (
	"context"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	gconstants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bobRepo "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	mock_repositories "github.com/manabie-com/backend/mock/yasuo/repositories"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := &mock_repositories.MockLessonRepo{}
	courseRepo := &bobRepo.MockCourseRepo{}
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	topicRepo := &mock_repositories.MockTopicRepo{}
	presetStudyPlanWeeklyRepo := &mock_repositories.MockPresetStudyPlanWeeklyRepo{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	jsm := new(mock_nats.JetStreamManagement)
	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	courseService := &CourseService{
		DBTrace:                   &database.DBTrace{DB: mockDB},
		UnleashClientIns:          mockUnleashClient,
		CourseRepo:                courseRepo,
		LessonRepo:                lessonRepo,
		TeacherRepo:               teacherRepo,
		TopicRepo:                 topicRepo,
		PresetStudyPlanWeeklyRepo: presetStudyPlanWeeklyRepo,
		JSM:                       jsm,
		Env:                       "local",
	}

	start := time.Now()
	end := time.Now().Add(time.Hour * 24 * 30)
	reqLesson := &pb.CreateLiveLessonRequest_Lesson{
		StartDate: &types.Timestamp{Seconds: start.Unix()},
		EndDate:   &types.Timestamp{Seconds: end.Unix()},
		Name:      "live-lesson-name",
		TeacherId: "teacher-id",
		Attachments: []*pb.Attachment{
			{
				Name: "att-name",
				Url:  "att-url",
			},
		},
		ControlSettings: &pb.ControlSettingLiveLesson{
			TeacherObversers:          []string{"teacher-obversers"},
			Lectures:                  []string{"teacher-teach"},
			DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
			PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
			UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
		},
	}
	course := &entities_bob.Course{
		ID:                database.Text("course-id"),
		SchoolID:          database.Int4(1),
		PresetStudyPlanID: database.Text("preset-study-plan-id"),
		CourseType:        database.Text(pb_bob.COURSE_TYPE_LIVE.String()),
		TeacherIDs:        database.TextArray([]string{reqLesson.TeacherId}),
		StartDate:         pgtype.Timestamptz{Time: time.Now()},
		EndDate:           pgtype.Timestamptz{Time: time.Now()},
	}

	req := &pb.CreateLiveLessonRequest{
		CourseId: course.ID.String,
		Lessons:  []*pb.CreateLiveLessonRequest_Lesson{reqLesson},
	}

	testCases := map[string]TestCase{
		"can find course id": {
			ctx:         ctx,
			req:         req,
			expectedErr: status.Error(codes.InvalidArgument, "cannot find course"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				courseRepo.On("FindByID", ctx, mock.Anything, course.ID).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"dont have live lesson setting control": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: course.ID.String,
				Lessons: []*pb.CreateLiveLessonRequest_Lesson{{
					StartDate: &types.Timestamp{Seconds: start.Unix()},
					EndDate:   &types.Timestamp{Seconds: end.Unix()},
					Name:      "live-lesson-name",
					TeacherId: "teacher-id",
					Attachments: []*pb.Attachment{
						{
							Name: "att-name",
							Url:  "att-url",
						},
					},
					ControlSettings: nil,
				}},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				courseRepo.On("FindByID", ctx, mock.Anything, course.ID).Once().Return(course, nil)
				teacherRepo.On("ManyTeacherIsInSchool", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				topicRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				lessonRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				presetStudyPlanWeeklyRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				var start, end *time.Time
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, mock.Anything, mock.Anything).Once().Return(start, end, nil)

				courseRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				jsm.On("PublishAsyncContext", ctx, gconstants.SubjectLessonCreated, mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := testCase.ctx
			testCase.setup(ctx)
			_, err := courseService.CreateLiveLesson(ctx, testCase.req.(*pb.CreateLiveLessonRequest))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
		})
	}
}

func CreateLiveLesson_InvalidArgument(t *testing.T) {
	ctx := context.Background()
	courseService := &CourseService{}

	expectedCode := codes.InvalidArgument
	testCases := map[string]TestCase{
		"missing course id": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: "",
				Lessons: []*pb.CreateLiveLessonRequest_Lesson{{
					StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
					EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
					Name:      "live-lesson-name",
					TeacherId: "teacher-id",
					Attachments: []*pb.Attachment{{
						Name: "att-name",
						Url:  "att-url",
					}},
					ControlSettings: &pb.ControlSettingLiveLesson{
						TeacherObversers:          []string{"teacher-obversers"},
						Lectures:                  []string{"teacher-teach"},
						DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
						PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
						UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
					},
				}},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "course id cannot be empty",
		},
		"missing student publish video": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: "",
				Lessons: []*pb.CreateLiveLessonRequest_Lesson{{
					StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
					EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
					Name:      "live-lesson-name",
					TeacherId: "teacher-id",
					Attachments: []*pb.Attachment{{
						Name: "att-name",
						Url:  "att-url",
					}},
					ControlSettings: &pb.ControlSettingLiveLesson{
						TeacherObversers:          []string{"teacher-obversers"},
						Lectures:                  []string{"teacher-teach"},
						DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
						PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_NONE,
						UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
					},
				}},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "missing publish student video status",
		},
		"missing live lesson view": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: "",
				Lessons: []*pb.CreateLiveLessonRequest_Lesson{{
					StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
					EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
					Name:      "live-lesson-name",
					TeacherId: "teacher-id",
					Attachments: []*pb.Attachment{{
						Name: "att-name",
						Url:  "att-url",
					}},
					ControlSettings: &pb.ControlSettingLiveLesson{
						TeacherObversers:          []string{"teacher-obversers"},
						Lectures:                  []string{"teacher-teach"},
						DefaultView:               pb_bob.LIVE_LESSON_VIEW_NONE,
						PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
						UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
					},
				}},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "missing live lesson view",
		},
		"missing teacher ids": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: "",
				Lessons: []*pb.CreateLiveLessonRequest_Lesson{{
					StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
					EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
					Name:      "live-lesson-name",
					TeacherId: "",
					Attachments: []*pb.Attachment{{
						Name: "att-name",
						Url:  "att-url",
					}},
					ControlSettings: &pb.ControlSettingLiveLesson{
						TeacherObversers:          []string{"teacher-obversers"},
						Lectures:                  []string{"teacher-teach"},
						DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
						PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
						UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
					},
				}},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "teacher ids cannot be empty",
		},
		"can find course id": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: "not-found-course",
				Lessons: []*pb.CreateLiveLessonRequest_Lesson{{
					StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
					EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
					Name:      "live-lesson-name",
					TeacherId: "teacher-id",
					Attachments: []*pb.Attachment{{
						Name: "att-name",
						Url:  "att-url",
					}},
					ControlSettings: &pb.ControlSettingLiveLesson{
						TeacherObversers:          []string{"teacher-obversers"},
						Lectures:                  []string{"teacher-teach"},
						DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
						PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
						UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
					},
				}},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "cannot find course",
		},
		"missing lesson": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: "course-id",
				Lessons:  []*pb.CreateLiveLessonRequest_Lesson{},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "lessons cannot be empty",
		},
		"start date must before end date": {
			ctx: ctx,
			req: &pb.CreateLiveLessonRequest{
				CourseId: "course-id",
				Lessons: []*pb.CreateLiveLessonRequest_Lesson{{
					EndDate:   &types.Timestamp{Seconds: time.Now().Unix()},
					StartDate: &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
					Name:      "live-lesson-name",
					TeacherId: "teacher-id",
					Attachments: []*pb.Attachment{{
						Name: "att-name",
						Url:  "att-url",
					}},
					ControlSettings: &pb.ControlSettingLiveLesson{
						TeacherObversers:          []string{"teacher-obversers"},
						Lectures:                  []string{"teacher-teach"},
						DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
						PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
						UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
					},
				}},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "start date must before end date",
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			rsp, err := courseService.CreateLiveLesson(context.Background(), testCase.req.(*pb.CreateLiveLessonRequest))
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, expectedCode, status.Code(err))
			assert.Nil(t, rsp, "expecting nil response")
		})
	}
}

func TestUpdateLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = interceptors.ContextWithUserGroup(ctx, pb.USER_GROUP_ADMIN.String())

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	lessonRepo := &mock_repositories.MockLessonRepo{}
	courseRepo := &bobRepo.MockCourseRepo{}
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	topicRepo := &mock_repositories.MockTopicRepo{}
	liveLessonSentNotificationRepo := &mock_virtual_repo.MockLiveLessonSentNotificationRepo{}
	presetStudyPlanWeeklyRepo := &mock_repositories.MockPresetStudyPlanWeeklyRepo{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	courseService := &CourseService{
		DBTrace:                        &database.DBTrace{DB: mockDB},
		UnleashClientIns:               mockUnleashClient,
		CourseRepo:                     courseRepo,
		LessonRepo:                     lessonRepo,
		TeacherRepo:                    teacherRepo,
		TopicRepo:                      topicRepo,
		PresetStudyPlanWeeklyRepo:      presetStudyPlanWeeklyRepo,
		LiveLessonSentNotificationRepo: liveLessonSentNotificationRepo,
		Env:                            "local",
	}

	start := time.Now()
	end := time.Now().Add(time.Hour * 24 * 30)

	course := &entities_bob.Course{
		ID:                database.Text("course-id"),
		SchoolID:          database.Int4(1),
		PresetStudyPlanID: database.Text("preset-study-plan-id"),
		CourseType:        database.Text(pb_bob.COURSE_TYPE_LIVE.String()),
		StartDate:         pgtype.Timestamptz{Time: time.Now()},
		EndDate:           pgtype.Timestamptz{Time: time.Now()},
	}
	lesson := &entities_bob.Lesson{
		LessonID:  database.Text("lesson-id"),
		CourseID:  course.ID,
		TeacherID: database.Text("text"),
	}
	topic := &entities_bob.Topic{
		ID: database.Text("topic-id"),
	}

	presetStudyPlanWeekly := &entities_bob.PresetStudyPlanWeekly{
		ID:                database.Text("preset-id"),
		LessonID:          lesson.LessonID,
		PresetStudyPlanID: database.Text("preset-study-plan-id"),
		StartDate:         pgtype.Timestamptz{Time: time.Now()},
		EndDate:           pgtype.Timestamptz{Time: time.Now()},
		TopicID:           topic.ID,
	}

	req := &pb.UpdateLiveLessonRequest{
		LessonId:  lesson.LessonID.String,
		StartDate: &types.Timestamp{Seconds: start.Unix()},
		EndDate:   &types.Timestamp{Seconds: end.Unix()},
		Name:      "live-lesson-name",
		TeacherId: "teacher-id",
		Attachments: []*pb.Attachment{
			{
				Name: "att-name",
				Url:  "att-url",
			},
		},
		ControlSettings: &pb.ControlSettingLiveLesson{
			TeacherObversers:          []string{"teacher-obversers"},
			Lectures:                  []string{"teacher-teach"},
			DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
			PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
			UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
		},
	}

	errInternal := errors.New("internal")

	testCases := map[string]TestCase{
		"can find lesson id": {
			ctx:         ctx,
			req:         req,
			expectedErr: status.Error(codes.NotFound, "cannot find lesson"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"can find course id": {
			ctx:         ctx,
			req:         req,
			expectedErr: status.Error(codes.NotFound, "cannot find course"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(lesson, nil)
				courseRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"dont have live lesson setting control": {
			ctx: ctx,
			req: &pb.UpdateLiveLessonRequest{
				LessonId:  lesson.LessonID.String,
				StartDate: &types.Timestamp{Seconds: start.Unix()},
				EndDate:   &types.Timestamp{Seconds: end.Unix()},
				Name:      "live-lesson-name",
				TeacherId: "teacher-id",
				Attachments: []*pb.Attachment{
					{
						Name: "att-name",
						Url:  "att-url",
					},
				},
				ControlSettings: nil,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(lesson, nil)
				courseRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(course, nil)

				presetStudyPlanWeeklyRepo.On("FindByLessonID", ctx, mock.Anything, mock.Anything).Once().Return(presetStudyPlanWeekly, nil)
				topicRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(topic, nil)

				teacherRepo.On("ManyTeacherIsInSchool", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				topicRepo.On("Update", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				presetStudyPlanWeeklyRepo.On("Update", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				lessonRepo.On("Update", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				var start, end *time.Time
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, mock.Anything, mock.Anything).Once().Return(start, end, nil)
				courseRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		"live lesson sent notifications should be called when new start date is set to a future date": {
			ctx: ctx,
			req: &pb.UpdateLiveLessonRequest{
				LessonId: lesson.LessonID.String,
				// SoftDeleteLiveLessonSentNotificationRecord will only get called when request start date is set to a future date
				StartDate:       &types.Timestamp{Seconds: start.Add(1 * time.Hour).Unix()},
				EndDate:         &types.Timestamp{Seconds: end.Add(1 * time.Hour).Unix()},
				Name:            "live-lesson-name",
				TeacherId:       "teacher-id",
				ControlSettings: nil,
			},
			expectedErr: status.Error(codes.Unknown, errInternal.Error()),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(lesson, nil)
				courseRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(course, nil)

				presetStudyPlanWeeklyRepo.On("FindByLessonID", ctx, mock.Anything, mock.Anything).Once().Return(presetStudyPlanWeekly, nil)
				topicRepo.On("FindByID", ctx, mock.Anything, mock.Anything).Once().Return(topic, nil)

				teacherRepo.On("ManyTeacherIsInSchool", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				topicRepo.On("Update", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				presetStudyPlanWeeklyRepo.On("Update", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				liveLessonSentNotificationRepo.On("SoftDeleteLiveLessonSentNotificationRecord", ctx, mock.Anything, mock.Anything).Once().Return(errInternal)
				mockTxer.On("Rollback", mock.Anything).Return(nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := testCase.ctx
			testCase.setup(ctx)
			_, err := courseService.UpdateLiveLesson(ctx, testCase.req.(*pb.UpdateLiveLessonRequest))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
		})
	}
}

func UpdateLiveLesson_InvalidArgument(t *testing.T) {
	ctx := context.Background()
	courseService := &CourseService{}

	expectedCode := codes.InvalidArgument
	testCases := map[string]TestCase{
		"missing live lesson view": {
			ctx: ctx,
			req: &pb.UpdateLiveLessonRequest{
				LessonId:  "lesson-id",
				StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
				Name:      "live-lesson-name",
				TeacherId: "teacher-id",
				Attachments: []*pb.Attachment{{
					Name: "att-name",
					Url:  "att-url",
				}},
				ControlSettings: &pb.ControlSettingLiveLesson{
					TeacherObversers:          []string{"teacher-obversers"},
					Lectures:                  []string{"teacher-teacher"},
					DefaultView:               pb_bob.LIVE_LESSON_VIEW_NONE,
					PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
					UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "missing live lesson view",
		},
		"missing lesson id": {
			ctx: ctx,
			req: &pb.UpdateLiveLessonRequest{
				LessonId:  "",
				StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
				Name:      "live-lesson-name",
				TeacherId: "teacher-id",
				Attachments: []*pb.Attachment{{
					Name: "att-name",
					Url:  "att-url",
				}},
				ControlSettings: &pb.ControlSettingLiveLesson{
					TeacherObversers:          []string{"teacher-obversers"},
					Lectures:                  []string{"teacher-teacher"},
					DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
					PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
					UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "lesson id cannot be empty",
		},
		"missing teacher id": {
			ctx: ctx,
			req: &pb.UpdateLiveLessonRequest{
				LessonId:  "lesson-id",
				StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
				Name:      "live-lesson-name",
				TeacherId: "",
				Attachments: []*pb.Attachment{{
					Name: "att-name",
					Url:  "att-url",
				}},
				ControlSettings: &pb.ControlSettingLiveLesson{
					TeacherObversers:          []string{"teacher-obversers"},
					Lectures:                  []string{"teacher-teacher"},
					DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
					PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
					UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "teacher ids cannot be empty",
		},
		"cannot add teacher of another school": {
			ctx: ctx,
			req: &pb.UpdateLiveLessonRequest{
				LessonId:  "lesson-id",
				StartDate: &types.Timestamp{Seconds: time.Now().Unix()},
				EndDate:   &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
				Name:      "live-lesson-name",
				TeacherId: "teacher-id-another-school",
				Attachments: []*pb.Attachment{{
					Name: "att-name",
					Url:  "att-url",
				}},
				ControlSettings: &pb.ControlSettingLiveLesson{
					TeacherObversers:          []string{"teacher-obversers"},
					Lectures:                  []string{"teacher-teacher"},
					DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
					PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
					UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "cannot add teacher of another school",
		},
		"start date must before end date": {
			ctx: ctx,
			req: &pb.UpdateLiveLessonRequest{
				LessonId:  "lesson-id",
				EndDate:   &types.Timestamp{Seconds: time.Now().Unix()},
				StartDate: &types.Timestamp{Seconds: time.Now().Add(time.Hour * 24 * 30).Unix()},
				Name:      "live-lesson-name",
				TeacherId: "teacher-id",
				Attachments: []*pb.Attachment{{
					Name: "att-name",
					Url:  "att-url",
				}},
				ControlSettings: &pb.ControlSettingLiveLesson{
					TeacherObversers:          []string{"teacher-obversers"},
					Lectures:                  []string{"teacher-teacher"},
					DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
					PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
					UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
				},
			},
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "start date must before end date",
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			rsp, err := courseService.UpdateLiveLesson(context.Background(), testCase.req.(*pb.UpdateLiveLessonRequest))
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, expectedCode, status.Code(err))
			assert.Nil(t, rsp, "expecting nil response")
		})
	}
}

func DeleteLiveLesson_InvalidArgument(t *testing.T) {
	ctx := context.Background()
	courseService := &CourseService{}

	expectedCode := codes.InvalidArgument
	testCases := map[string]TestCase{
		"missing lesson id": {
			ctx: ctx,
			req: &pb.DeleteLiveLessonRequest{
				LessonIds: []string{},
			},
			expectedErrMsg: "missing lesson ids",
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			rsp, err := courseService.DeleteLiveLesson(context.Background(), testCase.req.(*pb.DeleteLiveLessonRequest))
			assert.Equal(t, testCase.expectedCode, status.Code(err), "%s - expecting InvalidArgument", caseName)
			assert.Equal(t, expectedCode, status.Code(err))
			assert.Nil(t, rsp, "expecting nil response")
		})
	}
}
