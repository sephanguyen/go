package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	enigma_entities "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type UserRegistrationService struct {
	DB     database.Ext
	Logger *zap.Logger

	UserRepo interface {
		FindByIDUnscope(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.LegacyUser, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
	}
	StudentRepo interface {
		Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*entity.LegacyStudent, error)
		Create(context.Context, database.QueryExecer, *entity.LegacyStudent) error
		Update(ctx context.Context, db database.QueryExecer, s *entity.LegacyStudent) error
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error
	}
	TeacherRepo interface {
		SoftDeleteMultiple(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray) error
		Find(ctx context.Context, db database.QueryExecer, teacherID pgtype.Text) (*entity.Teacher, error)
		Update(ctx context.Context, db database.QueryExecer, teacher *entity.Teacher) error
		Create(ctx context.Context, db database.QueryExecer, teacher *entity.Teacher) error
	}
	StaffRepo interface {
		Find(ctx context.Context, db database.QueryExecer, staffID pgtype.Text) (*entity.Staff, error)
		Create(ctx context.Context, db database.QueryExecer, staff *entity.Staff) error
		Update(ctx context.Context, db database.QueryExecer, staff *entity.Staff) (*entity.Staff, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, staffIDs pgtype.TextArray) error
	}
	PartnerSyncDataLogService interface {
		UpdateLogStatus(ctx context.Context, id, status string) error
	}
	UserGroupV2Repo interface {
		FindUserGroupByRoleName(ctx context.Context, db database.QueryExecer, roleName string) (*entity.UserGroupV2, error)
	}
	UserGroupMemberRepo interface {
		UpsertBatch(ctx context.Context, db database.QueryExecer, userGroupsMembers []*entity.UserGroupMember) error
		SoftDelete(ctx context.Context, db database.QueryExecer, userGroupsMembers []*entity.UserGroupMember) error
	}
	LocationRepo interface {
		GetLocationOrg(ctx context.Context, db database.Ext, resourcePath string) (*domain.Location, error)
	}
	UserAccessPathRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userAccessPaths []*entity.UserAccessPath) error
		FindLocationIDsFromUserID(ctx context.Context, db database.QueryExecer, userID string) ([]string, error)
		Delete(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) error
	}
	StudentEnrollmentStatusHistoryRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, e *entity.StudentEnrollmentStatusHistory) error
		SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error
	}
}

func (u *UserRegistrationService) SyncStudentHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventUserRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncStudentHandler proto.Unmarshal: %w", err)
	}
	u.Logger.Info("UserRegistrationService.syncStudentHandler",
		zap.String("signature", req.Signature),
	)
	if err := u.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("UserRegistrationService.syncStudentHandler update log status to processing: %w", err)
	}
	if len(req.Students) == 0 {
		return false, nil
	}
	if err := nats.ChunkHandler(len(req.Students), constants.MaxRecordProcessPertime, func(start, end int) error {
		return u.syncStudent(ctx, req.Students[start:end])
	}); err != nil {
		return true, fmt.Errorf("syncStudentHandler err SyncStudent: %w", err)
	}
	if err := u.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("UserRegistrationService.syncStudentHandler update log status to success: %w", err)
	}

	return false, nil
}

func (u *UserRegistrationService) SyncStaffHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventUserRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncStaffHandler proto.Unmarshal: %w", err)
	}
	u.Logger.Info("UserRegistrationService.syncStaffHandler",
		zap.String("signature", req.Signature),
	)
	if err := u.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("UserRegistrationService.syncStaffHandler update log status to processing: %w", err)
	}
	if len(req.Staffs) == 0 {
		return false, fmt.Errorf("syncStaffHandler length of Staffs = 0")
	}
	if err := nats.ChunkHandler(len(req.Staffs), constants.MaxRecordProcessPertime, func(start, end int) error {
		return u.syncStaff(ctx, req.Staffs[start:end])
	}); err != nil {
		return true, fmt.Errorf("syncStaffHandler err syncStaff: %w", err)
	}
	if err := u.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("UserRegistrationService.syncStaffHandler update log status to success: %w", err)
	}

	return false, nil
}

