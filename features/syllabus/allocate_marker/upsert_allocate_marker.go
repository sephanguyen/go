package allocate_marker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) someCourseTypes(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	listCourseType := []struct {
		ID   string
		name string
	}{
		{ID: "1" + idutil.ULIDNow(), name: "name 1" + idutil.ULIDNow()},
		{ID: "2" + idutil.ULIDNow(), name: "name 2" + idutil.ULIDNow()},
	}
	for _, c := range listCourseType {
		stmt := `INSERT INTO course_type (course_type_id,name,created_at,updated_at) VALUES($1,$2,now(),now()) 
		ON CONFLICT ON CONSTRAINT course_type__pk DO UPDATE SET deleted_at = null`
		_, err := s.BobDBTrace.Exec(ctx, stmt, c.ID,
			c.name)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert course type with `id:%s`, %v", c.ID, err)
		}
		stepState.CourseTypeIDs = append(stepState.CourseTypeIDs, c.ID)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func genPbCourse(locationID, courseTypeID string) *mpb.UpsertCoursesRequest_Course {
	r := &mpb.UpsertCoursesRequest_Course{
		Id:             "course_id_" + idutil.ULIDNow(),
		Name:           "course name " + idutil.ULIDNow(),
		SchoolId:       constants.ManabieSchool,
		Icon:           "link-icon",
		LocationIds:    []string{locationID},
		TeachingMethod: mpb.CourseTeachingMethod_COURSE_TEACHING_METHOD_GROUP,
		CourseType:     courseTypeID,
	}
	return r
}

func (s *Suite) userUpsertCoursesDataWithLocation(ctx context.Context, locationID, courseTypeID string) context.Context {
	stepState := utils.StepStateFromContext[StepState](ctx)

	course := genPbCourse(locationID, courseTypeID)
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.MasterMgmtConn).UpsertCourses(s.AuthHelper.SignedCtx(ctx, stepState.Token), &mpb.UpsertCoursesRequest{
		Courses: []*mpb.UpsertCoursesRequest_Course{course},
	})

	stepState.CourseIDs = append(stepState.CourseIDs, course.Id)
	return utils.StepStateToContext(ctx, stepState)
}

