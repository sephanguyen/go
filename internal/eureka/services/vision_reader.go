package services

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/vision"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Language string

var (
	English  Language = "en"
	Japanese Language = "ja"
)

type VisionReaderService struct {
	VisionFactory vision.Factory
}

func (s *VisionReaderService) DetectTextFromImage(ctx context.Context, req *epb.DetectTextFromImageRequest) (*epb.DetectTextFromImageResponse, error) {
	if req.GetLang() != string(English) && req.GetLang() != string(Japanese) {
		return nil, status.Errorf(codes.InvalidArgument, "lang must be en or ja")
	}
	b64data := req.GetSrc()[strings.IndexByte(req.GetSrc(), ',')+1:]
	content, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "src must be base64: %v", err.Error())
	}

	textAnnotation, err := s.VisionFactory.DetectTextFromImage(ctx, content, req.GetLang())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &epb.DetectTextFromImageResponse{
		Text: textAnnotation.GetText(),
	}, nil
}
