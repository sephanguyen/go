package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gocarina/gocsv"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	nameParentCSVHeader                 = "name"
	userNameParentCSVHeader             = "username"
	emailParentCSVHeader                = "email"
	phoneNumberParentCSVHeader          = "phone_number"
	studentEmailsParentCSVHeader        = "student_email"
	relationshipsParentCSVHeader        = "relationship"
	primaryPhoneNumberParentCSVHeader   = "primary_phone_number"
	secondaryPhoneNumberParentCSVHeader = "secondary_phone_number"
	remarksParentCSVHeader              = "remarks"
	tagParentCSVHeader                  = "parent_tag"
	firstNameParentCSVHeader            = "first_name"
	lastNameParentCSVHeader             = "last_name"
	externalUserIDParentCSVHeader       = "external_user_id"
)

type ImportParentCSVField struct {
	ExternalUserID       *CsvField `csv:"external_user_id"`
	Name                 *CsvField `csv:"name"`
	UserName             *CsvField `csv:"username"`
	Email                *CsvField `csv:"email"`
	PhoneNumber          *CsvField `csv:"phone_number"`
	StudentEmail         *CsvField `csv:"student_email"`
	Relationship         *CsvField `csv:"relationship"`
	PrimaryPhoneNumber   *CsvField `csv:"primary_phone_number"`
	SecondaryPhoneNumber *CsvField `csv:"secondary_phone_number"`
	Remarks              *CsvField `csv:"remarks"`
	ParentTag            *CsvField `csv:"parent_tag"`
	FirstName            *CsvField `csv:"first_name"`
	LastName             *CsvField `csv:"last_name"`
	FirstNamePhonetic    *CsvField `csv:"first_name_phonetic"`
	LastNamePhonetic     *CsvField `csv:"last_name_phonetic"`
}

type ParentCSV struct {
	Parent           entity.Parent
	UserPhoneNumbers entity.UserPhoneNumbers
	StudentIDs       []string
	StudentNames     []string
	Relationship     []string
	Tags             entity.DomainTags
	RowNumber        int
}

func (s *UserModifierService) checkDuplicateData(ctx context.Context, parents entity.Parents) ([]*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError, error) {
	var errorCSVs []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError
	zapLogger := ctxzap.Extract(ctx).Sugar()

	existEmailMap := map[string]string{}
	existEPhoneNumberMap := map[string]string{}
	existExternalUserIDMap := map[string]string{}

	phoneNumbers := parents.PhoneNumbers()

	users := parents.Users()

	// validate email in parent data
	existingUsers, err := s.UserRepo.GetByEmailInsensitiveCase(ctx, s.DB, parents.Emails())
	if err != nil {
		return errorCSVs, status.Errorf(codes.Internal, "s.UserRepo.GetByEmailInsensitiveCase: %v", err)
	}

	// validate phone_number in student data
	phoneNumberExistingStudents, err := s.UserRepo.GetByPhone(ctx, s.DB, database.TextArray(phoneNumbers))
	if err != nil {
		return errorCSVs, status.Errorf(codes.Internal, "s.UserRepo.GetByPhone: %v", err)
	}

	externalUserIDs := users.ExternalUserIDs()
	externalUserIDExistingParents, err := s.DomainUserRepo.GetByExternalUserIDs(ctx, s.DB, externalUserIDs)
	if err != nil {
		return errorCSVs, status.Errorf(codes.Internal, "s.DomainUserRepo.GetByExternalUserIDs: %v", err)
	}

	for _, user := range existingUsers {
		existEmailMap[strings.ToLower(user.Email.String)] = user.GetUID()
	}

	for _, user := range phoneNumberExistingStudents {
		existEPhoneNumberMap[user.GetPhoneNumber()] = user.GetUID()
	}

	for _, user := range externalUserIDExistingParents {
		existExternalUserIDMap[user.ExternalUserID().String()] = user.UserID().String()
	}

	for index, parent := range parents {

		if _, ok := existEmailMap[strings.ToLower(parent.Email.String)]; ok {
			zapLogger.Errorf("email in parent data validation failed email: %s", parent.Email.String)
			errorCSVs = append(errorCSVs, convertImportParentsRestErrToErrResForEachLineCSV(errAlreadyRegisteredRow, emailParentCSVHeader, index))
		}

		if _, ok := existEPhoneNumberMap[parent.GetPhoneNumber()]; ok {
			zapLogger.Errorf("phone_number in parent data validation failed phone_number: %v", parent.GetPhoneNumber())
			errorCSVs = append(errorCSVs, convertImportParentsRestErrToErrResForEachLineCSV(errAlreadyRegisteredRow, phoneNumberParentCSVHeader, index))
		}

		trimmedExternalUserID := strings.TrimSpace(parent.ExternalUserID.String)

		if trimmedExternalUserID != "" {
			if _, ok := existExternalUserIDMap[trimmedExternalUserID]; ok {
				zapLogger.Errorf("external_user_id in parent data validation failed external_user_id: %v", trimmedExternalUserID)
				errorCSVs = append(errorCSVs, convertImportParentsRestErrToErrResForEachLineCSV(errAlreadyRegisteredRow, externalUserIDParentCSVHeader, index))
			}
		}
	}
	return errorCSVs, nil
}