func (u *UserRegistrationService) syncStudent(ctx context.Context, req []*npb.EventUserRegistration_Student) error {
	zapLogger := ctxzap.Extract(ctx)
	var errs error
	deleteIDs := []string{}

	studentUserGroup, err := u.UserGroupV2Repo.FindUserGroupByRoleName(ctx, u.DB, constant.RoleStudent)
	if err != nil {
		zapLogger.Sugar().Warn(fmt.Errorf("can not find student user group: %w", err))
	}

	for _, studentEvent := range req {
		switch studentEvent.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			if err := u.upsertStudent(ctx, studentEvent, studentUserGroup); err != nil {
				errs = multierr.Append(errs, fmt.Errorf("u.upsertStudent studentID %s: %w", studentEvent.StudentId, err))
			}

		case npb.ActionKind_ACTION_KIND_DELETED:
			deleteIDs = append(deleteIDs, studentEvent.StudentId)
		}
	}

	// soft delete student
	if err := u.deleteStudent(ctx, deleteIDs, studentUserGroup); err != nil {
		errs = multierr.Append(errs, fmt.Errorf("u.delete studentIDs %v: %w", deleteIDs, err))
	}

	return errs
}

func (u *UserRegistrationService) syncStaff(ctx context.Context, req []*npb.EventUserRegistration_Staff) error {
	zapLogger := ctxzap.Extract(ctx)
	teacherUserGroup, err := u.UserGroupV2Repo.FindUserGroupByRoleName(ctx, u.DB, constant.RoleTeacher)
	if err != nil {
		zapLogger.Sugar().Warn(fmt.Errorf("can not find teacher user group: %w", err))
	}

	var errs error
	deleteIDs := []string{}
	for _, staffEvent := range req {
		switch staffEvent.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			if err := u.upsertStaff(ctx, staffEvent, teacherUserGroup); err != nil {
				errs = multierr.Append(errs, fmt.Errorf("s.upsertStaff teacherID %s: %w", staffEvent.StaffId, err))
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			deleteIDs = append(deleteIDs, staffEvent.StaffId)
		}
	}

	if err := u.deleteStaff(ctx, deleteIDs, teacherUserGroup); err != nil {
		errs = multierr.Append(err, fmt.Errorf("u.deleteStaff teacherIDs %v: %w", deleteIDs, err))
	}

	return errs
}

func (u *UserRegistrationService) deleteStudent(ctx context.Context, studentIDs []string, studentUserGroup *entity.UserGroupV2) error {
	if len(studentIDs) == 0 {
		return nil
	}

	err := database.ExecInTx(ctx, u.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := u.StudentRepo.SoftDelete(ctx, tx, database.TextArray(studentIDs)); err != nil {
			return fmt.Errorf("u.StudentRepo.SoftDelete: %w", err)
		}

		if err := u.UserRepo.SoftDelete(ctx, tx, database.TextArray(studentIDs)); err != nil {
			return fmt.Errorf("u.UserRepo.SoftDelete: %w", err)
		}

		if err := u.StudentEnrollmentStatusHistoryRepo.SoftDelete(ctx, tx, database.TextArray(studentIDs)); err != nil {
			return fmt.Errorf("u.StudentEnrollmentStatusHistoryRepo.SoftDelete: %w", err)
		}

		if err := u.UserAccessPathRepo.Delete(ctx, tx, database.TextArray(studentIDs)); err != nil {
			return fmt.Errorf("u.UserAccessPathRepo.Delete: %w", err)
		}

		// revoke permission
		return u.revokeUserGroupMember(ctx, tx, studentIDs, studentUserGroup)
	})

	return err
}

func (u *UserRegistrationService) deleteStaff(ctx context.Context, staffIDs []string, teacherUserGroup *entity.UserGroupV2) error {
	if len(staffIDs) == 0 {
		return nil
	}

	err := database.ExecInTx(ctx, u.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := u.TeacherRepo.SoftDeleteMultiple(ctx, tx, database.TextArray(staffIDs)); err != nil {
			return fmt.Errorf("u.Teacher.SoftDeleteMultiple: %w", err)
		}

		if err := u.UserRepo.SoftDelete(ctx, tx, database.TextArray(staffIDs)); err != nil {
			return fmt.Errorf("u.UserRepo.SoftDelete: %w", err)
		}

		if err := u.StaffRepo.SoftDelete(ctx, tx, database.TextArray(staffIDs)); err != nil {
			return fmt.Errorf("u.StaffRepo.SoftDelete: %w", err)
		}

		// revoke permission
		return u.revokeUserGroupMember(ctx, tx, staffIDs, teacherUserGroup)
	})

	return err
}