func (s *Suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	listLocation := []struct {
		locationID       string
		name             string
		parentLocationID string
		archived         bool
		expected         bool
	}{ // satisfied
		{locationID: "1", archived: false, expected: true},
		{locationID: "2", parentLocationID: "1", archived: false, expected: true},
		{locationID: "3", parentLocationID: "2", archived: false, expected: true},
		{locationID: "7", archived: false, expected: true},
		// unsatisfied
		{locationID: "4", archived: true},
		{locationID: "5", parentLocationID: "4", archived: false, expected: false},
		{locationID: "6", parentLocationID: "5", archived: false, expected: false},
		{locationID: "8", parentLocationID: "7", archived: true, expected: false},
	}
	addedRandom := "-" + idutil.ULIDNow()

	for _, l := range listLocation {
		l.locationID += addedRandom
		if l.parentLocationID != "" {
			l.parentLocationID += addedRandom
		}

		stmt := `INSERT INTO locations (location_id,name,parent_location_id, is_archived) VALUES($1,$2,$3,$4) 
				ON CONFLICT ON CONSTRAINT locations_pkey DO NOTHING`
		_, err := s.BobDBTrace.Exec(ctx, stmt, l.locationID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validUserAccessPath(ctx context.Context, userID, locationID string) error {
	userAccessPathRepo := &repository.UserAccessPathRepo{}
	userAccessPathEnt := &entity.UserAccessPath{}
	database.AllNullEntity(userAccessPathEnt)

	if err := multierr.Combine(
		userAccessPathEnt.UserID.Set(userID),
		userAccessPathEnt.LocationID.Set(locationID),
	); err != nil {
		return err
	}

	if err := userAccessPathRepo.Upsert(ctx, s.BobDBTrace, []*entity.UserAccessPath{userAccessPathEnt}); err != nil {
		return errors.Wrap(err, "userAccessPathRepo.Upsert")
	}

	return nil
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (s *Suite) teacherAccessCoursesByLocation(ctx context.Context, numTeacher, numCourse int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// init location
	ctx, err := s.aListOfLocationsInDB(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not create location")
	}

	ctx, err = s.someCourseTypes(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not course type")
	}
	for i := 0; i < numTeacher; i++ {
		teacherID, token, _ := s.AuthHelper.AUserSignedInAsRole(ctx, "teacher")
		err := s.validUserAccessPath(ctx, teacherID, stepState.LocationIDs[0])
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not create user access path")
		}

		stepState.Token = token
		stepState.TeacherIDs = append(stepState.TeacherIDs, teacherID)
	}

	for i := 0; i < numCourse; i++ {
		ctx = s.userUpsertCoursesDataWithLocation(ctx, stepState.LocationIDs[0], stepState.CourseTypeIDs[0])
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can not create course with location")
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminListAllocateTeacher(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	time.Sleep(1 * time.Second)
	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).ListAllocateTeacher(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListAllocateTeacherRequest{
		LocationIds: stepState.LocationIDs,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemStoresAllocateMarkerCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	am := &entities.AllocateMarker{}
	for i, teacherId := range stepState.TeacherIDs {
		var rowCount int32
		query := fmt.Sprintf("SELECT count(*) FROM %s WHERE teacher_id = $1", am.TableName())
		rows := s.EurekaDB.QueryRow(ctx, query, teacherId)
		rows.Scan(&rowCount)
		if rowCount != stepState.NumberAllocatedSubmissions[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected %v submission but got %v", stepState.NumberAllocatedSubmissions[i], rowCount)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnsAllocateTeacherCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.ListAllocateTeacherResponse)
	if len(stepState.TeacherIDs) != len(response.GetAllocateTeachers()) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected %v teachers but got %v", len(stepState.TeacherIDs), len(response.GetAllocateTeachers()))
	}

	for _, i := range response.GetAllocateTeachers() {
		if i.TeacherId == stepState.TeacherIDs[0] && i.NumberAssignedSubmission != stepState.NumberAllocatedSubmissions[0] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected %v assigned submission but got %v", i.NumberAssignedSubmission, stepState.NumberAllocatedSubmissions[0])
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertAllocateMarker(ctx context.Context, numberSubmissions, numberAllocatedSubmission1, numberAllocatedSubmission2 int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.TeacherIDs = []string{fmt.Sprintf("teacher_id_1_%v", idutil.ULIDNow()), fmt.Sprintf("teacher_id_2_%v", idutil.ULIDNow())}
	stepState.NumberAllocatedSubmissions = []int32{int32(numberAllocatedSubmission1), int32(numberAllocatedSubmission2)}
	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertAllocateMarker(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpsertAllocateMarkerRequest{
		Submissions: s.toSubmissionList(ctx, numberSubmissions),
		AllocateMarkers: []*sspb.UpsertAllocateMarkerRequest_AllocateMarkerItem{
			toAllocateMarker(stepState.TeacherIDs[0], numberAllocatedSubmission1),
			toAllocateMarker(stepState.TeacherIDs[1], numberAllocatedSubmission2),
		},
		CreatedBy: fmt.Sprintf("school_admin_id_%v", idutil.ULIDNow()),
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertAllocateMarkerForFirstTeacher(ctx context.Context, numberSubmissions int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.NumberAllocatedSubmissions = []int32{int32(numberSubmissions)}
	ctx, _ = s.aSignedIn(ctx, "school admin")

	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertAllocateMarker(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpsertAllocateMarkerRequest{
		Submissions: s.toSubmissionList(ctx, numberSubmissions),
		AllocateMarkers: []*sspb.UpsertAllocateMarkerRequest_AllocateMarkerItem{
			toAllocateMarker(stepState.TeacherIDs[0], numberSubmissions),
		},
		CreatedBy: fmt.Sprintf("school_admin_id_%v", idutil.ULIDNow()),
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) toSubmissionList(ctx context.Context, numberItem int) []*sspb.UpsertAllocateMarkerRequest_SubmissionItem {
	stepState := utils.StepStateFromContext[StepState](ctx)
	submissions := make([]*sspb.UpsertAllocateMarkerRequest_SubmissionItem, 0)
	for i := 0; i < numberItem; i++ {
		submissions = append(submissions, &sspb.UpsertAllocateMarkerRequest_SubmissionItem{
			SubmissionId: fmt.Sprintf("submission_id_%v_%v", stepState.Token[:20], i),
			StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
				StudyPlanId:        fmt.Sprintf("study_plan_id_%v_%v", stepState.Token[:20], i),
				LearningMaterialId: fmt.Sprintf("lm_id_%v_%v", stepState.Token[:20], i),
				StudentId:          wrapperspb.String(fmt.Sprintf("student_id_%v_%v", stepState.Token[:20], i)),
			},
		})
	}

	return submissions
}

func toAllocateMarker(teacherID string, numberAllocatedSubmission int) *sspb.UpsertAllocateMarkerRequest_AllocateMarkerItem {
	return &sspb.UpsertAllocateMarkerRequest_AllocateMarkerItem{
		TeacherId:       teacherID,
		NumberAllocated: int32(numberAllocatedSubmission),
	}
}
