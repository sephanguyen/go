package classes

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services"
	global_constant "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	classAvatar = "class_avatar"

	classPlan          = "class_plan"
	classPlanName      = "planName"
	classPlanPeriod    = "planPeriod"
	defaultClassAvatar = "https://storage.googleapis.com/manabie-backend/class/ico_class_default.png"
)

type ClassService struct {
	l               sync.RWMutex
	Cfg             *configurations.Config
	ClassCodeLength int
	DB              database.Ext
	JSM             nats.JetStreamManagement
	HTTPClient      interface {
		Do(req *http.Request) (*http.Response, error)
	}

	LessonModifierServices interface {
		ResetAllLiveLessonStatesInternal(ctx context.Context, lessonID string) error
	}

	UserRepo interface {
		Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error)
		UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]*entities.User, error)
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.UserFindFilter, fields ...string) ([]*entities.User, error)
	}

	ClassRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.Class) error
		Update(ctx context.Context, db database.QueryExecer, e *entities.Class) error
		FindByID(ctx context.Context, db database.QueryExecer, ID pgtype.Int4) (*entities.Class, error)
		FindJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entities.Class, error)
		UpdateClassCode(ctx context.Context, db database.QueryExecer, ID pgtype.Int4, code pgtype.Text) error
		FindByCode(ctx context.Context, db database.QueryExecer, code pgtype.Text) (*entities.Class, error)
		GetNextID(ctx context.Context, db database.QueryExecer) (*pgtype.Int4, error)
	}

	ClassMemberRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.ClassMember) error
		IsOwner(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userID pgtype.Text) (bool, error)
		Get(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userID, status pgtype.Text) (*entities.ClassMember, error)
		UpdateStatus(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userIDs pgtype.TextArray, status pgtype.Text) ([]*entities.ClassMember, error)
		FindOwner(ctx context.Context, db database.QueryExecer, classIDs pgtype.Int4Array) (mapUserIDByClass map[pgtype.Int4][]pgtype.Text, err error)
		Count(ctx context.Context, db database.QueryExecer, classIDs pgtype.Int4Array, userGroup pgtype.Text) (mapTotalUserByClass map[pgtype.Int4]pgtype.Int8, err error)
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.FindClassMemberFilter) ([]*entities.ClassMember, error)
		FindActiveStudentMember(ctx context.Context, db database.QueryExecer, classID pgtype.Int4) ([]string, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userIDs pgtype.TextArray, status pgtype.Text) ([]*entities.ClassMember, error)
		ClassJoinNotIn(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, classIDs pgtype.Int4Array) ([]int32, error)
	}

	ConfigRepo interface {
		Find(ctx context.Context, db database.QueryExecer, country pgtype.Text, group pgtype.Text) ([]*entities.Config, error)
	}

	StudentOrderRepo interface {
		Create(context.Context, database.QueryExecer, *entities.StudentOrder) error
		FindOrderByPromotionCode(ctx context.Context, db database.QueryExecer, studentID, promoCode pgtype.Text) (*entities.StudentOrder, error)
		UpdateReferenceNumber(context.Context, database.QueryExecer, pgtype.Int4, pgtype.Text) error
	}

	SchoolConfigRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, schoolID pgtype.Int4) (*entities.SchoolConfig, error)
	}

	TeacherRepo interface {
		Create(ctx context.Context, db database.QueryExecer, t *entities.Teacher) error
		Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error)
		IsInSchool(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray, schoolID pgtype.Int4) (bool, error)
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Teacher, error)
	}

	TopicRepo interface {
		Create(context.Context, database.QueryExecer, *entities.Topic) (string, error)
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
	}

	LearningObjectiveRepo interface {
		BulkImport(context.Context, database.QueryExecer, []*entities.LearningObjective) error
		RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error)
	}

	QuestionRepo interface {
		RetrieveQuizSetsFromLoId(ctx context.Context, db database.QueryExecer, loID pgtype.Text, topicType pgtype.Text, limit, page int) (*repositories.QuestionPagination, error)
	}

	StudentEventLogRepo interface {
		RetrieveAllSubmitionsOfStudent(ctx context.Context, db database.QueryExecer, studentID pgtype.TextArray, loIDs *pgtype.TextArray) (map[string]map[string][]*repositories.SubmissionResult, error)
		Retrieve(ctx context.Context, db database.QueryExecer, studentID, sessionID string, from, to *pgtype.Timestamptz) ([]*entities.StudentEventLog, error)
		RetrieveBySessions(ctx context.Context, db database.QueryExecer, sessionIDs pgtype.TextArray) (map[string][]*entities.StudentEventLog, error)
	}

	SchoolRepo interface {
		Create(ctx context.Context, db database.QueryExecer, s *entities.School) error
		RetrieveCountries(ctx context.Context, db database.QueryExecer, schoolIDs pgtype.Int4Array) ([]string, error)
	}

	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error)
	}

	QuestionSetRepo interface {
		FindByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) ([]*entities.QuestionSets, error)
		CreateAll(ctx context.Context, db database.QueryExecer, quizsets []*entities.QuestionSets) error
	}

	ActivityLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.ActivityLog) error
	}

	LessonRepo interface {
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.LessonFilter) ([]*entities.Lesson, error)
		EndLiveLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, endTime pgtype.Timestamptz) error
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
		UpdateRoomID(ctx context.Context, db database.QueryExecer, lessonID, roomID pgtype.Text) error
	}
	CourseClassRepo interface {
		Find(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (mapCourseIDsByClassID map[pgtype.Int4]pgtype.TextArray, err error)
	}
	YasuoCourseClassRepo interface {
		SoftDeleteClass(ctx context.Context, db database.QueryExecer, classID pgtype.Int4) error
		UpsertV2(ctx context.Context, db database.Ext, cc []*entities.CourseClass) error
	}
	PresetStudyPlanRepo interface {
		RetrievePresetStudyPlanWeeklies(ctx context.Context, db database.QueryExecer, presetStudyPlanID pgtype.Text) ([]*entities.PresetStudyPlanWeekly, error)
	}
	CourseRepo interface {
		RetrieveCourses(context.Context, database.QueryExecer, *repositories.CourseQuery) (entities.Courses, error)
		FindByID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*entities.Course, error)
	}
	MediaRepo interface {
		UpsertMediaBatch(ctx context.Context, db database.QueryExecer, media entities.Medias) error
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, mediaIDs pgtype.TextArray) ([]*entities.Media, error)
	}
	WhiteboardSvc interface {
		FetchRoomToken(ctx context.Context, roomUUID string) (string, error)
		CreateRoom(context.Context, *whiteboard.CreateRoomRequest) (*whiteboard.CreateRoomResponse, error)
	}
	LessonMemberRepo interface {
		CourseAccessible(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]string, error)
	}
	LessonGroupRepo interface {
		Get(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities.LessonGroup, error)
	}
	MasterClassRepo interface {
		UpsertClasses(ctx context.Context, db database.Ext, classes []*domain.Class) error
		GetByID(ctx context.Context, db database.QueryExecer, id string) (*domain.Class, error)
	}
	MasterClassMemberRepo interface {
		UpsertClassMembers(ctx context.Context, db database.QueryExecer, classMembers []*domain.ClassMember) error
		GetByClassIDAndUserIDs(ctx context.Context, db database.QueryExecer, classID string, userIDs []string) (map[string]*domain.ClassMember, error)
	}
}

