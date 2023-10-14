package classes

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services/log"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_cloudconvert "github.com/manabie-com/backend/mock/golibs/cloudconvert"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

func TestConvertMedia(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// mediaRepo := new(mock_repositories.MockMediaRepo)
	conversionTaskRepo := new(mock_repositories.MockConversionTaskRepo)
	cloudConvertSvc := new(mock_cloudconvert.MockService)

	svc := &ClassModifierService{
		ConversionTaskRepo: conversionTaskRepo,
		ConversionSvc:      cloudConvertSvc,
	}

	t.Run("medias don't have PDF resources", func(t *testing.T) {
		medias := []*bpb.Media{
			{
				Type:     bpb.MediaType_MEDIA_TYPE_IMAGE,
				Resource: "img1",
			},
			{
				Type:     bpb.MediaType_MEDIA_TYPE_VIDEO,
				Resource: "vid1",
			},
		}
		_, err := svc.ConvertMedia(ctx, &bpb.ConvertMediaRequest{Media: medias})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("medias have PDF resources", func(t *testing.T) {
		urls := []string{
			"https://1",
			"https://2",
			"https://3",
			"https://4",
			"https://5",
		}
		urlsMapUUID := make(map[string]string)
		for _, url := range urls {
			urlsMapUUID[url] = strings.Replace(url, "https://", "", -1)
		}

		cloudConvertSvc.On(
			"CreateConversionTasks",
			ctx,
			urls,
		).Once().Return([]string{"1", "2", "3", "4", "5"}, nil)

		taskEntities := make([]*entities.ConversionTask, 0, len(urls))
		for _, url := range urls {
			e := new(entities.ConversionTask)
			database.AllNullEntity(e)
			e.TaskUUID.Set(urlsMapUUID[url])
			e.ResourceURL.Set(url)
			e.Status.Set(bpb.ConversionTaskStatus_CONVERSION_TASK_STATUS_WAITING.String())

			taskEntities = append(taskEntities, e)
		}
		conversionTaskRepo.On(
			"CreateTasks",
			ctx,
			mock.Anything,
			taskEntities,
		).Once().Return(nil)

		medias := []*bpb.Media{
			{
				Type:     bpb.MediaType_MEDIA_TYPE_IMAGE,
				Resource: "img1",
			},
			{
				Type:     bpb.MediaType_MEDIA_TYPE_VIDEO,
				Resource: "vid1",
			},
			{
				Type:     bpb.MediaType_MEDIA_TYPE_IMAGE,
				Resource: "img2",
			},
			{
				Type:     bpb.MediaType_MEDIA_TYPE_VIDEO,
				Resource: "vid2",
			},
		}
		for _, url := range urls {
			medias = append(medias, &bpb.Media{
				Type:     bpb.MediaType_MEDIA_TYPE_PDF,
				Resource: url,
			})
		}
		_, err := svc.ConvertMedia(ctx, &bpb.ConvertMediaRequest{Media: medias})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestClassModifierService_RetrieveWhiteboardToken(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	whiteboardSvc := new(mock_whiteboard.MockService)

	lessonID := "lesson-id"
	userID := idutil.ULIDNow()
	roomID := "room-id"
	appID := "app-id"
	token := "token"
	course := "course-id"

	lesson := &entities.Lesson{}
	err := multierr.Combine(
		lesson.TeacherID.Set(userID),
		lesson.LessonID.Set(lessonID),
		lesson.CourseID.Set(course),
		lesson.RoomID.Set(roomID))
	if err != nil {
		t.Fatalf("error creating lesson entity: %s", err)
	}

	lessons := []*entities.Lesson{
		lesson,
	}

	ctx = interceptors.ContextWithUserID(ctx, userID)

	classService := &ClassModifierService{
		OldClassService: &ClassServiceABAC{
			ClassService: &ClassService{
				UserRepo:         userRepo,
				LessonMemberRepo: lessonMemberRepo,
				LessonRepo:       lessonRepo,
				WhiteboardSvc:    whiteboardSvc,
				Cfg: &configurations.Config{
					Whiteboard: configs.WhiteboardConfig{AppID: appID},
				},
			},
		},
	}
	testCases := map[string]TestCase{
		"student subscribe token": {
			ctx: ctx,
			req: &bpb.RetrieveWhiteboardTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: &bpb.RetrieveWhiteboardTokenResponse{},
			expectedErr:  fmt.Errorf("rcv.UserRepo.UserGroup: no rows in result set"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(entities.UserGroupStudent, pgx.ErrNoRows)
			},
		},
		"err check StreamSubscriberPermission": {
			ctx: ctx,
			req: &bpb.RetrieveWhiteboardTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: &bpb.RetrieveWhiteboardTokenResponse{},
			expectedErr:  fmt.Errorf("err rcv.LessonMemberRepo.CourseAccessible: tx is closed"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(entities.UserGroupStudent, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, database.Text(userID)).
					Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err no permission": {
			ctx: ctx,
			req: &bpb.RetrieveWhiteboardTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: &bpb.RetrieveWhiteboardTokenResponse{},
			expectedErr:  fmt.Errorf("rpc error: code = PermissionDenied desc = student not allowed to join lesson"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(entities.UserGroupStudent, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, database.Text(userID)).Once().Return([]string{"course_id"}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Lesson{}, nil)
			},
		},
		"student subscribe token success": {
			ctx: ctx,
			req: &bpb.RetrieveWhiteboardTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: &bpb.RetrieveWhiteboardTokenResponse{
				WhiteboardToken: token,
				RoomId:          roomID,
				WhiteboardAppId: appID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				err := multierr.Combine(
					filter.LessonID.Set([]string{lessonID}),
					filter.TeacherID.Set(nil),
					filter.CourseID.Set(nil))
				if err != nil {
					t.Fatalf("error creating LessonFilter entity: %s", err)
				}

				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return(token, nil)
			},
		},
		"lesson have no room id": {
			ctx: ctx,
			req: &bpb.RetrieveWhiteboardTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: &bpb.RetrieveWhiteboardTokenResponse{
				WhiteboardToken: token,
				RoomId:          roomID,
				WhiteboardAppId: appID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {

				filterForStudentGroup := &repositories.LessonFilter{}
				err := multierr.Combine(
					filterForStudentGroup.LessonID.Set([]string{lessonID}),
					filterForStudentGroup.TeacherID.Set(nil),
					filterForStudentGroup.CourseID.Set([]string{course}))
				if err != nil {
					t.Fatalf("error creating filterForStudentGroup entity: %s", err)
				}

				filter := &repositories.LessonFilter{}
				err = multierr.Combine(
					filter.LessonID.Set([]string{lessonID}),
					filter.TeacherID.Set(nil),
					filter.CourseID.Set(nil))
				if err != nil {
					t.Fatalf("error creating LessonFilter entity: %s", err)
				}

				lessonNoRoom := &entities.Lesson{}
				err = multierr.Combine(
					lessonNoRoom.TeacherID.Set(userID),
					lessonNoRoom.LessonID.Set(lessonID),
					lessonNoRoom.CourseID.Set(course))
				if err != nil {
					t.Fatalf("error creating Lesson entity: %s", err)
				}

				userRepo.On("UserGroup", ctx, mock.Anything, lessonNoRoom.TeacherID).Once().Return(entities.UserGroupStudent, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, database.Text(userID)).Once().Return([]string{course}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, filterForStudentGroup).Once().Return([]*entities.Lesson{
					lessonNoRoom,
				}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return([]*entities.Lesson{
					lessonNoRoom,
				}, nil)
				whiteboardSvc.On("CreateRoom", ctx, &whiteboard.CreateRoomRequest{
					Name:     lessonNoRoom.LessonID.String,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{
					UUID: roomID,
				}, nil)
				lessonRepo.On("UpdateRoomID", ctx, mock.Anything, lessonNoRoom.LessonID, database.Text(roomID)).
					Once().Return(nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*bpb.RetrieveWhiteboardTokenRequest)
			res := testCase.expectedResp.(*bpb.RetrieveWhiteboardTokenResponse)
			rsp, err := classService.RetrieveWhiteboardToken(ctx, req)

			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
				assert.EqualError(t, err, testCase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.EqualValues(t, res, rsp)
			}
			mock.AssertExpectationsForObjects(t, userRepo, lessonMemberRepo, lessonRepo, whiteboardSvc)
		})
	}
}

func TestClassModifierService_JoinLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	whiteboardSvc := new(mock_whiteboard.MockService)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	logRepo := new(mock_repositories.MockVirtualClassroomLogRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)

	db := &mock_database.Ext{}
	lessonID := "lesson-id"
	userID := idutil.ULIDNow()
	roomID := "room-id"
	appID := "app-id"
	token := "token"
	courseId := "course-id"
	teacherID := "teacher-id"

	lesson := &entities.Lesson{}
	err := multierr.Combine(
		lesson.TeacherID.Set(userID),
		lesson.LessonID.Set(lessonID),
		lesson.CourseID.Set(courseId),
		lesson.RoomID.Set(roomID))
	if err != nil {
		t.Fatalf("error creating lesson entity: %s", err)
	}

	teacher := &entities.Teacher{}
	_ = teacher.ID.Set(teacherID)
	_ = teacher.SchoolIDs.Set([]int{1})

	lessons := []*entities.Lesson{
		lesson,
	}
	course := &entities.Course{}
	_ = course.ID.Set(courseId)
	courses := entities.Courses{course}

	jsm := new(mock_nats.JetStreamManagement)

	ctx = interceptors.ContextWithUserID(ctx, userID)

	classService := &ClassModifierService{
		OldClassService: &ClassServiceABAC{
			ClassService: &ClassService{
				UserRepo:         userRepo,
				LessonMemberRepo: lessonMemberRepo,
				LessonRepo:       lessonRepo,
				WhiteboardSvc:    whiteboardSvc,
				Cfg: &configurations.Config{
					Whiteboard: configs.WhiteboardConfig{AppID: appID},
				},
				TeacherRepo: teacherRepo,
				CourseRepo:  courseRepo,
				JSM:         jsm,
			},
		},
		VirtualClassRoomLogService: &log.VirtualClassRoomLogService{DB: db, Repo: logRepo},
	}
	testCases := map[string]TestCase{
		"student subscribe token": {
			ctx: ctx,
			req: &bpb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &bpb.JoinLessonResponse{},
			expectedErr:  fmt.Errorf("rcv.UserRepo.UserGroup: no rows in result set"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(entities.UserGroupStudent, pgx.ErrNoRows)
			},
		},
		"student subscribe token success": {
			ctx: ctx,
			req: &bpb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &bpb.JoinLessonResponse{
				WhiteboardToken: token,
				RoomId:          roomID,
				WhiteboardAppId: appID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				err := multierr.Combine(
					filter.LessonID.Set([]string{lessonID}),
					filter.TeacherID.Set(nil),
					filter.CourseID.Set(nil))
				if err != nil {
					t.Fatalf("error creating LessonFilter entity: %s", err)
				}

				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return(token, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
				logRepo.On("GetLatestByLessonID", ctx, db, database.Text("lesson-id")).
					Return(&entities.VirtualClassRoomLog{
						LogID:       database.Text("log-id-1"),
						LessonID:    database.Text("lesson-id-1"),
						IsCompleted: database.Bool(false),
					}, nil).Once()
				logRepo.On("AddAttendeeIDByLessonID", ctx, db, database.Text("lesson-id"), database.Text(userID)).
					Return(nil).Once()
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*bpb.JoinLessonRequest)
			res := testCase.expectedResp.(*bpb.JoinLessonResponse)
			rsp, err := classService.JoinLesson(ctx, req)

			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
				assert.EqualError(t, err, testCase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
				assert.EqualValues(t, res.WhiteboardToken, rsp.WhiteboardToken)
				assert.EqualValues(t, res.RoomId, rsp.RoomId)
				assert.EqualValues(t, res.WhiteboardAppId, rsp.WhiteboardAppId)
				assert.NotEmpty(t, rsp.StmToken)
				assert.NotEmpty(t, rsp.VideoToken)
				assert.NotEmpty(t, rsp.ScreenRecordingToken)
				assert.NotEmpty(t, rsp.StreamToken)
			}
			mock.AssertExpectationsForObjects(t, userRepo, lessonMemberRepo, lessonRepo, whiteboardSvc)
		})
	}
}
