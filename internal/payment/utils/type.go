package utils

import (
	"context"
	"io"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/payment/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type OrderType int

const (
	OrderCreate OrderType = iota
	OrderUpdate
	OrderCancel
	OrderEnrollment
	OrderWithdraw
	OrderGraduate
	OrderResume
	OrderLOA
	OrderCustom
)

type OrderItemData struct {
	Order              entities.Order
	StudentInfo        entities.Student
	ProductInfo        entities.Product
	PackageInfo        PackageInfo
	StudentProduct     entities.StudentProduct
	RootStudentProduct entities.StudentProduct
	ProductSetting     entities.ProductSetting

	StudentName            string
	LocationName           string
	IsEnrolledInLocation   bool
	IsOneTimeProduct       bool
	IsDisableProRatingFlag bool
	ProductType            pb.ProductType
	OrderItem              *pb.OrderItem
	BillItems              []BillingItemData
	GradeName              string
	Timezone               int32
	PriceType              string
}

type BillingItemData struct {
	BillingItem *pb.BillingItem
	IsUpcoming  bool
}

type MessageSyncData struct {
	OrderType                 OrderType
	Order                     entities.Order
	Student                   entities.Student
	StudentCourseMessage      map[string][]*pb.EventSyncStudentPackageCourse
	StudentPackages           []*npb.EventStudentPackage
	StudentProducts           []entities.StudentProduct
	SystemNotificationMessage *payload.UpsertSystemNotification
}

type UpsertSystemNotificationData struct {
	StudentDetailPath string
	LocationName      string
	EndDate           time.Time
	StartDate         time.Time
	Timezone          int32
}

type ElasticSearchData struct {
	Order      entities.Order
	Products   []entities.Product
	OrderItems []entities.OrderItem
}

type PackageInfo struct {
	MapCourseInfo     map[string]*pb.CourseItem
	Package           entities.Package
	QuantityType      pb.QuantityType
	StudentCourseSync []*pb.EventSyncStudentPackageCourse
	Quantity          int32
}

type BillItemForRetrieveApi struct {
	BillItemDescription *pb.BillItemDescription
	BillItemEntity      entities.BillItem
}

type IBillingService interface {
	CreateBillItemForOrderCreate(ctx context.Context, db database.QueryExecer, orderItemData OrderItemData) (err error)
	CreateBillItemForOrderUpdate(ctx context.Context, db database.QueryExecer, orderItemData OrderItemData) (err error)
	CreateBillItemForOrderWithdrawal(ctx context.Context, db database.QueryExecer, orderItemData OrderItemData) (err error)
	CreateBillItemForOrderLOA(ctx context.Context, db database.QueryExecer, orderItemData OrderItemData) (err error)
	CreateBillItemForOrderGraduate(ctx context.Context, db database.QueryExecer, orderItemData OrderItemData) (err error)
	CreateBillItemForOrderCancel(ctx context.Context, db database.QueryExecer, orderItemData OrderItemData) (err error)
}

type IStorage interface {
	UploadFromFile(ctx context.Context, file io.Reader, fileID string, contentType string, fileSize int64) (downloadUrl string, err error)
}

type ImportedStudentCourseRow struct {
	Row                      int32
	StudentPackage           *entities.StudentPackages
	StudentPackageAccessPath *entities.StudentPackageAccessPath
	StudentPackageEvent      *npb.EventStudentPackage
}

type ImportedStudentClassRow struct {
	Row       int32
	StudentID string
	CourseID  string
	ClassID   string
}

type TestCase struct {
	Name                string
	Ctx                 context.Context
	Req                 interface{}
	ExpectedResp        interface{}
	ExpectedErr         error
	ExpectErrorMessages interface{}
	Setup               func(ctx context.Context)
}

type VoidStudentPackageArgs struct {
	Order          entities.Order
	StudentProduct entities.StudentProduct
	Product        entities.Product
	IsCancel       bool
}

type TimeRange struct {
	FromTime time.Time
	ToTime   time.Time
}
