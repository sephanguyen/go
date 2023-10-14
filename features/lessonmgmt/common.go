package lessonmgmt

import (
	"context"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreateLessonRequest struct {
	*lpb.CreateLessonRequest
}

func DefaultCreateLessonRequest(ctx context.Context, attendanceStatus string) *CreateLessonRequest {
	stepState := StepStateFromContext(ctx)
	now := time.Now().Round(time.Second)

	locationID := stepState.CenterIDs[len(stepState.CenterIDs)-1]

	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(now.Add(-2 * time.Hour)),
		EndTime:         timestamppb.New(now.Add(2 * time.Hour)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      locationID,
		StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*lpb.Material{},
		SavingOption: &lpb.CreateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
		SchedulingStatus: lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	// For lesson group
	switch stepState.CurrentTeachingMethod {
	case "group":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP
			req.ClassId = stepState.CurrentClassId
			req.CourseId = stepState.CurrentCourseID
		}
	case "individual":
		{
			req.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL
		}
	}

	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentID := stepState.StudentIDWithCourseID[i]
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		status := lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT
		if len(attendanceStatus) > 0 {
			status = lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[attendanceStatus])
		}
		addedStudentIDs[studentID] = true
		req.StudentInfoList = append(req.StudentInfoList, &lpb.CreateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: status,
			LocationId:       locationID,
		})
	}

	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &lpb.Material{
			Resource: &lpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}
	return &CreateLessonRequest{req}
}

func (c *CreateLessonRequest) WithLocation(locationID string) *CreateLessonRequest {
	c.LocationId = locationID
	return c
}

func (c *CreateLessonRequest) WithTime(startTime, endTime time.Time) *CreateLessonRequest {
	c.StartTime = timestamppb.New(startTime)
	c.EndTime = timestamppb.New(endTime)
	return c
}

func (c *CreateLessonRequest) WithStudentInfo(studentInfos []*lpb.CreateLessonRequest_StudentInfo) *CreateLessonRequest {
	c.StudentInfoList = studentInfos
	return c
}

func (c *CreateLessonRequest) AddStudent(studentInfos *lpb.CreateLessonRequest_StudentInfo) *CreateLessonRequest {
	c.StudentInfoList = append(c.StudentInfoList, studentInfos)
	return c
}

func (c *CreateLessonRequest) WithTeacher(teacherIds []string) *CreateLessonRequest {
	c.TeacherIds = teacherIds
	return c
}