func (u *UserRegistrationService) revokeUserGroupMember(ctx context.Context, db database.QueryExecer, userIDs []string, userGroup *entity.UserGroupV2) error {
	zapLogger := ctxzap.Extract(ctx)

	if userGroup != nil {
		userGroupMembersToRevoke, err := toUserGroupMemberEntity(userIDs, userGroup.UserGroupID.String)
		if err != nil {
			return fmt.Errorf("toUserGroupMemberEntity: %s", err.Error())
		}
		if err := u.UserGroupMemberRepo.SoftDelete(ctx, db, userGroupMembersToRevoke); err != nil {
			zapLogger.Sugar().Warn(fmt.Errorf("can not revoke student user group: %w", err))
		}
	}

	return nil
}

func (u *UserRegistrationService) upsertStudent(ctx context.Context, studentEvent *npb.EventUserRegistration_Student, studentUserGroup *entity.UserGroupV2) error {
	zapLogger := ctxzap.Extract(ctx)

	student, err := u.StudentRepo.Find(ctx, u.DB, database.Text(studentEvent.StudentId))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("err GetUser: %w", err)
	}

	err = database.ExecInTx(ctx, u.DB, func(ctx context.Context, tx pgx.Tx) error {
		// update student
		if student != nil {
			user, err := u.UserRepo.FindByIDUnscope(ctx, tx, student.ID)
			if err != nil {
				return fmt.Errorf("err FindUser: %w", err)
			}

			student.LegacyUser = *user
			if err := reqToUpdateStudentEntity(studentEvent, student); err != nil {
				return fmt.Errorf("reqToUpdateStudentEntity: %s", err.Error())
			}

			if err := u.StudentRepo.Update(ctx, tx, student); err != nil {
				return fmt.Errorf("updateStudent: %s", err.Error())
			}
		} else {
			// create new student
			student, err = reqToCreateStudentEntity(studentEvent)
			if err != nil {
				return fmt.Errorf("reqToCreateStudentEntity: %s", err.Error())
			}
			if err := u.StudentRepo.Create(ctx, tx, student); err != nil {
				return fmt.Errorf("createStudent: %s", err.Error())
			}
			if err := u.createUserAccessPath(ctx, tx, student.ID.String); err != nil {
				return fmt.Errorf("u.createUserAccessPath: %w", err)
			}
			if err := u.createStudentEnrollmentStatusHistory(ctx, tx, student); err != nil {
				return fmt.Errorf("u.createStudentEnrollmentStatusHistory: %w", err)
			}
		}

		// upsert assign user_group for student
		if studentUserGroup != nil {
			userGroupMember, err := toUserGroupMemberEntity([]string{student.ID.String}, studentUserGroup.UserGroupID.String)
			if err != nil {
				return fmt.Errorf("toUserGroupMemberEntity: %s", err.Error())
			}
			if err := u.UserGroupMemberRepo.UpsertBatch(ctx, tx, userGroupMember); err != nil {
				zapLogger.Sugar().Warn(fmt.Errorf("can not assign student user group to user %s: %s", student.GetUID(), err.Error()))
			}
		}

		return nil
	})

	return err
}

