package utils

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
)

type ObjectInfo struct {
	ByteContent []byte
	ObjectName  string
	ContentType filestorage.ContentType
}

type ObjectUploader struct {
	fileStorage         filestorage.FileStorage
	tempFile            *TempFile
	formattedObjectName string
	downloadFileURL     string
	objectInfo          *ObjectInfo
}

func NewObjectUploader(fileStorage filestorage.FileStorage, tempFileCreator ITempFileCreator, objectInfo *ObjectInfo) (*ObjectUploader, error) {
	// Generate the formatted object name and download file URL
	formattedObjectName := fileStorage.FormatObjectName(objectInfo.ObjectName)
	downloadFileURL := fileStorage.GetDownloadURL(formattedObjectName)

	// Create a temporary file
	tempFile, err := tempFileCreator.CreateTempFile(objectInfo.ObjectName)
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file err: %v", err)
	}

	// Write the byte content to the temporary file
	if _, err = tempFile.File.Write(objectInfo.ByteContent); err != nil {
		return nil, fmt.Errorf("cannot write create payment request file: %v error: %v", tempFile.ObjectPath, err)
	}

	return &ObjectUploader{
		fileStorage:         fileStorage,
		tempFile:            tempFile,
		objectInfo:          objectInfo,
		formattedObjectName: formattedObjectName,
		downloadFileURL:     downloadFileURL,
	}, nil
}

func (u *ObjectUploader) DoUploadFile(ctx context.Context) error {
	// Upload the file to storage
	// The object name will be the formatted object name
	// The path name is the path of the temporary file
	if err := u.fileStorage.UploadFile(ctx, filestorage.FileToUploadInfo{
		ObjectName:  u.formattedObjectName,
		PathName:    u.tempFile.ObjectPath,
		ContentType: u.objectInfo.ContentType,
	}); err != nil {
		return fmt.Errorf("UploadFile error: %v", err)
	}
	return nil
}

func (u *ObjectUploader) GetDownloadFileURL() string {
	return u.downloadFileURL
}

func (u *ObjectUploader) GetFormattedObjectName() string {
	return u.formattedObjectName
}

func (u *ObjectUploader) Close() error {
	err := u.tempFile.Close()
	if err != nil {
		return err
	}

	err = u.tempFile.CleanUp()
	if err != nil {
		return err
	}

	return nil
}
