package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) exportCourseAccessPaths(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.ExportCourseAccessPathsRequest{}
	stepState.Response, stepState.ResponseErr = mpb.NewCourseAccessPathServiceClient(s.Connections.MasterMgmtConn).
		ExportCourseAccessPaths(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkCourseAccessPathCSV(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export subjects: %s", stepState.ResponseErr.Error())
	}

	expectedRows := stepState.ExpectedCSV
	resp := stepState.Response.(*mpb.ExportCourseAccessPathsResponse)
	respLines := strings.Split(string(resp.Data), "\n")

	if expectedRows[0] != respLines[0] {
		return ctx, fmt.Errorf("course access path csv header is not valid.\nexpected: %s\ngot: %s", expectedRows[0], respLines[0])
	}
	for _, v := range expectedRows[1:] {
		if !sliceutils.ContainsFunc(respLines, func(resLine string) bool {
			resSplit := strings.Split(resLine, ",")
			expectSplit := strings.Split(v, ",")
			courseID := resSplit[1]
			locationID := resSplit[2]

			expectCourseID := expectSplit[0]
			expectLocationID := expectSplit[1]

			return courseID == expectCourseID && locationID == expectLocationID
		}) {
			return ctx, fmt.Errorf("course access path csv is not valid.\nexpected:%s\ngot %s", strings.Join(expectedRows, "\n"), string(resp.Data))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