func (s *UserModifierService) checkDuplicateRow(parents entity.Parents) []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError {
	var errorCSVs []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError

	// this var is for checking duplicate email and phone_number in CSV file
	knowEmailAndPhone := map[string]bool{}

	knowExternalUserID := map[string]bool{}

	for i, parent := range parents {
		_, ok := knowEmailAndPhone[strings.ToLower(parent.Email.String)]
		if !ok {
			knowEmailAndPhone[strings.ToLower(parent.Email.String)] = !ok
		} else {
			errorCSVs = append(errorCSVs, convertImportParentsRestErrToErrResForEachLineCSV(errDuplicationRow, emailParentCSVHeader, i))
			continue
		}

		if parent.PhoneNumber.String != "" {
			_, ok = knowEmailAndPhone[parent.PhoneNumber.String]
			if !ok {
				knowEmailAndPhone[parent.PhoneNumber.String] = !ok
			} else {
				errorCSVs = append(errorCSVs, convertImportParentsRestErrToErrResForEachLineCSV(errDuplicationRow, phoneNumberParentCSVHeader, i))
			}
		}

		trimmedExternalUserID := strings.TrimSpace(parent.ExternalUserID.String)

		if trimmedExternalUserID != "" {
			_, ok = knowExternalUserID[trimmedExternalUserID]
			if !ok {
				knowExternalUserID[trimmedExternalUserID] = !ok
			} else {
				errorCSVs = append(errorCSVs, convertImportParentsRestErrToErrResForEachLineCSV(errDuplicationRow, externalUserIDParentCSVHeader, i))
			}
		}
	}

	return errorCSVs
}

