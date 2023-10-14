package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	common_constants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	LOEventType = "learning_objective"

	// LO event names
	LOEventStarted   = "started"
	LOEventPaused    = "paused"
	LOEventResumed   = "resumed"
	LOEventCompleted = "completed"
	LOEventExited    = "exited"

	LOCompletenessQuizEventType       = "quiz_finished"
	LOCompletenessVideoEventType      = "video_finished"
	LOCompletenessStudyGuideEventType = "study_guide_finished"
)

type StudentRepo interface {
	Create(context.Context, database.QueryExecer, *entities.Student) error
	Retrieve(context.Context, database.QueryExecer, pgtype.TextArray) ([]repositories.StudentProfile, error)
	UpdateStudentProfile(ctx context.Context, db database.QueryExecer, s *entities.Student) error
	Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*entities.Student, error)
	Update(ctx context.Context, db database.QueryExecer, s *entities.Student) error
	FindByPhone(ctx context.Context, db database.QueryExecer, phone pgtype.Text) (*entities.Student, error)
	GetCountryByStudent(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (string, error)
}

type StudentEventLogRepo interface {
	Create(context.Context, database.QueryExecer, []*entities.StudentEventLog) error
	Retrieve(ctx context.Context, db database.QueryExecer, studentID, sessionID string, from, to *pgtype.Timestamptz) ([]*entities.StudentEventLog, error)
	RetrieveLOEvents(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*entities.StudentEventLog, error)
}

type UserRepo interface {
	UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
	Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error)
	Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error)
}

type PresetStudyPlanRepo interface {
	RetrievePresetStudyPlans(ctx context.Context, db database.QueryExecer, name, country, subject string, grade int) ([]*entities.PresetStudyPlan, error)
	RetrievePresetStudyPlanWeeklies(ctx context.Context, db database.QueryExecer, presetStudyPlanID pgtype.Text) ([]*entities.PresetStudyPlanWeekly, error)
	RetrieveStudentPresetStudyPlans(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*repositories.PlanWithStartDate, error)
	RetrieveStudyAheadTopicsOfPresetStudyPlan(ctx context.Context, db database.QueryExecer, studentID, presetStudyPlanID pgtype.Text) ([]repositories.AheadTopic, error)
	RetrieveStudentCompletedTopics(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, topicsId pgtype.TextArray) ([]*entities.Topic, error)
}

type StudentLOCompletenessRepo interface {
	TotalLOFinished(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) (total int, err error)
	UpsertLOCompleteness(context.Context, database.QueryExecer, []*entities.StudentsLearningObjectivesCompleteness) error
	DailyLOFinished(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*repositories.DailyLOFinished, error)
	RetrieveFinishedLOs(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.StudentsLearningObjectivesCompleteness, error)
	CountTotalLOsFinished(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) (int, error)
}

type StudentStatRepo interface {
	Stat(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*entities.StudentStat, error)
	Upsert(context.Context, database.QueryExecer, *entities.StudentStat) error
}

type StudentCommentRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, comment *entities.StudentComment) error
	RetrieveByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, fields ...string) ([]entities.StudentComment, error)
}

type LearningObjectiveRepo interface {
	RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error)
	RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error)
}

type SchoolRepo interface {
	Create(context.Context, *entities.School) error
}

type StudentOrderRepo interface {
	Create(context.Context, database.QueryExecer, *entities.StudentOrder) error
	FindOrderByPromotionCode(ctx context.Context, db database.QueryExecer, studentID, promoCode pgtype.Text) (*entities.StudentOrder, error)
	UpdateReferenceNumber(context.Context, database.QueryExecer, pgtype.Int4, pgtype.Text) error
}

type StudentTopicCompletenessRepo interface {
	Upsert(context.Context, database.Ext, []*entities.StudentTopicCompleteness) error
	RetrieveCompletedByStudentIDWeeklies(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*entities.StudentTopicCompleteness, error)
	RetrieveByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, topicIDs *pgtype.TextArray) (map[string][]*entities.StudentTopicCompleteness, error)
}

