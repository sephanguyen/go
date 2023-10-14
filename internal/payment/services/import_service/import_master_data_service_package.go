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

func (s *ImportMasterDataService) packageModifier(
	ctx context.Context,
	data []byte,
) ([]*pb.ImportProductResponse_ImportProductError, error) {
	r := csv.NewReader(bytes.NewReader(data))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	err = s.validatePackageHeaderCSV(lines[0])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var errors []*pb.ImportProductResponse_ImportProductError
	defaultProductSetting := entities.ProductSetting{
		IsPausable:                   pgtype.Bool{Bool: constant.ProductSettingDefaultIsPausable, Status: pgtype.Present},
		IsEnrollmentRequired:         pgtype.Bool{Bool: constant.ProductSettingDefaultIsEnrollmentRequired, Status: pgtype.Present},
		IsAddedToEnrollmentByDefault: pgtype.Bool{Bool: constant.ProductSettingDefaultIsAddedToEnrollmentByDefault, Status: pgtype.Present},
		IsOperationFee:               pgtype.Bool{Bool: constant.ProductSettingDefaultIsOperationFee, Status: pgtype.Present},
	}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for i, line := range lines[1:] {
			var pkg entities.Package
			pkg, err = ProductAndPackageFromCsv(line)
			if err != nil {
				errors = append(errors, &pb.ImportProductResponse_ImportProductError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse package item: %s", err),
				})
				continue
			}
			if pkg.PackageType.String == pb.PackageType_PACKAGE_TYPE_ONE_TIME.String() ||
				pkg.PackageType.String == pb.PackageType_PACKAGE_TYPE_SLOT_BASED.String() {
				if pkg.PackageStartDate.Status == pgtype.Null && pkg.PackageEndDate.Status == pgtype.Null {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("package_start_date, package_end_date are missing"),
					})
					continue
				}
				if pkg.PackageStartDate.Status == pgtype.Null {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("package_start_date is missing"),
					})
					continue
				}
				if pkg.PackageEndDate.Status == pgtype.Null {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("package_end_date is missing"),
					})
					continue
				}
			}
			if pkg.PackageID.Get() == nil {
				err := s.PackageRepo.Create(ctx, tx, &pkg)
				if err != nil {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to insert package item: %s", err),
					})
				} else {
					err = multierr.Combine(
						defaultProductSetting.ProductID.Set(pkg.PackageID),
						s.ProductSettingRepo.Create(ctx, tx, &defaultProductSetting),
					)
					if err != nil {
						errors = append(errors, &pb.ImportProductResponse_ImportProductError{
							RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
							Error:     fmt.Sprintf("unable to set product setting for package item: %s", err),
						})
					}
				}
			} else {
				err := s.PackageRepo.Update(ctx, tx, &pkg)
				if err != nil {
					errors = append(errors, &pb.ImportProductResponse_ImportProductError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to update package item: %s", err),
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
		log.Printf("Error when importing package: %s", err.Error())
	}
	return errors, nil
}

func (s *ImportMasterDataService) validatePackageHeaderCSV(header []string) error {
	headerTitles := []string{
		"package_id",
		"name",
		"package_type",
		"tax_id",
		"product_tag",
		"product_partner_id",
		"available_from",
		"available_until",
		"max_slot",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"package_start_date",
		"package_end_date",
		"remarks",
		"is_archived",
		"is_unique",
	}

	err := utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return err
	}
	return nil
}

func ProductAndPackageFromCsv(line []string) (pkg entities.Package, err error) {
	const (
		PackageID = iota
		Name
		PackageType
		TaxID
		ProductTag
		ProductPartnerID
		AvailableFrom
		AvailableUntil
		MaxSlot
		CustomBillingPeriod
		BillingScheduleID
		DisableProRatingFlag
		PackageStartDate
		PackageEndDate
		Remarks
		IsArchived
		IsUniq
	)

	if err = multierr.Combine(
		utils.StringToFormatString("package_id", line[PackageID], true, pkg.Product.ProductID.Set),
		utils.StringToFormatString("package_id", line[PackageID], true, pkg.PackageID.Set),
		utils.StringToFormatString("name", line[Name], false, pkg.Name.Set),
		pkg.ProductType.Set(pb.ProductType_PRODUCT_TYPE_PACKAGE),
		utils.StringToFormatString("tax_id", line[TaxID], true, pkg.TaxID.Set),
		utils.StringToFormatString("product_tag", line[ProductTag], true, pkg.ProductTag.Set),
		utils.StringToFormatString("product_partner_id", line[ProductPartnerID], true, pkg.ProductPartnerID.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, pkg.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, pkg.AvailableUntil.Set),
		utils.StringToDate("custom_billing_period", line[CustomBillingPeriod], true, pkg.CustomBillingPeriod.Set),
		utils.StringToFormatString("billing_schedule_id", line[BillingScheduleID], true, pkg.BillingScheduleID.Set),
		utils.StringToBool("disable_pro_rating_flag", line[DisableProRatingFlag], true, pkg.DisableProRatingFlag.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, pkg.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, pkg.IsArchived.Set),
		utils.StringToBool("is_unique", line[IsUniq], true, pkg.IsUnique.Set),
		utils.StringToPackageType("package_type", line[PackageType], pkg.PackageType.Set),
		utils.StringToInt("max_slot", line[MaxSlot], true, pkg.MaxSlot.Set),
		utils.StringToDate("package_start_date", line[PackageStartDate], true, pkg.PackageStartDate.Set),
		utils.StringToDate("package_end_date", line[PackageEndDate], true, pkg.PackageEndDate.Set),
	); err != nil {
		return pkg, err
	}

	return pkg, nil
}