func (s *UserModifierService) ImportParentsAndAssignToStudent(ctx context.Context, req *pb.ImportParentsAndAssignToStudentRequest) (*pb.ImportParentsAndAssignToStudentResponse, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	res := &pb.ImportParentsAndAssignToStudentResponse{}
	var errorCSVs []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError

	if err := readAndValidatePayload(req.Payload); err != nil {
		return nil, err
	}

	importParentData, err := convertPayloadToImportParentData(req.Payload)
	if err != nil {
		return nil, err
	}

	if len(importParentData) == 0 {
		return res, nil
	}

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resourcePath := int64(organization.SchoolID().Int32())
	isUserNameStudentParentToggle := unleash.IsFeatureUserNameStudentParentEnabled(s.UnleashClient, s.Env, organization)
	authUsernameConfig, err := s.InternalConfigurationRepo.GetByKey(ctx, s.DB, constant.KeyAuthUsernameConfig)
	isAuthUsernameConfigOn := false
	if err != nil {
		if !strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		isAuthUsernameConfigOn = authUsernameConfig.ConfigValue().String() == constant.ConfigValueOn
	}
	isEnableUsername := isUserNameStudentParentToggle && isAuthUsernameConfigOn
	parentCSVs, errorCSVs, err := s.generatedParentCSVs(ctx, importParentData, isEnableUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport s.generatedParentCSVs: %v", err)
	}
	if len(errorCSVs) > 0 {
		res.Errors = errorCSVs
		return res, nil
	}

	if len(parentCSVs) == 0 {
		return res, nil
	}

	legacyParentUsers, parentUserProfiles, userPhoneNumbers, parentsWithTags := generatedParents(parentCSVs, isEnableUsername)

	errorCSVs = s.checkDuplicateRow(legacyParentUsers)
	if len(errorCSVs) > 0 {
		res.Errors = errorCSVs
		return res, nil
	}

	for idx, parentUserProfile := range parentUserProfiles {
		if err := entity.ValidUser(isEnableUsername, parentUserProfile); err != nil {
			err := convertToImportError(err)
			res.Errors = append(res.Errors, convertImportParentsRestErrToErrResForEachLineCSV(err, userNameParentCSVHeader, idx))
			return res, nil
		}
	}

	if err := ValidateUserDuplicatedFields(parentUserProfiles); err != nil {
		err := err.(entity.DuplicatedFieldError)
		res.Errors = append(res.Errors, convertImportParentsRestErrToErrResForEachLineCSV(errDuplicationRow, err.DuplicatedField, err.Index))
		return res, nil
	}

	errorCSVs, err = s.checkDuplicateData(ctx, legacyParentUsers)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport s.checkDuplicateData: %v", err)
	}
	if len(errorCSVs) > 0 {
		res.Errors = errorCSVs
		return res, nil
	}

	if isEnableUsername {
		if err := ValidateUserNamesExistedInSystem(ctx, s.DomainUserRepo, s.DB, parentUserProfiles); err != nil {
			switch err := err.(type) {
			case entity.ExistingDataError:
				res.Errors = append(res.Errors, convertImportParentsRestErrToErrResForEachLineCSV(errAlreadyRegisteredRow, userNameParentCSVHeader, err.Index))
				return res, nil
			default:
				return nil, status.Errorf(codes.Internal, "otherErrorImport ValidateUserNamesExistedInSystem: %v", err)
			}
		}
	}

	err = s.insertUsrEmailsFromParents(ctx, legacyParentUsers)
	if err != nil {
		err = status.Errorf(codes.Internal, "otherErrorImport s.insertUsrEmailsFromParents: %v", err)
		return nil, err
	}

	parentUsers := legacyParentUsers.Users()
	importUserEvents, err := toImportParentUserEvents(ctx, parentCSVs)
	if err != nil {
		err = status.Errorf(codes.Internal, "otherErrorImport toImportUserEvents: %v", err)
		return nil, err
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// find parent user group id
		parentUserGroup, err := s.UserGroupV2Repo.FindUserGroupByRoleName(ctx, tx, constant.RoleParent)
		if err != nil {
			return fmt.Errorf("s.UserGroupV2Repo.FindUserGroupByRoleName: %v", err)
		}

		if err = s.createParents(ctx, tx, legacyParentUsers); err != nil {
			return fmt.Errorf("s.createParents: %v", err)
		}

		if err = s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
			return fmt.Errorf("s.UserPhoneNumberRepo.Upsert: %v", err)
		}

		// Check if student to assign exist
		for _, parentCSV := range parentCSVs {
			if len(parentCSV.StudentIDs) != 0 && len(parentCSV.Relationship) != 0 {
				err = s.assignMultiParentsToMultiStudent(ctx, tx, parentCSV)
				if err != nil {
					return fmt.Errorf("s.assignMultiParentsToMultiStudent: %v", err)
				}
			}
		}

		// assign parent user group to parent users
		if err := s.UserGroupsMemberRepo.AssignWithUserGroup(ctx, tx, parentUsers, parentUserGroup.UserGroupID); err != nil {
			return fmt.Errorf("s.UserGroupsMemberRepo.AssignWithUserGroup: %v", err)
		}

		if err := s.UpsertTaggedUsers(ctx, tx, parentsWithTags, nil); err != nil {
			return fmt.Errorf("UpsertTaggedUsers: %v", err)
		}

		// Import to identity platform
		tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, tx, strconv.FormatInt(resourcePath, 10))
		if err != nil {
			zapLogger.Error(
				"cannot get tenant id",
				zap.Error(err),
				zap.Int64("organizationID", resourcePath),
			)
			switch err {
			case pgx.ErrNoRows:
				return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: strconv.Itoa(int(resourcePath))}.Error())
			default:
				return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
			}
		}

		err = s.CreateUsersInIdentityPlatform(ctx, tenantID, parentUsers, resourcePath)
		if err != nil {
			zapLogger.Error(
				"failed to create users on identity platform",
				zap.Error(err),
				zap.Int64("organizationID", resourcePath),
				zap.String("tenantID", tenantID),
				zap.Strings("emails", parentUsers.Limit(10).Emails()),
			)
			// convert to switch case to handle error types
			return status.Error(codes.Internal, errors.Wrap(err, "failed to create users on identity platform").Error())
		}

		importUserEvents, err = s.ImportUserEventRepo.Upsert(ctx, tx, importUserEvents)
		if err != nil {
			return fmt.Errorf("s.ImportUserEventRepo.Upsert err: %v", err)
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport database.ExecInTx: %v", err)
	}

	err = s.TaskQueue.Add(nats.PublishImportUserEventsTask(ctx, s.DB, s.JSM, &nats.PublishImportUserEventsTaskOptions{
		ImportUserEventIDs: importUserEvents.IDs(),
		ResourcePath:       golibs.ResourcePathFromCtx(ctx),
	}))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "otherErrorImport task.TaskQueue.Add: %v", err)
	}

	return res, nil
}

