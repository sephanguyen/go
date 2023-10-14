package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *UserModifierService) UpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) (*pb.UpdateStudentResponse, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	// validate request
	if err := s.validUpdateStudentRequest(ctx, req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tags, err := s.DomainTagRepo.GetByIDs(ctx, s.DB, req.StudentProfile.TagIds)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "DomainTagRepo.GetByIDs").Error())
	}
	if err := validUserTags(constant.RoleStudent, req.GetStudentProfile().GetTagIds(), tags); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	locations, err := s.GetLocations(ctx, req.StudentProfile.LocationIds)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	toUpdateStudent, err := studentPbInUpdateStudentRequestToStudentEnt(req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// validate student
	existingStudent, err := s.StudentRepo.Find(ctx, s.DB, toUpdateStudent.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "student id is not exists")
	}

	existingUser, err := s.UserRepo.Get(ctx, s.DB, toUpdateStudent.ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user id is not exists")
	}
	existingStudent.LegacyUser = *existingUser

	gradeMaster, err := s.getGradeMaster(ctx, req.StudentProfile.GradeId)
	if err != nil {
		return nil, err
	}
	if len(gradeMaster) > 0 {
		for k, v := range gradeMaster {
			currentGrade := existingStudent.CurrentGrade.Int
			gradeID := k.GradeID().String()
			if !(field.IsNull(v) && field.IsUndefined(v)) {
				currentGrade = int16(v.Int32())
			}
			if err := multierr.Combine(
				existingStudent.CurrentGrade.Set(currentGrade),
				existingStudent.GradeID.Set(gradeID),
			); err != nil {
				return nil, status.Errorf(codes.Internal, "multierr.Combine err: %v", err)
			}
		}
	}

	var studentPB *pb.UpdateStudentResponse_StudentProfile
	var studentPhoneNumberPB *pb.StudentPhoneNumber
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// if student email edited
		if toUpdateStudent.Email != existingStudent.Email {
			// Check if edited email already exists
			emailExistingStudents, err := s.UserRepo.GetByEmail(ctx, tx, database.TextArray([]string{toUpdateStudent.Email.String}))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", err).Error())
			}
			if len(emailExistingStudents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot edit student with email existing in system: %s", toUpdateStudent.Email.String))
			}
			existingStudent.LegacyUser.Email = toUpdateStudent.LegacyUser.Email

			err = s.UsrEmailRepo.UpdateEmail(ctx, tx, toUpdateStudent.ID, database.Text(strconv.Itoa(int(resourcePath))), toUpdateStudent.Email)
			switch err {
			case nil:
				break
			case repository.ErrNoRowAffected:
				// it's ok to have no row affected
				break
			default:
				return status.Error(codes.Internal, errors.Wrap(err, "s.UserModifierService.UsrEmailRepo.UpdateEmail").Error())
			}

			/*//Use centralized func to update email
			repo := userservice.RepoForDomainUser{
				DomainUserRepo:     s.DomainUserRepo,
				DomainUsrEmailRepo: s.DomainUsrEmailRepo,
				UserRepo:           s.UserRepo,
				OrganizationRepo:   s.OrganizationRepo,
			}
			userservice.UpdateEmail(ctx, tx, s.TenantManager, s.DB, repo, )*/

			// Import to identity platform
			tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, s.DB, strconv.FormatInt(resourcePath, 10))
			if err != nil {
				zapLogger.Error(
					"cannot get tenant id",
					zap.Error(err),
					zap.Int64("organizationID", resourcePath),
				)
				switch err {
				case pgx.ErrNoRows:
					return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: strconv.FormatInt(resourcePath, 10)}.Error())
				default:
					return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
				}
			}

			err = s.UpdateUserEmailInIdentityPlatform(ctx, tenantID, existingStudent.ID.String, toUpdateStudent.LegacyUser.Email.String)
			if err != nil {
				zapLogger.Error(
					"cannot update users on identity platform",
					zap.Error(err),
					zap.Int64("organizationID", resourcePath),
					zap.String("tenantID", tenantID),
					zap.String("uid", existingStudent.ID.String),
					zap.String("email", existingStudent.Email.String),
					zap.String("emailToUpdate", toUpdateStudent.Email.String),
				)
				switch err {
				case user.ErrUserNotFound:
					return status.Error(codes.NotFound, errcode.NewUserNotFoundErr(existingStudent.ID.String).Error())
				default:
					return status.Error(codes.Internal, err.Error())
				}
			}
		}

		assignParameterToUpdateStudent(existingStudent, req.StudentProfile.StudentPhoneNumber, req.StudentProfile.StudentPhoneNumbers, toUpdateStudent)
		if err := s.StudentRepo.Update(ctx, tx, existingStudent); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.StudentRepo.Update: %w", err).Error())
		}

		studentPB = studentToStudentPBInUpdateStudentResponse(existingStudent)
		if err := UpsertUserAccessPath(ctx, s.UserAccessPathRepo, tx, locations, existingStudent.GetUID()); err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "UpsertUserAccessPath").Error())
		}

		// Upsert school_history
		if err := s.validateSchoolHistoriesReq(ctx, req.SchoolHistories); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("validateSchoolHistoriesReq: %v", err).Error())
		}
		schoolHistories, err := schoolHistoryPbToStudentSchoolHistory(req.SchoolHistories, existingStudent.ID.String, fmt.Sprint(resourcePath))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("schoolHistoryPbToStudentSchoolHistory: %v", err).Error())
		}
		if err := s.SchoolHistoryRepo.SoftDeleteByStudentIDs(ctx, tx, database.TextArray([]string{existingStudent.ID.String})); err != nil {
			return errorx.ToStatusError(err)
		}
		if err := s.SchoolHistoryRepo.Upsert(ctx, tx, schoolHistories); err != nil {
			return errorx.ToStatusError(err)
		}

		// Upsert user_address
		if err := s.validateUserAddressesReq(ctx, req.UserAddresses); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("validateUserAddressesReq: %v", err).Error())
		}
		homeAddresses, err := userAddressPbToStudentHomeAddress(req.UserAddresses, existingStudent.LegacyUser.ID.String, fmt.Sprint(resourcePath))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("userAddressPbToStudentHomeAddress: %v", err).Error())
		}
		if err := s.UserAddressRepo.SoftDeleteByUserIDs(ctx, tx, database.TextArray([]string{existingStudent.LegacyUser.ID.String})); err != nil {
			return errorx.ToStatusError(err)
		}
		if err := s.UserAddressRepo.Upsert(ctx, tx, homeAddresses); err != nil {
			return errorx.ToStatusError(err)
		}
		if req.StudentProfile.StudentPhoneNumbers != nil {
			if err := s.UserPhoneNumberRepo.SoftDeleteByUserIDs(ctx, tx, database.TextArray([]string{existingStudent.ID.String})); err != nil {
				return errorx.ToStatusError(err)
			}
			userPhoneNumbers, phoneNumber, homePhoneNumber, err := updateUserPhoneNumbersPbToStudentPhoneNumbers(req.StudentProfile.StudentPhoneNumbers.StudentPhoneNumber, existingStudent.ID.String, fmt.Sprint(resourcePath))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("updateUserPhoneNumbersPbToStudentPhoneNumbers: %v", err).Error())
			}
			if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
				return errorx.ToStatusError(err)
			}
			studentPB.StudentPhoneNumber = &pb.StudentPhoneNumber{
				PhoneNumber:       phoneNumber,
				HomePhoneNumber:   homePhoneNumber,
				ContactPreference: req.StudentProfile.StudentPhoneNumbers.ContactPreference,
			}
		}
		// to keep backward compatible we still allow FE send old block and new block, but priority new block,
		// so we have to check line bellow
		if req.StudentProfile.StudentPhoneNumber != nil && req.StudentProfile.StudentPhoneNumbers == nil {
			if err := validateStudentPhoneNumber(req.StudentProfile.StudentPhoneNumber); err != nil {
				return status.Error(codes.InvalidArgument, fmt.Errorf("validateStudentPhoneNumber: %v", err).Error())
			}
			if err := s.UserPhoneNumberRepo.SoftDeleteByUserIDs(ctx, tx, database.TextArray([]string{existingStudent.ID.String})); err != nil {
				return errorx.ToStatusError(err)
			}
			userPhoneNumbers, err := userPhoneNumbersPbToStudentPhoneNumbers(req.StudentProfile.StudentPhoneNumber, existingStudent.ID.String, fmt.Sprint(resourcePath))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("userPhoneNumbersPbToStudentPhoneNumbers: %v", err).Error())
			}
			if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
				return errorx.ToStatusError(err)
			}
			studentPhoneNumberPB = &pb.StudentPhoneNumber{
				PhoneNumber:       req.StudentProfile.StudentPhoneNumber.PhoneNumber,
				HomePhoneNumber:   req.StudentProfile.StudentPhoneNumber.HomePhoneNumber,
				ContactPreference: req.StudentProfile.StudentPhoneNumber.ContactPreference,
			}
			studentPB.StudentPhoneNumber = studentPhoneNumberPB
		}

		taggedUsers, err := s.DomainTaggedUserRepo.GetByUserIDs(ctx, tx, []string{req.StudentProfile.GetId()})
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("DomainTaggedUserRepo.GetByUserIDs: %v", err).Error())
		}

		domainUser := grpc.NewUpdateStudentRequest(req)
		userWithTags := map[entity.User][]entity.DomainTag{domainUser: tags}
		if err := s.UpsertTaggedUsers(ctx, tx, userWithTags, taggedUsers); err != nil {
			return errors.Wrap(err, "UpsertTaggedUsers")
		}

		if err := s.StudentParentRepo.UpsertParentAccessPathByStudentIDs(ctx, tx, []string{existingStudent.ID.String}); err != nil {
			return err
		}

		studentPB.LocationIds = req.StudentProfile.LocationIds
		userEvents := newUpdateStudentEvents(locations, existingStudent)
		if err = s.publishUserEvent(ctx, constants.SubjectUserUpdated, userEvents...); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(req.SchoolHistories) != 0 {
		currentSchools, err := s.SchoolHistoryRepo.GetCurrentSchoolByStudentID(ctx, s.DB, database.Text(toUpdateStudent.ID.String))
		if err != nil {
			return nil, err
		}
		if len(currentSchools) != 0 {
			err = s.SchoolHistoryRepo.UnsetCurrentSchoolByStudentID(ctx, s.DB, database.Text(toUpdateStudent.ID.String))
			if err != nil {
				return nil, err
			}
		}
		currentSchoolsFromRequest, err := s.SchoolHistoryRepo.GetSchoolHistoriesByGradeIDAndStudentID(ctx, s.DB, database.Text(req.StudentProfile.GradeId), database.Text(toUpdateStudent.ID.String), database.Bool(false))
		if err != nil {
			return nil, err
		}
		if len(currentSchoolsFromRequest) != 0 {
			schoolIDs := []string{}
			for _, currentSchool := range currentSchoolsFromRequest {
				for _, schoolHistory := range req.SchoolHistories {
					if currentSchool.SchoolID.String == schoolHistory.SchoolId {
						schoolIDs = append(schoolIDs, currentSchool.SchoolID.String)
						continue
					}
				}
			}
			if len(schoolIDs) != 0 {
				err = s.SchoolHistoryRepo.SetCurrentSchoolByStudentIDAndSchoolID(ctx, s.DB, database.Text(schoolIDs[0]), database.Text(toUpdateStudent.ID.String))
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		err = s.SchoolHistoryRepo.RemoveCurrentSchoolByStudentID(ctx, s.DB, database.Text(toUpdateStudent.ID.String))
		if err != nil {
			return nil, err
		}
	}

	_ = s.publishAsyncUserDeviceToken(ctx, existingStudent, locations)

	return &pb.UpdateStudentResponse{
		StudentProfile: studentPB,
	}, nil
}

func (s *UserModifierService) validUpdateStudentRequest(ctx context.Context, req *pb.UpdateStudentRequest) error {
	switch {
	case req.StudentProfile.Id == "":
		return errors.New("student id cannot be empty")
	case req.StudentProfile.Email == "":
		return errors.New("student email cannot be empty")
	case req.StudentProfile.Name == "" && req.StudentProfile.FirstName == "" && req.StudentProfile.LastName == "":
		return errors.New("student name cannot be empty")
	case req.StudentProfile.Name == "" && req.StudentProfile.FirstName == "":
		return errors.New("student first name cannot be empty")
	case req.StudentProfile.Name == "" && req.StudentProfile.LastName == "":
		return errors.New("student last name cannot be empty")
	case len(req.StudentProfile.LocationIds) < 1:
		return errors.New("student location length must be at least 1")
	}

	if _, ok := pb.StudentEnrollmentStatus_value[req.StudentProfile.EnrollmentStatus.String()]; !ok {
		return ErrStudentEnrollmentStatusUnknown
	}
	if req.StudentProfile.EnrollmentStatus == pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
		return ErrStudentEnrollmentStatusNotAllowedTobeNone
	}

	if err := s.validateLocationsForUpdateStudent(ctx, req); err != nil {
		return err
	}

	if req.StudentProfile.StudentPhoneNumbers != nil {
		if err := validateStudentPhoneNumbersUpdateStudent(req.StudentProfile.StudentPhoneNumbers.StudentPhoneNumber); err != nil {
			return err
		}
	}

	return nil
}

