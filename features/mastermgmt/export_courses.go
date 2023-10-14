package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"k8s.io/utils/strings/slices"
)

func (s *suite) coursesExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	iStmt := `INSERT INTO courses
		(course_id, name, course_type_id, course_partner_id, teaching_method, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())`
	delStmt := `update courses set deleted_at = now()`
	cts, err := s.seedCourseTypes(ctx, 3)
	cs := make([]*domain.Course, 3)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	_, err = s.BobDBTrace.Exec(ctx, delStmt)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot delete course, err: %s", err)
	}
	for i := 0; i < len(cts); i++ {
		cID := idutil.ULIDNow()
		cName := fmt.Sprintf("course name %d", i)
		pID := "p" + cID
		cTID := cts[i].CourseTypeID
		remarks := fmt.Sprintf("remarks %d", i)
		teachingMethod := domain.CourseTeachingMethodGroup
		csvTeachingMethodVal := "Group"
		if i%2 == 0 {
			teachingMethod = domain.CourseTeachingMethodIndividual
			csvTeachingMethodVal = "Individual"
		}
		_, err := s.BobDBTrace.Exec(ctx, iStmt, cID, cName, cTID, pID, teachingMethod, remarks)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert course, err: %s", err)
		}

		cs[i] = &domain.Course{
			CourseID:     cID,
			Name:         cName,
			CourseTypeID: cTID,
			Remarks:      remarks,
			PartnerID:    pID,
			Icon:         csvTeachingMethodVal, // HACK to use string type
		}
	}
	stepState.CoursesExportCSV = `"course_id","course_name","course_type_id","course_partner_id","teaching_method","remarks"` + "\n" +
		fmt.Sprintf(`"%s","%s","%s","%s","%s","%s"`, cs[0].CourseID, cs[0].Name, cs[0].CourseTypeID, cs[0].PartnerID, cs[0].Icon, cs[0].Remarks) + "\n" +
		fmt.Sprintf(`"%s","%s","%s","%s","%s","%s"`, cs[1].CourseID, cs[1].Name, cs[1].CourseTypeID, cs[1].PartnerID, cs[1].Icon, cs[1].Remarks) + "\n" +
		fmt.Sprintf(`"%s","%s","%s","%s","%s","%s"`, cs[2].CourseID, cs[2].Name, cs[2].CourseTypeID, cs[2].PartnerID, cs[2].Icon, cs[2].Remarks) + "\n"

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedCourseTypes(ctx context.Context, count int) ([]*domain.CourseType, error) {
	iStmt := `INSERT INTO course_type
	(course_type_id, name, created_at, updated_at)
	VALUES ($1, $2, now(), now())`
	cts := make([]*domain.CourseType, count)
	for i := 0; i < count; i++ {
		rd := idutil.ULIDNow()
		cts[i] = &domain.CourseType{
			CourseTypeID: rd,
			Name:         "Course Type " + rd,
		}
		_, err := s.BobDBTrace.Exec(ctx, iStmt, cts[i].CourseTypeID, cts[i].Name)
		if err != nil {
			return nil, fmt.Errorf("cannot insert course type, err: %s", err)
		}
	}
	return cts, nil
}

func (s *suite) exportCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.ExportCoursesRequest{}
	stepState.Response, stepState.ResponseErr = mpb.NewMasterDataCourseServiceClient(s.Connections.MasterMgmtConn).
		ExportCourses(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsCoursesInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export courses: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*mpb.ExportCoursesResponse)
	csvLines := strings.Split(stepState.CoursesExportCSV, "\n")
	respLines := strings.Split(string(resp.Data), "\n")

	if csvLines[0] != respLines[0] {
		return ctx, fmt.Errorf("course csv header is not valid: %s", respLines[0])
	}
	for _, v := range csvLines[1:] {
		if !slices.Contains(respLines, v) {
			return ctx, fmt.Errorf("course csv is not valid: %s\nexpected:\n%s", string(resp.Data), stepState.CoursesExportCSV)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