func (rcv *ClassService) permissionAccessControl(ctx context.Context, schoolID pgtype.Int4, classID int32, checkOwner bool) (*entities.User, error) {
	userID := interceptors.UserIDFromContext(ctx)
	currentUser, err := rcv.UserRepo.Get(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "can't find current user")
	}
	switch currentUser.Group.String {
	case pb.USER_GROUP_ADMIN.String():
		break
	case pb.USER_GROUP_TEACHER.String():
		if schoolID.Int == 0 {
			return nil, status.Error(codes.InvalidArgument, "missing school id")
		}
		if checkOwner {
			isOwner, err := rcv.ClassMemberRepo.IsOwner(ctx, rcv.DB, database.Int4(classID), database.Text(userID))
			if err != nil {
				return nil, services.ToStatusError(err)
			}

			if !isOwner {
				return nil, status.Error(codes.PermissionDenied, "you are not owner of this class")
			}
		} else {
			isInSchool, err := rcv.TeacherRepo.IsInSchool(ctx, rcv.DB, database.TextArray([]string{currentUser.ID.String}), schoolID)
			if err != nil {
				return nil, fmt.Errorf("rcv.TeacherRepo.IsInSchool: %w", err)
			}

			if !isInSchool {
				return nil, status.Error(codes.PermissionDenied, "cannot handle class of other school")
			}
		}
	case pb.USER_GROUP_SCHOOL_ADMIN.String():
		if schoolID.Int == 0 {
			return nil, status.Error(codes.InvalidArgument, "missing school id")
		}
		schoolAdmin, err := rcv.SchoolAdminRepo.Get(ctx, rcv.DB, currentUser.ID)
		if err != nil {
			return nil, fmt.Errorf("rcv.SchoolAdminRepo.Get: %w", err)
		}

		if schoolAdmin == nil || schoolAdmin.SchoolID.Int != schoolID.Int {
			return nil, status.Error(codes.PermissionDenied, "cannot handle class of other school")
		}
	default:
		return nil, status.Error(codes.PermissionDenied, "user group not allowed")
	}

	return currentUser, err
}

func (rcv *ClassService) UpdateClassCode(ctx context.Context, req *pb.UpdateClassCodeRequest) (*pb.UpdateClassCodeResponse, error) {
	if req.ClassId == 0 {
		return nil, status.Error(codes.InvalidArgument, "class id cannot be empty")
	}

	pgClassID := database.Int4(req.ClassId)
	class, err := rcv.ClassRepo.FindByID(ctx, rcv.DB, pgClassID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot find class")
	}

	_, err = rcv.permissionAccessControl(ctx, class.SchoolID, class.ID.Int, true)
	if err != nil {
		return nil, err
	}

	classCode := database.Text(entities.GenerateClassCode(rcv.ClassCodeLength))
	err = rcv.ClassRepo.UpdateClassCode(ctx, rcv.DB, pgClassID, classCode)
	if err != nil {
		return nil, fmt.Errorf("updateClassCode: %w", err)
	}

	return &pb.UpdateClassCodeResponse{
		ClassId:   class.ID.Int,
		ClassCode: class.Code.String,
	}, nil
}

func (rcv *ClassService) CreateClass(ctx context.Context, req *pb.CreateClassRequest) (*pb.CreateClassResponse, error) {
	if req.ClassName == "" {
		return nil, status.Error(codes.InvalidArgument, "missing className")
	}

	user, err := rcv.permissionAccessControl(ctx, database.Int4(req.SchoolId), 0, false)
	if err != nil {
		return nil, err
	}

	if user.Group.String == pb.USER_GROUP_TEACHER.String() {
		req.OwnerIds = []string{user.ID.String}
	}

	if len(req.OwnerIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing ownerId")
	}

	users := []*entities.User{user}

	if user.Group.String == pb.USER_GROUP_ADMIN.String() ||
		user.Group.String == pb.USER_GROUP_SCHOOL_ADMIN.String() {
		filter := repositories.UserFindFilter{}
		err := multierr.Combine(
			filter.UserGroup.Set(nil),
			filter.Email.Set(nil),
			filter.Phone.Set(nil),
			filter.IDs.Set(nil),
		)

		if len(req.OwnerIds) > 0 {
			filter.IDs.Set(req.OwnerIds)
		}

		if err != nil {
			return nil, err
		}

		users, err = rcv.UserRepo.Find(ctx, rcv.DB, &filter, "user_id", "name", "user_group")
		if err != nil {
			return nil, services.ToStatusError(err)
		}

		if len(users) == 0 {
			return nil, status.Error(codes.InvalidArgument, "cannot find users")
		}
	}

	userGroup := pb.USER_GROUP_TEACHER.String()
	for _, u := range users {
		if u.Group.String != userGroup {
			return nil, status.Error(codes.InvalidArgument, "owner must be teacher")
		}
	}

	classID, err := rcv.createClass(ctx, &CreateClassOpts{
		CreateClassRequest: req,
		Country:            pb.Country(pb.Country_value[user.Country.String]),
		OwnerUserGroup:     pb.UserGroup(pb.UserGroup_value[userGroup]),
	})
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	return &pb.CreateClassResponse{
		ClassId: classID,
	}, nil
}

