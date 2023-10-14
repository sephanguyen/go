package services

import (
	"context"
	"fmt"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	"github.com/manabie-com/backend/internal/yasuo/entities"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/go-pg/pg"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SchoolService struct {
	DBTrace *database.DBTrace
	JSM     nats.JetStreamManagement

	UserRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities_bob.User, error)
	}
	SchoolRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolIDs []int32) (map[int32]*entities_bob.School, error)
		Update(ctx context.Context, db database.QueryExecer, school *entities_bob.School) (*entities_bob.School, error)
	}
	ActivityLogRepo interface {
		CreateV2(context.Context, database.Ext, *entities_bob.ActivityLog) error
	}
	TeacherRepo interface {
		JoinSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) error
		LeaveSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) error
		IsInSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) (bool, error)
	}
	ClassMemberRepo interface {
		UpdateStatus(ctx context.Context, db database.QueryExecer, userIDs []string, classIDs []int32, userGroup, status string) error
	}
	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities_bob.SchoolAdmin, error)
	}
	ClassRepo interface {
		FindBySchool(ctx context.Context, db database.QueryExecer, schoolID int32) ([]*entities_bob.Class, error)
	}
}

// MergeSchools merge school, which created from student
func (s *SchoolService) MergeSchools(ctx context.Context, req *pb.MergeSchoolsRequest) (*pb.MergeSchoolsResponse, error) {
	return &pb.MergeSchoolsResponse{
		Successful: false,
	}, nil
}

func convertPbToSchoolEn(src *pb.School) (*entities_bob.School, error) {
	if src.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "school id cannot be empty")
	}

	if src.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if src.CityId == 0 {
		return nil, status.Error(codes.InvalidArgument, "city cannot be empty")
	}
	if src.DistrictId == 0 {
		return nil, status.Error(codes.InvalidArgument, "district cannot be empty")
	}
	if src.Country.String() == "" {
		return nil, status.Error(codes.InvalidArgument, "country cannot be empty")
	}
	var p pgtype.Point
	if src.Point.Lat != 0 && src.Point.Long != 0 {
		p.Status = pgtype.Present
		p.P = pgtype.Vec2{
			X: src.Point.Lat,
			Y: src.Point.Long,
		}
	}
	school := &entities_bob.School{}
	database.AllNullEntity(school)
	school.Point = p
	err := multierr.Combine(
		school.Name.Set(src.Name),
		school.PhoneNumber.Set(src.Phone),
		school.CityID.Set(src.CityId),
		school.DistrictID.Set(src.DistrictId),
		school.Country.Set(src.Country.String()),
	)
	if err != nil {
		return nil, err
	}
	if src.Id != 0 {
		err = school.ID.Set(src.Id)
		if err != nil {
			return nil, err
		}
	}
	return school, nil
}

// UpdateSchools edit school for admin, school admin/staff
func (s *SchoolService) UpdateSchool(ctx context.Context, req *pb.UpdateSchoolRequest) (*pb.UpdateSchoolResponse, error) {
	currentID := interceptors.UserIDFromContext(ctx)

	currentUser, err := s.UserRepo.Get(ctx, s.DBTrace, database.Text(currentID))
	if err != nil {
		return nil, errors.Wrap(err, "s.UserRepo.GetProfile")
	}

	if currentUser.Group.String != constant.UserGroupAdmin {
		return nil, status.Error(codes.PermissionDenied, "only admin can update school information")
	}

	schoolEn, err := convertPbToSchoolEn(req.School)
	if err != nil {
		return nil, err
	}

	schools, err := s.SchoolRepo.Get(ctx, s.DBTrace, []int32{req.School.Id})
	if err != nil {
		return nil, errors.Wrap(err, "s.SchoolRepo.Get")
	}
	if len(schools) == 0 || schools[req.School.Id] == nil {
		return nil, status.Error(codes.InvalidArgument, "school not found")
	}

	_, err = s.SchoolRepo.Update(ctx, s.DBTrace, schoolEn)
	if err != nil {
		return nil, errors.Wrap(err, "s.SchoolRepo.Update")
	}

	return &pb.UpdateSchoolResponse{
		Successful: true,
	}, nil
}

func (s *SchoolService) handlePermissionToAddAndRemoveTeacherFromSchool(ctx context.Context, teacherID string, schoolID int32) (bool, error) {
	if teacherID == "" {
		return false, status.Error(codes.InvalidArgument, "missing teacher id")
	}

	if schoolID == 0 {
		return false, status.Error(codes.InvalidArgument, "missing school id")
	}

	schools, err := s.SchoolRepo.Get(ctx, s.DBTrace, []int32{schoolID})
	if err != nil {
		return false, fmt.Errorf("s.SchoolRepo.Get: %w", err)
	}
	if len(schools) == 0 || schools[schoolID] == nil {
		return false, status.Error(codes.NotFound, "cannot find school")
	}

	teacher, err := s.UserRepo.Get(ctx, s.DBTrace, database.Text(teacherID))
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return false, status.Error(codes.NotFound, "cannot find teacher")
		}
		return false, fmt.Errorf("s.UserRepo.GetProfile: %w", err)
	}

	if teacher.Group.String != constant.UserGroupTeacher {
		return false, status.Error(codes.InvalidArgument, "this is not a teacher")
	}

	isInSchool, err := s.TeacherRepo.IsInSchool(ctx, s.DBTrace, teacherID, schoolID)
	if err != nil {
		return false, fmt.Errorf("s.TeacherRepo.IsInSchool: %w", err)
	}

	return isInSchool, nil
}

