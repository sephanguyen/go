package services

import (
	"context"
	"crypto/md5" // nolint
	"fmt"
	"io"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadModifierService struct {
	DBTrace database.Ext
	Logger  *zap.Logger
	Config  *configurations.Config

	Uploader
}

func (s *UploadModifierService) UploadHtmlContent(ctx context.Context, req *pb.UploadHtmlContentRequest) (*pb.UploadHtmlContentResponse, error) {
	url, fileName := s.generateUploadURL(s.Config.Storage.Endpoint, s.Config.Storage.Bucket, req.GetContent())

	if s.Config.Common.Environment != "local" {
		client, err := storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("err storage.NewClient: %w", err)
		}
		wc := client.Bucket(s.Config.Storage.Bucket).Object(fileName[1:]).NewWriter(ctx)
		err = uploadToCloudStorage(wc, req.GetContent(), "text/html; charset=utf-8")
		if err != nil {
			return nil, fmt.Errorf("err uploadToCloudStorage: %w", err)
		}
	} else {
		err := uploadToS3(ctx, s.Uploader, req.GetContent(), s.Config.Storage.Bucket, fileName, "text/html; charset=UTF-8")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "err uploadToS3: %v", err.Error())
		}
	}

	return &pb.UploadHtmlContentResponse{
		Url: url,
	}, nil
}

func (s *UploadModifierService) BulkUploadHtmlContent(ctx context.Context, req *pb.BulkUploadHtmlContentRequest) (*pb.BulkUploadHtmlContentResponse, error) {
	objects := make([]*UploadObject, 0, len(req.Contents))
	urls := make([]string, 0, len(req.Contents))
	if s.Config.Common.Environment != "local" {
		client, err := storage.NewClient(ctx)
		for _, content := range req.Contents {
			url, fileName := s.generateUploadURL(s.Config.Storage.Endpoint, s.Config.Storage.Bucket, content)
			if err != nil {
				return nil, fmt.Errorf("err storage.NewClient: %w", err)
			}
			wc := client.Bucket(s.Config.Storage.Bucket).Object(fileName[1:]).NewWriter(ctx)
			err = uploadToCloudStorage(wc, content, "text/html; charset=utf-8")
			if err != nil {
				return nil, fmt.Errorf("err uploadToCloudStorage: %w", err)
			}
			urls = append(urls, url)
		}
	} else {
		for _, content := range req.Contents {
			url, fileName := s.generateUploadURL(s.Config.Storage.Endpoint, s.Config.Storage.Bucket, content)
			objects = append(objects, &UploadObject{
				Data:        content,
				Bucket:      s.Config.Storage.Bucket,
				Path:        fileName,
				ContentType: "text/html",
			})
			urls = append(urls, url)
		}

		if err := bulkUploadToS3(ctx, s.Uploader, objects); err != nil {
			return nil, status.Errorf(codes.Internal, "err bulkUploadToS3: %v", err.Error())
		}
	}
	return &pb.BulkUploadHtmlContentResponse{
		Urls: urls,
	}, nil
}

// nolint
func (s *UploadModifierService) generateUploadURL(endpoint, bucket, content string) (url, fileName string) {
	h := md5.New()
	io.WriteString(h, content)
	fileName = "/content/" + fmt.Sprintf("%x.html", h.Sum(nil))

	return endpoint + "/" + bucket + fileName, fileName
}

func (s *UploadModifierService) BulkUploadFile(ctx context.Context, req *pb.BulkUploadFileRequest) (*pb.BulkUploadFileResponse, error) {
	objects := make([]*UploadPayloadObject, 0, len(req.Files))
	fileResp := make([]*pb.BulkUploadFileResponse_File, 0, len(req.Files))
	for _, file := range req.Files {
		url, path := s.generateUploadFileURL(s.Config.Storage.Endpoint, s.Config.Storage.Bucket, file.FileName)
		objects = append(objects, &UploadPayloadObject{
			Data:        file.Payload,
			Bucket:      s.Config.Storage.Bucket,
			Path:        path,
			ContentType: file.ContentType,
		})
		fileResp = append(fileResp, &pb.BulkUploadFileResponse_File{
			FileName: file.FileName,
			Url:      url,
		})
	}
	if err := bulkUploadPayloadToS3(ctx, s.Uploader, objects); err != nil {
		return nil, status.Errorf(codes.Internal, "err bulkUploadPayloadToS3: %v", err.Error())
	}
	return &pb.BulkUploadFileResponse{
		Files: fileResp,
	}, nil
}

func (s *UploadModifierService) generateUploadFileURL(endpoint, bucket, fileName string) (url, path string) {
	path = "/file/" + fileName
	return endpoint + "/" + bucket + path, path
}
