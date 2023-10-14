package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GenerateProductWrongColumNameTestCases(ctx context.Context, productType pb.ProductType, columnNames []string, lines string) []utils.TestCase {
	numberColumns := 1
	testCases := make([]utils.TestCase, 0, numberColumns)
	for idx := 0; idx < numberColumns; idx++ {
		wrongIdx := idx
		wrongColumnNames := make([]string, 0, numberColumns)
		wrongColumnNames = append(
			wrongColumnNames,
			columnNames[:wrongIdx]...,
		)
		wrongColumnNames = append(
			wrongColumnNames,
			fmt.Sprintf("%s_wrong_name", columnNames[wrongIdx]),
		)
		wrongColumnNames = append(
			wrongColumnNames,
			columnNames[wrongIdx+1:]...,
		)

		testCases = append(testCases, utils.TestCase{
			Name: fmt.Sprintf("invalid file - %s column name (toLowerCase) != %s", utils.NumberNames[wrongIdx], columnNames[wrongIdx]),
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			ExpectedErr: status.Error(
				codes.InvalidArgument,
				fmt.Sprintf("csv file invalid format - %s column (toLowerCase) should be '%s'", utils.NumberNames[wrongIdx], columnNames[wrongIdx]),
			),
			Req: &pb.ImportProductRequest{
				ProductType: productType,
				Payload: []byte(
					fmt.Sprintf(
						`%s
						%s`,
						strings.Join(wrongColumnNames, ","),
						lines,
					),
				),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		})
	}
	return testCases
}
