package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure/repo"
	infrastructure_lesson "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClassDoAccountService struct {
	cfg                *configs.ClassDoConfig
	LessonmgmtDB       database.Ext
	ClassDoAccountRepo infrastructure.ClassDoAccountRepo
	LessonRepo         infrastructure_lesson.LessonRepo
}

func NewClassDoAccountService(cfg *configs.ClassDoConfig, lessonmgmtDB database.Ext, classDoAccountRepo infrastructure.ClassDoAccountRepo, lessonRepo infrastructure_lesson.LessonRepo) *ClassDoAccountService {
	return &ClassDoAccountService{
		cfg:                cfg,
		LessonmgmtDB:       lessonmgmtDB,
		ClassDoAccountRepo: classDoAccountRepo,
		LessonRepo:         lessonRepo,
	}
}

func (c *ClassDoAccountService) ImportClassDoAccount(ctx context.Context, req *lpb.ImportClassDoAccountRequest) (res *lpb.ImportClassDoAccountResponse, err error) {
	classDoAccounts, resError, err := c.buildImportAccountArgs(req.Payload)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(resError.Errors) == 0 && len(classDoAccounts) > 0 {
		err = database.ExecInTx(ctx, c.LessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			err = c.ClassDoAccountRepo.UpsertClassDoAccounts(ctx, tx, classDoAccounts)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("error in ClassDoAccountRepo.UpsertClassDoAccounts: %s", err.Error()))
			}
			classDoAccountDeleted, _ := sliceutils.Reduce(classDoAccounts, func(classDoAccountsDeleted []string, c *domain.ClassDoAccount) ([]string, error) {
				if c.Action == domain.ActionDelete {
					classDoAccountsDeleted = append(classDoAccountsDeleted, c.ClassDoID)
				}
				return classDoAccountsDeleted, nil
			}, make([]string, 0, len(classDoAccounts)))
			if len(classDoAccountDeleted) > 0 {
				return c.LessonRepo.RemoveClassDoLinkOfLesson(ctx, tx, classDoAccountDeleted)
			}
			return nil
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return resError, nil
}

func (c *ClassDoAccountService) buildImportAccountArgs(data []byte) (domain.ClassDoAccounts, *lpb.ImportClassDoAccountResponse, error) {
	sc := scanner.NewCSVScanner(bytes.NewReader(data))
	if len(sc.GetRow()) == 0 {
		return nil, &lpb.ImportClassDoAccountResponse{}, fmt.Errorf("no data in csv file")
	}
	errors := []*lpb.ImportClassDoAccountResponse_ImportClassDoAccountError{}
	classDoAccounts := make([]*domain.ClassDoAccount, 0, 10)
	for sc.Scan() {
		classDoAccount, err := domain.NewClassDoAccountBuilder().
			WithClassDoID(sc.Text(domain.ClassDoIDLabel)).
			WithClassDoEmail(sc.Text(domain.ClassDoEmailLabel)).
			WithClassDoAPIKey(sc.Text(domain.ClassDoAPIKeyLabel)).
			WithAction(sc.Text(domain.ClassDoActionLabel)).
			EncryptAPIKey(c.cfg.SecretKey).
			Build()

		if err != nil {
			icErr := &lpb.ImportClassDoAccountResponse_ImportClassDoAccountError{
				RowNumber: int32(sc.GetCurRow()),
				Error:     err.Error(),
			}
			errors = append(errors, icErr)
			continue
		}
		classDoAccounts = append(classDoAccounts, classDoAccount)
	}
	return classDoAccounts, &lpb.ImportClassDoAccountResponse{Errors: errors}, nil
}

func (c *ClassDoAccountService) ExportClassDoAccount(ctx context.Context, _ *lpb.ExportClassDoAccountRequest) (res *lpb.ExportClassDoAccountResponse, err error) {
	data, err := c.ClassDoAccountRepo.GetAllClassDoAccounts(ctx, c.LessonmgmtDB)
	if err != nil {
		return &lpb.ExportClassDoAccountResponse{}, status.Error(codes.Internal, fmt.Sprintf("error in ClassDoAccountRepo.GetAllClassDoAccounts: %s", err.Error()))
	}

	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn:  "classdo_id",
			CSVColumn: domain.ClassDoIDLabel,
		},
		{
			DBColumn:  "classdo_email",
			CSVColumn: domain.ClassDoEmailLabel,
		},
		{
			DBColumn:  "classdo_api_key",
			CSVColumn: domain.ClassDoAPIKeyLabel,
		},
	}
	exportable := sliceutils.Map(data, func(d *repo.ClassDoAccount) database.Entity {
		generatedAPIDecrypted, err := crypt.Decrypt(d.ClassDoAPIKey.String, c.cfg.SecretKey)
		if err != nil {
			d.ClassDoAPIKey = pgtype.Text{String: "", Status: pgtype.Present}
		} else {
			d.ClassDoAPIKey = pgtype.Text{String: generatedAPIDecrypted, Status: pgtype.Present}
		}
		return d
	})

	str, err := exporter.ExportBatch(exportable, exportCols)
	if err != nil {
		return nil, fmt.Errorf("error in ExportBatch: %w", err)
	}

	res = &lpb.ExportClassDoAccountResponse{
		Data: exporter.ToCSV(str),
	}

	return res, nil
}