type ConfigRepo interface {
	Retrieve(context.Context, database.QueryExecer, pgtype.Text, pgtype.Text, pgtype.TextArray) ([]*entities.Config, error)
}

type ActivityLogRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.ActivityLog) error
}

type StudentLearningTimeDaiyRepo interface {
	Retrieve(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, queryEnhancers ...repositories.QueryEnhancer) ([]*entities.StudentLearningTimeDaily, error)
}
type StudentTopicOverdueRepo interface {
	RetrieveStudentTopicOverdue(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*repositories.TopicDueDate, error)
	RemoveStudentTopicOverdue(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, topicIDs pgtype.TextArray) error
}

type AssignmentRepo interface {
	FindAssignmentByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Assignment, error)
	DeleteAssignment(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) error
	RetrieveStudentAssignment(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz, isActive bool, userGroup pgtype.Text) ([]repositories.Topic, error)
	ExecQueueAssignment(ctx context.Context, tx database.QueryExecer, assignments []*entities.Assignment) error
	ExecQueueStudentAssignment(ctx context.Context, tx database.QueryExecer, studentAssignments []*entities.StudentAssignment) error
	FindStudentAssignmentWithStudyPlan(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]string, error)
	FindStudentOverdueAssignment(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*repositories.AssignmentWithTopic, error)
	RetrieveStudentAssignmentByTopic(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, topicIDs pgtype.TextArray) ([]string, error)
	FindStudentCompletedAssignmentWeeklies(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, from, to *pgtype.Timestamptz) ([]*repositories.AssignmentWithTopic, error)
}

type TopicRepo interface {
	RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
}

type StudentAssignmentRepo interface {
	UpdateStudentAssignmentStatus(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, assignmentIDs pgtype.TextArray, status pgtype.Text) error
}

type StudentSubmissionRepo interface {
	Create(context.Context, database.QueryExecer, *entities.StudentSubmission) error
	List(context.Context, database.QueryExecer, *repositories.StudentSubmissionFilter) ([]*entities.StudentSubmission, error)
}

type MediaRepo interface {
	CreateBatch(context.Context, database.QueryExecer, *entities.Comment) error
}

type UserMgmtStudentSvc interface {
	GetStudentProfile(context.Context, *upb.GetStudentProfileRequest, ...grpc.CallOption) (*upb.GetStudentProfileResponse, error)
	UpsertStudentComment(context.Context, *upb.UpsertStudentCommentRequest, ...grpc.CallOption) (*upb.UpsertStudentCommentResponse, error)
	DeleteStudentComments(ctx context.Context, in *upb.DeleteStudentCommentsRequest, opts ...grpc.CallOption) (*upb.DeleteStudentCommentsResponse, error)
	RetrieveStudentComment(ctx context.Context, in *upb.RetrieveStudentCommentRequest, opts ...grpc.CallOption) (*upb.RetrieveStudentCommentResponse, error)
}

type UserReaderStudentSvc interface {
	RetrieveStudentAssociatedToParentAccount(context.Context, *upb.RetrieveStudentAssociatedToParentAccountRequest, ...grpc.CallOption) (*upb.RetrieveStudentAssociatedToParentAccountResponse, error)
}

type StudyPlanReaderSvc interface {
	RetrieveStat(ctx context.Context, req *epb.RetrieveStatRequest, opts ...grpc.CallOption) (*epb.RetrieveStatResponse, error)
}

type StudentLearningTimeSvc interface {
	RetrieveLearningProgress(ctx context.Context, req *epb.RetrieveLearningProgressRequest, opts ...grpc.CallOption) (*epb.RetrieveLearningProgressResponse, error)
}

type EurekaStudentEventLogModifierSvc interface {
	CreateStudentEventLogs(ctx context.Context, req *epb.CreateStudentEventLogsRequest, opts ...grpc.CallOption) (*epb.CreateStudentEventLogsResponse, error)
}

