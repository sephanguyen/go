package entities

import (
	"time"

	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type (
	UpdateProductDiscount struct {
		StudentID             string                       `json:"student_id"`
		LocationID            string                       `json:"location_id"`
		ProductID             string                       `json:"product_id"`
		StudentProductID      string                       `json:"student_product_id"`
		EffectiveDate         time.Time                    `json:"effective_date"`
		StudentProductEndDate time.Time                    `json:"student_product_end_date"`
		DiscountID            string                       `json:"discount_id"`
		DiscountType          paymentPb.DiscountType       `json:"discount_type"`
		DiscountAmountType    paymentPb.DiscountAmountType `json:"discount_amount_type"`
		DiscountAmountValue   float32                      `json:"discount_amount_value"`
	}
)
