package classes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"

	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAbacStudentRetrieveStreamToken_Error(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := new(mock_repositories.MockLessonRepo)
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	classService := &ClassService{
		LessonRepo:       lessonRepo,
		LessonMemberRepo: lessonMemberRepo,
	}
	classServiceABAC := &ClassServiceABAC{
		classService,
	}
	lessonID := "lesson-id"
	studentID := "student-id"

	lesson := &entities.Lesson{}
	_ = lesson.TeacherID.Set(studentID)
	_ = lesson.LessonID.Set(lessonID)

	lessons := []*entities.Lesson{
		lesson,
	}

	var courses pgtype.TextArray
	_ = courses.Set([]string{"course-1", "course-2"})

	var classIDs pgtype.Int4Array
	_ = classIDs.Set([]int{1})

	emptyLessons := []*entities.Lesson{}

	ctx = interceptors.ContextWithUserID(ctx, studentID)
	testCases := map[string]TestCase{
		"error class repo": {
			ctx: ctx,
			req: &pb.StudentRetrieveStreamTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("ClassServiceABAC.StudentRetrieveStreamToken: err rcv.LessonMemberRepo.CourseAccessible: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, lesson.TeacherID).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"error lesson repo": {
			ctx: ctx,
			req: &pb.StudentRetrieveStreamTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("ClassServiceABAC.StudentRetrieveStreamToken: err rcv.LessonRepo.Find: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, lesson.TeacherID).Once().Return([]string{"course_id"}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, pgx.ErrNoRows)
			},
		},
		"permission deny": {
			ctx: ctx,
			req: &pb.StudentRetrieveStreamTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, "student is not assigned to course"),
			setup: func(ctx context.Context) {
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, lesson.TeacherID).Once().Return([]string{"course_id"}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(emptyLessons, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.StudentRetrieveStreamTokenRequest)
			rsp, err := classServiceABAC.StudentRetrieveStreamToken(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
				assert.EqualError(t, err, testCase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestAbacEndLiveLesson_Error(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := new(mock_repositories.MockLessonRepo)
	classService := &ClassService{
		LessonRepo: lessonRepo,
	}
	classServiceABAC := &ClassServiceABAC{
		classService,
	}
	lessonID := "teacher-id"
	teacherID := "teacher-id"

	emptyLessons := []*entities.Lesson{}
	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	testCases := map[string]TestCase{
		"error find lesson": {
			ctx: ctx,
			req: &pb.EndLiveLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"permission deny": {
			ctx: ctx,
			req: &pb.EndLiveLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, ""),
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(emptyLessons, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.EndLiveLessonRequest)
			rsp, err := classServiceABAC.EndLiveLesson(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestClassService_JoinLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := new(mock_repositories.MockLessonRepo)
	classRepo := new(mock_repositories.MockClassRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	schoolRepo := new(mock_repositories.MockSchoolRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	jsm := new(mock_nats.JetStreamManagement)

	whiteboardSvc := new(mock_whiteboard.MockService)

	var courseIDs pgtype.TextArray
	_ = courseIDs.Set([]string{"course-1", "course-2"})

	var classIDs pgtype.Int4Array
	_ = classIDs.Set([]int{1})

	classService := &ClassService{
		Cfg:             &configurations.Config{},
		UserRepo:        userRepo,
		LessonRepo:      lessonRepo,
		CourseClassRepo: courseClassRepo,
		ClassRepo:       classRepo,
		TeacherRepo:     teacherRepo,
		SchoolRepo:      schoolRepo,
		CourseRepo:      courseRepo,
		WhiteboardSvc:   whiteboardSvc,
		JSM:             jsm,
	}
	lessonID := "teacher-id"
	teacherID := "teacher-id"
	roomID := "room-id"

	lesson := &entities.Lesson{}
	_ = lesson.TeacherID.Set(teacherID)
	_ = lesson.LessonID.Set(lessonID)
	_ = lesson.CourseID.Set("course-2")
	_ = lesson.RoomID.Set(roomID)

	lessons := []*entities.Lesson{
		lesson,
	}

	teacher := &entities.Teacher{}
	_ = teacher.ID.Set(teacherID)
	_ = teacher.SchoolIDs.Set([]int{1})

	course := &entities.Course{}
	_ = course.ID.Set("course-2")
	courses := entities.Courses{course}

	invalidCourse := &entities.Course{}
	_ = invalidCourse.ID.Set("course-1")
	invalidCourses := entities.Courses{invalidCourse}

	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	testCases := map[string]TestCase{
		"student subscribe token": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupStudent, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		"lesson have no room id": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupStudent, nil)
				whiteboardSvc.On("CreateRoom", ctx, &whiteboard.CreateRoomRequest{
					Name:     lesson.LessonID.String,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{
					UUID: roomID,
				}, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
			},
		},
		"error finding lesson": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, pgx.ErrNoRows)
			},
		},
		"invalid RetrieveCourses": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				teacherRepo.On("FindByID", ctx, mock.Anything, lesson.TeacherID).Once().Return(teacher, nil)
				schoolRepo.On("RetrieveCountries", ctx, mock.Anything, mock.Anything).Once().Return([]string{"COUNTRY_VN"}, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(invalidCourses, pgx.ErrNoRows)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
			},
		},
		"invalid subscribe token": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, ""),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				teacherRepo.On("FindByID", ctx, mock.Anything, lesson.TeacherID).Once().Return(teacher, nil)
				schoolRepo.On("RetrieveCountries", ctx, mock.Anything, mock.Anything).Once().Return([]string{"COUNTRY_VN"}, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(entities.Courses{}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
			},
		},
		"valid token": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				teacherRepo.On("FindByID", ctx, mock.Anything, lesson.TeacherID).Once().Return(teacher, nil)
				schoolRepo.On("RetrieveCountries", ctx, mock.Anything, mock.Anything).Once().Return([]string{"COUNTRY_VN"}, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.JoinLessonRequest)
			rsp, err := classService.JoinLesson(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestClassServiceABAC_JoinLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo := new(mock_repositories.MockUserRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)

	c := &ClassService{
		UserRepo:         userRepo,
		LessonMemberRepo: lessonMemberRepo,
		LessonRepo:       lessonRepo,
	}

	classService := &ClassServiceABAC{
		ClassService: c,
	}

	lessonID := idutil.ULIDNow()
	studentID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, studentID)
	testCases := map[string]TestCase{
		"student subscribe token": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{},
			expectedErr:  fmt.Errorf("rcv.UserRepo.UserGroup: no rows in result set"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(studentID)).Once().Return(entities.UserGroupStudent, pgx.ErrNoRows)
			},
		},
		"err check StreamSubscriberPermission": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{},
			expectedErr:  fmt.Errorf("err rcv.LessonMemberRepo.CourseAccessible: tx is closed"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(studentID)).Once().Return(entities.UserGroupStudent, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, database.Text(studentID)).
					Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err no permission": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{},
			expectedErr:  fmt.Errorf("rpc error: code = PermissionDenied desc = student not allowed to join lesson"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(studentID)).Once().Return(entities.UserGroupStudent, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, database.Text(studentID)).Once().Return([]string{"course_id"}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Lesson{}, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.JoinLessonRequest)
			rsp, err := classService.JoinLesson(ctx, req)

			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
				assert.EqualError(t, err, testCase.expectedErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