func toImportParentUserEvents(ctx context.Context, parentCSVs []*ParentCSV) (entity.ImportUserEvents, error) {
	importUserEvents := make([]*entity.ImportUserEvent, 0)

	for _, parentCSV := range parentCSVs {
		if len(parentCSV.StudentIDs) != 0 {
			for idx := range parentCSV.StudentIDs {
				studentID := parentCSV.StudentIDs[idx]
				studentName := parentCSV.StudentNames[idx]
				parentID := parentCSV.Parent.ID.String
				schoolID := strconv.FormatInt(int64(parentCSV.Parent.SchoolID.Int), 10)

				importUserEvent, err := CreateImportUserEvents(ctx, studentID, studentName, parentID, schoolID)
				if err != nil {
					return nil, err
				}

				importUserEvents = append(importUserEvents, importUserEvent)
			}
		} else {
			studentID := ""
			studentName := ""
			parentID := parentCSV.Parent.ID.String
			schoolID := strconv.FormatInt(int64(parentCSV.Parent.SchoolID.Int), 10)

			importUserEvent, err := CreateImportUserEvents(ctx, studentID, studentName, parentID, schoolID)
			if err != nil {
				return nil, err
			}

			importUserEvents = append(importUserEvents, importUserEvent)
		}
	}

	return importUserEvents, nil
}

func CreateImportUserEvents(ctx context.Context, studentID, studentName, parentID, schoolID string) (*entity.ImportUserEvent, error) {
	createParentEvent := &pb.EvtUser{
		Message: &pb.EvtUser_CreateParent_{
			CreateParent: &pb.EvtUser_CreateParent{
				StudentId:   studentID,
				StudentName: studentName,
				ParentId:    parentID,
				SchoolId:    schoolID,
			},
		},
	}
	importUserEvent := &entity.ImportUserEvent{}
	database.AllNullEntity(importUserEvent)

	payload, err := protojson.Marshal(createParentEvent)
	if err != nil {
		return nil, fmt.Errorf("protojson.Marshal err: %v", err)
	}

	err = multierr.Combine(
		importUserEvent.ImporterID.Set(interceptors.UserIDFromContext(ctx)),
		importUserEvent.Status.Set(cpb.ImportUserEventStatus_IMPORT_USER_EVENT_STATUS_WAITING.String()),
		importUserEvent.UserID.Set(parentID),
		importUserEvent.Payload.Set(payload),
		importUserEvent.ResourcePath.Set(golibs.ResourcePathFromCtx(ctx)),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine err: %v", err)
	}

	return importUserEvent, nil
}

