package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	infrastructure_lesson "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure/repo"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ZoomAccountService struct {
	zoomAccountRepo   infrastructure.ZoomAccountRepo
	lessonRepo        infrastructure_lesson.LessonRepo
	zoomService       ZoomServiceInterface
	wrapperConnection *support.WrapperDBConnection
}

func NewZoomAccountService(wrapperConnection *support.WrapperDBConnection, zoomService ZoomServiceInterface, zoomAccountRepo infrastructure.ZoomAccountRepo, lessonRepo infrastructure_lesson.LessonRepo) *ZoomAccountService {
	return &ZoomAccountService{
		zoomAccountRepo:   zoomAccountRepo,
		wrapperConnection: wrapperConnection,
		zoomService:       zoomService,
		lessonRepo:        lessonRepo,
	}
}

func (z *ZoomAccountService) buildImportAccountArgs(data []byte) (domain.ZoomAccounts, *lpb.ImportZoomAccountResponse, error) {
	sc := scanner.NewCSVScanner(bytes.NewReader(data))
	if len(sc.GetRow()) == 0 {
		return nil, &lpb.ImportZoomAccountResponse{}, fmt.Errorf("no data in csv file")
	}
	errors := []*lpb.ImportZoomAccountResponse_ImportZoomAccountError{}
	zoomAccounts := []*domain.ZoomAccount{}
	for sc.Scan() {
		zoomAccount, err := domain.NewZoomAccountBuilder().
			WithID(sc.Text(domain.ZoomIDLabel)).
			WithEmail(sc.Text(domain.ZoomUsernameLabel)).
			WithAction(sc.Text(domain.ZoomAccountActionLabel)).
			Build()

		if err != nil {
			icErr := &lpb.ImportZoomAccountResponse_ImportZoomAccountError{
				RowNumber: int32(sc.GetCurRow()),
				Error:     err.Error(),
			}
			errors = append(errors, icErr)
			continue
		}
		zoomAccounts = append(zoomAccounts, zoomAccount)
	}
	return zoomAccounts, &lpb.ImportZoomAccountResponse{Errors: errors}, nil
}

func (z *ZoomAccountService) getMapZoomUser(ctx context.Context) (map[string]bool, error) {
	mapZoomUser := make(map[string]bool)
	pageNumber := 1
	pageSize := 300
	pageCount := 1
	for {
		zoomUserResponse, err := z.zoomService.RetryGetListUsers(ctx, &domain.ZoomGetListUserRequest{
			PageNumber: pageNumber,
			PageSize:   pageSize,
		})
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		for _, user := range zoomUserResponse.Users {
			if user.Status == domain.ZoomUserStatusActive {
				mapZoomUser[user.Email] = true
			}
		}
		pageNumber++
		if zoomUserResponse.PageCount > 1 {
			pageCount = zoomUserResponse.PageCount
		}
		if pageNumber > pageCount {
			break
		}
	}
	return mapZoomUser, nil
}

func (z *ZoomAccountService) ImportZoomAccount(ctx context.Context, req *lpb.ImportZoomAccountRequest) (res *lpb.ImportZoomAccountResponse, err error) {
	zoomAccounts, resError, err := z.buildImportAccountArgs(req.Payload)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mapZoomUsers, err := z.getMapZoomUser(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	zoomAccountNeedVerify := sliceutils.Filter(zoomAccounts, func(account *domain.ZoomAccount) bool {
		return account.Action != domain.ActionDelete
	})
	for index, zoomAccount := range zoomAccountNeedVerify {
		_, ok := mapZoomUsers[zoomAccount.Email]

		if !ok {
			icErr := &lpb.ImportZoomAccountResponse_ImportZoomAccountError{
				RowNumber: int32(index + 1),
				Error:     "Zoom User not found",
			}
			resError.Errors = append(resError.Errors, icErr)
		}
	}
	if len(resError.Errors) == 0 {
		conn, err := z.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
		if err != nil {
			return nil, err
		}
		err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
			if len(zoomAccounts) > 0 {
				err = z.zoomAccountRepo.Upsert(ctx, tx, zoomAccounts)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}
			zoomAccountDeleted, _ := sliceutils.Reduce(zoomAccounts, func(zoomAccountsDeleted []string, z *domain.ZoomAccount) ([]string, error) {
				if z.Action == domain.ActionDelete {
					zoomAccountsDeleted = append(zoomAccountsDeleted, z.ID)
				}
				return zoomAccountsDeleted, nil
			}, make([]string, 0, len(zoomAccounts)))
			if len(zoomAccountDeleted) > 0 {
				return z.lessonRepo.RemoveZoomLinkOfLesson(ctx, tx, zoomAccountDeleted)
			}
			return nil
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return resError, nil
}

func (z *ZoomAccountService) ExportZoomAccount(ctx context.Context, req *lpb.ExportZoomAccountRequest) (res *lpb.ExportZoomAccountResponse, err error) {
	conn, err := z.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	data, err := z.zoomAccountRepo.GetAllZoomAccount(ctx, conn)
	if err != nil {
		return &lpb.ExportZoomAccountResponse{}, status.Error(codes.Internal, err.Error())
	}
	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn:  "zoom_id",
			CSVColumn: domain.ZoomIDLabel,
		},
		{
			DBColumn:  "email",
			CSVColumn: domain.ZoomUsernameLabel,
		},
	}
	exportable := sliceutils.Map(data, func(d *repo.ZoomAccount) database.Entity {
		return d
	})

	str, err := exporter.ExportBatch(exportable, exportCols)
	if err != nil {
		return nil, fmt.Errorf("ExportBatch: %w", err)
	}

	res = &lpb.ExportZoomAccountResponse{
		Data: exporter.ToCSV(str),
	}
	return res, nil
}
