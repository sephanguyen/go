package lesson

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/gogo/protobuf/types"
)

func (s *suite) schoolAdminCreatesANewLessonWithExactInformationAsThatLessonCreatedBefore() error {
	iSession, err := s.SessionStack.Peek()
	if err != nil {
		return fmt.Errorf("s.ResponseStack.Peek(): %v", err)
	}
	session, ok := iSession.(*Session)
	if !ok {
		return fmt.Errorf("failed to get session")
	}

	lesson := session.Request.(*bpb.CreateLiveLessonRequest)
	res, err := s.schoolAdminCreateLesson(lesson)
	s.SessionStack.Push(&Session{
		Request:  lesson,
		Response: res,
		Error:    err,
	})

	return nil
}

func (s *suite) seeNewLessonInRespectiveCourseOnTeacherApp(userName string) error {
	iSession, err := s.SessionStack.Peek()
	if err != nil {
		return fmt.Errorf("s.ResponseStack.Peek(): %v", err)
	}
	session, ok := iSession.(*Session)
	if !ok {
		return fmt.Errorf("failed to get session")
	}
	lessonID := session.Response.(*bpb.CreateLiveLessonResponse).Id
	expectedLesson := session.Request.(*bpb.CreateLiveLessonRequest)

	// get auth token of school admin who logged in before
	teacherCredential, err := s.GetUserCredentialByUserName(userName)
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

func (s *suite) seeNewLessonInLessonListOnLearnerApp(userName string) error {
	iSession, err := s.SessionStack.Peek()
	if err != nil {
		return fmt.Errorf("s.ResponseStack.Peek(): %v", err)
	}
	session, ok := iSession.(*Session)
	if !ok {
		return fmt.Errorf("failed to get session")
	}
	lessonID := session.Response.(*bpb.CreateLiveLessonResponse).Id
	expectedLesson := session.Request.(*bpb.CreateLiveLessonRequest)

	// get auth token of school admin who logged in before
	studentCredential, err := s.GetUserCredentialByUserName(userName)
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
