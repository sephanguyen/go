package entities

import "time"

type (
	ElasticOrder struct {
		OrderID             string    `json:"order_id"`
		StudentID           string    `json:"student_id"`
		StudentName         string    `json:"student_full_name"`
		LocationID          string    `json:"location_id"`
		OrderSequenceNumber int32     `json:"order_sequence_number"`
		OrderComment        string    `json:"order_comment"`
		OrderStatus         string    `json:"order_status"`
		OrderType           string    `json:"order_type"`
		UpdatedAt           time.Time `json:"updated_at"`
		CreatedAt           time.Time `json:"created_at"`
		IsReviewed          bool      `json:"is_reviewed"`
	}

	ElasticOrderItem struct {
		OrderID     string    `json:"order_id"`
		ProductID   string    `json:"product_id"`
		OrderItemID string    `json:"order_item_id"`
		ProductName string    `json:"product_name"`
		DiscountID  string    `json:"discount_id"`
		StartDate   time.Time `json:"start_date"`
		CreatedAt   time.Time `json:"created_at"`
	}

	ElasticProduct struct {
		ProductID            string    `json:"product_id"`
		Name                 string    `json:"name"`
		ProductType          string    `json:"product_type"`
		TaxID                string    `json:"tax_id"`
		AvailableFrom        time.Time `json:"available_from"`
		AvailableUntil       time.Time `json:"available_until"`
		CustomBillingPeriod  time.Time `json:"custom_billing_period"`
		BillingScheduleID    string    `json:"billing_schedule_id"`
		DisableProRatingFlag bool      `json:"disable_pro_rating_flag"`
		Remarks              string    `json:"remarks"`
		IsArchived           bool      `json:"is_archived"`
		UpdatedAt            time.Time `json:"updated_at"`
		CreatedAt            time.Time `json:"created_at"`
	}
)