func (s *UserModifierService) insertUsrEmailsFromParents(ctx context.Context, parents []*entity.Parent) error {
	// Insert UsrEmail
	users := make(entity.LegacyUsers, 0, len(parents))
	for _, parent := range parents {
		users = append(users, &parent.LegacyUser)
	}
	usrEmails, err := s.UsrEmailRepo.CreateMultiple(ctx, s.DB, users)
	if err != nil {
		return fmt.Errorf("s.UsrEmailRepo.CreateMultiple: %v", err)
	}
	for i := range usrEmails {
		parents[i].ID = usrEmails[i].UsrID
		parents[i].LegacyUser.ID = usrEmails[i].UsrID
	}
	return nil
}

func generatedParents(parentCSVs []*ParentCSV, isEnableUsername bool) (entity.Parents, entity.Users, entity.UserPhoneNumbers, map[entity.User][]entity.DomainTag) {
	parents := make([]*entity.Parent, 0, len(parentCSVs))
	userProfiles := make(entity.Users, 0, len(parentCSVs))
	userPhoneNumbers := make(entity.UserPhoneNumbers, 0, len(parentCSVs))
	userWithTags := make(map[entity.User][]entity.DomainTag)

	for _, parentCSV := range parentCSVs {
		profile := &pb.UserProfile{
			UserId:         parentCSV.Parent.GetUID(),
			ExternalUserId: parentCSV.Parent.ExternalUserID.String,
			Email:          parentCSV.Parent.Email.String,
			Username:       parentCSV.Parent.UserName.String,
			FirstName:      parentCSV.Parent.FirstName.String,
			LastName:       parentCSV.Parent.LastName.String,
		}

		if !isEnableUsername {
			profile.Username = profile.Email
			parentCSV.Parent.UserName = database.Text(profile.Email)
		}
		userProfile := grpc.NewUserProfile(profile)
		parents = append(parents, &parentCSV.Parent)
		userProfiles = append(userProfiles, userProfile)
		userPhoneNumbers = append(userPhoneNumbers, parentCSV.UserPhoneNumbers...)
		userWithTags[userProfile] = parentCSV.Tags
	}

	return parents, userProfiles, userPhoneNumbers, userWithTags
}

func (s *UserModifierService) generatedParentCSVs(ctx context.Context, importParentData []*ImportParentCSVField, isEnableUsername bool) ([]*ParentCSV, []*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError, error) {
	errorCSVs := make([]*pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError, 0, len(importParentData))
	parentCSVs := make([]*ParentCSV, 0, len(importParentData))

	currentUserID := interceptors.UserIDFromContext(ctx)
	currentUser, err := s.UserRepo.Get(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return nil, errorCSVs, fmt.Errorf("s.UserRepo.Get: %v", err)
	}
	resourcePath := golibs.ResourcePathFromCtx(ctx)

	for idx, data := range importParentData {
		parentCSV, errLineRes, err := s.convertLineCSVToParent(ctx, data, idx, currentUser.Country.String, resourcePath, isEnableUsername)
		if err != nil {
			err = status.Errorf(codes.Internal, "otherErrorImport s.convertLineCSVToParent: %v", err)
			return nil, errorCSVs, fmt.Errorf("s.convertLineCSVToParent: %v", err)
		}
		if errLineRes != nil {
			errorCSVs = append(errorCSVs, errLineRes)
			continue
		}
		parentCSVs = append(parentCSVs, &parentCSV)
	}
	return parentCSVs, errorCSVs, nil
}

