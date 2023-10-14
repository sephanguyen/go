package uploader

import (
	"context"
	"log"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type Uploader interface {
	InitUploader(ctx context.Context, req *UploadRequest) (*UploadInfo, error)
}

type UploadRequest struct {
	ObjectName    string
	FileExtension string
	ContentType   string
}

type UploadInfo struct {
	DownloadURL      string
	DoUploadFromFile func(ctx context.Context, filePathName string) error
}

type SDKUploaderService struct {
	Cfg       *configs.StorageConfig
	FileStore interface {
		UploadFromFile(ctx context.Context, objectName, pathName, contentType string) error
		GetDownloadURL(objectName string) string
	}
}

func (s *SDKUploaderService) formatObjectName(objectName, fileExtension string) string {
	// Formats the object name and add an ID
	if len(fileExtension) != 0 {
		fileExtension = "." + strings.ToLower(fileExtension)
	}
	objectName = objectName + idutil.ULIDNow() + fileExtension

	// Check for file folder path
	if len(s.Cfg.FileUploadFolderPath) != 0 {
		objectName = s.Cfg.FileUploadFolderPath + "/" + objectName
	}

	return objectName
}

func (s *SDKUploaderService) InitUploader(ctx context.Context, req *UploadRequest) (*UploadInfo, error) {
	log.Println("SDKUploaderService invoked")

	objectName := s.formatObjectName(req.ObjectName, req.FileExtension)
	downloadURL := s.FileStore.GetDownloadURL(objectName)

	return &UploadInfo{
		DownloadURL: downloadURL,
		DoUploadFromFile: func(ctx context.Context, filePathName string) error {
			return s.FileStore.UploadFromFile(ctx, objectName, filePathName, req.ContentType)
		},
	}, nil
}