type CreateClassOpts struct {
	*pb.CreateClassRequest
	Country pb.Country
	// use if pb.CreateClassRequest.OwnerIds != nil
	OwnerUserGroup pb.UserGroup
	PresetClassID  uint
}

func (rcv *ClassService) createClass(ctx context.Context, req *CreateClassOpts) (int32, error) {
	// TODO: change classID to string
	rcv.l.Lock()
	defer rcv.l.Unlock()

	var classID int32
	err := database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		var (
			grades   []int
			subjects []string
		)

		for _, g := range req.Grades {
			gradeInt, err := i18n.ConvertStringGradeToInt(req.Country, g)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Invalid grade %s", g))
			}

			grades = append(grades, gradeInt)
		}

		for _, s := range req.Subjects {
			subjects = append(subjects, s.String())
		}

		avatars, err := rcv.ConfigRepo.Find(ctx, tx, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar))
		if err != nil {
			return errors.Wrap(err, "rcv.ConfigRepo.Find")
		}
		randomAvatar := pgtype.Text{String: defaultClassAvatar}
		if len(avatars) > 0 {
			randomAvatar = avatars[rand.Intn(len(avatars))].Value
		}

		planID, planExpiredAt, planDuration, err := rcv.GetSchoolConfig(ctx, tx, req.SchoolId, req.Country.String())
		if err != nil {
			return errors.Wrap(err, "rcv.GetSchoolConfig")
		}

		e := new(entities.Class)
		database.AllNullEntity(e)
		err = multierr.Combine(
			e.SchoolID.Set(req.SchoolId),
			// follow specs, this is random avt
			e.Avatar.Set(randomAvatar.String),
			e.Name.Set(req.ClassName),
			e.Subjects.Set(subjects),
			e.Grades.Set(grades),
			e.PlanID.Set(planID),
			e.Country.Set(req.Country.String()),
			e.PlanDuration.Set(planDuration),
			e.Status.Set(entities.ClassStatusActive),
			e.Code.Set(entities.GenerateClassCode(rcv.ClassCodeLength)),
		)

		if req.PresetClassID != 0 {
			err = multierr.Combine(err, e.ID.Set(req.PresetClassID))
		} else {
			nextID, nErr := rcv.ClassRepo.GetNextID(ctx, tx)
			if nErr != nil {
				return fmt.Errorf("err getNextClassID: %w", err)
			}

			e.ID = *nextID
		}

		if err != nil {
			return err
		}

		if !planExpiredAt.IsZero() {
			_ = e.PlanExpiredAt.Set(planExpiredAt)
		}

		err = rcv.ClassRepo.Create(ctx, tx, e)
		if err != nil {
			return fmt.Errorf("rcv.ClassRepo.Create: %w", err)
		}

		classID = e.ID.Int

		err = rcv.PublishClassEvt(ctx, &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_CreateClass_{
				CreateClass: &pb.EvtClassRoom_CreateClass{
					ClassId:    classID,
					ClassName:  req.ClassName,
					TeacherIds: req.OwnerIds,
				},
			},
		})

		if err != nil {
			return errors.Wrap(err, "rcv.PublishClassEvt")
		}

		if len(req.OwnerIds) != 0 {
			err = rcv.createClassMember(ctx, tx, e.ID.Int, req.OwnerIds, req.OwnerUserGroup.String(), "", true, false)
			if err != nil {
				return errors.Wrap(err, "rcv.createClassMember")
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	err = rcv.PublishClassEvt(ctx, &pb.EvtClassRoom{
		Message: &pb.EvtClassRoom_ActiveConversation_{
			ActiveConversation: &pb.EvtClassRoom_ActiveConversation{
				ClassId: classID,
				Active:  true,
			},
		},
	})

	if err != nil {
		ctxzap.Extract(ctx).Warn("rcv.PublishClassEvt", zap.Error(err))
	}

	return classID, nil
}

func (rcv *ClassService) EditClass(ctx context.Context, req *pb.EditClassRequest) (*pb.EditClassResponse, error) {
	if req.ClassId == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing classId")
	}

	if req.ClassName == "" {
		return nil, status.Error(codes.InvalidArgument, "missing className")
	}

	classID := database.Int4(req.ClassId)
	class, err := rcv.ClassRepo.FindByID(ctx, rcv.DB, classID)
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	_, err = rcv.permissionAccessControl(ctx, class.SchoolID, req.ClassId, true)
	if err != nil {
		return nil, err
	}

	_ = class.Name.Set(req.ClassName)
	err = rcv.ClassRepo.Update(ctx, rcv.DB, class)
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	err = rcv.PublishClassEvt(ctx, &pb.EvtClassRoom{
		Message: &pb.EvtClassRoom_EditClass_{
			EditClass: &pb.EvtClassRoom_EditClass{
				ClassId:   req.ClassId,
				ClassName: req.ClassName,
			},
		},
	})

	if err != nil {
		ctxzap.Extract(ctx).Warn("rcv.PublishClassEvt", zap.Error(err))
	}

	return &pb.EditClassResponse{}, nil
}