func studentPbInUpdateStudentRequestToStudentEnt(req *pb.UpdateStudentRequest) (*entity.LegacyStudent, error) {
	student := new(entity.LegacyStudent)
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)
	enrollmentStatus := req.StudentProfile.EnrollmentStatus.String()
	if req.StudentProfile.EnrollmentStatusStr != "" {
		enrollmentStatus = req.StudentProfile.EnrollmentStatusStr
	}
	if err := multierr.Combine(
		student.ID.Set(req.StudentProfile.Id),
		student.FullName.Set(req.StudentProfile.Name),
		student.FirstName.Set(req.StudentProfile.FirstName),
		student.LastName.Set(req.StudentProfile.LastName),
		student.FirstNamePhonetic.Set(req.StudentProfile.FirstNamePhonetic),
		student.LastNamePhonetic.Set(req.StudentProfile.LastNamePhonetic),
		student.CurrentGrade.Set(req.StudentProfile.Grade),
		student.EnrollmentStatus.Set(enrollmentStatus),
		student.StudentNote.Set(req.StudentProfile.StudentNote),
		student.Email.Set(req.StudentProfile.Email),
		student.UserName.Set(req.StudentProfile.Email),
		student.LoginEmail.Set(req.StudentProfile.Email),
	); err != nil {
		return nil, err
	}

	if req.StudentProfile.FirstName != "" && req.StudentProfile.LastName != "" {
		if err := multierr.Combine(
			student.LegacyUser.FullName.Set(CombineFirstNameAndLastNameToFullName(req.StudentProfile.FirstName, req.StudentProfile.LastName)),
			student.LegacyUser.FirstName.Set(req.StudentProfile.FirstName),
			student.LegacyUser.LastName.Set(req.StudentProfile.LastName),
		); err != nil {
			return nil, err
		}
	}

	fullNamePhonetic := CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.StudentProfile.FirstNamePhonetic, req.StudentProfile.LastNamePhonetic)
	if fullNamePhonetic != "" {
		if err := student.LegacyUser.FullNamePhonetic.Set(fullNamePhonetic); err != nil {
			return nil, err
		}
	}

	if err := req.StudentProfile.Birthday.CheckValid(); err == nil {
		_ = student.LegacyUser.Birthday.Set(req.StudentProfile.Birthday.AsTime())
	}
	if req.StudentProfile.Gender != pb.Gender_NONE {
		_ = student.LegacyUser.Gender.Set(req.StudentProfile.Gender.String())
	}

	if req.StudentProfile.StudentExternalId != "" {
		if err := student.StudentExternalID.Set(req.StudentProfile.StudentExternalId); err != nil {
			return nil, err
		}
	}
	return student, nil
}

