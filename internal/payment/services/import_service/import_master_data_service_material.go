package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) importMaterial(ctx context.Context, payload []byte) (*pb.ImportProductResponse, error) {
	errors := []*pb.ImportProductResponse_ImportProductError{}

	r := csv.NewReader(bytes.NewReader(payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	header := lines[0]
	headerTitles := []string{
		"material_id",
		"name",
		"material_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"custom_billing_period",
		"custom_billing_date",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"remarks",
		"is_archived",
		"is_unique",
	}
	err = utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	defaultProductSetting := entities.ProductSetting{
		IsPausable:                   pgtype.Bool{Bool: constant.ProductSettingDefaultIsPausable, Status: pgtype.Present},
		IsEnrollmentRequired:         pgtype.Bool{Bool: constant.ProductSettingDefaultIsEnrollmentRequired, Status: pgtype.Present},
		IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: constant.ProductSettingDefaultIsAddedToEnrollmentByDefault, Status: pgtype.Present},
		IsOperationFee:               pgtype.Bool{Bool: constant.ProductSettingDefaultIsOperationFee, Status: pgtype.Present},
	}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// first line is header
		for i, line := range lines[1:] {
			material, err := ReadMaterialFromCsv(line, headerTitles)
			if err != nil {
				errors = append(errors, &pb.ImportProductResponse_ImportProductError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse material item: %s", err),
				})
				continue
			}

			if material.MaterialID.Get() == nil {
				err := s.MaterialRepo.Create(ctx, tx, &material)
				if err != nil {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to create material item: %s", err),
					})
				} else {
					err = multierr.Combine(
						defaultProductSetting.ProductID.Set(material.MaterialID),
						s.ProductSettingRepo.Create(ctx, tx, &defaultProductSetting),
					)
					if err != nil {
						errors = append(errors, &pb.ImportProductResponse_ImportProductError{
							RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
							Error:     fmt.Sprintf("unable to set product setting for material item: %s", err),
						})
					}
				}
			} else {
				err := s.MaterialRepo.Update(ctx, tx, &material)
				if err != nil {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to update material item: %s", err),
					})
				}
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error when importing material: %s", err.Error())
	}
	return &pb.ImportProductResponse{
		Errors: errors,
	}, nil
}

func ReadMaterialFromCsv(line []string, columnNames []string) (material entities.Material, err error) {
	const (
		MaterialID = iota
		Name
		MaterialType
		TaxID
		ProductTag
		ProductPartnerID
		AvailableFrom
		AvailableUntil
		CustomBillingPeriod
		CustomBillingDate
		BillingScheduleID
		DisableProRatingFlag
		Remarks
		IsArchived
		IsUniq
	)

	mandatory := []int{
		Name,
		MaterialType,
		AvailableFrom,
		AvailableUntil,
		IsArchived,
	}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		err = fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
		return
	}

	err = multierr.Combine(
		utils.StringToFormatString("material_id", line[MaterialID], true, material.MaterialID.Set),
		utils.StringToMaterialType("material_type", line[MaterialType], material.MaterialType.Set),
		utils.StringToDate("custom_billing_date", line[CustomBillingDate], true, material.CustomBillingDate.Set),
		utils.StringToFormatString("material_id", line[MaterialID], true, material.Product.ProductID.Set),
		utils.StringToFormatString("name", line[Name], false, material.Name.Set),
		material.ProductType.Set(pb.ProductType_PRODUCT_TYPE_MATERIAL),
		utils.StringToFormatString("tax_id", line[TaxID], true, material.TaxID.Set),
		utils.StringToFormatString("product_tag", line[ProductTag], true, material.ProductTag.Set),
		utils.StringToFormatString("product_partner_id", line[ProductPartnerID], true, material.ProductPartnerID.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, material.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, material.AvailableUntil.Set),
		utils.StringToDate("custom_billing_period", line[CustomBillingPeriod], true, material.CustomBillingPeriod.Set),
		utils.StringToFormatString("billing_schedule_id", line[BillingScheduleID], true, material.BillingScheduleID.Set),
		utils.StringToBool("disable_pro_rating_flag", line[DisableProRatingFlag], true, material.DisableProRatingFlag.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, material.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, material.IsArchived.Set),
		utils.StringToBool("is_unique", line[IsUniq], true, material.IsUnique.Set),
	)

	return
}
