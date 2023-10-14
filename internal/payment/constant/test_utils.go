package constant

import (
	"fmt"

	"github.com/jackc/pgtype"

	"github.com/jackc/pgconn"
)

const (
	StudentID                           = "Student-1234"
	ProductID                           = "Product-1"
	LocationID                          = "Location-1234"
	UserID                              = "User-1"
	OrderID                             = "Order-123"
	StudentProductID                    = "Student-Product-123"
	StudentName                         = "Student Learner"
	LocationName                        = "Location-Name"
	OrderComment                        = "This is an order comment"
	TotalSuccessForSearchEngine         = 1
	DefaultPrice                        = 1000
	CustomBillingItemName               = "Default custom billing item"
	DiscountID                          = "Discount-1"
	ProductName                         = "Product-name-1"
	TaxID                               = "Tax-1"
	CourseID                            = "Course-1"
	BillingSchedulePeriodID             = "billing_schedule_period_id_12"
	StudentPackageID                    = "Student-package-1234"
	CourseName                          = "Course-name-1"
	PackageID                           = "package-1"
	OrderItemID                         = "order-item-1"
	HappyCase                           = "Happy Case"
	FailCaseErrorQuery                  = "Fail case: Error when query"
	FailCaseErrorRow                    = "Fail case: Error when scan rows"
	FailCaseErrorCreateActionLog        = "Fail case: Error when create order action log"
	InvalidOrderID                      = "invalid-order-id"
	BillingScheduleID                   = "billing_schedule_id_12"
	FeeID                               = "Fee-1"
	MaterialID                          = "Material-1"
	UpcomingStudentPackageID            = "Upcoming-student-package-id-1"
	UpcomingStudentCourseID             = "Upcoming-student-course-id-1"
	StudentSubscriptionStudentPackageID = "student-subscription-student-package-id-1"
	OrderVersionNumber                  = 1
	LeavingReasonID                     = "leaving-reason-id-1"
	StudentDetailPath                   = "/admin/student/student-id-123"
	StudentPackageOrderID               = "student-package-order-id-1"
	FromStudentPackageOrderID           = "student-package-order-id-2"
)

var (
	SuccessCommandTag = pgconn.CommandTag([]byte(`1`))
	FailCommandTag    = pgconn.CommandTag([]byte(`0`))
	ErrDefault        = fmt.Errorf("error something")
	DefaultGrade      = pgtype.Text{String: "1", Status: pgtype.Present}
)
