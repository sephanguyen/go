package uploads

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type UploadReaderService struct {
	bpb.UnimplementedUploadServiceServer
	Cfg       configs.StorageConfig
	FileStore interface {
		GeneratePresignedPutObjectURL(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error)
		GenerateResumableObjectURL(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error)
		GeneratePublicObjectURL(objectName string) string
	}
}

func (s *UploadReaderService) getFilePath(name string) string {
	if len(s.Cfg.FileUploadFolderPath) != 0 {
		name = s.Cfg.FileUploadFolderPath + "/" + name
	}

	return name
}

func (s *UploadReaderService) generateFileName(prefix, fileExtension string) string {
	if len(fileExtension) != 0 {
		fileExtension = "." + strings.ToLower(fileExtension)
	}

	return prefix + idutil.ULIDNow() + fileExtension
}

func (s *UploadReaderService) normalizeExpiry(expiry time.Duration) time.Duration {
	if expiry > s.Cfg.MaximumURLExpiryDuration || expiry < s.Cfg.MinimumURLExpiryDuration {
		expiry = s.Cfg.DefaultURLExpiryDuration
	}

	return expiry
}

func (s *UploadReaderService) GeneratePresignedPutObjectURL(ctx context.Context, req *bpb.PresignedPutObjectRequest) (*bpb.PresignedPutObjectResponse, error) {
	expiry := req.Expiry.AsDuration()
	expiry = s.normalizeExpiry(expiry)
	name := s.generateFileName(req.PrefixName, req.FileExtension)
	name = s.getFilePath(name)

	url, err := s.FileStore.GeneratePresignedPutObjectURL(ctx, name, expiry)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bpb.PresignedPutObjectResponse{
		PresignedUrl: url.String(),
		Expiry:       durationpb.New(expiry),
		Name:         name,
		DownloadUrl:  s.FileStore.GeneratePublicObjectURL(name),
	}, nil
}

func (s *UploadReaderService) GenerateResumableUploadURL(ctx context.Context, req *bpb.ResumableUploadURLRequest) (*bpb.ResumableUploadURLResponse, error) {
	expiry := req.Expiry.AsDuration()
	expiry = s.normalizeExpiry(expiry)
	name := s.generateFileName(req.PrefixName, req.FileExtension)
	name = s.getFilePath(name)

	url, err := s.FileStore.GenerateResumableObjectURL(ctx, name, expiry, req.AllowOrigin, req.ContentType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bpb.ResumableUploadURLResponse{
		ResumableUploadUrl: url.String(),
		Expiry:             durationpb.New(expiry),
		FileName:           name,
		DownloadUrl:        s.FileStore.GeneratePublicObjectURL(name), // replace presigned url soon
	}, nil
}
