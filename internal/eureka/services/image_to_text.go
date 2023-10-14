package services

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/mathpix"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImageToText struct {
	MathpixFactory interface {
		DetectLatexFromImage(ctx context.Context, content string) ([]mathpix.Data, error)
	}
}

func (s *ImageToText) DetectFormula(ctx context.Context, req *epb.DetectFormulaRequest) (*epb.DetectFormulaResponse, error) {
	content := req.GetSrc()
	if content == "" || !strings.Contains(content, ",") {
		return nil, status.Errorf(codes.InvalidArgument, "invalid src")
	}

	mathpixData, err := s.MathpixFactory.DetectLatexFromImage(ctx, content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "MathpixFactory.DetectLatexFromImage: %v", err.Error())
	}

	if len(mathpixData) == 0 {
		return &epb.DetectFormulaResponse{}, nil
	}

	formulaEpb := make([]*epb.DetectFormulaResponse_Formula, 0, len(mathpixData))
	for _, data := range mathpixData {
		formulaEpb = append(formulaEpb, &epb.DetectFormulaResponse_Formula{
			Type:  data.Type,
			Value: data.Value,
		})
	}
	return &epb.DetectFormulaResponse{
		Formulas: formulaEpb,
	}, nil
}