// look deprecated,
// No. JPREP still use it
func (rcv *ClassService) JoinClass(ctx context.Context, req *pb.JoinClassRequest) (*pb.JoinClassResponse, error) {
	if req.ClassCode == "" {
		ctxzap.Extract(ctx).Warn("invalid classCode", zap.String("class-code", req.ClassCode))
		return nil, services.StatusErrWithDetail(codes.InvalidArgument, "invalid classCode", errWrongClassCodeMsg)
	}

	class, err := rcv.ClassRepo.FindByCode(ctx, rcv.DB, database.Text(req.ClassCode))
	if err != nil {
		if errors.Cause(err) != pgx.ErrNoRows {
			return nil, services.ToStatusError(err)
		}
		return nil, status.Error(codes.NotFound, "invalid classCode")
	}

	now := timeutil.Now()
	if !class.PlanExpiredAt.Time.IsZero() && now.After(class.PlanExpiredAt.Time) {
		return nil, services.StatusErrWithDetail(codes.FailedPrecondition, "class code is expired", errClassCodeExpired)
	}

	userID := interceptors.UserIDFromContext(ctx)
	member, err := rcv.ClassMemberRepo.Get(ctx, rcv.DB, class.ID, database.Text(userID), database.Text(entities.ClassMemberStatusActive))
	if err != nil {
		if errors.Cause(err) != pgx.ErrNoRows {
			return nil, services.ToStatusError(err)
		}
	}

	if member != nil {
		return nil, status.Error(codes.AlreadyExists, "You’ve activated this code already")
	}

	user, err := rcv.UserRepo.Get(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	switch user.Group.String {
	case pb.USER_GROUP_TEACHER.String():
		err := rcv.handleJoinClassForTeacher(ctx, class, []string{userID})
		if err != nil {
			return nil, err
		}

		return &pb.JoinClassResponse{
			ClassId: class.ID.Int,
		}, nil
	case pb.USER_GROUP_STUDENT.String():
		break
	default:
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	err = database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		// When use createClassMember with blank student subscription id, it will create class member with null student subscription id
		err = rcv.createClassMember(ctx, tx, class.ID.Int, []string{userID}, user.Group.String, "", false, true)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, services.ToStatusError(err)
	}

	return &pb.JoinClassResponse{
		ClassId: class.ID.Int,
	}, nil
}

func (rcv *ClassService) handleJoinClassForTeacher(ctx context.Context, class *entities.Class, teacherIDs []string) error {
	if len(teacherIDs) == 0 {
		return status.Error(codes.InvalidArgument, "missing teacher ids")
	}

	isInSchool, err := rcv.TeacherRepo.IsInSchool(ctx, rcv.DB, database.TextArray(teacherIDs), class.SchoolID)
	if err != nil {
		return errors.Wrap(err, "rcv.TeacherRepo.IsInSchool")
	}

	if !isInSchool {
		return services.StatusErrWithDetail(codes.FailedPrecondition, "invalid class", errClassDoesNotBelongToYourSchoolMsg)
	}

	err = database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		return rcv.createClassMember(ctx, tx, class.ID.Int, teacherIDs, pb.USER_GROUP_TEACHER.String(), "", true, true)
	})

	if err != nil {
		return services.ToStatusError(err)
	}

	return nil
}

// Seem use for manabie old, look deprecated
func (rcv *ClassService) AddClassMember(ctx context.Context, req *pb.AddClassMemberRequest) (*pb.AddClassMemberResponse, error) {
	if req.ClassId == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing class id")
	}
	if len(req.TeacherIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing teacher ids")
	}

	class, err := rcv.ClassRepo.FindByID(ctx, rcv.DB, database.Int4(req.ClassId))
	if err != nil {
		if errors.Cause(err) != pgx.ErrNoRows {
			return nil, services.ToStatusError(err)
		}
		return nil, status.Error(codes.NotFound, "invalid class id")
	}

	now := timeutil.Now()
	if !class.PlanExpiredAt.Time.IsZero() && now.After(class.PlanExpiredAt.Time) {
		return nil, services.StatusErrWithDetail(codes.FailedPrecondition, "class is expired", errClassCodeExpired)
	}

	currentUser, err := rcv.permissionAccessControl(ctx, class.SchoolID, req.ClassId, false)
	if err != nil {
		return nil, err
	}

	if currentUser.Group.String == pb.USER_GROUP_TEACHER.String() {
		return nil, status.Error(codes.PermissionDenied, "user group not allowed")
	}

	teachers, err := rcv.TeacherRepo.Retrieve(ctx, rcv.DB, database.TextArray(req.TeacherIds), "user_id", "user_group")
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	if len(teachers) == 0 || len(teachers) != len(req.TeacherIds) {
		return nil, status.Error(codes.InvalidArgument, "cannot find all teacher")
	}

	members, err := rcv.ClassMemberRepo.FindByIDs(ctx, rcv.DB, class.ID, database.TextArray(req.TeacherIds), database.Text(entities.ClassMemberStatusActive))
	if err != nil {
		if errors.Cause(err) != pgx.ErrNoRows {
			return nil, services.ToStatusError(err)
		}
	}

	if len(members) != 0 {
		return nil, status.Error(codes.AlreadyExists, "some teacher activated this code already")
	}

	err = rcv.handleJoinClassForTeacher(ctx, class, req.TeacherIds)
	if err != nil {
		return nil, err
	}

	logger := ctxzap.Extract(ctx)
	go rcv.trackLogWhenClassChange(logger, currentUser.ID.String, []string{}, class.ID.Int, "add")

	return &pb.AddClassMemberResponse{}, nil
}

// Seem use for manabie old, look deprecated
func (rcv *ClassService) RemoveMember(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	if req.ClassId == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing classId")
	}

	if len(req.UserIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing userIDs")
	}

	classID := database.Int4(req.ClassId)
	class, err := rcv.ClassRepo.FindByID(ctx, rcv.DB, classID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot find class")
	}

	user, err := rcv.permissionAccessControl(ctx, class.SchoolID, req.ClassId, true)
	if err != nil {
		return nil, services.ToStatusError(err)
	}
	userID := user.ID.String
	userGroup := user.Group.String

	if userGroup == pb.USER_GROUP_TEACHER.String() {
		for _, uId := range req.UserIds {
			if userID == uId {
				return nil, status.Error(codes.PermissionDenied, "can not remove yourself")
			}
		}
	}

	err = database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := rcv.removeMember(ctx, tx, int64(req.ClassId), req.UserIds, true)
		if err != nil {
			return errors.Wrap(err, "rcv.removeMember")
		}

		return nil
	})
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	logger := ctxzap.Extract(ctx)
	go rcv.trackLogWhenClassChange(logger, user.ID.String, req.UserIds, class.ID.Int, "remove")

	return &pb.RemoveMemberResponse{}, nil
}