// StudentService implement proto StudentServer
type StudentService struct {
	*pb.UnimplementedStudentServer
	EurekaDBTrace                database.Ext
	DB                           database.Ext
	Env                          string
	JSM                          nats.JetStreamManagement
	StudentRepo                  StudentRepo
	StudentEventLogRepo          StudentEventLogRepo
	UserRepo                     UserRepo
	PresetStudyPlanRepo          PresetStudyPlanRepo
	StudentLOCompletenessRepo    StudentLOCompletenessRepo
	StudentStatRepo              StudentStatRepo
	StudentCommentRepo           StudentCommentRepo
	LearningObjectiveRepo        LearningObjectiveRepo
	SchoolRepo                   SchoolRepo
	StudentOrderRepo             StudentOrderRepo
	StudentTopicCompletenessRepo StudentTopicCompletenessRepo
	ConfigRepo                   ConfigRepo
	ActivityLogRepo              ActivityLogRepo
	StudentLearningTimeDaiyRepo  StudentLearningTimeDaiyRepo
	StudentTopicOverdueRepo      StudentTopicOverdueRepo
	StudentAssignmentRepo        StudentAssignmentRepo
	TopicRepo                    TopicRepo
	UserMgmtStudentSvc           UserMgmtStudentSvc
	ClassMemberRepo              interface {
		InClass(ctx context.Context, db database.QueryExecer, userID pgtype.Text, userIDs pgtype.TextArray) ([]string, error)
	}
	AppleUserRepo interface {
		Create(ctx context.Context, db database.QueryExecer, a *entities.AppleUser) error
	}
	StudyPlanReaderSvc               StudyPlanReaderSvc
	StudentLearningTimeSvc           StudentLearningTimeSvc
	EurekaStudentEventLogModifierSvc EurekaStudentEventLogModifierSvc
}

func NewStudentService(eurekaDBTrace *database.DBTrace, db database.Ext,
	env string,
	jsm nats.JetStreamManagement,
	sr StudentRepo,
	el StudentEventLogRepo,
	ur UserRepo,
	ppr PresetStudyPlanRepo,
	loCompl StudentLOCompletenessRepo,
	ssr StudentStatRepo,
	lo LearningObjectiveRepo,
	scmr StudentCommentRepo,
	str StudentTopicCompletenessRepo,
	stodr StudentTopicOverdueRepo,
	tr TopicRepo,
	userMgmtStudentSvc UserMgmtStudentSvc,
	studyPlanReaderSvc StudyPlanReaderSvc,
	studentLearningTimeSvc StudentLearningTimeSvc,
	eurekaStudentEventLogModifierSvc epb.StudentEventLogModifierServiceClient,
) *StudentService {
	s := &StudentService{
		EurekaDBTrace:                    eurekaDBTrace,
		DB:                               db,
		Env:                              env,
		JSM:                              jsm,
		StudentRepo:                      sr,
		StudentEventLogRepo:              el,
		UserRepo:                         ur,
		PresetStudyPlanRepo:              ppr,
		StudentLOCompletenessRepo:        loCompl,
		StudentStatRepo:                  ssr,
		StudentCommentRepo:               scmr,
		LearningObjectiveRepo:            lo,
		StudentTopicCompletenessRepo:     str,
		StudentTopicOverdueRepo:          stodr,
		TopicRepo:                        tr,
		UserMgmtStudentSvc:               userMgmtStudentSvc,
		StudyPlanReaderSvc:               studyPlanReaderSvc,
		StudentLearningTimeSvc:           studentLearningTimeSvc,
		EurekaStudentEventLogModifierSvc: eurekaStudentEventLogModifierSvc,
	}

	return s
}

