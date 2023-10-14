package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/entities"
	classService "github.com/manabie-com/backend/internal/payment/services/domain_service/class"
	courseService "github.com/manabie-com/backend/internal/payment/services/domain_service/course"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	studentPackageService "github.com/manabie-com/backend/internal/payment/services/domain_service/student_package"
	subscriptionService "github.com/manabie-com/backend/internal/payment/services/domain_service/subscription"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type ISubscriptionServiceForCourseMgMte interface {
	PublishStudentPackage(ctx context.Context, eventMessages []*npb.EventStudentPackage) (err error)
	PublishStudentClass(ctx context.Context, eventMessages []*npb.EventStudentPackageV2) (err error)
}

type IStudentServiceForCourseMgMt interface {
	GetMapLocationAccessStudentByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (mapLocationAccessStudent map[string]interface{}, err error)
}

type ICourseServiceForCourseMgMt interface {
	GetMapLocationAccessCourseForCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (mapLocationAccessCourse map[string]interface{}, err error)
}

type IClassServiceForCourseMgMt interface {
	GetMapClassWithLocationByClassIDs(ctx context.Context, db database.QueryExecer, classIDs []string) (mapClass map[string]entities.Class, err error)
}

type IStudentPackageForCourseMgMt interface {
	UpsertStudentPackage(
		ctx context.Context,
		db database.QueryExecer,
		studentID string,
		mapLocationAccessWithStudentID map[string]interface{},
		mapLocationAccessWithCourseID map[string]interface{},
		mapStudentCourseWithStudentPackageAccessPath map[string]entities.StudentPackageAccessPath,
		importedStudentCourseRows []utils.ImportedStudentCourseRow,
	) (
		events []*npb.EventStudentPackage,
		errors []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError,
	)
	GetMapStudentCourseWithStudentPackageIDByIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (mapStudentCourse map[string]entities.StudentPackageAccessPath, err error)
	UpsertStudentClass(
		ctx context.Context,
		db database.QueryExecer,
		mapStudentCourse map[string]entities.StudentPackageAccessPath,
		mapClass map[string]entities.Class,
		importedStudentClass []utils.ImportedStudentClassRow,
	) (
		events []*npb.EventStudentPackageV2,
		errors []*pb.ImportStudentClassesResponse_ImportStudentClassesError,
	)
	DeleteStudentClass(
		ctx context.Context,
		db database.QueryExecer,
		mapStudentCourse map[string]entities.StudentPackageAccessPath,
		mapClass map[string]entities.Class,
		importedStudentClass []utils.ImportedStudentClassRow,
	) (
		events []*npb.EventStudentPackageV2,
		errors []*pb.ImportStudentClassesResponse_ImportStudentClassesError,
	)
	UpsertStudentPackageForManualFlow(
		ctx context.Context,
		db database.QueryExecer,
		studentID string,
		studentCourse *pb.StudentCourseData,
	) (event *npb.EventStudentPackage, err error)
	UpdateTimeStudentPackageForManualFlow(
		ctx context.Context,
		db database.QueryExecer,
		studentID string,
		studentCourse *pb.StudentCourseData,
	) (event *npb.EventStudentPackage, err error)
}

type CourseMgMt struct {
	DB                  database.Ext
	SubscriptionService ISubscriptionServiceForCourseMgMte
	StudentService      IStudentServiceForCourseMgMt
	StudentPackage      IStudentPackageForCourseMgMt
	CourseService       ICourseServiceForCourseMgMt
	ClassService        IClassServiceForCourseMgMt
}

func NewCourseMgMt(db database.Ext, jsm nats.JetStreamManagement, kafka kafka.KafkaManagement, config configs.CommonConfig) *CourseMgMt {
	return &CourseMgMt{
		DB:                  db,
		SubscriptionService: subscriptionService.NewSubscriptionService(jsm, db, kafka, config),
		StudentService:      studentService.NewStudentService(),
		StudentPackage:      studentPackageService.NewStudentPackage(),
		CourseService:       courseService.NewCourseService(),
		ClassService:        classService.NewClassService(),
	}
}