func (rcv *ClassService) LeaveClass(ctx context.Context, req *pb.LeaveClassRequest) (*pb.LeaveClassResponse, error) {
	if req.ClassId == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing classId")
	}

	userID := interceptors.UserIDFromContext(ctx)

	err := database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := rcv.removeMember(ctx, tx, int64(req.ClassId), []string{userID}, false)
		if err != nil {
			return errors.Wrap(err, "rcv.removeMember")
		}

		return nil
	})

	if err != nil {
		return nil, services.ToStatusError(err)
	}

	return &pb.LeaveClassResponse{}, nil
}

func (rcv *ClassService) RetrieveAssignedPresetStudyPlan(context.Context, *pb.RetrieveAssignedPresetStudyPlanRequest) (*pb.RetrieveAssignedPresetStudyPlanResponse, error) {
	panic("implement me")
}

func (rcv *ClassService) createClassMember(ctx context.Context, tx database.QueryExecer, classID int32, userIDs []string, group, studentSubscriptionID string, isOwner, sendToNats bool) error {
	err := rcv.upsertMasterClassMember(ctx, tx, classID, userIDs, false)
	if err != nil {
		ctxzap.Extract(ctx).Warn("upsertMasterClassMember", zap.Error(err))
	}
	classMembers, err := rcv.ClassMemberRepo.FindByIDs(ctx, tx, database.Int4(classID), database.TextArray(userIDs), database.Text(entities.ClassMemberStatusActive))
	if err != nil && errors.Cause(err) != pgx.ErrNoRows {
		return errors.Wrap(err, "rcv.ClassMemberRepo.FindByIDs")
	}

	if len(classMembers) != 0 {
		return status.Error(codes.AlreadyExists, "already joined this class")
	}
	for _, userID := range userIDs {
		classMember := new(entities.ClassMember)
		database.AllNullEntity(classMember)
		if studentSubscriptionID != "" {
			_ = classMember.StudentSubscriptionID.Set(studentSubscriptionID)
		}
		err := multierr.Combine(
			classMember.ID.Set(ksuid.New().String()),
			classMember.ClassID.Set(classID),
			classMember.UserID.Set(userID),
			classMember.UserGroup.Set(group),
			classMember.IsOwner.Set(isOwner),
			classMember.Status.Set(entities.ClassMemberStatusActive),
		)
		if err != nil {
			return err
		}
		err = rcv.ClassMemberRepo.Create(ctx, tx, classMember)
		if err != nil {
			return errors.Wrap(err, "rcv.ClassMemberRepo.Create")
		}
		if !sendToNats {
			continue
		}
		err = rcv.PublishClassEvt(ctx, &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_JoinClass_{
				JoinClass: &pb.EvtClassRoom_JoinClass{
					ClassId:   classID,
					UserId:    userID,
					UserGroup: pb.UserGroup(pb.UserGroup_value[group]),
				},
			},
		})

		if err != nil {
			ctxzap.Extract(ctx).Warn("rcv.PublishClassEvt", zap.Error(err))
		}
	}
	return nil
}

func (rcv *ClassService) upsertMasterClassMember(ctx context.Context, tx database.QueryExecer, classID int32, userIDs []string, deleted bool) error {
	class, err := rcv.MasterClassRepo.GetByID(ctx, tx, strconv.Itoa(int(classID)))
	if err != nil {
		return errors.Wrap(err, "rcv.MasterClassRepo.GetByID")
	}
	classMembers := []*domain.ClassMember{}
	now := timeutil.Now()
	members, err := rcv.MasterClassMemberRepo.GetByClassIDAndUserIDs(ctx, tx, class.ClassID, userIDs)
	if err != nil {
		return errors.Wrap(err, "rcv.MasterClassRepo.GetByClassIDAndUserIDs")
	}
	for _, userID := range userIDs {
		classMemberID := idutil.ULIDNow()
		if row, ok := members[userID]; ok {
			classMemberID = row.ClassMemberID
		}
		classMember := &domain.ClassMember{
			ClassMemberID: classMemberID,
			ClassID:       class.ClassID,
			UserID:        userID,
			CreatedAt:     now,
			UpdatedAt:     now,
			StartDate:     now, // JPREP JoinClassRequest don't have start_date, end_date for class_member
			EndDate:       now,
			DeletedAt:     nil,
		}
		if deleted {
			classMember.DeletedAt = &now
		}
		classMembers = append(classMembers, classMember)
	}
	if len(classMembers) > 0 {
		err = rcv.MasterClassMemberRepo.UpsertClassMembers(ctx, tx, classMembers)
		if err != nil {
			return errors.Wrap(err, "rcv.MasterClassMemberRepo.UpsertClassMembers")
		}
	}

	return nil
}

func (rcv *ClassService) GetSchoolConfig(ctx context.Context, db database.QueryExecer, schoolID int32, country string) (planID string, planExpiredAt time.Time, planDuration int32, err error) {
	var schoolConfig *entities.SchoolConfig
	schoolConfig, err = rcv.SchoolConfigRepo.FindByID(ctx, db, pgtype.Int4{Int: schoolID, Status: pgtype.Present})
	if err == nil {
		planID = schoolConfig.PlanID.String
		planExpiredAt = schoolConfig.PlanExpiredAt.Time
		planDuration = int32(schoolConfig.PlanDuration.Int)
		return
	}

	// fallback get from common config
	var (
		configs    []*entities.Config
		planConfig *entities.Config
		ok         bool
	)

	configs, err = rcv.ConfigRepo.Find(ctx, db, database.Text(country), database.Text(classPlan))
	if err != nil {
		err = errors.Wrap(err, "rcv.ConfigRepo.Find")
		return
	}

	mapConfig := services.ConfigMap(configs)

	planConfig, ok = mapConfig[classPlanName]
	if !ok {
		err = errors.New("missing config classPlan in db")
		return
	}

	planID = planConfig.Value.String

	planPeriod, ok := mapConfig[classPlanPeriod]
	if !ok {
		err = errors.New("missing config planPeriod in db")
		return
	}
	planExpiredAt, err = time.Parse("2006-01-02 15:04:05", planPeriod.Value.String)
	if err != nil {
		if n, errParseInt := strconv.Atoi(planPeriod.Value.String); errParseInt == nil {
			planDuration = int32(n)
			err = nil
			return
		}

		err = errors.Wrap(err, "time.Parse")
		return
	}

	return
}

