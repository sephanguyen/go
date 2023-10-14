package file_storage

import (
	"context"
	"io"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/services/domain_service/file_storage/storage"
	"github.com/manabie-com/backend/internal/payment/utils"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileStorageService struct {
	Store    utils.IStorage
	FileRepo interface {
		GetByFileName(ctx context.Context, db database.QueryExecer, fileName string) (entities.File, error)
		Create(ctx context.Context, db database.QueryExecer, e *entities.File) error
	}
}

func (s *FileStorageService) UploadFile(
	ctx context.Context,
	reader io.Reader,
	db database.QueryExecer,
	fileName string,
	fileType string,
	fileSize int64,
) (downloadUrl string, err error) {
	var (
		file      entities.File
		fileID    string
		isNewFile bool
	)
	file, err = s.FileRepo.GetByFileName(ctx, db, fileName)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		err = status.Errorf(codes.Internal, "get file by name have error %v", err.Error())
		return
	}
	if file.FileID.Status == pgtype.Present {
		fileID = file.FileID.String
	} else {
		fileID = idutil.ULIDNow()
		isNewFile = true
	}

	downloadUrl, err = s.Store.UploadFromFile(ctx, reader, fileID, fileType, fileSize)
	if err != nil {
		err = status.Errorf(codes.Internal, "uploading file have error %v", err.Error())
		return
	}

	fileEntity := entities.File{}
	err = multierr.Combine(
		fileEntity.FileID.Set(fileID),
		fileEntity.FileName.Set(fileName),
		fileEntity.FileType.Set(fileType),
		fileEntity.DownloadLink.Set(downloadUrl),
		fileEntity.DeletedAt.Set(nil),
	)

	if err != nil {
		err = status.Errorf(codes.Internal, "assigning data to file entity have error %v", err.Error())
		return
	}

	if !isNewFile {
		return
	}

	err = s.FileRepo.Create(ctx, db, &fileEntity)
	if err != nil {
		err = status.Errorf(codes.Internal, "creating file in repo  have error %v", err.Error())
	}
	return
}

func (s *FileStorageService) GetDownloadFileByName(ctx context.Context, db database.QueryExecer, fileName string) (downloadUrl string, err error) {
	var (
		file entities.File
	)
	file, err = s.FileRepo.GetByFileName(ctx, db, fileName)
	if err != nil {
		err = status.Errorf(codes.Internal, "getting file by name have error %v", err.Error())
		return
	}

	downloadUrl = file.DownloadLink.String
	return
}

func NewFileStorageService(storageConfig configs.StorageConfig) (*FileStorageService, error) {
	storageService, err := storage.NewStorageService(storageConfig)
	return &FileStorageService{
		Store:    storageService,
		FileRepo: &repositories.FileRepo{},
	}, err
}
