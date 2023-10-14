package lesson

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/gogo/protobuf/types"
	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) schoolAdminCreateLesson(req *bpb.CreateLiveLessonRequest) (*bpb.CreateLiveLessonResponse, error) {
	// get auth token of school admin who logged in before
	schoolAdminCredential, err := s.GetUserCredentialByUserGroup(constant.UserGroupSchoolAdmin)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = eibanam.ContextWithTokenForGrpcCall(ctx, schoolAdminCredential.AuthToken)
	res, err := bpb.NewLessonModifierServiceClient(s.helper.BobConn).CreateLiveLesson(eibanam.ContextWithToken(ctx, schoolAdminCredential.AuthToken), req)

	return res, err
}

func (s *suite) schoolAdminCreatesLessonWithAllRequiredFields() error {
	// create teacher
	if len(s.TeacherIDs) == 0 {
		if err := s.createTeacher(); err != nil {
			return err
		}
	}

	// create student
	if len(s.StudentIDs) == 0 {
		if err := s.createStudent(); err != nil {
			return err
		}
	}

	// get auth token of school admin who logged in before
	schoolAdminCredential, err := s.GetUserCredentialByUserGroup(constant.UserGroupSchoolAdmin)
	if err != nil {
		return err
	}

	// create course
	if len(s.CourseIDs) == 0 {
		coursesReq, err := s.helper.CreateACourseViaGRPC(schoolAdminCredential.AuthToken, s.CurrentSchoolID)
		if err != nil {
			return fmt.Errorf("could not create new course: %v", err)
		}
		s.AddCourseIDs(coursesReq.Courses[0].Id)
	}

	now := time.Now()
	now = now.Add(-time.Duration(now.Nanosecond()) * time.Nanosecond)
	req := &bpb.CreateLiveLessonRequest{
		Name:       "lesson name " + idutil.ULIDNow(),
		StartTime:  timestamppb.New(now.Add(-1 * time.Hour)),
		EndTime:    timestamppb.New(now.Add(1 * time.Hour)),
		TeacherIds: s.TeacherIDs,
		CourseIds:  s.CourseIDs,
		LearnerIds: s.StudentIDs,
	}

	res, err := s.schoolAdminCreateLesson(req)
	s.SessionStack.Push(&Session{
		Request:  req,
		Response: res,
		Error:    err,
	})

	return nil
}

