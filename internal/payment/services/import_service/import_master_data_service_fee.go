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

func (s *ImportMasterDataService) importFee(ctx context.Context, payload []byte) (*pb.ImportProductResponse, error) {
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
		"fee_id",
		"name",
		"fee_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"custom_billing_period",
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
	// first line is header
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for i, line := range lines[1:] {
			fee, err := ReadFeeFromCsv(line, headerTitles)
			if err != nil {
				errors = append(errors, &pb.ImportProductResponse_ImportProductError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse fee item: %s", err),
				})
				continue
			}
			if fee.FeeID.Get() == nil {
				err := s.FeeRepo.Create(ctx, tx, &fee)
				if err != nil {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to create fee item: %s", err),
					})
				} else {
					err = multierr.Combine(
						defaultProductSetting.ProductID.Set(fee.FeeID),
						s.ProductSettingRepo.Create(ctx, tx, &defaultProductSetting),
					)
					if err != nil {
						errors = append(errors, &pb.ImportProductResponse_ImportProductError{
							RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
							Error:     fmt.Sprintf("unable to set product setting for fee item: %s", err),
						})
					}
				}
			} else {
				err := s.FeeRepo.Update(ctx, tx, &fee)
				if err != nil {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to update fee item: %s", err),
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
		log.Printf("Error when importing fee: %s", err.Error())
	}
	return &pb.ImportProductResponse{
		Errors: errors,
	}, nil
}

func ReadFeeFromCsv(line []string, columnNames []string) (fee entities.Fee, err error) {
	const (
		FeeID = iota
		Name
		FeeType
		TaxID
		ProductTag
		ProductPartnerID
		AvailableFrom
		AvailableUntil
		CustomBillingPeriod
		BillingScheduleID
		DisableProRatingFlag
		Remarks
		IsArchived
		IsUniq
	)

	mandatory := []int{
		Name,
		FeeType,
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
		utils.StringToFormatString("fee_id", line[FeeID], true, fee.FeeID.Set),
		utils.StringToFeeType("fee_type", line[FeeType], fee.FeeType.Set),
		utils.StringToFormatString("fee_id", line[FeeID], true, fee.Product.ProductID.Set),
		utils.StringToFormatString("name", line[Name], false, fee.Name.Set),
		fee.ProductType.Set(pb.ProductType_PRODUCT_TYPE_FEE),
		utils.StringToFormatString("tax_id", line[TaxID], true, fee.TaxID.Set),
		utils.StringToFormatString("product_tag", line[ProductTag], true, fee.ProductTag.Set),
		utils.StringToFormatString("product_partner_id", line[ProductPartnerID], true, fee.ProductPartnerID.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, fee.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, fee.AvailableUntil.Set),
		utils.StringToDate("custom_billing_period", line[CustomBillingPeriod], true, fee.CustomBillingPeriod.Set),
		utils.StringToFormatString("billing_schedule_id", line[BillingScheduleID], true, fee.BillingScheduleID.Set),
		utils.StringToBool("disable_pro_rating_flag", line[DisableProRatingFlag], true, fee.DisableProRatingFlag.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, fee.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], true, fee.IsArchived.Set),
		utils.StringToBool("is_unique", line[IsUniq], true, fee.IsUnique.Set),
	)

	return
}