func studentToStudentPBInUpdateStudentResponse(student *entity.LegacyStudent) *pb.UpdateStudentResponse_StudentProfile {
	studentPB := &pb.UpdateStudentResponse_StudentProfile{
		Id:                student.ID.String,
		Name:              student.GetName(),
		FirstName:         student.FirstName.String,
		LastName:          student.LastName.String,
		FirstNamePhonetic: student.FirstNamePhonetic.String,
		LastNamePhonetic:  student.LastNamePhonetic.String,
		FullNamePhonetic:  student.FullNamePhonetic.String,
		Grade:             int32(student.CurrentGrade.Int),
		EnrollmentStatus:  pb.StudentEnrollmentStatus(student.EnrollmentStatus.Status),
		StudentExternalId: student.StudentExternalID.String,
		StudentNote:       student.StudentNote.String,
		Email:             student.Email.String,
		GradeId:           student.GradeID.String,
	}
	if student.Birthday.Status != pgtype.Null {
		studentPB.Birthday = timestamppb.New(student.Birthday.Time)
	}
	if student.Gender.Status != pgtype.Null {
		studentPB.Gender = pb.Gender(pb.Gender_value[student.Gender.String])
	}

	return studentPB
}

func assignParameterToUpdateStudent(existingStudent *entity.LegacyStudent, studentPhoneNumberReq *pb.StudentPhoneNumber, updateStudentPhoneNumberReq *pb.UpdateStudentPhoneNumber, toUpdateStudent *entity.LegacyStudent) {
	existingStudent.LegacyUser.FullName = toUpdateStudent.LegacyUser.FullName
	existingStudent.LegacyUser.FirstName = toUpdateStudent.LegacyUser.FirstName
	existingStudent.LegacyUser.LastName = toUpdateStudent.LegacyUser.LastName
	existingStudent.LegacyUser.FirstNamePhonetic = toUpdateStudent.LegacyUser.FirstNamePhonetic
	existingStudent.LegacyUser.LastNamePhonetic = toUpdateStudent.LegacyUser.LastNamePhonetic
	existingStudent.LegacyUser.FullNamePhonetic = toUpdateStudent.LegacyUser.FullNamePhonetic
	existingStudent.CurrentGrade = toUpdateStudent.CurrentGrade
	existingStudent.EnrollmentStatus = toUpdateStudent.EnrollmentStatus
	existingStudent.StudentExternalID = toUpdateStudent.StudentExternalID
	existingStudent.StudentNote = toUpdateStudent.StudentNote
	existingStudent.LegacyUser.Birthday = toUpdateStudent.Birthday
	existingStudent.LegacyUser.Gender = toUpdateStudent.Gender

	if studentPhoneNumberReq != nil {
		existingStudent.ContactPreference = database.Text(studentPhoneNumberReq.ContactPreference.String())

		if studentPhoneNumberReq.PhoneNumber != "" {
			existingStudent.PhoneNumber = database.Text(studentPhoneNumberReq.PhoneNumber)
		}
	}
	if updateStudentPhoneNumberReq != nil {
		existingStudent.ContactPreference = database.Text(updateStudentPhoneNumberReq.ContactPreference.String())
		for _, studentPhoneNumber := range updateStudentPhoneNumberReq.StudentPhoneNumber {
			if studentPhoneNumber.PhoneNumberType == pb.StudentPhoneNumberType_PHONE_NUMBER {
				existingStudent.PhoneNumber = database.Text(studentPhoneNumber.PhoneNumber)
			}
		}
	}
}

