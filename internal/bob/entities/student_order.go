package entities

import (
	"github.com/jackc/pgtype"
)

type StudentOrder struct {
	ID                  pgtype.Int4 `sql:"student_order_id,pk"`
	Amount              pgtype.Numeric
	Currency            pgtype.Text
	PaymentMethod       pgtype.Text
	StudentID           pgtype.Text `sql:"student_id"`
	PackageID           pgtype.Int4 `sql:"package_id"`
	PackageName         pgtype.Text
	Status              pgtype.Text
	Coupon              pgtype.Text
	CouponAmount        pgtype.Numeric
	GatewayResponse     pgtype.Text
	GatewayFullFeedback pgtype.Text
	GatewayLink         pgtype.Text
	Country             pgtype.Text
	GatewayName         pgtype.Text
	IsManualCreated     pgtype.Bool `sql:",notnull"`
	CreatedByEmail      pgtype.Text
	IosTransactionID    pgtype.Text `sql:"ios_transaction_id"`
	InAppTransactionID  pgtype.Text `sql:"inapp_transaction_id"`
	ReferenceNumber     pgtype.Text
	UpdatedAt           pgtype.Timestamptz
	CreatedAt           pgtype.Timestamptz
}

func (rcv *StudentOrder) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"student_order_id", "amount", "currency", "payment_method", "student_id", "package_id", "package_name", "status", "coupon", "coupon_amount", "gateway_response", "gateway_full_feedback", "gateway_link", "country", "gateway_name", "is_manual_created", "created_by_email", "ios_transaction_id", "inapp_transaction_id", "reference_number", "updated_at", "created_at"}
	values = []interface{}{&rcv.ID, &rcv.Amount, &rcv.Currency, &rcv.PaymentMethod, &rcv.StudentID, &rcv.PackageID, &rcv.PackageName, &rcv.Status, &rcv.Coupon, &rcv.CouponAmount, &rcv.GatewayResponse, &rcv.GatewayFullFeedback, &rcv.GatewayLink, &rcv.Country, &rcv.GatewayName, &rcv.IsManualCreated, &rcv.CreatedByEmail, &rcv.IosTransactionID, &rcv.InAppTransactionID, &rcv.ReferenceNumber, &rcv.UpdatedAt, &rcv.CreatedAt}
	return
}

func (*StudentOrder) TableName() string {
	return "student_orders"
}
