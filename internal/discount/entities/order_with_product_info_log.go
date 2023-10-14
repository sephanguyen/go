package entities

type (
	OrderWithProductInfoLog struct {
		OrderID           string   `json:"order_id"`
		StudentID         string   `json:"student_id"`
		LocationID        string   `json:"location_id"`
		OrderStatus       string   `json:"order_status"`
		OrderType         string   `json:"order_type"`
		StudentProductIDs []string `json:"student_product_ids"`
	}
)