func (rcv *ClassService) removeMember(ctx context.Context, tx database.QueryExecer, classID int64, userIDs []string, isKicked bool) error {
	err := rcv.upsertMasterClassMember(ctx, tx, int32(classID), userIDs, true)
	if err != nil {
		ctxzap.Extract(ctx).Warn("upsertMasterClassMember", zap.Error(err))
	}
	var pgClassID pgtype.Int4
	_ = pgClassID.Set(classID)

	members, err := rcv.ClassMemberRepo.UpdateStatus(ctx, tx, pgClassID, database.TextArray(userIDs), database.Text(entities.ClassMemberStatusInactive))
	if err != nil {
		return errors.Wrap(err, "rcv.ClassMemberRepo.UpdateStatus")
	}

	var subscriptionIDs []string
	for _, m := range members {
		if m.StudentSubscriptionID.Status == pgtype.Present {
			subscriptionIDs = append(subscriptionIDs, m.StudentSubscriptionID.String)
		}
	}

	err = rcv.PublishClassEvt(ctx, &pb.EvtClassRoom{
		Message: &pb.EvtClassRoom_LeaveClass_{
			LeaveClass: &pb.EvtClassRoom_LeaveClass{
				ClassId:  int32(classID),
				UserIds:  userIDs,
				IsKicked: isKicked,
			},
		},
	})

	if err != nil {
		ctxzap.Extract(ctx).Warn("rcv.PublishClassEvt", zap.Error(err))
	}

	return nil
}

func (rcv *ClassService) RetrieveClassMember(ctx context.Context, req *pb.RetrieveClassMemberRequest) (*pb.RetrieveClassMemberResponse, error) {
	filter := &repositories.FindClassMemberFilter{
		ClassIDs: database.Int4Array([]int32{req.ClassId}),
		Status:   database.Text(entities.ClassMemberStatusActive),
	}
	if err := multierr.Combine(
		filter.Group.Set(nil),
		filter.OffsetID.Set(nil),
		filter.Limit.Set(nil),
		filter.UserName.Set(nil),
	); err != nil {
		return nil, fmt.Errorf("RetrieveClassMember.SetFilter: %w", err)
	}

	members, err := rcv.ClassMemberRepo.Find(ctx, rcv.DB, filter)
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	data := make([]*pb.RetrieveClassMemberResponse_Member, 0, len(members))
	for _, m := range members {
		joinedAt, _ := types.TimestampProto(m.CreatedAt.Time)
		data = append(data, &pb.RetrieveClassMemberResponse_Member{
			UserId:    m.UserID.String,
			UserGroup: pb.UserGroup(pb.UserGroup_value[m.UserGroup.String]),
			JoinAt:    joinedAt,
		})
	}

	return &pb.RetrieveClassMemberResponse{
		Members: data,
	}, nil
}

func (rcv *ClassService) makeStudentAssignmentList(studentIDs []string, assignmentId string) []*entities.StudentAssignment {
	studentAssignments := make([]*entities.StudentAssignment, 0, len(studentIDs))
	for _, studentID := range studentIDs {
		studentAssignment := &entities.StudentAssignment{}
		database.AllNullEntity(studentAssignment)
		studentAssignment.StudentID.Set(studentID)
		studentAssignment.AssignmentID.Set(assignmentId)
		studentAssignment.AssignmentStatus.Set(entities.StudentAssignmentStatusActive)
		studentAssignments = append(studentAssignments, studentAssignment)
	}
	return studentAssignments
}

func (rcv *ClassService) RegisterTeacher(ctx context.Context, req *pb.RegisterTeacherRequest) (*pb.RegisterTeacherResponse, error) {
	return nil, nil
}

func (rcv *ClassService) EditAssignedTopic(ctx context.Context, req *pb.EditAssignedTopicRequest) (*pb.EditAssignedTopicResponse, error) {
	return nil, nil
}

func (rcv *ClassService) RemoveAssignedTopic(ctx context.Context, req *pb.RemoveAssignedTopicRequest) (*pb.RemoveAssignedTopicResponse, error) {
	return &pb.RemoveAssignedTopicResponse{}, nil
}

func (rcv *ClassService) toPbAssignment(assignments []*entities.Assignment, topics []*entities.Topic, studentAssignments []*entities.StudentAssignment, isPassed bool) []*pb.Assignment {
	status := pb.ASSIGNMENT_STATUS_NONE

	topicMap := make(map[string]*pb.Topic, len(topics))
	for _, topic := range topics {
		topicMap[topic.ID.String] = services.ToTopicPb(topic)
	}

	studentIDsMapByAssignment := make(map[string][]string)
	for _, studentAssignment := range studentAssignments {
		studentIDs := studentIDsMapByAssignment[studentAssignment.AssignmentID.String]
		studentIDsMapByAssignment[studentAssignment.AssignmentID.String] = append(studentIDs, studentAssignment.StudentID.String)
	}

	pbAssignments := make([]*pb.Assignment, 0, len(assignments))

	for _, assignment := range assignments {
		if !isPassed && assignment.StartDate.Time.Unix() <= time.Now().Unix() {
			status = pb.ASSIGNMENT_STATUS_STARTED
		}
		if !isPassed && assignment.StartDate.Time.Unix() > time.Now().Unix() {
			status = pb.ASSIGNMENT_STATUS_UPCOMING
		}
		startDate, _ := types.TimestampProto(assignment.StartDate.Time)
		endDate, _ := types.TimestampProto(assignment.EndDate.Time)
		pbAssignment := &pb.Assignment{
			AssignmentId: assignment.AssignmentID.String,
			Status:       status,
			StartDate:    startDate,
			EndDate:      endDate,
			Topic:        topicMap[assignment.TopicID.String],
			StudentIds:   studentIDsMapByAssignment[assignment.AssignmentID.String],
		}
		pbAssignments = append(pbAssignments, pbAssignment)
	}
	return pbAssignments
}