func (u *UserRegistrationService) upsertStaff(ctx context.Context, req *npb.EventUserRegistration_Staff, teacherUserGroup *entity.UserGroupV2) error {
	zapLogger := ctxzap.Extract(ctx)
	staff, err := u.StaffRepo.Find(ctx, u.DB, database.Text(req.StaffId))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("u.StaffRepo.Find: %w", err)
	}

	err = database.ExecInTx(ctx, u.DB, func(ctx context.Context, tx pgx.Tx) error {
		if staff != nil {
			user, err := u.UserRepo.FindByIDUnscope(ctx, tx, staff.ID)
			if err != nil {
				return fmt.Errorf("err FindUser: %w", err)
			}

			// update staff and user entity
			staff.LegacyUser = *user
			if err := reqToUpdateStaffEntity(req, staff); err != nil {
				return fmt.Errorf("reqToUpdateTeacherEntity: %s", err.Error())
			}

			if _, err := u.StaffRepo.Update(ctx, tx, staff); err != nil {
				return fmt.Errorf("u.StaffRepo.Update: %w", err)
			}
		} else {
			staff, err = reqToCreateStaffEntity(req)
			if err != nil {
				return fmt.Errorf("reqToCreateStaffEntity: %s", err.Error())
			}
			teacher, err := reqToCreateTeacherEntity(req)
			if err != nil {
				return fmt.Errorf("reqToCreateTeacherEntity: %s", err.Error())
			}

			if err := u.StaffRepo.Create(ctx, tx, staff); err != nil {
				return fmt.Errorf("u.StaffRepo.Create: %w", err)
			}
			if err := u.TeacherRepo.Create(ctx, tx, teacher); err != nil {
				return fmt.Errorf("u.TeacherRepo.Create: %w", err)
			}
			if err := u.createUserAccessPath(ctx, tx, staff.ID.String); err != nil {
				return fmt.Errorf("u.createUserAccessPath: %w", err)
			}
		}

		// upsert assign user_group for teacher
		if teacherUserGroup != nil {
			userGroupMember, err := toUserGroupMemberEntity([]string{staff.ID.String}, teacherUserGroup.UserGroupID.String)
			if err != nil {
				return fmt.Errorf("toUserGroupMemberEntity: %s", err.Error())
			}
			if err := u.UserGroupMemberRepo.UpsertBatch(ctx, tx, userGroupMember); err != nil {
				zapLogger.Sugar().Warn(fmt.Errorf("can not assign teacher user group to user %s: %s", staff.GetUID(), err.Error()))
			}
		}

		return nil
	})

	return err
}

func reqToCreateStudentEntity(studentEvent *npb.EventUserRegistration_Student) (*entity.LegacyStudent, error) {
	student := &entity.LegacyStudent{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)

	if studentEvent.StudentId == "" {
		studentEvent.StudentId = idutil.ULIDNow()
	}
	additionalData := &entity.StudentAdditionalData{
		JprefDivs: studentEvent.StudentDivs,
	}

	err := multierr.Combine(
		student.ID.Set(studentEvent.StudentId),
		student.GivenName.Set(studentEvent.GivenName),
		student.LastName.Set(studentEvent.LastName),
		student.FirstName.Set(studentEvent.GivenName),
		student.FullName.Set(CombineFirstNameAndLastNameToFullName(studentEvent.GivenName, studentEvent.LastName)),
		student.PhoneNumber.Set(studentEvent.StudentId), // to by pass not null contraint since JPREF does not send phoneNumber
		student.Country.Set(cpb.Country_COUNTRY_JP.String()),
		student.AdditionalData.Set(additionalData),
		student.SchoolID.Set(constants.JPREPSchool),
		student.ResourcePath.Set(fmt.Sprint(constants.JPREPSchool)),
		student.UserRole.Set(constant.UserRoleStudent),
		// student.CurrentGrade.Set()
	)

	return student, err
}

func reqToUpdateStudentEntity(studentEvent *npb.EventUserRegistration_Student, student *entity.LegacyStudent) error {
	err := multierr.Combine(
		student.GivenName.Set(studentEvent.GivenName),
		student.LastName.Set(studentEvent.LastName),
		student.FirstName.Set(studentEvent.GivenName),
		student.FullName.Set(CombineFirstNameAndLastNameToFullName(studentEvent.GivenName, studentEvent.LastName)),
		student.DeletedAt.Set(nil),
		student.LegacyUser.DeletedAt.Set(nil),
	)
	return err
}

func reqToCreateTeacherEntity(teacherEvent *npb.EventUserRegistration_Staff) (*entity.Teacher, error) {
	teacher := &entity.Teacher{}
	database.AllNullEntity(teacher)
	database.AllNullEntity(&teacher.LegacyUser)

	err := multierr.Combine(
		teacher.ID.Set(teacherEvent.StaffId),
		teacher.SchoolIDs.Set([]int32{constants.JPREPSchool}),
		teacher.ResourcePath.Set(fmt.Sprint(constants.JPREPSchool)),
	)

	return teacher, err
}