// GetStudentProfile returns a list of students.
func (s *StudentService) GetStudentProfile(ctx context.Context, req *pb.GetStudentProfileRequest) (*pb.GetStudentProfileResponse, error) {
	respMgmt, err := s.UserMgmtStudentSvc.GetStudentProfile(
		signCtx(ctx),
		&upb.GetStudentProfileRequest{StudentIds: req.StudentIds},
	)
	if err != nil {
		return nil, toStatusError(err)
	}

	resp := make([]*pb.GetStudentProfileResponse_Data, 0)

	for _, student := range respMgmt.GetProfiles() {
		if student == nil {
			continue
		}
		birthDay, _ := types.TimestampProto(student.Birthday.AsTime())
		createdAt, _ := types.TimestampProto(student.CreatedAt.AsTime())

		profileData := &pb.GetStudentProfileResponse_Data{
			Profile: &pb.StudentProfile{
				Id:        student.Id,
				Name:      student.Name,
				Country:   pb.Country(student.Country),
				Phone:     student.Phone,
				Email:     student.Email,
				Grade:     student.Grade,
				Avatar:    student.Avatar,
				Birthday:  birthDay,
				CreatedAt: createdAt,
				Divs:      student.Divs,
				School:    toBobSchoolPb(student.School),
			},
		}

		resp = append(resp, profileData)
	}

	return &pb.GetStudentProfileResponse{
		Datas: resp,
	}, nil
}

func (s *StudentService) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)

	currentStudent, _ := s.UserRepo.Retrieve(ctx, s.DB, database.TextArray([]string{currentUserID}))
	country := pb.Country(pb.Country_value[currentStudent[0].Country.String])
	grade, err := i18n.ConvertStringGradeToInt(country, req.Grade)
	if err != nil {
		return nil, err
	}
	e := toStudentProfileEnity(req)
	e.ID.Set(currentUserID)
	e.CurrentGrade.Set(grade)
	e.Country = currentStudent[0].Country
	e.User.ID = e.ID

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.StudentRepo.UpdateStudentProfile(ctx, tx, e); err != nil {
			return errors.Wrap(err, "s.StudentRepo.UpdateStudentProfileTx")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if req.Name != "" {
		data := &pb.EvtUserInfo{
			UserId: currentUserID,
			Name:   req.Name,
		}
		msg, _ := data.Marshal()

		var msgId string
		msgId, err = s.JSM.PublishAsyncContext(ctx, common_constants.SubjectUserDeviceTokenUpdated, msg)
		if err != nil {
			ctxzap.Extract(ctx).Error("UpdateUserDeviceToken s.JSM.PublishAsync failed", zap.String("msg-id", msgId), zap.Error(err))
		}
	}

	return &pb.UpdateProfileResponse{
		Successful: true,
	}, nil
}

func toStudentProfileEnity(src *pb.UpdateProfileRequest) *entities.Student {
	e := new(entities.Student)
	database.AllNullEntity(e)
	database.AllNullEntity(&e.User)
	e.TargetUniversity.Set(src.TargetUniversity)
	e.Birthday.Set(time.Unix(src.Birthday.Seconds, int64(src.Birthday.Nanos)))
	e.Biography.Set(src.Biography)
	e.User.Avatar.Set(src.Avatar)
	e.User.LastName.Set(src.Name)
	if src.School != nil {
		se := toSchoolEntity(src.School)
		se.IsSystemSchool.Set(false)
		// in case student selects existed school
		if src.School.Id != 0 {
			se.ID.Set(src.School.Id)
		}
		e.School = se
	}
	return e
}

func (s *StudentService) RetrieveLearningProgress(ctx context.Context, req *pb.RetrieveLearningProgressRequest) (*pb.RetrieveLearningProgressResponse, error) {
	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, codes.Unauthenticated.String())
	}

	eurekaResp, err := s.StudentLearningTimeSvc.RetrieveLearningProgress(mdCtx, &epb.RetrieveLearningProgressRequest{
		StudentId: req.StudentId,
		From:      timestamppb.New(time.Unix(req.GetFrom().GetSeconds(), int64(req.GetFrom().GetNanos()))),
		To:        timestamppb.New(time.Unix(req.GetTo().GetSeconds(), int64(req.GetTo().GetNanos()))),
	})

	if err != nil {
		return nil, err
	}

	var resp pb.RetrieveLearningProgressResponse
	respBytes, err1 := json.Marshal(eurekaResp)
	err2 := json.Unmarshal(respBytes, &resp)

	if err := multierr.Combine(err1, err2); err != nil {
		return nil, status.Errorf(codes.Internal, "Marshal/Unmarshal error %s", err.Error())
	}
	return &resp, nil
}

