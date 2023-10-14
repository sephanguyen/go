package lessonmgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) exportClassrooms(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.ExportClassroomsRequest{}
	stepState.Response, stepState.ResponseErr = lpb.NewLessonExecutorServiceClient(s.LessonMgmtConn).
		ExportClassrooms(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func getExportColumnsFromString(strCols string) []exporter.ExportColumnMap {
	result := []exporter.ExportColumnMap{}
	listCols := strings.Split(strCols, ",")
	for _, column := range listCols {
		result = append(result, exporter.ExportColumnMap{
			DBColumn: column,
		})
	}
	return result
}

func (s *Suite) returnsClassroomsInCsv(ctx context.Context, strCols string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export classroom: %s", stepState.ResponseErr.Error())
	}
	cols := getExportColumnsFromString(strCols)
	resp := stepState.Response.(*lpb.ExportClassroomsResponse)
	classroomRepo := repo.ClassroomRepo{}
	expectedData, err := classroomRepo.ExportAllClassrooms(ctx, s.BobDBTrace, cols)

	if err != nil {
		return ctx, fmt.Errorf("can not get expected classroom: %s", err)
	}
	if string(expectedData) != string(resp.GetData()) {
		return ctx, fmt.Errorf("classroom csv is not valid:\ngot:\n%s \nexpected: \n%s", resp.Data, expectedData)
	}
	return StepStateToContext(ctx, stepState), nil
}