func (s *UserModifierService) convertLineCSVToParent(
	ctx context.Context,
	importParentData *ImportParentCSVField,
	order int,
	countryCode,
	resourcePath string,
	isEnableUsername bool,
) (ParentCSV, *pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	parentCSV := ParentCSV{
		RowNumber: order,
	}

	parent := &parentCSV.Parent
	database.AllNullEntity(parent)
	database.AllNullEntity(&parent.LegacyUser)

	var (
		newUserID        string
		externalUserID   string
		email            string
		phoneNumber      string
		arrStudentEmails []string
		arrRelationships []string
		tags             []string
	)

	if err := validateRequiredFieldImportParentCSV(importParentData, order); err != nil {
		return parentCSV, err, nil
	}

	// validate parent email

	email = importParentData.Email.String()
	match, err := regexp.MatchString(emailPattern, email)
	if err != nil {
		return parentCSV, nil, fmt.Errorf("regexp.MatchString: %v, row: %v", err, order)
	}
	if !match {
		zapLogger.Warn("email validation failed")
		return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, emailParentCSVHeader, order), nil
	}

	// validate student email
	studentEmail := importParentData.StudentEmail.String()
	parentRelationshipRaw := importParentData.Relationship.String()

	arrStudentEmails = strings.Split(studentEmail, ";")
	arrRelationships = strings.Split(parentRelationshipRaw, ";")

	switch {
	case
		studentEmail == "" && parentRelationshipRaw != "",
		studentEmail != "" && parentRelationshipRaw == "",
		studentEmail == "" && parentRelationshipRaw == "",
		len(arrStudentEmails) != len(arrRelationships):
		return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(fmt.Errorf("notMatchRelationshipAndEmailStudent"), studentEmailsParentCSVHeader, order), nil

	case studentEmail != "" && parentRelationshipRaw != "":
		{
			// Create slice studentIDs for get student profile
			var studentIds, studentNames []string

			studentUserProfiles, err := s.DomainStudentRepo.GetByEmails(ctx, s.DB, arrStudentEmails)
			if err != nil || (len(studentUserProfiles) != len(arrStudentEmails)) {
				return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, studentEmailsParentCSVHeader, order), nil
			}

			for _, studentProfiles := range studentUserProfiles {
				studentIds = append(studentIds, studentProfiles.UserID().String())
				studentNames = append(studentNames, studentProfiles.FullName().String())
			}

			parentCSV.StudentNames = studentNames
			parentCSV.StudentIDs = studentIds

			for _, relationship := range arrRelationships {
				relationshipInt, err := strconv.ParseInt(relationship, 10, 32)
				if err != nil || relationshipInt == int64(pb.FamilyRelationship_FAMILY_RELATIONSHIP_NONE) {
					return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, relationshipsParentCSVHeader, order), nil
				}

				relaValue, ok := pb.FamilyRelationship_name[int32(relationshipInt)]
				if !ok {
					zapLogger.Warnf("relationship validation failed, relationship: %v, row: %v", relaValue, order)
					return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, relationshipsParentCSVHeader, order), nil
				}
				parentCSV.Relationship = append(parentCSV.Relationship, relaValue)
			}
		}
	}

	// Validation parent's phone number
	phoneNumber = importParentData.PhoneNumber.String()
	if phoneNumber != "" {
		// validate phone number format
		num, err := phonenumbers.Parse(phoneNumber, strings.Split(countryCode, "_")[1])
		if err != nil || !phonenumbers.IsValidNumber(num) {
			zapLogger.Warnf("phone number validation failed phone: %s, err: %v, row: %v", phoneNumber, err, order)
			return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, phoneNumberParentCSVHeader, order), nil
		}
	}

	if importParentData.ParentTag.CheckExist() {
		tagSequence := importParentData.ParentTag.String()
		tags = strings.Split(tagSequence, ";")
		if isDuplicatedIDs(tags) {
			return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, tagParentCSVHeader, order), nil
		}
		var err error
		parentCSV.Tags, err = s.DomainTagRepo.GetByPartnerInternalIDs(ctx, s.DB, tags)
		if err != nil {
			zapLogger.Warnf("get tag by ids err: %v, row: %v", err, order)
			return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, tagParentCSVHeader, order), nil
		}

		if err := importCsvValidateTag(constant.RoleParent, tags, parentCSV.Tags); err != nil {
			return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(err, tagParentCSVHeader, order), nil
		}
	}

	schoolID, err := strconv.Atoi(resourcePath)
	if err != nil {
		return parentCSV, nil, fmt.Errorf("strconv.Atoi: %v, row: %v", err, order)
	}

	var errLineRes *pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError
	newUserID = idutil.ULIDNow()

	username := importParentData.UserName
	if username != nil {
		username := strings.TrimSpace(username.String())
		username = strings.ToLower(username)
		parent.LegacyUser.UserName = database.Text(username)
	}
	_ = parent.LegacyUser.LoginEmail.Set(email)
	if isEnableUsername {
		_ = parent.LegacyUser.LoginEmail.Set(newUserID + constant.LoginEmailPostfix)
	}
	if err = multierr.Combine(
		parent.LegacyUser.ID.Set(newUserID),
		parent.LegacyUser.Email.Set(email),
		parent.LegacyUser.FullName.Set(importParentData.Name.String()),
		parent.LegacyUser.FirstName.Set(""),
		parent.LegacyUser.LastName.Set(""),
		parent.LegacyUser.Group.Set(entity.UserGroupParent),
		parent.LegacyUser.Country.Set(countryCode),
		parent.LegacyUser.ResourcePath.Set(resourcePath),
		parent.LegacyUser.Remarks.Set(importParentData.Remarks),
		parent.ID.Set(newUserID),
		parent.SchoolID.Set(schoolID),
		parent.ResourcePath.Set(resourcePath),
		parent.LegacyUser.UserRole.Set(constant.UserRoleParent),
	); err != nil {
		return parentCSV, nil, fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
	}

	if err = multierr.Combine(
		parent.LegacyUser.FullName.Set(CombineFirstNameAndLastNameToFullName(importParentData.FirstName.String(), importParentData.LastName.String())),
		parent.LegacyUser.FirstName.Set(importParentData.FirstName),
		parent.LegacyUser.LastName.Set(importParentData.LastName),
	); err != nil {
		return parentCSV, nil, fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
	}

	if err := setParentUserEntityCSVNormalize(parent, importParentData.FirstNamePhonetic, importParentData.LastNamePhonetic, order); err != nil {
		return parentCSV, nil, err
	}

	if phoneNumber != "" {
		if err := parent.LegacyUser.PhoneNumber.Set(phoneNumber); err != nil {
			return parentCSV, nil, fmt.Errorf("parent.User.PhoneNumber.Set: %v, row: %v", err, order)
		}
	}

	externalUserID = strings.TrimSpace(importParentData.ExternalUserID.String())
	if externalUserID != "" {
		if err := parent.LegacyUser.ExternalUserID.Set(externalUserID); err != nil {
			return parentCSV, nil, fmt.Errorf("parent.User.ExternalUserID.Set: %v, row: %v", err, order)
		}
	}

	parentCSV, errLineRes, err = getAndValidUserPhoneNumberCSV(importParentData, parentCSV, order)

	return parentCSV, errLineRes, err
}

