package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aValidEvent_Upsert(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	timeNow := time.Now()
	stepState.StartAt = &timestamppb.Timestamp{
		Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
	}
	stepState.EndAt = &timestamppb.Timestamp{
		Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
	}

	ctx, event := s.aEvtSyncStudentPackage_Upsert(ctx)
	stepState.Event = event
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidEvent_Delete(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidEvent_Upsert(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, errPublishEvent := s.sendEventToNatsJS(ctx, "SyncStudentPackageEvent", constants.SubjectSyncStudentPackage)
	if errPublishEvent != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// give some time to let student packages inserted to DB
	time.Sleep(100 * time.Millisecond)

	for index := 0; index < len(stepState.Event.(*npb.EventSyncStudentPackage).StudentPackages); index++ {
		stepState.Event.(*npb.EventSyncStudentPackage).StudentPackages[index].ActionKind = npb.ActionKind_ACTION_KIND_DELETED
	}

	if ctx, err := s.sendEventToNatsJS(ctx, "SyncStudentPackageEvent", constants.SubjectSyncStudentPackage); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aEvtSyncStudentPackage_Upsert(ctx context.Context) (context.Context, *npb.EventSyncStudentPackage) {
	stepState := StepStateFromContext(ctx)
	var StudentPackages []*npb.EventSyncStudentPackage_StudentPackage
	for i := 0; i < 3; i++ {
		StudentPackages = append(StudentPackages, &npb.EventSyncStudentPackage_StudentPackage{
			StudentId:  idutil.ULIDNow(),
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			Packages: []*npb.EventSyncStudentPackage_Package{
				{
					CourseIds: stepState.CourseIDs,
					StartDate: stepState.StartAt,
					EndDate:   stepState.EndAt,
				},
			},
		})
	}
	event := &npb.EventSyncStudentPackage{
		StudentPackages: StudentPackages,
	}

	return StepStateToContext(ctx, stepState), event
}

func (s *suite) getCourseOfStudent(ctx context.Context) (context.Context, map[string][]string) {
	stepState := StepStateFromContext(ctx)
	courses := make(map[string]bool)
	courseOfStudents := make(map[string][]string)
	for _, studentPackage := range stepState.Event.(*npb.EventSyncStudentPackage).StudentPackages {
		studentId := studentPackage.StudentId
		for _, item := range studentPackage.Packages {
			for _, course := range item.CourseIds {
				if isExist := courses[fmt.Sprintf("%s-%s", studentId, course)]; !isExist {
					courses[fmt.Sprintf("%s-%s", studentId, course)] = true
					courseOfStudents[studentId] = append(courseOfStudents[studentId], course)
				}
			}
		}
	}
	return StepStateToContext(ctx, stepState), courseOfStudents
}

func (s *suite) eurekaMustCreateCourseStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, courseOfStudents := s.getCourseOfStudent(ctx)
	totalCourses := 0
	studentIds := []string{}
	for studentId, courses := range courseOfStudents {
		studentIds = append(studentIds, studentId)
		totalCourses += len(courses)
	}

	count := 0
	query := fmt.Sprintf("SELECT count(*) FROM course_students WHERE student_id = ANY($1)")
	if err := try.Do(func(attempt int) (retry bool, err error) {
		err = s.DB.QueryRow(ctx, query, studentIds).Scan(&count)
		if err != nil {
			return true, err
		}
		if count != totalCourses {
			time.Sleep(1 * time.Second)
			return true, fmt.Errorf("Eureka does not create course student correctly")
		}
		return attempt < 5, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustUpdateCoursestudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, courseOfStudents := s.getCourseOfStudent(ctx)
	totalCourses := 0
	studentIds := []string{}
	for studentId, courses := range courseOfStudents {
		studentIds = append(studentIds, studentId)
		totalCourses += len(courses)
	}

	count := 0
	query := fmt.Sprintf("SELECT count(*) FROM course_students WHERE student_id = ANY($1) AND deleted_at is not null")
	if err := try.Do(func(attempt int) (retry bool, err error) {
		err = s.DB.QueryRow(ctx, query, studentIds).Scan(&count)
		if err != nil {
			return false, err
		}

		if count != totalCourses {
			time.Sleep(1 * time.Second)
			return true, fmt.Errorf("Eureka does not update course student correctly")
		}
		return attempt < 5, err
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

const stmtGetNumberOfCoursesStudentStudyPlan = `SELECT count(*) FROM study_plan_items spi
		JOIN student_study_plans ssp ON spi.study_plan_id = ssp.study_plan_id AND ssp.student_id =$1
		JOIN study_plans sp ON spi.study_plan_id = ssp.study_plan_id
		WHERE sp.course_id =ANY($2) AND sp.deleted_at IS NULL AND spi.deleted_at IS NULL AND ssp.deleted_at IS NULL`

func (s *suite) getNumberOfCourseStudentStudyPlan(ctx context.Context, studentID string, courseIds []string) (context.Context, int, error) {
	stepState := StepStateFromContext(ctx)
	var count int
	err := db.QueryRow(ctx, stmtGetNumberOfCoursesStudentStudyPlan, &studentID, &courseIds).Scan(&count)
	return StepStateToContext(ctx, stepState), count, err
}

func (s *suite) ourSystemMustRemoveAllCourseStudentStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, courseOfStudents := s.getCourseOfStudent(ctx)

	for studentID, courseIds := range courseOfStudents {
		ctx, count, err := s.getNumberOfCourseStudentStudyPlan(ctx, studentID, courseIds)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not all student's study plan item in course deleted")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

const stmtNumberStudyPlansEachCourseStudentPlt = `SELECT count(*) FROM student_study_plans ssp
		JOIN study_plans sp ON ssp.study_plan_id = sp.study_plan_id
		WHERE student_id = $1 AND sp.master_study_plan_id = ANY($2) AND sp.deleted_at IS NULL and ssp.deleted_at IS NULL`

func (s *suite) getNumberOfNewStudyPlanForEachCoursesStudent(ctx context.Context, studentID string, studyPlanIDs []string) (context.Context, int, error) {
	stepState := StepStateFromContext(ctx)
	var count int
	err := s.DB.QueryRow(ctx, stmtNumberStudyPlansEachCourseStudentPlt, &studentID, &studyPlanIDs).Scan(&count)
	return StepStateToContext(ctx, stepState), count, err
}

func (s *suite) ourSystemMustCreateNewStudyPlanForEachCourseStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, courseOfStudents := s.getCourseOfStudent(ctx)
	for studentID, courses := range courseOfStudents {
		r := repositories.CourseStudyPlanRepo{}
		courseStudyPlan, err := r.FindByCourseIDs(ctx, s.DB, database.TextArray(courses))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studyPlanIDs := make([]string, 0, len(courseStudyPlan))
		for _, csp := range courseStudyPlan {
			studyPlanIDs = append(studyPlanIDs, csp.StudyPlanID.String)
		}
		ctx, count, err := s.getNumberOfNewStudyPlanForEachCoursesStudent(ctx, studentID, studyPlanIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if count != len(studyPlanIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect student %s has %d study plan but got %d", studentID, len(studyPlanIDs), count)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