// look deprecated
func (s *StudentService) AssignPresetStudyPlans(ctx context.Context, req *pb.AssignPresetStudyPlansRequest) (*pb.AssignPresetStudyPlansResponse, error) {
	return &pb.AssignPresetStudyPlansResponse{}, nil
}

func (s *StudentService) RetrievePresetStudyPlans(ctx context.Context, req *pb.RetrievePresetStudyPlansRequest) (*pb.RetrievePresetStudyPlansResponse, error) {
	var country string
	if req.Country != pb.COUNTRY_NONE {
		country = req.Country.String()
	}

	var subject string
	if req.Subject != pb.SUBJECT_NONE {
		subject = req.Subject.String()
	}

	var grade int
	var err error
	if req.Grade == "" {
		grade = -1
	} else {
		grade, err = i18n.ConvertStringGradeToInt(req.Country, req.Grade)
		if err != nil {
			return nil, err
		}
	}

	plans, err := s.PresetStudyPlanRepo.RetrievePresetStudyPlans(ctx, s.DB, req.Name, country, subject, grade)
	if err != nil {
		return nil, status.Error(codes.Unknown, errors.Wrapf(err, "c.PresetStudyPlanRepo.RetrievePresetStudyPlans: grade: %v", req.Grade).Error())
	}

	ret := make([]*pb.PresetStudyPlan, 0, len(plans))
	for _, p := range plans {
		ret = append(ret, toPresetStudyPlanPb(p))
	}
	return &pb.RetrievePresetStudyPlansResponse{PresetStudyPlans: ret}, nil
}

func (s *StudentService) RetrievePresetStudyPlanWeeklies(ctx context.Context, req *pb.RetrievePresetStudyPlanWeekliesRequest) (*pb.RetrievePresetStudyPlanWeekliesResponse, error) {
	weeklies, err := s.PresetStudyPlanRepo.RetrievePresetStudyPlanWeeklies(ctx, s.DB, database.Text(req.PresetStudyPlanId))
	if err != nil {
		return nil, status.Error(codes.Unknown, errors.Wrapf(err, "c.PresetStudyPlanRepo.RetrievePresetStudyPlanWeeklies").Error())
	}

	ret := make([]*pb.PresetStudyPlanWeekly, 0, len(weeklies))
	for _, w := range weeklies {
		ret = append(ret, toPresetStudyPlanWeeklyPb(w))
	}
	return &pb.RetrievePresetStudyPlanWeekliesResponse{PresetStudyPlanWeeklies: ret}, nil
}

func (s *StudentService) RetrieveStudentStudyPlans(ctx context.Context, req *pb.RetrieveStudentStudyPlansRequest) (*pb.RetrieveStudentStudyPlansResponse, error) {
	var from *pgtype.Timestamptz
	if req.From != nil {
		from = new(pgtype.Timestamptz)
		from.Set(time.Unix(req.From.Seconds, int64(req.From.Nanos)).UTC())
	}

	var to *pgtype.Timestamptz
	if req.To != nil {
		to = new(pgtype.Timestamptz)
		to.Set(time.Unix(req.To.Seconds, int64(req.To.Nanos)).UTC())
	}

	plans, err := s.PresetStudyPlanRepo.RetrieveStudentPresetStudyPlans(ctx, s.DB, database.Text(req.StudentId), from, to)
	if err != nil {
		return nil, status.Error(codes.Unknown, errors.Wrap(err, "s.PresetStudyPlanRepo.RetrieveStudentPresetStudyPlans").Error())
	}

	ret := make([]*pb.RetrieveStudentStudyPlansResponse_PlanWithStartDate, 0, len(plans))
	for _, p := range plans {
		ret = append(ret, &pb.RetrieveStudentStudyPlansResponse_PlanWithStartDate{
			Plan:      toPresetStudyPlanPb(p.PresetStudyPlan),
			Week:      int32(p.Week.Int),
			StartDate: &types.Timestamp{Seconds: p.StartDate.Time.Unix()},
		})
	}
	return &pb.RetrieveStudentStudyPlansResponse{PlanWithStartDates: ret}, nil
}