func (rcv *ClassService) RetrieveActiveClassAssignment(ctx context.Context, req *pb.RetrieveActiveClassAssignmentRequest) (*pb.RetrieveActiveClassAssignmentResponse, error) {
	return &pb.RetrieveActiveClassAssignmentResponse{}, nil
}

func (rcv *ClassService) RetrievePastClassAssignment(ctx context.Context, req *pb.RetrievePastClassAssignmentRequest) (*pb.RetrievePastClassAssignmentResponse, error) {
	return &pb.RetrievePastClassAssignmentResponse{}, nil
}

func (rcv *ClassService) TeacherAssignClassWithTopic(ctx context.Context, req *pb.TeacherAssignClassWithTopicRequest) (*pb.TeacherAssignClassWithTopicResponse, error) {
	return nil, nil
}

// look deprecated
func (rcv *ClassService) RetrieveClassLearningStatistics(ctx context.Context, req *pb.RetrieveClassLearningStatisticsRequest) (*pb.RetrieveClassLearningStatisticsResponse, error) {
	return &pb.RetrieveClassLearningStatisticsResponse{}, nil
}

type studentLearningStats struct {
	studentName string

	completion     float64
	completionDate time.Time

	accuracy  float64
	timeSpent int64 // in seconds
}

func (rcv *ClassService) RetrieveStudentLearningStatistics(ctx context.Context, req *pb.RetrieveStudentLearningStatisticsRequest) (*pb.RetrieveStudentLearningStatisticsResponse, error) {
	return &pb.RetrieveStudentLearningStatisticsResponse{}, nil
}

// groupSubmissions groups submissions histories by session id together,
// e.g. if we have a list of submissions like this (sorted by created date):
//
//   - question id: 1, correct: true, session id: s1
//   - question id: 2, correct: false, session id: s1
//   - question id: 1, correct: false, session id: s2
//   - question id: 3, correct: true, session id: s1
//   - question id: 2, correct: false, session id: s2
//   - question id: 1, correct: true, session id: s3
//
// then this function will return:
//
//	[
//		[
//			question id: 1, correct: true, session id: s1,
//			question id: 2, correct: false, session id: s1,
//			question id: 3, correct: true, session id: s1
//		],
//		[
//			question id: 1, correct: false, session id: s2,
//			question id: 2, correct: false, session id: s2
//		],
//		[
//			question id: 1, correct: true, session id: s3
//		]
//	]
func groupSubmissions(submissions []*repositories.SubmissionResult) (ret [][]*repositories.SubmissionResult) {
	var i int
	sessions := make(map[string]int)

	for _, s := range submissions {
		if _, ok := sessions[s.SessionID]; !ok {
			sessions[s.SessionID] = i
			ret = append(ret, nil)
			i++
		}
		ret[sessions[s.SessionID]] = append(ret[sessions[s.SessionID]], s)
	}
	return
}

func (rcv *ClassService) PublishClassEvt(ctx context.Context, msg *pb.EvtClassRoom) error {
	var msgID string
	data, _ := msg.Marshal()

	msgID, err := rcv.JSM.PublishAsyncContext(ctx, global_constant.SubjectClassUpserted, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishClassEvt rcv.JSM.PublishAsyncContext Class.Upserted failed, msgID: %s, %w", msgID, err))
	}

	return err
}

func (rcv *ClassService) trackLogWhenClassChange(logger *zap.Logger, userID string, classMemberIDs []string, classID int32, logAction string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a := new(entities.ActivityLog)
	a.ID.Set(ksuid.New().String())
	a.UserID.Set(userID)
	if logAction == "add" {
		a.ActionType.Set(entities.LogActionAddClassMember)
	}
	if logAction == "remove" {
		a.ActionType.Set(entities.LogActionRemoveClassMember)
	}
	payload := map[string]interface{}{
		"class_id":         classID,
		"class_member_ids": classMemberIDs,
	}
	a.Payload.Set(payload)

	if cerr := rcv.ActivityLogRepo.Create(ctx, rcv.DB, a); cerr != nil {
		logger.Error("s.ActivityLogRepo.Create", zap.Error(cerr))
	}
}

type CreateAssignmentArgument struct {
	TeacherID      string
	ClassID        int32
	StudentIDs     []string
	TopicIDs       []string
	StartDate      time.Time
	EndDate        time.Time
	AssignmentType string
}

func (rcv *ClassService) TeacherAssignStudentWithTopic(ctx context.Context, req *pb.TeacherAssignStudentWithTopicRequest) (*pb.TeacherAssignStudentWithTopicResponse, error) {
	return nil, nil
}

