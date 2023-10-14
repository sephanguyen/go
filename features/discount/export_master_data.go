package discount

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	export "github.com/manabie-com/backend/internal/discount/services/export_service"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
)

func (s *suite) addDataForExportMasterData(ctx context.Context, dataType string) (context context.Context, err error) {
	stepState := StepStateFromContext(ctx)
	if dataType == "discount tag" {
		err = s.insertSomeDiscountTags(ctx)
	}
	context = StepStateToContext(ctx, stepState)
	if err != nil {
		return
	}
	return
}

func (s *suite) theUserExportMasterData(ctx context.Context, user string, dataType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	clientExport := pb.NewExportServiceClient(s.DiscountConn)
	var req *pb.ExportMasterDataRequest
	if dataType == "discount tag" {
		req = &pb.ExportMasterDataRequest{ExportDataType: pb.ExportMasterDataType_EXPORT_DISCOUNT_TAG}
	}
	resp, err := clientExport.ExportMasterData(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = req
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theMasterDataCSVHasCorrectContent(ctx context.Context, _ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request := stepState.Request.(*pb.ExportMasterDataRequest)
	response := stepState.Response.(*pb.ExportMasterDataResponse)

	r := csv.NewReader(bytes.NewReader(response.Data))
	lines, err := r.ReadAll()
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("r.ReadAll() err: %v", err)
	}
	colMap, _ := export.GetExportColMapAndEntityType(request.ExportDataType)
	csvColumn := make([]string, 0, len(colMap))
	for _, col := range colMap {
		csvColumn = append(csvColumn, col.CSVColumn)
	}
	// check the header record
	err = checkCSVHeaderForExport(
		csvColumn,
		lines[0],
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