func (s *SchoolService) addLogWhenAddAndRemoveTeacherFromSchool(ctx context.Context, actionType, adminID, teacherID string, schoolID int32) {
	logger := ctxzap.Extract(ctx)
	payload := map[string]interface{}{
		"school_id":  schoolID,
		"teacher_id": teacherID,
	}

	activityLog := &entities_bob.ActivityLog{}
	database.AllNullEntity(activityLog)
	// activityLog.Payload = payload
	err := multierr.Combine(
		activityLog.UserID.Set(adminID),
		activityLog.ActionType.Set(actionType),
		activityLog.Payload.Set(payload),
	)
	if err != nil {
		logger.Error("multierr.Combine", zap.Error(err))
	}

	if cerr := s.ActivityLogRepo.CreateV2(ctx, s.DBTrace, activityLog); cerr != nil {
		logger.Error("s.ActivityLogRepo.Create", zap.Error(cerr))
	}
}

func (s *SchoolService) RemoveTeacherFromSchool(ctx context.Context, req *pb.RemoveTeacherFromSchoolRequest) (*pb.RemoveTeacherFromSchoolResponse, error) {
	adminID := interceptors.UserIDFromContext(ctx)

	isInSchool, err := s.handlePermissionToAddAndRemoveTeacherFromSchool(ctx, req.TeacherId, req.SchoolId)
	if err != nil {
		return nil, err
	}
	if isInSchool == false {
		return nil, status.Error(codes.InvalidArgument, "the teacher is not affiliated with this school")
	}
	admin, err := s.UserRepo.Get(ctx, s.DBTrace, database.Text(adminID))
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find users")
		}
		return nil, fmt.Errorf("s.UserRepo.GetProfile: %w", err)
	}
	if admin.Group.String == pb.USER_GROUP_SCHOOL_ADMIN.String() {
		schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DBTrace, database.Text(adminID))
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return nil, status.Error(codes.NotFound, "cannot find school admin")
			}
			return nil, fmt.Errorf("s.SchoolAdminRepo.Get: %w", err)
		}
		if schoolAdmin.SchoolID.Int != req.SchoolId {
			return nil, status.Error(codes.PermissionDenied, "the teacher is not affiliated with your school")
		}
	}
	err = database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		classes, err := s.ClassRepo.FindBySchool(ctx, s.DBTrace, req.SchoolId)
		if err != nil {
			return fmt.Errorf("s.ClassRepo.FindBySchool: %w", err)
		}

		classIDs := []int32{}
		for _, v := range classes {
			classIDs = append(classIDs, v.ID.Int)
		}

		err = s.ClassMemberRepo.UpdateStatus(ctx, tx, []string{req.TeacherId}, classIDs, pb.USER_GROUP_TEACHER.String(), entities.ClassMemberStatusInactive)
		if err != nil {
			return fmt.Errorf("s.ClassMemberRepo.UpdateStatus: %w", err)
		}

		for _, classID := range classIDs {
			err = s.PublishClassEvt(ctx, &pb_bob.EvtClassRoom{
				Message: &pb_bob.EvtClassRoom_LeaveClass_{
					LeaveClass: &pb_bob.EvtClassRoom_LeaveClass{
						ClassId:  classID,
						UserIds:  []string{req.TeacherId},
						IsKicked: true,
					},
				},
			})
			if err != nil {
				ctxzap.Extract(ctx).Warn("rcv.PublishClassEvt", zap.Error(err))
			}
		}
		err = s.TeacherRepo.LeaveSchool(ctx, tx, req.TeacherId, req.SchoolId)
		if err != nil {
			return fmt.Errorf("s.TeacherRepo.LeaveSchool: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.addLogWhenAddAndRemoveTeacherFromSchool(ctx, entities_bob.LogActionTypeRemoveTeacherFromSchoolFromSchool, adminID, req.TeacherId, req.SchoolId)

	return &pb.RemoveTeacherFromSchoolResponse{}, nil
}

func (s *SchoolService) AddTeacher(ctx context.Context, req *pb.AddTeacherRequest) (*pb.AddTeacherResponse, error) {
	adminID := interceptors.UserIDFromContext(ctx)
	isInSchool, err := s.handlePermissionToAddAndRemoveTeacherFromSchool(ctx, req.TeacherId, req.SchoolId)
	if err != nil {
		return nil, err
	}
	if isInSchool {
		return nil, status.Error(codes.InvalidArgument, "the teacher already is part of this school")
	}

	err = s.TeacherRepo.JoinSchool(ctx, s.DBTrace, req.TeacherId, req.SchoolId)
	if err != nil {
		return nil, fmt.Errorf("s.TeacherRepo.JoinSchool: %w", err)
	}

	s.addLogWhenAddAndRemoveTeacherFromSchool(ctx, entities_bob.LogActionTypeAddTeacherToSchool, adminID, req.TeacherId, req.SchoolId)

	return &pb.AddTeacherResponse{}, nil
}

func (s *SchoolService) PublishClassEvt(ctx context.Context, msg *pb_bob.EvtClassRoom) error {
	var msgID string
	data, _ := msg.Marshal()

	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectClassUpserted, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishClassEvt rcv.JSM.PublishAsyncContext Class.Upserted failed, msgID: %s, %w", msgID, err))
	}

	return err
}

func (s *SchoolService) CreateSchoolConfig(ctx context.Context, req *pb.CreateSchoolConfigRequest) (*pb.CreateSchoolConfigResponse, error) {
	return &pb.CreateSchoolConfigResponse{}, nil
}

func (s *SchoolService) UpdateSchoolConfig(ctx context.Context, req *pb.UpdateSchoolConfigRequest) (*pb.UpdateSchoolConfigResponse, error) {
	return &pb.UpdateSchoolConfigResponse{}, nil
}
