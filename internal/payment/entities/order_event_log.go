package entities

import "time"

type (
	OrderEventLog struct {
		OrderStatus         string    `json:"order_status"`
		OrderType           string    `json:"order_type"`
		StudentID           string    `json:"student_id"`
		LocationID          string    `json:"location_id"`
		EnrollmentStatus    string    `json:"enrollment_status"`
		StartDate           time.Time `json:"start_date"`
		EndDate             time.Time `json:"end_date"`
		OrderID             string    `json:"order_id"`
		OrderSequenceNumber int32     `json:"order_sequence_number"`
	}
)
