package constant

import (
	"time"

	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

const (
	NoDataInCsvFile                                            = "no data in csv file"
	DateFormatYYYYMMDD                                         = "2006-01-02"
	DayDuration                                                = 24 * time.Hour
	StudentBillingTabPathTemplate                              = "/user/students_erp/%s/show?tab=StudentDetail__billing"
	EngNotificationContentTempForStudentProductWithScheduleTag = "The system cannot create an update order - %s for %s due to the pending “scheduled“ tag."
	JpNotificationContentTempForStudentProductWithScheduleTag  = "%sの%sは保留中の予約事項があるため、変更オーダーの作成ができませんでした。"
)

var (
	SpecialDiscounts = map[string]bool{
		paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String():   true,
		paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String(): true,
	}
)