func (s *ClassService) SyncClass(ctx context.Context, req []*npb.EventMasterRegistration_Class) error {
	masterClasses := []*domain.Class{}
	now := timeutil.Now()
	for _, r := range req {
		class, err := s.ClassRepo.FindByID(ctx, s.DB, database.Int4(int32(r.ClassId)))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("err findClass: %w", err)
		}

		masterClass := &domain.Class{
			ClassID:    strconv.Itoa(int(r.ClassId)),
			Name:       r.ClassName,
			SchoolID:   strconv.Itoa(global_constant.JPREPSchool),
			CreatedAt:  now,
			UpdatedAt:  now,
			LocationID: global_constant.JPREPOrgLocation,
			DeletedAt:  nil,
		}
		switch r.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			if class == nil {
				_, err = s.createClass(ctx, &CreateClassOpts{
					CreateClassRequest: &pb.CreateClassRequest{
						ClassName: r.ClassName,
						// consider: create a new school for jpref or still use ManabieSchool
						SchoolId: global_constant.JPREPSchool,
					},
					Country:       pb.COUNTRY_JP,
					PresetClassID: uint(r.ClassId),
				})
				if err != nil {
					return fmt.Errorf("s.createClass: %w", err)
				}
			} else {
				performActiveConversation := class.Status.String == entities.ClassStatusInactive

				class.Name = database.Text(r.ClassName)
				class.Status = database.Text(entities.ClassStatusActive)
				err = s.ClassRepo.Update(ctx, s.DB, class)
				if err != nil {
					return fmt.Errorf("err UpdateClass: %w", err)
				}

				if performActiveConversation {
					// publish event active conversation
					err = s.PublishClassEvt(ctx, &pb.EvtClassRoom{
						Message: &pb.EvtClassRoom_ActiveConversation_{
							ActiveConversation: &pb.EvtClassRoom_ActiveConversation{
								ClassId: class.ID.Int,
								Active:  true,
							},
						},
					})

					if err != nil {
						ctxzap.Extract(ctx).Warn("rcv.PublishClassEvt", zap.Error(err))
					}
				}
			}

			err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
				err = s.YasuoCourseClassRepo.SoftDeleteClass(ctx, tx, database.Int4(int32(r.ClassId)))
				if err != nil {
					return fmt.Errorf("SoftDeleteCourseClass: %w", err)
				}

				// create course class
				courseClass := &entities.CourseClass{}
				database.AllNullEntity(courseClass)
				err = multierr.Combine(
					courseClass.CourseID.Set(r.CourseId),
					courseClass.ClassID.Set(r.ClassId),
					courseClass.Status.Set(entities.CourseClassStatusActive),
				)
				if err != nil {
					return fmt.Errorf("err set CourseClass: %w", err)
				}

				err = s.YasuoCourseClassRepo.UpsertV2(ctx, tx, []*entities.CourseClass{courseClass})
				if err != nil {
					return fmt.Errorf("err s.CourseClassRepo.UpsertV2: %w", err)
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("database.ExecInTx: %w", err)
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			if class == nil {
				continue
			}
			masterClass.DeletedAt = &now
			class.Status = database.Text(entities.ClassStatusInactive)
			err = s.ClassRepo.Update(ctx, s.DB, class)
			if err != nil {
				return fmt.Errorf("err UpdateClass: %w", err)
			}

			// publish event inactive conversation
			err = s.PublishClassEvt(ctx, &pb.EvtClassRoom{
				Message: &pb.EvtClassRoom_ActiveConversation_{
					ActiveConversation: &pb.EvtClassRoom_ActiveConversation{
						ClassId: class.ID.Int,
						Active:  false,
					},
				},
			})

			if err != nil {
				ctxzap.Extract(ctx).Warn("rcv.PublishClassEvt", zap.Error(err))
			}

			err = s.YasuoCourseClassRepo.SoftDeleteClass(ctx, s.DB, database.Int4(int32(r.ClassId)))
			if err != nil {
				return fmt.Errorf("SoftDeleteCourseClass: %w", err)
			}
		}
		course, err := s.CourseRepo.FindByID(ctx, s.DB, database.Text(r.CourseId))
		if err != nil {
			ctxzap.Extract(ctx).Warn("CourseRepo.FindByID", zap.Error(err))
		}
		if course != nil {
			masterClass.CourseID = course.ID.String
			masterClasses = append(masterClasses, masterClass)
		}
	}
	if len(masterClasses) > 0 {
		s.MasterClassRepo.UpsertClasses(ctx, s.DB, masterClasses)
	}
	return nil
}

func (s *ClassService) SyncClassMember(ctx context.Context, req []*npb.EventUserRegistration_Student) error {
	upsertedClassMembers := map[int32][]string{}
	deletedClassMemebers := map[int32][]string{}

	var cErr error
	for _, r := range req {
		switch r.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			classIDs := []int32{}
			for _, c := range r.Packages {
				classID := int32(c.ClassId)
				classIDs = append(classIDs, classID)
				upsertedClassMembers[classID] = append(upsertedClassMembers[classID], r.StudentId)
			}

			// check class ID not present in classIDs then remove student from this class
			classIDShouldDelete, err := s.ClassMemberRepo.ClassJoinNotIn(ctx, s.DB, database.Text(r.StudentId), database.Int4Array(classIDs))
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("err find old class to remove member: %w", err))
			}

			for _, classID := range classIDShouldDelete {
				deletedClassMemebers[classID] = append(deletedClassMemebers[classID], r.StudentId)
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			for _, c := range r.Packages {
				deletedClassMemebers[int32(c.ClassId)] = append(deletedClassMemebers[int32(c.ClassId)], r.StudentId)
			}
		}
	}

	for classID, memberIDs := range upsertedClassMembers {
		for _, id := range memberIDs {
			class, err := s.ClassRepo.FindByID(ctx, s.DB, database.Int4(classID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				cErr = multierr.Append(cErr, fmt.Errorf("err findClass: %w", err))
				continue
			}

			if class == nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("classID %d does not exist, studentID %s can not join", classID, id))
				continue
			}

			_, err = s.JoinClass(interceptors.ContextWithUserID(ctx, id), &pb.JoinClassRequest{
				ClassCode: class.Code.String,
			})
			if err != nil && status.Code(err) != codes.AlreadyExists {
				cErr = multierr.Combine(cErr, fmt.Errorf("err JoinClass: %d, studentID: %s, err: %w", classID, id, err))
				continue
			}
		}
	}

	for classID, memberIDs := range deletedClassMemebers {
		for _, id := range memberIDs {
			class, err := s.ClassRepo.FindByID(ctx, s.DB, database.Int4(classID))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				cErr = multierr.Append(cErr, fmt.Errorf("err findClass: %w", err))
				continue
			}

			if class == nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("classID %d does not exist, studentID %s can not leave", classID, id))
				continue
			}

			_, err = s.LeaveClass(interceptors.ContextWithUserID(ctx, id), &pb.LeaveClassRequest{
				ClassId: classID,
			})
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("err JoinClass: %d, studentID: %s, err: %w", classID, id, err))
				continue
			}
		}
	}

	return cErr
}