func getAndValidUserPhoneNumberCSV(importParentData *ImportParentCSVField, parentCSV ParentCSV, order int) (ParentCSV, *pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError, error) {
	primaryPhoneNumber := importParentData.PrimaryPhoneNumber.String()
	secondaryPhoneNumber := importParentData.SecondaryPhoneNumber.String()

	if primaryPhoneNumber != "" {
		newID := idutil.ULIDNow()
		newParentPhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(newParentPhoneNumber)

		if err := MatchingRegex(constant.PhoneNumberPattern, primaryPhoneNumber); err != nil {
			return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, primaryPhoneNumberParentCSVHeader, order), nil
		}

		if err := multierr.Combine(
			newParentPhoneNumber.ID.Set(newID),
			newParentPhoneNumber.UserID.Set(parentCSV.Parent.ID),
			newParentPhoneNumber.PhoneNumber.Set(primaryPhoneNumber),
			newParentPhoneNumber.PhoneNumberType.Set(pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER),
			newParentPhoneNumber.ResourcePath.Set(parentCSV.Parent.ResourcePath),
		); err != nil {
			return parentCSV, nil, fmt.Errorf("userPhoneNumber.PhoneNumber multierr.Combine: %v, row: %v", err, order)
		}

		parentCSV.UserPhoneNumbers = append(parentCSV.UserPhoneNumbers, newParentPhoneNumber)
	}

	if secondaryPhoneNumber != "" {
		newID := idutil.ULIDNow()
		newParentPhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(newParentPhoneNumber)

		if err := MatchingRegex(constant.PhoneNumberPattern, secondaryPhoneNumber); err != nil {
			return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errNotFollowParentTemplate, secondaryPhoneNumberParentCSVHeader, order), nil
		}

		if err := multierr.Combine(
			newParentPhoneNumber.ID.Set(newID),
			newParentPhoneNumber.UserID.Set(parentCSV.Parent.ID),
			newParentPhoneNumber.PhoneNumber.Set(secondaryPhoneNumber),
			newParentPhoneNumber.PhoneNumberType.Set(pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER),
			newParentPhoneNumber.ResourcePath.Set(parentCSV.Parent.ResourcePath),
		); err != nil {
			return parentCSV, nil, fmt.Errorf("userPhoneNumber.PhoneNumber multierr.Combine: %v, row: %v", err, order)
		}

		parentCSV.UserPhoneNumbers = append(parentCSV.UserPhoneNumbers, newParentPhoneNumber)
	}

	if primaryPhoneNumber != "" && secondaryPhoneNumber != "" && primaryPhoneNumber == secondaryPhoneNumber {
		return parentCSV, convertImportParentsRestErrToErrResForEachLineCSV(errDuplicationRow, primaryPhoneNumberParentCSVHeader, order), nil
	}

	return parentCSV, nil, nil
}

