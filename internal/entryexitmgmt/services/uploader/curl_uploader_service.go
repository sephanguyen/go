package uploader

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/curl"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/protobuf/types/known/durationpb"
)

type CurlUploaderService struct {
	UploadReaderService interface {
		GenerateResumableUploadURL(context.Context, *bpb.ResumableUploadURLRequest) (*bpb.ResumableUploadURLResponse, error)
	}
	HTTP curl.IHTTP
}

func (s *CurlUploaderService) InitUploader(ctx context.Context, req *UploadRequest) (*UploadInfo, error) {
	log.Println("CurlUploaderService invoked")

	resumableUploadResponse, err := s.UploadReaderService.GenerateResumableUploadURL(ctx, &bpb.ResumableUploadURLRequest{
		PrefixName:    req.ObjectName,
		Expiry:        durationpb.New(3 * time.Minute),
		FileExtension: req.FileExtension,
		ContentType:   req.ContentType,
	})

	if err != nil {
		return nil, err
	}

	return &UploadInfo{
		DownloadURL: resumableUploadResponse.DownloadUrl,
		DoUploadFromFile: func(ctx context.Context, filePathName string) error {
			if req.FileExtension == constant.PNG {
				return s.upload(filePathName, resumableUploadResponse.ResumableUploadUrl)
			}
			return fmt.Errorf("file extension %s is not supported by this uploader", req.FileExtension)
		},
	}, nil
}

func (s *CurlUploaderService) upload(objectPath string, url string) error {
	imageByte, err := generateImagePngByte(objectPath)
	if err != nil {
		return fmt.Errorf("err generateImagePngByte: %w", err)
	}

	if err := s.HTTP.Request(
		curl.PUT,
		url,
		map[string]string{
			"Content-Type": constant.ContentType,
		},
		bytes.NewReader(imageByte),
		nil,
	); err != nil {
		return fmt.Errorf("err http.Request: %w", err)
	}

	return nil
}

func generateImagePngByte(objectPath string) ([]byte, error) {
	buff := new(bytes.Buffer)
	file, err := os.Open(objectPath)
	if err != nil {
		return nil, fmt.Errorf("err os.Open: %w", err)
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("err image.Decode %w", err)
	}

	err = png.Encode(buff, image)
	if err != nil {
		return nil, fmt.Errorf("err png.Encode %w", err)
	}

	return buff.Bytes(), nil
}
