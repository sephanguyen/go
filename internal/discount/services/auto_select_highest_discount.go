package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

func (s *InternalService) AutoSelectHighestDiscount(ctx context.Context, req *pb.AutoSelectHighestDiscountRequest) (res *pb.AutoSelectHighestDiscountResponse, err error) {
	var (
		studentProducts      []*entities.StudentProduct
		highestDiscount      entities.Discount
		currentDiscount      entities.Discount
		totalUpdatedProducts int32
	)

	zlogger := ctxzap.Extract(ctx).Sugar()
	orgID := req.OrganizationId
	automationErrors := []*pb.AutoSelectHighestDiscountResponse_AutoSelectHighestDiscountError{}

	studentsCandidateForDiscountUpdate, err := s.RetrieveStudentsCandidateForDiscountUpdateOnDate(ctx, time.Now())
	zlogger.Info(fmt.Sprintf("Automation log for organization %v: %v candidate students for discount update and with error %v", orgID, len(studentsCandidateForDiscountUpdate), err))
	if err != nil {
		automationErrors = append(automationErrors, &pb.AutoSelectHighestDiscountResponse_AutoSelectHighestDiscountError{
			Error: fmt.Sprintf("failed to retrieve candidate students for discount automation with error %v", err),
		})
	}

	totalUpdatedProducts = 0
	for _, student := range studentsCandidateForDiscountUpdate {
		studentProducts, err = s.RetrieveActiveStudentProductsOfStudentInLocation(ctx, student.StudentID, student.LocationID)
		zlogger.Info(fmt.Sprintf("Automation log for organization %v: RetrieveStudentProductsOfStudentInLocation with error %v", orgID, err))
		if err != nil {
			automationErrors = append(automationErrors, &pb.AutoSelectHighestDiscountResponse_AutoSelectHighestDiscountError{
				StudentId: student.StudentID,
				Error:     fmt.Sprintf("failed to retrieve student product of student with error %v", err),
			})
		}

		zlogger.Info(fmt.Sprintf("Automation log for organization %v: %v candidate products for student %v in location %v and with error %v", orgID, len(studentProducts), student.StudentID, student.LocationID, err))
		for _, product := range studentProducts {
			highestDiscount, err = s.RetrieveHighestDiscountOfStudentProduct(ctx, student.StudentID, student.LocationID, product.ProductID.String)
			zlogger.Info(fmt.Sprintf("Automation log for organization %v: RetrieveHighestDiscountOfStudentProduct %v with discount_id %v and with error %v", orgID, product.StudentProductID.String, highestDiscount.DiscountID.String, err))
			if err != nil {
				automationErrors = append(automationErrors, &pb.AutoSelectHighestDiscountResponse_AutoSelectHighestDiscountError{
					StudentId:        student.StudentID,
					StudentProductId: product.ProductID.String,
					Error:            fmt.Sprintf("failed for student %v product %v: retrieve highest discount of student product with error %v", student.StudentID, product.StudentProductID, err),
				})
			}

			currentDiscount, err = s.RetrieveCurrentDiscountOfStudentProduct(ctx, product.StudentProductID.String)
			zlogger.Info(fmt.Sprintf("Automation log for organization %v: RetrieveCurrentDiscountOfStudentProduct %v with discount_id %v and with error %v", orgID, product.StudentProductID.String, currentDiscount.DiscountID.String, err))
			if err != nil {
				automationErrors = append(automationErrors, &pb.AutoSelectHighestDiscountResponse_AutoSelectHighestDiscountError{
					StudentId:        student.StudentID,
					StudentProductId: product.ProductID.String,
					Error:            fmt.Sprintf("failed for student %v product %v: retrieve current discount of student product with error %v", student.StudentID, product.StudentProductID, err),
				})
			}

			if len(automationErrors) == 0 && highestDiscount.DiscountID.String != currentDiscount.DiscountID.String {
				err = s.ValidateProductAndPublishUpdateOrderEvent(ctx, product.StudentProductID.String, highestDiscount)
				zlogger.Info(fmt.Sprintf("Automation log for organization %v: ValidateProductAndPublishUpdateOrderEvent for student product %v with error %v", orgID, product.StudentProductID.String, err))
				if err != nil {
					automationErrors = append(automationErrors, &pb.AutoSelectHighestDiscountResponse_AutoSelectHighestDiscountError{
						StudentId:        student.StudentID,
						StudentProductId: product.ProductID.String,
						Error:            fmt.Sprintf("failed for student %v product %v: publish update product event with error %v", student.StudentID, product.StudentProductID, err),
					})
				} else {
					totalUpdatedProducts++
				}
			}
		}
	}

	res = &pb.AutoSelectHighestDiscountResponse{
		TotalUpdatedProducts: totalUpdatedProducts,
		Errors:               automationErrors,
	}

	return
}
