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

func (s *suite) subjectsExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	iStmt := `INSERT INTO subject
		(subject_id, name, created_at, updated_at)
		VALUES ($1, $2,  now(), now())`
	expectedRows := [][]string{{
		"subject_id", "name",
	}}
	for i := 0; i < 5; i++ {
		timeID := idutil.ULIDNow()
		seedName := fmt.Sprintf("subject name %s", timeID)
		_, err := s.BobDBTrace.Exec(ctx, iStmt, timeID, seedName)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert subject, err: %s", err)
		}
		expectedRows = append(expectedRows, []string{
			timeID, seedName,
		})
	}
	stepState.ExpectedCSV = s.getQuotedCSVRows(expectedRows)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) exportSubjects(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.ExportSubjectsRequest{}
	stepState.Response, stepState.ResponseErr = mpb.NewSubjectServiceClient(s.Connections.MasterMgmtConn).
		ExportSubjects(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsSubjectsInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export subjects: %s", stepState.ResponseErr.Error())
	}

	expectedRows := stepState.ExpectedCSV
	resp := stepState.Response.(*mpb.ExportSubjectsResponse)
	respLines := strings.Split(string(resp.Data), "\n")

	if expectedRows[0] != respLines[0] {
		return ctx, fmt.Errorf("subject csv header is not valid.\nexpected: %s\ngot: %s", expectedRows[0], respLines[0])
	}
	for _, v := range expectedRows[1:] {
		if !slices.Contains(respLines, v) {
			return ctx, fmt.Errorf("subject csv is not valid.\nexpected:%s\ngot %s", expectedRows, string(resp.Data))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