func (s *StudentService) RetrieveStudentStudyPlanWeeklies(ctx context.Context, req *pb.RetrieveStudentStudyPlanWeekliesRequest) (*pb.RetrieveStudentStudyPlanWeekliesResponse, error) {
	return &pb.RetrieveStudentStudyPlanWeekliesResponse{}, nil
}

// FindStudent by phone number
func (s *StudentService) FindStudent(ctx context.Context, req *pb.FindStudentRequest) (*pb.FindStudentResponse, error) {
	student, err := s.StudentRepo.FindByPhone(ctx, s.DB, database.Text(req.Phone))
	if err != nil {
		return nil, status.Error(codes.NotFound, "student not found")
	}

	return &pb.FindStudentResponse{
		Profile: student2Profile(student),
	}, nil
}

func student2Profile(student *entities.Student) *pb.StudentProfile {
	birthDay, _ := types.TimestampProto(student.Birthday.Time)
	billingDate, _ := types.TimestampProto(student.BillingDate.Time)
	createdAt, _ := types.TimestampProto(student.CreatedAt.Time)
	country := pb.Country(pb.Country_value[student.Country.String])
	grade, _ := i18n.ConvertIntGradeToString(country, int(student.CurrentGrade.Int))

	now := time.Now()
	var paymentStatus pb.PaymentStatus
	if now.Before(student.BillingDate.Time) {
		if student.OnTrial.Bool {
			paymentStatus = pb.PAYMENT_STATUS_ON_TRIAL
		} else {
			paymentStatus = pb.PAYMENT_STATUS_PAID
		}
	} else {
		if student.OnTrial.Bool {
			paymentStatus = pb.PAYMENT_STATUS_EXPIRED_TRIAL
		} else {
			paymentStatus = pb.PAYMENT_STATUS_LATE_PAYMENT
		}
	}

	var divs []int64
	data, _ := student.GetStudentAdditionalData()
	if data != nil {
		divs = data.JprefDivs
	}

	return &pb.StudentProfile{
		Id:               student.ID.String,
		Name:             student.GetName(),
		Country:          country,
		Phone:            student.PhoneNumber.String,
		Email:            student.Email.String,
		Grade:            grade,
		TargetUniversity: student.TargetUniversity.String,
		Avatar:           student.Avatar.String,
		Birthday:         birthDay,
		Biography:        student.Biography.String,
		PaymentStatus:    paymentStatus,
		BillingDate:      billingDate,
		CreatedAt:        createdAt,
		IsTester:         student.IsTester.Bool,
		FacebookId:       student.FacebookID.String,
		Divs:             divs,
	}
}

func (s *StudentService) RetrieveDailyLOFinished(ctx context.Context, req *pb.RetrieveDailyLOFinishedRequest) (*pb.RetrieveDailyLOFinishedResponse, error) {
	return &pb.RetrieveDailyLOFinishedResponse{}, nil
}

func toPresetStudyPlanPb(p *entities.PresetStudyPlan) *pb.PresetStudyPlan {
	var pID string
	p.ID.AssignTo(&pID)
	country := pb.Country(pb.Country_value[p.Country.String])
	grade, _ := i18n.ConvertIntGradeToString(country, int(p.Grade.Int))
	var startDate int64
	if p.StartDate.Status == pgtype.Present {
		startDate = p.StartDate.Time.Unix()
	} else {
		startDate = timeutil.DefaultPSPStartDate(country).Unix()
	}
	return &pb.PresetStudyPlan{
		Id:        pID,
		Name:      p.Name.String,
		Country:   country,
		Grade:     grade,
		Subject:   pb.Subject(pb.Subject_value[p.Subject.String]),
		CreatedAt: &types.Timestamp{Seconds: p.CreatedAt.Time.Unix()},
		UpdatedAt: &types.Timestamp{Seconds: p.UpdatedAt.Time.Unix()},
		StartDate: &types.Timestamp{Seconds: startDate},
	}
}