func convertImportParentsRestErrToErrResForEachLineCSV(err error, fieldName string, i int) *pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError {
	return &pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError{
		RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
		Error:     err.Error(),
		FieldName: fieldName,
	}
}

func convertPayloadToImportParentData(payload []byte) ([]*ImportParentCSVField, error) {
	importParentData := []*ImportParentCSVField{}

	if err := gocsv.UnmarshalBytes(payload, &importParentData); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("gocsv.UnmarshalBytes: %w", err).Error())
	}

	if len(importParentData) > constant.LimitRowsCSV {
		return nil, status.Error(codes.InvalidArgument, errInvalidNumberRow.Error())
	}

	return importParentData, nil
}

func validateRequiredFieldImportParentCSV(importParentCSVField *ImportParentCSVField, order int) *pb.ImportParentsAndAssignToStudentResponse_ImportParentsAndAssignToStudentError {
	switch {
	case (importParentCSVField.FirstName == nil || (importParentCSVField.FirstName != nil && !importParentCSVField.FirstName.Exist)):
		return convertImportParentsRestErrToErrResForEachLineCSV(errMissingMandatory, firstNameParentCSVHeader, order)
	case (importParentCSVField.LastName == nil || (importParentCSVField.LastName != nil && !importParentCSVField.LastName.Exist)):
		return convertImportParentsRestErrToErrResForEachLineCSV(errMissingMandatory, lastNameParentCSVHeader, order)
	case importParentCSVField.Email == nil || (importParentCSVField.Email != nil && !importParentCSVField.Email.Exist):
		return convertImportParentsRestErrToErrResForEachLineCSV(errMissingMandatory, emailParentCSVHeader, order)
	}

	return nil
}

func setParentUserEntityCSVNormalize(
	parent *entity.Parent,
	firstNamePhoneticCSV *CsvField,
	lastNamePhoneticCSV *CsvField,
	order int,
) error {
	if firstNamePhoneticCSV != nil && firstNamePhoneticCSV.Exist {
		if err := parent.LegacyUser.FirstNamePhonetic.Set(firstNamePhoneticCSV); err != nil {
			return fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
	}
	if lastNamePhoneticCSV != nil && lastNamePhoneticCSV.Exist {
		if err := parent.LegacyUser.LastNamePhonetic.Set(lastNamePhoneticCSV); err != nil {
			return fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
	}
	fullNamePhonetic := CombineFirstNamePhoneticAndLastNamePhoneticToFullName(firstNamePhoneticCSV.String(), lastNamePhoneticCSV.String())
	if fullNamePhonetic != "" {
		if err := parent.LegacyUser.FullNamePhonetic.Set(fullNamePhonetic); err != nil {
			return fmt.Errorf("multierr.Combine: %v, row: %v", err, order)
		}
	}
	return nil
}

func convertToImportError(err error) error {
	switch err.(type) {
	case entity.MissingMandatoryFieldError:
		return errMissingMandatory
	case entity.InvalidFieldError:
		return errNotFollowParentTemplate
	case entity.DuplicatedFieldError:
		return errDuplicationRow
	case entity.ExistingDataError:
		return errAlreadyRegisteredRow
	default:
		return err
	}
}
