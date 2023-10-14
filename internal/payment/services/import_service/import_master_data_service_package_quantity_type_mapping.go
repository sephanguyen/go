package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportPackageQuantityTypeMapping(
	ctx context.Context,
	req *pb.ImportPackageQuantityTypeMappingRequest) (
	*pb.ImportPackageQuantityTypeMappingResponse,
	error,
) {
	errors := []*pb.ImportPackageQuantityTypeMappingResponse_ImportPackageQuantityTypeMappingError{}

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	headerTitles := []string{
		"package_type",
		"quantity_type",
	}
	err = utils.ValidateCsvHeader(len(headerTitles), lines[0], headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}
	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for i, line := range lines[1:] {
			packageQuantityTypeMapping, err := PackageQuantityTypeMappingFromCsv(line)

			if err != nil {
				errors = append(errors, &pb.ImportPackageQuantityTypeMappingResponse_ImportPackageQuantityTypeMappingError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse package quantity type mapping item: %s", err),
				})
				continue
			}

			if err = s.PackageQuantityTypeMappingRepo.Upsert(ctx, tx, &packageQuantityTypeMapping); err != nil {
				errors = append(errors, &pb.ImportPackageQuantityTypeMappingResponse_ImportPackageQuantityTypeMappingError{
					RowNumber: int32(i) + 2,
					Error:     fmt.Sprintf("unable to import package quantity type mapping item: %s", err),
				})
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error importing package quantity type mapping")
		}
		return nil
	})

	return &pb.ImportPackageQuantityTypeMappingResponse{
		Errors: errors,
	}, nil
}

func PackageQuantityTypeMappingFromCsv(line []string) (packageQuantityTypeMapping entities.PackageQuantityTypeMapping, err error) {
	const (
		PackageType = iota
		QuantityType
	)

	err = multierr.Combine(
		utils.StringToPackageType("package_type", line[PackageType], packageQuantityTypeMapping.PackageType.Set),
		utils.StringToQuantityType("quantity_type", line[QuantityType], packageQuantityTypeMapping.QuantityType.Set),
	)
	return
}