func (s *suite) schoolAdminSeesNewLessonOnCMS() error {
	// Pre-setup for hasura query using admin secret
	if err := eibanam.TrackTableForHasuraQuery(
		s.helper.HasuraAdminUrl,
		"lessons",
		"lessons_courses",
		"courses",
		"lessons_teachers",
		"teachers",
		"users",
		"lesson_members",
		"students",
		"school_admins",
	); err != nil {
		return errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := eibanam.CreateSelectPermissionForHasuraQuery(
		s.helper.HasuraAdminUrl,
		constant.UserGroupSchoolAdmin,
		"lessons",
		"lessons_courses",
		"courses",
		"lessons_teachers",
		"teachers",
		"users",
		"lesson_members",
		"students",
		"school_admins",
	); err != nil {
		return errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($lesson_id: String!) {
		  lessons(where: {lesson_id: {_eq: $lesson_id}}) {
			lesson_id
			status
			start_time
			end_time
			name
			lesson_group_id
			lessons_courses {
			  course {
				course_id
				name
			  }
			}
			lessons_teachers {
			  teacher {
				users {
				  name
				  user_id
				  email
				}
			  }
			}
			lesson_members {
			  user {
				name
				user_id
				email
				student {
				  student_id
				  current_grade
				  enrollment_status
				}
			  }
			}
		  }
		}
`
	if err := eibanam.AddQueryToAllowListForHasuraQuery(s.helper.HasuraAdminUrl, query); err != nil {
		return fmt.Errorf("addQueryToAllowListForHasuraQuery(): %v", err)
	}

	var profileQuery struct {
		Lessons []struct {
			LessonID       string    `graphql:"lesson_id"`
			Status         string    `graphql:"status"`
			StartTime      time.Time `graphql:"start_time"`
			EndTime        time.Time `graphql:"end_time"`
			Name           string    `graphql:"name"`
			LessonGroupID  string    `graphql:"lesson_group_id"`
			LessonsCourses []struct {
				Courses struct {
					CourseID string `graphql:"course_id"`
					Name     string `graphql:"name"`
				} `graphql:"course"`
			} `graphql:"lessons_courses"`
			LessonsTeachers []struct {
				Teacher struct {
					Users struct {
						Name   string `graphql:"name"`
						UserID string `graphql:"user_id"`
						Email  string `graphql:"email"`
					} `graphql:"users"`
				} `graphql:"teacher"`
			} `graphql:"lessons_teachers"`
			LessonMembers []struct {
				User struct {
					Name    string `graphql:"name"`
					UserID  string `graphql:"user_id"`
					Email   string `graphql:"email"`
					Student struct {
						StudentID        string `graphql:"student_id"`
						CurrentGrade     int32  `graphql:"current_grade"`
						EnrollmentStatus string `graphql:"enrollment_status"`
					} `graphql:"student"`
				} `graphql:"user"`
			} `graphql:"lesson_members"`
		} `graphql:"lessons(where: {lesson_id: {_eq: $lesson_id}})"`
	}

	iSession, err := s.SessionStack.Peek()
	if err != nil {
		return errors.Wrap(err, "s.ResponseStack.Peek()")
	}
	session, ok := iSession.(*Session)
	if !ok {
		return fmt.Errorf("failed to get session")
	}

	if session.Error != nil {
		return fmt.Errorf("lesson is created failed: %v, %v", session.Error, s.StudentIDs)
	}
	lessonID := session.Response.(*bpb.CreateLiveLessonResponse).Id
	variables := map[string]interface{}{
		"lesson_id": graphql.String(lessonID),
	}

	// get auth token of school admin who logged in before
	schoolAdminCredential, err := s.GetUserCredentialByUserGroup(constant.UserGroupSchoolAdmin)
	if err != nil {
		return err
	}

	// Setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = eibanam.ContextWithToken(ctx, schoolAdminCredential.AuthToken)

	err = eibanam.QueryHasura(ctx, s.helper.HasuraAdminUrl, &profileQuery, variables)
	if err != nil {
		return fmt.Errorf("queryHasura: %v", err)
	}
	if len(profileQuery.Lessons) == 0 {
		return errors.New("failed to query lesson")
	}

	// check value lesson in db by hasura
	actualLessons := profileQuery.Lessons
	expectedLesson := session.Request.(*bpb.CreateLiveLessonRequest)
	if len(actualLessons) != 1 {
		return fmt.Errorf(`expect created lesson has %v but got %v`, 1, len(actualLessons))
	}

	actualLesson := actualLessons[0]
	switch {
	case actualLesson.LessonID != lessonID:
		return fmt.Errorf(`expect created lesson has "id": %v but got %v`, lessonID, actualLesson.LessonID)
	case actualLesson.Name != expectedLesson.Name:
		return fmt.Errorf(`expect created lesson has "name": %v but got %v`, expectedLesson.Name, actualLesson.Name)
	case !actualLesson.EndTime.Equal(expectedLesson.EndTime.AsTime()):
		return fmt.Errorf(`expect created lesson has "end time": %v but got %v`, expectedLesson.EndTime.AsTime(), actualLesson.EndTime)
	case !actualLesson.StartTime.Equal(expectedLesson.StartTime.AsTime()):
		return fmt.Errorf(`expect created lesson has "start time": %v but got %v`, expectedLesson.StartTime.AsTime(), actualLesson.StartTime)
	case len(actualLesson.LessonsCourses) != len(expectedLesson.CourseIds):
		return fmt.Errorf(`expect created lesson has %v "courses" but got %v`, len(expectedLesson.CourseIds), len(actualLesson.LessonsCourses))
	case actualLesson.LessonsCourses[0].Courses.CourseID != expectedLesson.CourseIds[0]:
		return fmt.Errorf(`expect created lesson has "courses": %v but got %v`, expectedLesson.CourseIds[0], actualLesson.LessonsCourses[0].Courses.CourseID)
	case len(actualLesson.LessonsTeachers) != len(expectedLesson.TeacherIds):
		return fmt.Errorf(`expect created lesson has %v "teachers" but got %v`, len(expectedLesson.TeacherIds), len(actualLesson.LessonsTeachers))
	case actualLesson.LessonsTeachers[0].Teacher.Users.UserID != expectedLesson.TeacherIds[0]:
		return fmt.Errorf(`expect created lesson has "teachers": %v but got %v`, expectedLesson.TeacherIds[0], actualLesson.LessonsTeachers[0].Teacher.Users.UserID)
	case len(actualLesson.LessonMembers) != len(expectedLesson.LearnerIds):
		return fmt.Errorf(`expect created lesson has %v "learners" but got %v`, len(expectedLesson.LearnerIds), len(actualLesson.LessonMembers))
	case actualLesson.LessonMembers[0].User.UserID != expectedLesson.LearnerIds[0]:
		return fmt.Errorf(`expect created lesson has "learners": %v but got %v`, expectedLesson.LearnerIds[0], actualLesson.LessonMembers[0].User.UserID)
	}

	return nil
}

func (s *suite) teacherSeeNewLessonInRespectiveCourseOnTeacherApp() error {
	iSession, err := s.SessionStack.Peek()
	if err != nil {
		return errors.Wrap(err, "s.ResponseStack.Peek()")
	}
	session, ok := iSession.(*Session)
	if !ok {
		return errors.New("failed to get session")
	}
	lessonID := session.Response.(*bpb.CreateLiveLessonResponse).Id
	expectedLesson := session.Request.(*bpb.CreateLiveLessonRequest)

	// get auth token of school admin who logged in before
	teacherCredential, err := s.GetUserCredentialByUserGroup(constant.UserGroupTeacher)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = eibanam.ContextWithTokenForGrpcCall(ctx, teacherCredential.AuthToken)

	req := &pb.RetrieveLiveLessonRequest{
		CourseIds: expectedLesson.CourseIds,
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  0,
		},
		From: &types.Timestamp{Seconds: time.Time{}.Unix()},
		To:   &types.Timestamp{Seconds: time.Now().Add(time.Hour).Unix()},
	}
	res, err := pb.NewCourseClient(s.helper.BobConn).RetrieveLiveLesson(eibanam.ContextWithToken(ctx, teacherCredential.AuthToken), req)
	var actual *pb.Lesson
	for _, lesson := range res.Lessons {
		if lesson.LessonId == lessonID {
			actual = lesson
			break
		}
	}
	if actual == nil {
		return fmt.Errorf("teacher could not see new lesson %s in respective course on teacher app", lessonID)
	}

	switch {
	case actual.Topic.Name != expectedLesson.Name:
		return fmt.Errorf(`expect created lesson has "name": %v but got %v`, expectedLesson.Name, actual.Topic.Name)
	case !actual.StartTime.Equal(&types.Timestamp{Seconds: expectedLesson.StartTime.Seconds}):
		return fmt.Errorf(`expect created lesson has "start time": %v but got %v`, expectedLesson.StartTime.AsTime(), actual.StartTime)
	case !actual.EndTime.Equal(&types.Timestamp{Seconds: expectedLesson.EndTime.Seconds}):
		return fmt.Errorf(`expect created lesson has "end time": %v but got %v`, expectedLesson.EndTime.AsTime(), actual.EndTime)
	case actual.CourseId != expectedLesson.CourseIds[0]:
		return fmt.Errorf(`expect created lesson has "course": %v but got %v`, expectedLesson.CourseIds[0], actual.CourseId)
	case actual.Teacher[0].UserId != expectedLesson.TeacherIds[0]:
		return fmt.Errorf(`expect created lesson has "teacher": %v but got %v`, expectedLesson.TeacherIds[0], actual.Teacher[0].UserId)
	}

	return nil
}

func (s *suite) studentSeeNewLessonInLessonListOnLearnerApp() error {
	iSession, err := s.SessionStack.Peek()
	if err != nil {
		return errors.Wrap(err, "s.ResponseStack.Peek()")
	}
	session, ok := iSession.(*Session)
	if !ok {
		return errors.New("failed to get session")
	}
	lessonID := session.Response.(*bpb.CreateLiveLessonResponse).Id
	expectedLesson := session.Request.(*bpb.CreateLiveLessonRequest)

	// get auth token of school admin who logged in before
	studentCredential, err := s.GetUserCredentialByUserGroup(constant.UserGroupStudent)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = eibanam.ContextWithTokenForGrpcCall(ctx, studentCredential.AuthToken)

	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  0,
		},
		From: &types.Timestamp{Seconds: time.Time{}.Unix()},
		To:   &types.Timestamp{Seconds: time.Now().Add(time.Hour).Unix()},
	}
	res, err := pb.NewCourseClient(s.helper.BobConn).RetrieveLiveLesson(eibanam.ContextWithToken(ctx, studentCredential.AuthToken), req)
	var actual *pb.Lesson
	for _, lesson := range res.Lessons {
		if lesson.LessonId == lessonID {
			actual = lesson
			break
		}
	}
	if actual == nil {
		return fmt.Errorf("student could not see new lesson %s in respective course on learner app", lessonID)
	}

	switch {
	case actual.Topic.Name != expectedLesson.Name:
		return fmt.Errorf(`expect created lesson has "name": %v but got %v`, expectedLesson.Name, actual.Topic.Name)
	case !actual.StartTime.Equal(&types.Timestamp{Seconds: expectedLesson.StartTime.Seconds}):
		return fmt.Errorf(`expect created lesson has "start time": %v but got %v`, expectedLesson.StartTime.AsTime(), actual.StartTime)
	case !actual.EndTime.Equal(&types.Timestamp{Seconds: expectedLesson.EndTime.Seconds}):
		return fmt.Errorf(`expect created lesson has "end time": %v but got %v`, expectedLesson.EndTime.AsTime(), actual.EndTime)
	case actual.CourseId != expectedLesson.CourseIds[0]:
		return fmt.Errorf(`expect created lesson has "course": %v but got %v`, expectedLesson.CourseIds[0], actual.CourseId)
	case actual.Teacher[0].UserId != expectedLesson.TeacherIds[0]:
		return fmt.Errorf(`expect created lesson has "teacher": %v but got %v`, expectedLesson.TeacherIds[0], actual.Teacher[0].UserId)
	}

	return nil
}
