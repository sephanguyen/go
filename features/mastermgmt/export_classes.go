package mastermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"k8s.io/utils/strings/slices"
)

func (s *suite) seedLocation(ctx context.Context, locID string, locName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	accessPath := buildAccessPath(s.LocationID, "", []string{locID})
	stmt := `INSERT INTO locations (location_id,name,parent_location_id, is_archived, access_path) VALUES($1,$2,$3,$4,$5) 
				ON CONFLICT DO NOTHING`
	_, err := s.BobDBTrace.Exec(ctx, stmt, locID,
		locName,
		s.LocationID,
		"0", accessPath)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed location with `id:%s`, %v", locID, err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedCourse(ctx context.Context, cID string, cName string, rpID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `INSERT INTO courses (course_id, name, resource_path, created_at, updated_at) VALUES($1,$2,$3,now(),now())`
	_, err := s.BobDBTrace.Exec(ctx,
		stmt,
		cID,
		cName,
		rpID,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot seed course with `id:%s`, %v", cID, err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) classesExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	schoolID := golibs.ResourcePathFromCtx(ctx)
	// seed location
	locID := idutil.ULIDNow()
	locName := locID + "-location-name"
	ctx, err := s.seedLocation(ctx, locID, locName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// seed course
	cID := idutil.ULIDNow()
	cName := cID + "-course-name"
	ctx, err = s.seedCourse(ctx, cID, cName, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	classList := []class_domain.ExportingClass{
		{
			ClassID:      idutil.ULIDNow(),
			Name:         idutil.ULIDNow() + "-class-1",
			CourseID:     cID,
			CourseName:   cName,
			LocationID:   locID,
			LocationName: locName,
		},
	}
	stepState.RequestSentAt = time.Now()
	expectedRows := [][]string{{
		"class_id", "class_name", "course_id", "location_id",
	}}
	for _, cl := range classList {
		fields := []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at"}
		query := fmt.Sprintf("INSERT INTO class (%s) VALUES ($1,$2,$3,$4,$5,$6,$7)",
			strings.Join(fields, ","))

		_, err := s.BobDBTrace.Exec(ctx, query, cl.ClassID, cl.Name, cl.CourseID, cl.LocationID, schoolID, stepState.RequestSentAt, stepState.RequestSentAt)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		expectedRows = append(expectedRows, []string{
			cl.ClassID, cl.Name, cl.CourseID, cl.LocationID,
		})
	}
	stepState.ExpectedCSV = s.getQuotedCSVRows(expectedRows)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) exportClasses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.ExportClassesRequest{}
	stepState.Response, stepState.ResponseErr = mpb.NewClassServiceClient(s.Connections.MasterMgmtConn).
		ExportClasses(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsClassesInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export classes: %s", stepState.ResponseErr.Error())
	}

	resp := stepState.Response.(*mpb.ExportClassesResponse)
	expectedRows := stepState.ExpectedCSV
	respLines := strings.Split(string(resp.Data), "\n")

	if expectedRows[0] != respLines[0] {
		return ctx, fmt.Errorf("class csv header is not valid.\nexpected: %s\ngot: %s", expectedRows[0], respLines[0])
	}
	for _, v := range expectedRows[1:] {
		if !slices.Contains(respLines, v) {
			return ctx, fmt.Errorf("class csv is not valid.\nexpected:%s\ngot %s", string(resp.Data), expectedRows)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getQuotedCSVRows(rows [][]string) []string {
	return sliceutils.Map(rows, func(s []string) string {
		return fmt.Sprintf("%s%s%s", `"`, strings.Join(s, `","`), `"`)
	})
}