func (s *UserModifierService) publishAsyncUserDeviceToken(ctx context.Context, student *entity.LegacyStudent, locations []*domain.Location) error {
	locationIDs := []string{}
	for _, location := range locations {
		locationIDs = append(locationIDs, location.LocationID)
	}

	data := &pb.EvtUserInfo{
		UserId:      student.ID.String,
		Name:        student.GetName(),
		LocationIds: locationIDs,
	}

	msg, err := proto.Marshal(data)
	if err != nil {
		return fmt.Errorf("publishAsyncUserDeviceToken: proto.Marshal: %w", err)
	}

	_, err = s.JSM.TracedPublish(ctx, "publishAsyncUserDeviceToken", constants.SubjectUserDeviceTokenUpdated, msg)
	if err != nil {
		return fmt.Errorf("publishAsyncUserDeviceToken: s.JSM.TracedPublish: %w", err)
	}

	return nil
}

func (s *UserModifierService) UpdateUserEmailInIdentityPlatform(ctx context.Context, tenantID string, userID string, email string) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	tenantClient, err := s.TenantManager.TenantClient(ctx, tenantID)
	if err != nil {
		zapLogger.Warnw(
			"cannot get tenant client",
			"tenantID", tenantID,
			"err", err.Error(),
		)
		return errors.Wrap(err, "TenantClient")
	}

	return UpdateUserEmail(ctx, tenantClient, userID, email)
}

