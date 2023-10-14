package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"k8s.io/utils/strings/slices"
)

func (s *suite) gradesExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	iStmt := `INSERT INTO grade
		(grade_id, name, partner_internal_id, sequence, remarks, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5,  now(), now())`
	expectedRows := [][]string{{
		"grade_id", "grade_partner_id", "name", "sequence", "remarks",
	}}
	for i := 0; i < 5; i++ {
		timeID := idutil.ULIDNow()
		gID := fmt.Sprintf("grade id %s", timeID)
		gName := fmt.Sprintf("grade name %s", timeID)
		gPID := fmt.Sprintf("100%d", i)
		remarks := fmt.Sprintf("remarks %d", i)
		_, err := s.MasterMgmtDBTrace.Exec(ctx, iStmt, gID, gName, gPID, i, remarks)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert grade, err: %s", err)
		}
		expectedRows = append(expectedRows, []string{
			gID, gPID, gName, fmt.Sprintf("%d", i), remarks,
		})
	}
	stepState.ExpectedCSV = s.getQuotedCSVRows(expectedRows)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) exportGrades(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.ExportGradesRequest{}
	stepState.Response, stepState.ResponseErr = mpb.NewGradeServiceClient(s.Connections.MasterMgmtConn).
		ExportGrades(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsGradesInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export grades: %s", stepState.ResponseErr.Error())
	}

	expectedRows := stepState.ExpectedCSV
	resp := stepState.Response.(*mpb.ExportGradesResponse)
	respLines := strings.Split(string(resp.Data), "\n")

	if expectedRows[0] != respLines[0] {
		return ctx, fmt.Errorf("grade csv header is not valid.\nexpected: %s\ngot: %s", expectedRows[0], respLines[0])
	}
	for _, v := range expectedRows[1:] {
		if !slices.Contains(respLines, v) {
			return ctx, fmt.Errorf("grade csv is not valid.\nexpected:%s\ngot %s", string(resp.Data), expectedRows)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
