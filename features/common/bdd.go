package common

import (
	"context"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/filestore"
	calendar_dto "github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/gandalf/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	invoice_entity "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	asg_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	classdo_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_report_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure/repo"
	zoom_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	course_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	timesheet_dto "github.com/manabie-com/backend/internal/timesheet/domain/dto"
	timesheet_entity "github.com/manabie-com/backend/internal/timesheet/domain/entity"
	user_repo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	virDomain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb_tom "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/go-kafka/connect"
	"github.com/nats-io/nats.go"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// StepState is the step state for all team
type StepState struct {
	// root account when running bdd
	RootAccount map[int]AuthInfo

	OldAuthToken  string
	AuthToken     string
	Request       interface{}
	Response      interface{}
	ResponseErr   error
	RequestSentAt time.Time

	// Error model
	ExpectedErrModel *errdetails.BadRequest

	// For REST API
	RestResponse     interface{}
	RestResponseBody []byte

	//
	Request1  interface{}
	Request2  interface{}
	Response1 interface{}
	Response2 interface{}
	Requests  []interface{}
	Responses []interface{}

	Subs                  []*nats.Subscription
	FoundChanForJetStream chan interface{}
	CurrentTeacherID      string
	CurrentLessonID       string
	CurrentStaffID        string

	AssignedStudentIDs        []string
	UnAssignedStudentIDs      []string
	OtherStudentIDs           []string
	CurrentSchoolID           int32
	DefaultLocationID         string
	DefaultLocationTypeID     string
	CurrentStudentID          string
	CurrentUserID             string
	StudentID                 string
	ApplicantID               string
	FirebaseAddress           string
	CurrentUserGroup          string
	Random                    string
	Schools                   []*bob_entities.School
	BookID                    string
	CourseID                  string
	StudyPlanID               string
	ChapterID                 string
	CurrentChapterIDs         []string
	Topics                    []*epb.Topic
	TopicID                   string
	TopicIDs                  []string
	CurrentClassCode          string
	CurrentClassID            int32
	Courses                   []interface{}
	CourseIDs                 []string
	ImportedClass             []*class_domain.Class
	JoinedClass               []string
	LeavedClass               []string
	MediaIDs                  []string
	Medias                    []*pb.Media
	StudentIds                []string
	CurrentCustomAssignmentID string
	AllStudentSubmissions     []string

	// lesson
	RoleName               string
	LessonIDs              []string
	Lessons                []*bob_entities.Lesson
	LessonDomains          []*domain.Lesson
	Lesson                 *domain.Lesson
	TeacherIDs             []string
	GradeIDs               []string
	TeacherIDsUpdateLesson []string
	StudentIDWithCourseID  []string // will like: [student_id, course_id, ... 2*n]
	StudentPackageID       string
	SubmitPollingAnswer    []string
	FoundChanForLessonES   chan interface{}
	MediaItems             []*lpb.Media
	OldSchedulerID         string
	OldEndDate             time.Time
	LessonDates            map[string]string
	SavingType             lpb.SavingType
	TeacherNames           []string
	StudentNames           []string
	ZoomAccount            *zoom_domain.ZoomAccount
	ZoomLink               string
	ClassDoAccount         *classdo_domain.ClassDoAccount
	ClassDoLink            string
	ClassDoRoomID          string
	// for get lesson student slot info
	StudentSlotInfoCount           int
	ImportLessonPartnerInternalIDs []string

	// student_subscription
	StartDate time.Time
	EndDate   time.Time

	// lesson_report
	LessonReportID                 string
	FormConfigID                   string
	DynamicFieldValuesByUser       map[string][]*bob_entities.PartnerDynamicFormFieldValue
	LessonDynamicFieldValuesByUser map[string][]*lesson_report_repo.PartnerDynamicFormFieldValueDTO

	// virtual_classroom
	CurrentWhiteboardZoomState  *virDomain.WhiteboardZoomState
	RecordedVideos              virDomain.RecordedVideos
	AgoraSignature              string
	IsExpectingASpotlightedUser bool
	NumberOfStream              int
	GetLessonsSelectedFilter    string
	ArrangedLessonIDs           []string
	OffsetIndex                 int
	CreateStressTestLocation    bool
	LiveLessonConversations     []virDomain.LiveLessonConversation
	ExpectedPrivConversationMap map[string]string
	ExpectedPrivConversationID  string

	// virtual_classroom - live room
	CurrentChannelName   string
	CurrentChannelID     string
	CurrentLiveRoomLogID string

	// communication
	SchoolID                    string
	ConversationIDs             []string
	TeacherToken                string
	TeacherID                   string
	SchoolIds                   []string
	JoinedConversationIDs       []string
	OldConversations            map[string]Message
	CurrentNotificationID       string
	CurrentNotificationTargetID string
	SubV2Clients                map[string]CancellableStream

	// new notification
	NotificationID             string
	Notification               *cpb.Notification
	NotificationNeedToSent     *cpb.Notification
	NotificationDontNeedToSent *cpb.Notification

	Grades           []int
	CurrentCourseID  string
	CourseStudentIDs map[string][]string
	StudentParent    map[string][]string
	LessonGroupID    string
	MaterialIds      []string
	GradeName        string

	// For retrieve lesson with filter & search
	FilterCourseIDs     []string
	FilterFromTime      time.Time
	FilterToTime        time.Time
	SchoolIDs           []int32
	CurrentCourseTypeID string
	// For retrieve lesson management with filter & search
	FilterTeacherIDs         []string
	FilterStudentIDs         []string
	FilterCenterIDs          []string
	FilterSchedulingStatuses []domain.LessonSchedulingStatus
	FilterFromDate           time.Time
	FilterToDate             time.Time
	RandSchoolID             int
	// for retrieve lesson student subscription
	FilterClassIDs    []string
	FilterLocationIDs []string
	FilterGradeIDs    []string

	// for retrieve student attendance
	FilterAttendanceStatus string

	// for retrieve assigned student list
	FilterAssignedStatus []asg_domain.AssignedStudentStatus
	PurchaseMethod       string
	PageNumber           int
	PageLimit            int
	Offset               string
	PreOffset            string
	NextOffset           string
	// for retrieve lessons on calendar
	FilterStudentsCount int
	FilterCoursesCount  int
	FilterTeachersCount int
	FilterClassesCount  int
	NoneAssignedTeacher bool
	// lesson classroom
	ClassroomIDs []string
	// For retrieve live lessons by locations
	LocationIDs            []string
	LocationTypesID        []string
	LowestLevelLocationIDs []string
	TimeRandom             time.Time
	// For lesson group CRUD
	CurrentTeachingMethod string
	// payment
	Examples           interface{}
	ValidCsvRows       []string
	InvalidCsvRows     []string
	OverwrittenCsvRows []string
	NameOfData         string
	NumberOfBillItems  int
	UserData           *bob_entities.User
	LocationData       *bob_entities.Location
	ScenarioData       map[string]*GenerateBillItemsUtils

	// usermgmt
	ShardID                      int64
	SrcUser                      user.User
	SrcTenant                    multitenant.Tenant
	DestTenant                   multitenant.Tenant
	DestUser                     user.User
	OrganizationID               string
	TenantID                     string
	ParentIDs                    []string
	UserIDs                      []string
	StudentEmails                []string
	Users                        []*entity.LegacyUser
	ExistingStudents             []*entity.LegacyStudent
	ParentPassword               string
	MapExistingPackageAndCourses map[string]string
	ExistingLocations            []*location_repo.Location
	UserProfiles                 []*cpb.BasicProfile
	NumberOfIds                  int
	ExistedUserGroupID           string
	FirebaseResourceIDs          []string
	EvtImportStudents            []*upb.EvtImportStudent_ImportStudent
	ImportUserEvents             entity.ImportUserEvents
	PartnerInternalIDs           []string
	TagInternalIDs               []string
	GradeInternalIDs             []string
	ManabieGradeIDs              []string
	TagIDs                       []string
	MapOrgStaff                  map[int]MapRoleAndAuthInfo
	SchoolHistories              []user_repo.SchoolHistory
	IsUserNameEnabled            bool
	NumberValidCsvRows           int
	NumberInvalidCsvRows         int
	ExpectedData                 interface{}
	// Hexagon entities
	AuthUser                             entity.User
	AuthUserFirebaseIDToken              string
	AuthUserFirebaseRefreshToken         string
	AuthUserIdentityPlatformTokenID      string
	AuthUserIdentityPlatformRefreshToken string

	UserIDToExternalID           map[string]string
	BucketNameJobMigrationStatus string
	ObjectNameJobMigrationStatus string

	// entryexitmgmt
	NotifyParentRequest   bool
	ParentNotified        bool
	StudentName           string
	RetrieveRecordCount   int32
	CurrentParentID       string
	TimeZone              string
	BatchQRCodeStudentIds []string

	// mastermgmt
	CenterIDs         []string
	ExpectedCSV       []string
	LocationTypeIDs   []string
	LocationTypes     []*location_domain.LocationType
	LocationTypeOrgID string
	FilterStudentSubs []string
	// nolint
	CurrentClassId     string
	ClassIds           []string
	CourseTypeIDs      []string
	SeedingCourseIDs   []string
	AuthKey            string
	AuthValue          string
	CoursesExportCSV   string
	ExpectedError      string
	NewOrgID           string
	NewConfigKey       string
	TreeLocation       *location_domain.TreeLocation
	ConfigKeys         []string
	LocationConfigKeys map[string]string
	SubjectIDs         []string
	CourseProcessing   []*course_domain.Course // just avoid name conflict
	AcademicYearIDs    []string
	WorkingHoursIDs    []string
	AuditConfigOrgID   string
	AuditConfigKey     string

	// unleash
	UnleashConfigFilePath string
	UnleashFileContent    interface{}

	// enigma
	JPREPSignature            string
	CurrentSchoolIDString     string
	PartnerSyncLogID          string
	BodyBytes                 string
	PartnerSyncDataLogID      string
	PartnerSyncDataLogSplitID string

	// student's comment
	CommentIDs []string

	// invoicemgmt
	ErrorList                     []error
	ConcurrentCount               int
	BillItemSequenceNumbers       []int32
	ProductID                     string
	LocationID                    string
	TaxID                         string
	BillingScheduleID             string
	BillingSchedulePeriodID       string
	ResourcePath                  string
	InvoiceID                     string
	CurrentVirtualClassroomLogID  string
	BillItemSequenceNumber        int32
	StudentProductID              string
	OrderID                       string
	InvoiceScheduleID             string
	InvoiceScheduleHistoryID      string
	OrganizationStudentNumberMap  map[string]int
	OrganizationStudentListMap    map[string][]string
	OrganizationInvoiceHistoryMap map[string]string
	StudentBillItemMap            map[string][]int32
	InvoiceStudentMap             map[string]string
	InvoiceIDs                    []string
	BankID                        string
	DiscountID                    string
	PartnerBankIDs                []string
	PartnerConvenienceStoreID     string
	PaymentID                     string
	PaymentIDs                    []string
	PaymentRequestFileIDs         []string
	OrganizationCountry           string
	InvoiceScheduleDates          []time.Time
	BulkPaymentRequestID          string
	BulkPaymentRequestFileID      string
	PartnerBankID                 string
	PrefectureID                  string
	PrefectureCode                string
	BulkPaymentValidationsID      string
	PaymentMethod                 string
	InvoiceTotal                  int64
	PaymentSeqNumbers             []int32
	PaymentFile                   []byte
	InvoiceTotalAmount            []float64
	NoOfStudentsInvoiceToCreate   int
	StudentPaymentDetailID        string
	BankBranchID                  string
	BankBranchIDs                 []string
	OrderIDs                      []string
	PaymentDate                   time.Time
	ValidatedDate                 time.Time
	InvoiceAdjustmentIDs          []string
	InvoiceAdjustMapAmount        map[string]string
	CurrentInvoice                *invoice_entity.Invoice
	StudentInvoiceTotalMap        map[string]int64
	GradeID                       string
	CurrentPayment                *invoice_entity.Payment
	BulkPaymentID                 string
	InvoiceIDInvoiceReferenceMap  map[string]string
	InvoiceIDInvoiceReference2Map map[string]string
	InvoiceIDInvoiceTotalMap      map[string]float64
	InvoiceReferenceID            string
	InvoiceReferenceID2           string
	InvoiceTotalFloat             float64
	StudentInvoiceReferenceMap    map[string]string
	StudentInvoiceReference2Map   map[string]string
	StudentBillItemTotalPrice     map[string]float64
	BillItemTotalFloat            float64
	StudentPaymentMethodMap       map[string]string
	CurrentBillingAddress         *invoice_entity.BillingAddress
	CurrentPrefecture             *invoice_entity.Prefecture
	LatestPaymentStatuses         []string
	InvoiceTypes                  []string
	CurrentPayerName              string
	BankCode                      string
	BankBranchCode                string
	BankOpenAPIPublicKey          string
	BankOpenAPIPrivateKey         string
	PaymentListToValidate         []*invoice_entity.Payment
	CurrentInvoiceStatus          string
	PaymentStatusIDsMap           map[string][]string
	CutoffDate                    time.Time

	// timesheet
	CurrentTimesheetID               string
	CurrentTimesheetIDs              []string
	TimesheetLessonIDs               []string
	NumberOfOtherWorkingHours        int32
	NumberOfTransportExpensesRecords int32
	Timesheets                       []*timesheet_entity.Timesheet
	TimesheetDateBeforeChange        time.Time
	TimesheetLocationBeforeChange    string
	StaffID                          string
	NumberLogRecordsOfCurrentStaffID int32
	CurrentListTimesheetLessonHours  []*timesheet_dto.TimesheetLessonHours

	// calendar
	DateTypes   []*calendar_dto.DateType
	DateInfos   []*calendar_dto.DateInfo
	DateTypeID  string
	SchedulerID string
	DateInfoIDs []string

	// accesscontrol
	HasuraBody         []byte
	HasuraURL          string
	HasuraAdminAccount string
	PermissionID       string
	TotalRecords       int
	RoleId             string
	GrantedRoleId      string
	ErrInfo            error
	CurrentTestId      string
	CurrentTestData    string
	PermissionReadId   string
	PermissionWriteId  string
	RowAffected        int64

	// syllabus
	FlashcardID                string
	LearningMaterialID         string
	LearningMaterialIDs        []string
	TopicLODisplayOrderCounter int32

	DeletedLessonCount int
	DeletedLessonIDs   []string

	UserId            string
	SchoolAdminToken  string
	StudentToken      string
	OfflineLearningID string

	// student name, teacher name
	CurrentStudentFirstName string
	CurrentStudentLastName  string

	FileStore filestore.FileStore

	// hephaestus
	ConnectClient                                   *connect.Client
	SrcConnector, SinkConnector                     connect.Connector
	TestDebeziumRecordIDs, TestDebeziumJobRecordIDs []string
	TableName                                       string

	// OpenAPI
	ManabiePublicKey string
	ManabieSignature string

	CurrentCenterID       string
	StudentCoursesClasses map[string]*lpb.GetStudentCoursesAndClassesResponse

	// discount
	DiscountTagID           string
	DiscountTagTypeAndIDMap map[string]string
	ProductGroupIDs         []string
}

type Message struct {
	ConversationID string
	MessageID      string
	MessageType    string
	Content        string
}

type StepStateKey struct{}

func StepStateFromContext(ctx context.Context) *StepState {
	state := ctx.Value(StepStateKey{})
	if state == nil {
		return &StepState{}
	}
	return state.(*StepState)
}

func StepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, StepStateKey{}, state)
}

type suite struct {
	*StepState
	*Connections
}

type Suite struct {
	suite
}

type CancellableStream struct {
	pb_tom.ChatService_SubscribeV2Client
	Cancel context.CancelFunc
}

type Config configurations.Config

type GenerateBillItemsUtils struct {
	Request           interface{}
	Response          interface{}
	ResponseErr       error
	OrderID           string
	NumberOfBillItems int
}

type AuthInfo struct {
	UserID string
	Token  string
}

type MapRoleAndAuthInfo map[string]AuthInfo