func reqToCreateStaffEntity(teacherEvent *npb.EventUserRegistration_Staff) (*entity.Staff, error) {
	staff := &entity.Staff{}
	database.AllNullEntity(staff)
	database.AllNullEntity(&staff.LegacyUser)

	err := multierr.Combine(
		staff.ID.Set(teacherEvent.StaffId),
		staff.ResourcePath.Set(fmt.Sprint(constants.JPREPSchool)),
		staff.WorkingStatus.Set("AVAILABLE"),

		staff.LegacyUser.Group.Set(entity.UserGroupTeacher),
		staff.LegacyUser.GivenName.Set(teacherEvent.Name),
		staff.LegacyUser.LastName.Set(teacherEvent.Name),
		staff.LegacyUser.FirstName.Set(teacherEvent.Name),
		staff.LegacyUser.FullName.Set(teacherEvent.Name),       // legacy API
		staff.LegacyUser.PhoneNumber.Set(teacherEvent.StaffId), // to by pass not null contraint since JPREP does not send phoneNumber
		staff.LegacyUser.Country.Set(cpb.Country_COUNTRY_JP.String()),
		staff.LegacyUser.ResourcePath.Set(fmt.Sprint(constants.JPREPSchool)),
		staff.LegacyUser.UserRole.Set(constant.UserRoleStaff),
	)

	return staff, err
}

func reqToUpdateStaffEntity(teacherEvent *npb.EventUserRegistration_Staff, staff *entity.Staff) error {
	return multierr.Combine(
		staff.LegacyUser.GivenName.Set(teacherEvent.Name),
		staff.LegacyUser.LastName.Set(teacherEvent.Name),
		staff.LegacyUser.FirstName.Set(teacherEvent.Name),
		staff.LegacyUser.FullName.Set(teacherEvent.Name), // legacy API
		staff.LegacyUser.DeletedAt.Set(nil),

		staff.DeletedAt.Set(nil),
	)
}

func toUserGroupMemberEntity(userIDs []string, userGroupID string) (entity.UserGroupMembers, error) {
	userGroupMembers := entity.UserGroupMembers{}
	for _, userID := range userIDs {
		userGroupMember := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMember)

		err := multierr.Combine(
			userGroupMember.UserID.Set(userID),
			userGroupMember.UserGroupID.Set(userGroupID),
			userGroupMember.ResourcePath.Set(fmt.Sprint(constants.JPREPSchool)),
		)
		if err != nil {
			return nil, err
		}

		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, nil
}

// create user_access_path for student/teacher
func (u *UserRegistrationService) createUserAccessPath(ctx context.Context, db database.Ext, userID string) error {
	resourcePath, err := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return fmt.Errorf("resource path is invalid")
	}

	orgLocation, err := u.LocationRepo.GetLocationOrg(ctx, db, fmt.Sprint(resourcePath))
	if err != nil {
		return fmt.Errorf("s.LocationRepo.GetLocationOrg: %w", err)
	}

	// create user_resource_path (add locations for user)
	if err := UpsertUserAccessPath(ctx, u.UserAccessPathRepo, db, []*domain.Location{orgLocation}, userID); err != nil {
		return fmt.Errorf("UpsertUserAccessPath: %w", err)
	}

	return nil
}

func toStudentEnrollmentStatusHistoryEntity(student *entity.LegacyStudent) (*entity.StudentEnrollmentStatusHistory, error) {
	enrollmentStatus := &entity.StudentEnrollmentStatusHistory{}
	database.AllNullEntity(enrollmentStatus)

	err := multierr.Combine(
		enrollmentStatus.StudentID.Set(student.ID),
		enrollmentStatus.LocationID.Set(constants.JPREPOrgLocation),
		enrollmentStatus.EnrollmentStatus.Set(entity.StudentEnrollmentStatusEnrolled),
		enrollmentStatus.StartDate.Set(student.CreatedAt),
	)

	return enrollmentStatus, err
}

func (u *UserRegistrationService) createStudentEnrollmentStatusHistory(ctx context.Context, db database.Ext, student *entity.LegacyStudent) error {
	enrollmentStatus, err := toStudentEnrollmentStatusHistoryEntity(student)
	if err != nil {
		return fmt.Errorf("cannot convert to StudentEnrollmentStatusHistoryEntity: %v", err)
	}

	if err := u.StudentEnrollmentStatusHistoryRepo.Upsert(ctx, db, enrollmentStatus); err != nil {
		return fmt.Errorf("u.StudentEnrollmentStatusHistoryRepo.Upsert: %v", err)
	}

	return nil
}