func newUpdateStudentEvents(locations []*domain.Location, students ...*entity.LegacyStudent) []*pb.EvtUser {
	updateStudentEvents := make([]*pb.EvtUser, 0, len(students))
	locationIDs := []string{}
	for _, location := range locations {
		locationIDs = append(locationIDs, location.LocationID)
	}

	for _, student := range students {
		updateStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_UpdateStudent_{
				UpdateStudent: &pb.EvtUser_UpdateStudent{
					StudentId:                student.ID.String,
					DeviceToken:              student.DeviceToken.String,
					AllowNotification:        student.AllowNotification.Bool,
					Name:                     student.GetName(),
					StudentFirstName:         student.FirstName.String,
					StudentLastName:          student.LastName.String,
					StudentFirstNamePhonetic: student.FirstNamePhonetic.String,
					StudentLastNamePhonetic:  student.LastNamePhonetic.String,
					LocationIds:              locationIDs,
				},
			},
		}
		updateStudentEvents = append(updateStudentEvents, updateStudentEvent)
	}
	return updateStudentEvents
}

func (s *UserModifierService) validateLocationsForUpdateStudent(ctx context.Context, req *pb.UpdateStudentRequest) error {
	dateFormat := "2006-01-02"
	listStudentPackages, err := s.FatimaClient.ListStudentPackage(signCtx(ctx), &fpb.ListStudentPackageRequest{StudentIds: []string{req.StudentProfile.Id}})
	if err != nil {
		return status.Error(codes.Internal, fmt.Errorf("validateLocationsForUpdateStudent: %w", err).Error())
	}

	for _, studentPackage := range listStudentPackages.StudentPackages {
		if studentPackage.EndAt.AsTime().Format(dateFormat) >= time.Now().Format(dateFormat) {
			for _, id := range studentPackage.LocationIds {
				if !golibs.InArrayString(id, req.StudentProfile.LocationIds) {
					return fmt.Errorf(constant.InvalidLocations)
				}
			}
		}
	}

	return nil
}
