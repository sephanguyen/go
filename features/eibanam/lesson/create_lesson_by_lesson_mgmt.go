package lesson

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) createCenter(name string) error {
	id, err := s.helper.CreateCenterInDB(name)
	if err != nil {
		return err
	}
	s.AddCenterIDs(id)

	return nil
}

func (s *suite) schoolAdminCreatesANewLessonWithTeachersLearnersCenterMedia(startTimeString, endTimeString, teachingMedium, teachingMethod, savingOpt string) error {
	// create center
	if err := s.createCenter("center test"); err != nil {
		return err
	}

	// create teachers
	for i := 0; i < 2; i++ {
		if err := s.createTeacher(); err != nil {
			return err
		}
	}

	// create students
	for i := 0; i < 2; i++ {
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

	// create student subscription
	courseID := s.CourseIDs[len(s.CourseIDs)-1]
	studentIDWithCourseID := make([]string, 0, len(s.StudentIDs)*2)
	for _, studentID := range s.StudentIDs {
		studentIDWithCourseID = append(studentIDWithCourseID, studentID, courseID)
	}
	if err = s.helper.InsertStudentSubscription(studentIDWithCourseID...); err != nil {
		return fmt.Errorf("could not insert student subscription: %w", err)
	}

	// create media
	mediaIDs, err := s.helper.CreateMediaViaGRPC(schoolAdminCredential.AuthToken, 2)
	if err != nil {
		return fmt.Errorf("could not create media: %w", err)
	}
	s.AddMediaIDs(mediaIDs...)

	startTime, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return err
	}

	endTime, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return err
	}

	req := &bpb.CreateLessonRequest{
		StartTime:        timestamppb.New(startTime),
		EndTime:          timestamppb.New(endTime),
		TeachingMedium:   cpb.LessonTeachingMedium(cpb.LessonTeachingMedium_value[teachingMedium]),
		TeachingMethod:   cpb.LessonTeachingMethod(cpb.LessonTeachingMethod_value[teachingMethod]),
		TeacherIds:       s.TeacherIDs,
		CenterId:         s.CenterIDs[len(s.CenterIDs)-1],
		StudentInfoList:  []*bpb.CreateLessonRequest_StudentInfo{},
		Materials:        []*bpb.Material{},
		SchedulingStatus: bpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	if savingOpt == "save one time" {
		req.SavingOption = &bpb.CreateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		}
	}

	for _, studentID := range s.StudentIDs {
		req.StudentInfoList = append(req.StudentInfoList, &bpb.CreateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
		})
	}

	for _, mediaID := range s.MediaIDs {
		req.Materials = append(req.Materials, &bpb.Material{
			Resource: &bpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = eibanam.ContextWithTokenForGrpcCall(ctx, schoolAdminCredential.AuthToken)
	res, err := bpb.NewLessonManagementServiceClient(s.helper.BobConn).CreateLesson(eibanam.ContextWithToken(ctx, schoolAdminCredential.AuthToken), req)
	s.SessionStack.Push(&Session{
		Request:  req,
		Response: res,
		Error:    err,
	})

	return nil
}

func (s *suite) schoolAdminSeesTheNewLessonOnLessonManagement() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	iSession, err := s.SessionStack.Peek()
	if err != nil {
		return errors.Wrap(err, "s.ResponseStack.Peek()")
	}
	session, ok := iSession.(*Session)
	if !ok {
		return fmt.Errorf("failed to get session")
	}
	if session.Error != nil {
		return fmt.Errorf("lesson is created failed: %v", session.Error)
	}
	lessonID := session.Response.(*bpb.CreateLessonResponse).Id

	lesson, err := s.helper.GetLessonByID(ctx, lessonID)
	if err != nil {
		return err
	}

	expectedLesson := session.Request.(*bpb.CreateLessonRequest)
	if expectedLesson.StudentInfoList[0].CourseId != lesson.CourseID.String {
		return fmt.Errorf("expected CourseId %s but got %s", expectedLesson.StudentInfoList[0].CourseId, lesson.CourseID.String)
	}
	if expectedLesson.TeacherIds[0] != lesson.TeacherID.String {
		return fmt.Errorf("expected TeacherID %s but got %s", expectedLesson.TeacherIds[0], lesson.TeacherID.String)
	}
	if lesson.CreatedAt.Status != pgtype.Present || lesson.CreatedAt.Time.IsZero() {
		return fmt.Errorf("CreatedAt is null")
	}
	if lesson.UpdatedAt.Status != pgtype.Present || lesson.UpdatedAt.Time.IsZero() {
		return fmt.Errorf("UpdatedAt is null")
	}
	if !expectedLesson.StartTime.AsTime().Equal(lesson.StartTime.Time) {
		return fmt.Errorf("expected StartTime %v but got %s", expectedLesson.StartTime.AsTime(), lesson.StartTime.Time)
	}
	if !expectedLesson.EndTime.AsTime().Equal(lesson.EndTime.Time) {
		return fmt.Errorf("expected EndTime %v but got %v", expectedLesson.EndTime.AsTime(), lesson.EndTime.Time)
	}
	if lesson.LessonGroupID.Status != pgtype.Present || len(lesson.LessonGroupID.String) == 0 {
		return fmt.Errorf("LessonGroupID is null")
	}
	if lesson.LessonType.Status != pgtype.Present || len(lesson.LessonType.String) == 0 {
		return fmt.Errorf("LessonType is null")
	}
	if lesson.Status.Status != pgtype.Present || len(lesson.Status.String) == 0 {
		return fmt.Errorf("Status is null")
	}
	if lesson.TeachingModel.Status != pgtype.Present || len(lesson.TeachingModel.String) == 0 {
		return fmt.Errorf("TeachingModel is null")
	}
	if expectedLesson.CenterId != lesson.CenterID.String {
		return fmt.Errorf("expected CenterId %s but got %s", expectedLesson.CenterId, lesson.CenterID.String)
	}

	if expectedLesson.TeachingMedium.String() != lesson.TeachingMedium.String {
		return fmt.Errorf("expected TeachingMedium %s but got %s", expectedLesson.TeachingMedium.String(), lesson.TeachingMedium.String)
	}

	if expectedLesson.TeachingMethod.String() != lesson.TeachingMethod.String {
		return fmt.Errorf("expected TeachingMethod %s but got %s", expectedLesson.TeachingMethod.String(), lesson.TeachingMethod.String)
	}

	if lesson.SchedulingStatus.String != expectedLesson.SchedulingStatus.String() {
		return fmt.Errorf("expected SchedulingStatus %s but got %s", expectedLesson.SchedulingStatus.String(), lesson.SchedulingStatus.String)
	}

	return nil
}
