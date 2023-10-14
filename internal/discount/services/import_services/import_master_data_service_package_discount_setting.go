package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportPackageDiscountSetting(ctx context.Context, req *pb.ImportPackageDiscountSettingRequest) (*pb.ImportPackageDiscountSettingResponse, error) {
	errors := []*pb.ImportPackageDiscountSettingResponse_ImportPackageDiscountSettingError{}

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	header := lines[0]
	headerTitles := []string{
		"package_id",
		"min_slot_trigger",
		"max_slot_trigger",
		"discount_tag_id",
		"is_archived",
		"product_group_id",
	}

	if err := utils.ValidateCsvHeader(len(headerTitles), header, headerTitles); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	type packageDiscountSettingWithPositions struct {
		positions                     []int32
		packageDiscountSettingMapping []*entities.PackageDiscountSetting
	}

	mapPackageDiscountSettingWithPositions := make(map[pgtype.Text]packageDiscountSettingWithPositions)
	for i, line := range lines[1:] {
		packageDiscountSettingMappingFromCSV, err := PackageDiscountSettingFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportPackageDiscountSettingResponse_ImportPackageDiscountSettingError{
				RowNumber: int32(i) + 2,
				Error:     fmt.Sprintf("unable to parse package discount setting: %s", err),
			})
			continue
		}
		v, ok := mapPackageDiscountSettingWithPositions[packageDiscountSettingMappingFromCSV.PackageID]
		if !ok {
			mapPackageDiscountSettingWithPositions[packageDiscountSettingMappingFromCSV.PackageID] = packageDiscountSettingWithPositions{
				positions:                     []int32{int32(i)},
				packageDiscountSettingMapping: []*entities.PackageDiscountSetting{packageDiscountSettingMappingFromCSV},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.packageDiscountSettingMapping = append(v.packageDiscountSettingMapping, packageDiscountSettingMappingFromCSV)
			mapPackageDiscountSettingWithPositions[packageDiscountSettingMappingFromCSV.PackageID] = v
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for packageID, packageDiscountSettingMappingWithPosition := range mapPackageDiscountSettingWithPositions {
			upsertErr := s.PackageDiscountSettingRepo.Upsert(ctx, tx, packageID, packageDiscountSettingMappingWithPosition.packageDiscountSettingMapping)
			if upsertErr != nil {
				for _, position := range packageDiscountSettingMappingWithPosition.positions {
					errors = append(errors, &pb.ImportPackageDiscountSettingResponse_ImportPackageDiscountSettingError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to import package discount setting: %s", upsertErr),
					})
				}
				continue
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error when importing package discount setting: %s", err.Error())
	}

	return &pb.ImportPackageDiscountSettingResponse{
		Errors: errors,
	}, nil
}

func PackageDiscountSettingFromCsv(line []string) (*entities.PackageDiscountSetting, error) {
	const (
		PackageID = iota
		MinSlotTrigger
		MaxSlotTrigger
		DiscountTagID
		IsArchived
		ProductGroupID
	)
	packageDiscountSetting := &entities.PackageDiscountSetting{}
	if err := multierr.Combine(
		utils.StringToFormatString("package_id", line[PackageID], false, packageDiscountSetting.PackageID.Set),
		utils.StringToInt("min_slot_trigger", line[MinSlotTrigger], true, packageDiscountSetting.MinSlotTrigger.Set),
		utils.StringToInt("max_slot_trigger", line[MaxSlotTrigger], true, packageDiscountSetting.MaxSlotTrigger.Set),
		utils.StringToFormatString("discount_tag_id", line[DiscountTagID], false, packageDiscountSetting.DiscountTagID.Set),
		utils.StringToBool("is_archived", line[IsArchived], true, packageDiscountSetting.IsArchived.Set),
		utils.StringToFormatString("product_group_id", line[ProductGroupID], false, packageDiscountSetting.ProductGroupID.Set),
	); err != nil {
		return nil, err
	}

	return packageDiscountSetting, nil
}
