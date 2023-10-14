package service

import (
	"context"
	"fmt"

	dbEntities "github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ExportService) ExportMasterData(ctx context.Context, req *pb.ExportMasterDataRequest) (exportDataResp *pb.ExportMasterDataResponse, err error) {
	colMap, entityType := GetExportColMapAndEntityType(req.ExportDataType)
	var entities []database.Entity
	entities, err = exporter.RetrieveAllData(ctx, s.DB, entityType)
	if err != nil {
		return nil, err
	}
	res, err := exporter.ExportBatch(entities, colMap)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("exporter.ExportBatch err: %v", err))
	}
	csvBytes := exporter.ToCSV(res)
	return &pb.ExportMasterDataResponse{
		Data: csvBytes,
	}, nil
}

func GetExportColMapAndEntityType(exportDataType pb.ExportMasterDataType) (colMap []exporter.ExportColumnMap, entityType database.Entity) {
	switch exportDataType {
	case pb.ExportMasterDataType_EXPORT_DISCOUNT_TAG:
		colMap = []exporter.ExportColumnMap{
			{
				CSVColumn: "discount_tag_id",
				DBColumn:  "discount_tag_id",
			},
			{
				CSVColumn: "discount_tag_name",
				DBColumn:  "discount_tag_name",
			},
			{
				CSVColumn: "selectable",
				DBColumn:  "selectable",
			},
			{
				CSVColumn: "is_archived",
				DBColumn:  "is_archived",
			},
		}
		entityType = &dbEntities.DiscountTag{}
	default:
	}
	return
}