func toPresetStudyPlanWeeklyPb(p *entities.PresetStudyPlanWeekly) *pb.PresetStudyPlanWeekly {
	return &pb.PresetStudyPlanWeekly{
		Id:                p.ID.String,
		TopicId:           p.TopicID.String,
		PresetStudyPlanId: p.PresetStudyPlanID.String,
		Week:              int32(p.Week.Get().(int16)),
	}
}

func (s *StudentService) UpsertStudentComment(ctx context.Context, req *pb.UpsertStudentCommentRequest) (*pb.UpsertStudentCommentResponse, error) {
	createdAt, _ := types.TimestampFromProto(req.StudentComment.CreatedAt)
	updatedAt, _ := types.TimestampFromProto(req.StudentComment.UpdatedAt)
	resp, err := s.UserMgmtStudentSvc.UpsertStudentComment(signCtx(ctx), &upb.UpsertStudentCommentRequest{
		StudentComment: &upb.StudentComment{
			CommentId:      req.StudentComment.CommentId,
			StudentId:      req.StudentComment.StudentId,
			CoachId:        req.StudentComment.CoachId,
			CommentContent: req.StudentComment.CommentContent,
			CreatedAt:      timestamppb.New(createdAt),
			UpdatedAt:      timestamppb.New(updatedAt),
		},
	})
	if err != nil {
		return nil, errors.Wrap(toStatusError(err), "s.UserMgmtStudentSvc.UpsertStudentComment")
	}

	return &pb.UpsertStudentCommentResponse{
		Successful: resp.Successful,
	}, nil
}

