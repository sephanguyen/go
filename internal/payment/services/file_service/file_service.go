package file_service

import (
	"bytes"
	"context"
	"io"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/services/domain_service/file_storage"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
)

type IFileStorageService interface {
	UploadFile(
		ctx context.Context,
		reader io.Reader,
		db database.QueryExecer,
		fileName string,
		fileType string,
		fileSize int64,
	) (downloadUrl string, err error)
	GetDownloadFileByName(ctx context.Context, db database.QueryExecer, fileName string) (downloadUrl string, err error)
}

type FileService struct {
	DB          database.Ext
	FileStorage IFileStorageService
}

func (s *FileService) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (res *pb.UploadFileResponse, err error) {
	var downloadUrl string
	contentReader := bytes.NewReader(req.Content)
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		downloadUrl, err = s.FileStorage.UploadFile(
			ctx,
			contentReader,
			tx,
			req.FileName.String(),
			req.FileType.String(),
			contentReader.Size(),
		)
		return
	})
	if err != nil {
		return
	}
	res = &pb.UploadFileResponse{DownloadUrl: HardDownloadURL(downloadUrl)}
	return
}

func (s *FileService) GetEnrollmentFile(ctx context.Context, req *pb.GetEnrollmentFileRequest) (res *pb.GetEnrollmentFileResponse, err error) {
	var downloadUrl string
	downloadUrl, _ = s.FileStorage.GetDownloadFileByName(ctx, s.DB, pb.FileName_ENROLLMENT.String())

	res = &pb.GetEnrollmentFileResponse{DownloadUrl: HardDownloadURL(downloadUrl)}
	return
}

func HardDownloadURL(url string) string {
	if len(url) == 0 {
		return ""
	}
	const host = "https://storage.googleapis.com"
	return host + url
}

func NewFileService(db database.Ext, storageConfig configs.StorageConfig) (exportService *FileService, err error) {
	fileStorage, err := file_storage.NewFileStorageService(storageConfig)
	return &FileService{
		DB:          db,
		FileStorage: fileStorage,
	}, err
}
