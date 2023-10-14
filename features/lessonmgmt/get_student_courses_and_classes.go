package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) insertStudentToClassMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	classesByStudentID := make(map[string][]*lpb.GetStudentCoursesAndClassesResponse_Class)
	courseIDsByStudentID := make(map[string][]*cpb.Course)
	for i := 0; i < len(stepState.StudentIDWithCourseID)/2+1; i += 2 {
		studentID := stepState.StudentIDWithCourseID[i]
		courseID := stepState.StudentIDWithCourseID[i+1]
		courseIDsByStudentID[studentID] = append(courseIDsByStudentID[studentID], &cpb.Course{
			Info: &cpb.ContentBasicInfo{
				Id: courseID,
			},
		})
		for _, class := range stepState.ImportedClass {
			if class.CourseID == courseID {
				stmt := `INSERT INTO class_member (class_member_id,class_id,user_id) VALUES($1,$2,$3) 
				ON CONFLICT DO NOTHING`
				_, err := s.BobDB.Exec(ctx, stmt, idutil.ULIDNow(), class.ClassID, studentID)
				if err != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert class_member with `studentID:%s`, %v", studentID, err)
				}
				classesByStudentID[studentID] = append(classesByStudentID[studentID], &lpb.GetStudentCoursesAndClassesResponse_Class{
					ClassId:  class.ClassID,
					CourseId: courseID,
				})
			}
		}
	}

	if stepState.StudentCoursesClasses == nil {
		stepState.StudentCoursesClasses = make(map[string]*lpb.GetStudentCoursesAndClassesResponse)
	}
	for studentID := range courseIDsByStudentID {
		stepState.StudentCoursesClasses[studentID] = &lpb.GetStudentCoursesAndClassesResponse{
			StudentId: studentID,
			Classes:   classesByStudentID[studentID],
			Courses:   courseIDsByStudentID[studentID],
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) GetStudentsCoursesAndClasses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for studentID := range stepState.StudentCoursesClasses {
		req := &lpb.GetStudentCoursesAndClassesRequest{
			StudentId: studentID,
		}
		stepState.Request = req
		res, err := lpb.NewStudentSubscriptionServiceClient(s.LessonMgmtConn).
			GetStudentCoursesAndClasses(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
		if err != nil {
			stepState.ResponseErr = err
		}
		stepState.Responses = append(stepState.Responses, res)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) MustGetCorrectCoursesAndClassOfStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, response := range stepState.Responses {
		res := response.(*lpb.GetStudentCoursesAndClassesResponse)
		if v, ok := stepState.StudentCoursesClasses[res.StudentId]; ok {
			if len(v.Courses) != len(res.Courses) {
				return ctx, fmt.Errorf("expected %d courses but got %d", len(v.Courses), len(res.Courses))
			}
			if len(v.Classes) != len(res.Classes) {
				return ctx, fmt.Errorf("expected %d classes but got %d", len(v.Classes), len(res.Classes))
			}
			// check courses list
		LoopLabel:
			for _, course := range v.Courses {
				for _, actual := range res.Courses {
					if actual.Info.Id == course.Info.Id {
						if len(actual.Info.Name) == 0 {
							return ctx, fmt.Errorf("course's name is empty %s", course.Info.Id)
						}
						continue LoopLabel
					}
				}
				return ctx, fmt.Errorf("could not found course info %s for student id %s", course.Info.Id, res.StudentId)
			}

			// check classes list
		LoopLabel2:
			for _, class := range v.Classes {
				for _, actual := range res.Classes {
					if actual.ClassId == class.ClassId {
						if len(actual.Name) == 0 {
							return ctx, fmt.Errorf("class's name is empty %s", class.ClassId)
						}
						if actual.CourseId != class.CourseId {
							return ctx, fmt.Errorf("expected course %s of class %s but got %s", class.CourseId, class.ClassId, actual.CourseId)
						}
						continue LoopLabel2
					}
				}
				return ctx, fmt.Errorf("could not found class info of class %s for student id %s", class.ClassId, res.StudentId)
			}
		} else {
			return ctx, fmt.Errorf("could not expected response for student id %s", res.StudentId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