func (s *StudentService) RetrieveStudentComment(ctx context.Context, req *pb.RetrieveStudentCommentRequest) (*pb.RetrieveStudentCommentResponse, error) {
	resp, err := s.UserMgmtStudentSvc.RetrieveStudentComment(signCtx(ctx), &upb.RetrieveStudentCommentRequest{
		StudentId: req.StudentId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.RetrieveStudentCommentResponse{
		Comment: fromUserCommentsToBobComments(resp.Comment),
	}, nil
}

func fromUserCommentsToBobComments(comments []*upb.CommentInfo) []*pb.CommentInfo {
	result := make([]*pb.CommentInfo, 0)
	for _, comment := range comments {
		result = append(result, &pb.CommentInfo{
			StudentComment: &pb.StudentComment{
				CommentId:      comment.StudentComment.CommentId,
				CoachId:        comment.StudentComment.CoachId,
				StudentId:      comment.StudentComment.StudentId,
				CommentContent: comment.StudentComment.CommentContent,
				UpdatedAt:      &types.Timestamp{Seconds: comment.StudentComment.UpdatedAt.Seconds},
				CreatedAt:      &types.Timestamp{Seconds: comment.StudentComment.CreatedAt.Seconds},
			},
		})
	}
	return result
}

func dateRangeValid(from, to *types.Timestamp) bool {
	if from == nil || to == nil {
		return true
	}

	tFrom := time.Unix(from.Seconds, int64(from.Nanos))
	tTo := time.Unix(to.Seconds, int64(to.Nanos))
	return tFrom.Before(tTo)
}

func (s *StudentService) RetrieveStudyAheadTopics(ctx context.Context, req *pb.RetrieveStudyAheadTopicsRequest) (*pb.RetrieveStudyAheadTopicsResponse, error) {
	return &pb.RetrieveStudyAheadTopicsResponse{}, nil
}

func (s *StudentService) RetrieveArchivedTopics(ctx context.Context, req *pb.RetrieveArchivedTopicsRequest) (*pb.RetrieveArchivedTopicsResponse, error) {
	return &pb.RetrieveArchivedTopicsResponse{}, nil
}

var allPrivileges = []pb.PlanPrivilege{
	pb.CAN_ACCESS_LEARNING_TOPICS,
	pb.CAN_ACCESS_PRACTICE_TOPICS,
	pb.CAN_ACCESS_MOCK_EXAMS,
	pb.CAN_ACCESS_ALL_LOS,
	pb.CAN_ACCESS_SOME_LOS,
	pb.CAN_WATCH_VIDEOS,
	pb.CAN_READ_STUDY_GUIDES,
	pb.CAN_SKIP_VIDEOS,
	pb.CAN_CHAT_WITH_TEACHER,
}

var allowAll = &pb.Permission{
	PermissionAllowGrades: []*pb.PermissionAllowGrade{
		{
			Subject:        pb.SUBJECT_MATHS,
			PlanPrivileges: allPrivileges,
		},
		{
			Subject:        pb.SUBJECT_BIOLOGY,
			PlanPrivileges: allPrivileges,
		},
		{
			Subject:        pb.SUBJECT_PHYSICS,
			PlanPrivileges: allPrivileges,
		},
		{
			Subject:        pb.SUBJECT_CHEMISTRY,
			PlanPrivileges: allPrivileges,
		},
		{
			Subject:        pb.SUBJECT_GEOGRAPHY,
			PlanPrivileges: allPrivileges,
		},
		{
			Subject:        pb.SUBJECT_ENGLISH,
			PlanPrivileges: allPrivileges,
		},
		{
			Subject:        pb.SUBJECT_ENGLISH_2,
			PlanPrivileges: allPrivileges,
		},
	},
}

func (s *StudentService) StudentPermission(ctx context.Context, req *pb.StudentPermissionRequest) (*pb.StudentPermissionResponse, error) {
	return &pb.StudentPermissionResponse{
		Permissions: map[int32]*pb.Permission{
			int32(12): allowAll,
			int32(11): allowAll,
			int32(10): allowAll,
			int32(9):  allowAll,
			int32(8):  allowAll,
			int32(7):  allowAll,
			int32(6):  allowAll,
			int32(5):  allowAll,
		},
	}, nil
}

func (s *StudentService) CountTotalLOsFinished(ctx context.Context, req *pb.CountTotalLOsFinishedRequest) (*pb.CountTotalLOsFinishedResponse, error) {
	from, err := database.TimestamptzFromProto(req.From)
	if err != nil {
		return nil, err
	}
	to, err := database.TimestamptzFromProto(req.To)
	if err != nil {
		return nil, err
	}

	totalLOsFinished, err := s.StudentLOCompletenessRepo.CountTotalLOsFinished(ctx, s.DB, database.Text(req.StudentId), from, to)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.StudentLOCompletenessRepo.TotalLOFinished: %w", err).Error())
	}

	return &pb.CountTotalLOsFinishedResponse{TotalLosFinished: int32(totalLOsFinished)}, nil
}

func (s *StudentService) RetrieveOverdueTopic(ctx context.Context, req *pb.RetrieveOverdueTopicRequest) (*pb.RetrieveOverdueTopicResponse, error) {
	return &pb.RetrieveOverdueTopicResponse{}, nil
}

func (s *StudentService) TopicWithStartDateToPb(src *repositories.AssignmentWithTopic) *pb.RetrieveCompletedTopicWeekliesResponse_TopicWithAssignBy {
	return &pb.RetrieveCompletedTopicWeekliesResponse_TopicWithAssignBy{
		Topics:      ToTopicPb(src.Topic),
		CompletedAt: &types.Timestamp{Seconds: src.CompletedAt.Time.Unix()},
		AssignedBy: &pb.BasicProfile{
			Name:      src.User.GetName(),
			UserId:    src.User.ID.String,
			Avatar:    src.User.Avatar.String,
			UserGroup: src.User.Group.String,
		},
	}
}

// look deprecated
func (s *StudentService) RetrieveCompletedTopicWeeklies(ctx context.Context, req *pb.RetrieveCompletedTopicWeekliesRequest) (*pb.RetrieveCompletedTopicWeekliesResponse, error) {
	return &pb.RetrieveCompletedTopicWeekliesResponse{}, nil
}
