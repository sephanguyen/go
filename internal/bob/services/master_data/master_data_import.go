package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

type MasterDataImporterService struct {
	DB database.Ext

	LocationImporterService interface {
		ImportLocation(ctx context.Context, req *mpb.ImportLocationRequest) (*mpb.ImportLocationResponse, error)
		ImportLocationType(ctx context.Context, req *mpb.ImportLocationTypeRequest) (*mpb.ImportLocationTypeResponse, error)
	}
}

func (m *MasterDataImporterService) ImportLocation(ctx context.Context, req *bpb.ImportLocationRequest) (*bpb.ImportLocationResponse, error) {
	res, err := m.LocationImporterService.ImportLocation(ctx, &mpb.ImportLocationRequest{Payload: req.Payload})
	if err != nil {
		return nil, err
	}

	errs := make([]*bpb.ImportLocationResponse_ImportLocationError, 0, len(res.Errors))

	for _, e := range res.Errors {
		err := &bpb.ImportLocationResponse_ImportLocationError{
			RowNumber: e.RowNumber,
			Error:     e.Error,
		}
		errs = append(errs, err)
	}

	return &bpb.ImportLocationResponse{
		Errors:       errs,
		TotalSuccess: res.TotalSuccess,
		TotalFailed:  res.TotalFailed,
	}, nil
}

func (m *MasterDataImporterService) ImportLocationType(ctx context.Context, req *bpb.ImportLocationTypeRequest) (*bpb.ImportLocationTypeResponse, error) {
	res, err := m.LocationImporterService.ImportLocationType(ctx, &mpb.ImportLocationTypeRequest{Payload: req.Payload})
	if err != nil {
		return nil, err
	}

	errs := make([]*bpb.ImportLocationTypeResponse_ImportLocationTypeError, 0, len(res.Errors))

	for _, e := range res.Errors {
		err := &bpb.ImportLocationTypeResponse_ImportLocationTypeError{
			RowNumber: e.RowNumber,
			Error:     e.Error,
		}
		errs = append(errs, err)
	}

	return &bpb.ImportLocationTypeResponse{
		Errors:       errs,
		TotalSuccess: res.TotalSuccess,
		TotalFailed:  res.TotalFailed,
	}, nil
}
