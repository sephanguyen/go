package entryexitmgmt

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	EntryExitMgmtConfigKeyPrefix = "entryexit.entryexitmgmt."
)

//nolint:gosec
func (s *suite) aValidUser(ctx context.Context, db database.Ext, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Set user entity
	num := rand.Int()

	u := &bob_entities.User{}
	database.AllNullEntity(u)

	ugm := &entity.UserGroupMember{}
	database.AllNullEntity(ugm)

	err := multierr.Combine(
		u.LastName.Set(fmt.Sprintf("valid-user-%d", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		u.Country.Set(bob_pb.COUNTRY_JP.String()),
		u.Group.Set(constant.UserGroupAdmin),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Set user options
	for _, opt := range opts {
		err := opt(ctx, db, u, ugm)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	// Create user in database
	err = s.createUserInDB(ctx, db, u, int64(stepState.CurrentSchoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Set user group entity
	uGroup := &bob_entities.UserGroup{}
	database.AllNullEntity(uGroup)
	err = multierr.Combine(
		uGroup.GroupID.Set(u.Group.String),
		uGroup.UserID.Set(u.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// Create user in user group
	userGroupRepo := &bob_repo.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, db, uGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroupRepo.Upsert: %w %s", err, u.Group.String)
	}

	stepState.StudentName = u.LastName.String

	// Add user to user group
	ctx, err = s.createUserGroupMember(ctx, db, u.ID.String, ugm)
	if err != nil {
		return ctx, err
	}

	// Add user to a location
	ctx, err = s.addUserToLocation(ctx, db, u.ID.String)
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserInDB(ctx context.Context, db database.Ext, user *bob_entities.User, schoolID int64) error {
	err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := bob_repo.UserRepo{}

		_, err := userRepo.Get(ctx, tx, user.ID)
		if err != pgx.ErrNoRows {
			return err
		}

		err = userRepo.Create(ctx, tx, user)
		if err != nil {
			return err
		}

		return createEntities(ctx, tx, user, schoolID)
	})

	if err != nil {
		return err
	}
	return nil
}

func (s *suite) addUserToLocation(ctx context.Context, db database.Ext, userID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().UTC()
	e := &entity.UserAccessPath{}
	database.AllNullEntity(e)
	_ = multierr.Combine(
		e.UserID.Set(userID),
		e.LocationID.Set("01FR4M51XJY9E77GSN4QZ1Q8N5"),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)

	_, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Error on inserting access path %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserGroupMember(ctx context.Context, db database.Ext, userID string, ugm *entity.UserGroupMember) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().UTC()
	_ = multierr.Combine(
		ugm.UserID.Set(userID),
		ugm.CreatedAt.Set(now),
		ugm.UpdatedAt.Set(now),
	)

	_, err := database.InsertExcept(ctx, ugm, []string{"resource_path"}, db.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Error on inserting user group %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func (s *suite) signedAsAccount(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	var (
		role      string
		userGroup string
		err       error
	)

	switch group {
	case "unauthenticated":
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case "teacher":
		userGroup = constant.UserGroupTeacher
		role = userConstant.RoleTeacher
	case "school admin":
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleSchoolAdmin
	case "parent":
		stepState.CurrentParentID = id
		userGroup = constant.UserGroupParent
		role = userConstant.RoleParent
	case "student":
		userGroup = constant.UserGroupStudent
		role = userConstant.RoleStudent
	case "hq staff":
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleHQStaff
	case "centre lead":
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleCentreLead
	case "centre manager":
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleCentreManager
	case "centre staff":
		userGroup = constant.UserGroupSchoolAdmin
		role = userConstant.RoleCentreStaff
	}

	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), s.BobDBTrace, withID(id), withUserGroup(userGroup), withResourcePath(stepState.ResourcePath), withRole(role))
	if err != nil {
		return ctx, err
	}

	time.Sleep(3 * time.Second) // wait for kafka sync

	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = userGroup
	stepState.AuthToken, err = s.generateExchangeToken(id, userGroup, int64(stepState.CurrentSchoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return ctx, nil
}

func (s *suite) aSignedInAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	var err error
	stepState.AuthToken, err = generateValidAuthenticationToken(id, constant.UserGroupAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupAdmin

	return s.aValidUser(StepStateToContext(ctx, stepState), s.BobDBTrace, withID(id), withUserGroup(constant.UserGroupAdmin), withRole(userConstant.RoleSchoolAdmin))
}

func (s *suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	var err error
	stepState.AuthToken, err = generateValidAuthenticationToken(id, "phone")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupStudent

	return s.aValidStudentInDB(StepStateToContext(ctx, stepState), id)
}

func (s *suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentRepo := bob_repo.StudentRepo{}
	now := time.Now()
	student := &bob_entities.Student{}
	database.AllNullEntity(student)
	err := multierr.Combine(
		student.ID.Set(id),
		student.CurrentGrade.Set(12),
		student.OnTrial.Set(true),
		student.TotalQuestionLimit.Set(10),
		student.SchoolID.Set(stepState.CurrentSchoolID),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.BillingDate.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = studentRepo.Create(ctx, s.BobDBTrace, student)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return s.aValidUser(StepStateToContext(ctx, stepState), s.BobDBTrace, withID(student.ID.String), withUserGroup(constant.UserGroupStudent), withRole(userConstant.RoleStudent))
}

func (s *suite) sendEntryExitRequest(ctx context.Context, req interface{}) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()

	switch r := req.(type) {
	case *eepb.CreateEntryExitRequest:
		stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).CreateEntryExit(contextWithToken(ctx), r)
		if stepState.ResponseErr == nil && stepState.Response != nil {
			stepState.ParentNotified = stepState.Response.(*eepb.CreateEntryExitResponse).ParentNotified
		}
	case *eepb.UpdateEntryExitRequest:
		stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).UpdateEntryExit(contextWithToken(ctx), r)
		if stepState.ResponseErr == nil && stepState.Response != nil {
			stepState.ParentNotified = stepState.Response.(*eepb.UpdateEntryExitResponse).ParentNotified
		}
	case *eepb.DeleteEntryExitRequest:
		stepState.Response, stepState.ResponseErr = eepb.NewEntryExitServiceClient(s.EntryExitMgmtConn).DeleteEntryExit(contextWithToken(ctx), r)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateDeviceToken(ctx context.Context, userID string) (context.Context, error) {
	deviceToken := idutil.ULIDNow()

	updateQuery := fmt.Sprintf("UPDATE users SET device_token = '%s', allow_notification = 'true' WHERE user_id = '%s'", deviceToken, userID)
	if _, err := s.BobDBTrace.Exec(ctx, updateQuery); err != nil {
		return ctx, fmt.Errorf("db.Exec %v", err)
	}

	insertQuery := "INSERT INTO user_device_tokens(user_id, device_token, allow_notification, created_at, updated_at) VALUES ($1, $2, true, NOW(), NOW())"
	if _, err := s.BobDBTrace.Exec(ctx, insertQuery, userID, deviceToken); err != nil {
		return ctx, fmt.Errorf("db.Exec %v", err)
	}

	return ctx, nil
}

func (s *suite) checkUserNotification(ctx context.Context, userIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, receiveID := range userIDs {

		var deviceToken string
		row := s.BobDBTrace.QueryRow(ctx, "SELECT device_token FROM public.user_device_tokens WHERE user_id = $1", receiveID)

		if err := row.Scan(&deviceToken); err != nil {
			return ctx, fmt.Errorf("error finding user device token with userid: %s: %w", receiveID, err)
		}

		resp, err := retrievePushedNotification(ctx, s.NotificationMgmtConn, deviceToken)
		if err != nil {
			return ctx, fmt.Errorf("error when call NotificationModifierService.RetrievePushedNotificationMessages: %w", err)
		}

		if len(resp.Messages) == 0 {
			return ctx, fmt.Errorf("wrong node: user receive id: " + receiveID + ", device_token: " + deviceToken)
		}

		gotNotification := resp.Messages[len(resp.Messages)-1]
		gotTitle := gotNotification.Title

		jpTitle, enTitle := "入退室記録", "Entry & Exit Activity"
		if gotTitle != jpTitle && gotTitle != enTitle {
			return ctx, fmt.Errorf("want notification title to be: %s or %s, got %s", jpTitle, enTitle, gotTitle)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func retrievePushedNotification(ctx context.Context, noti *grpc.ClientConn, deviceToken string) (*npb.RetrievePushedNotificationMessageResponse, error) {
	respNoti, err := npb.NewInternalServiceClient(noti).RetrievePushedNotificationMessages(
		ctx,
		&npb.RetrievePushedNotificationMessageRequest{
			DeviceToken: deviceToken,
			Limit:       1,
			Since:       timestamppb.Now(),
		})

	if err != nil {
		return nil, fmt.Errorf("Error from notificationmgmt: %v", err)
	}

	return respNoti, nil
}

func (s *suite) createStudentParentRelationship(ctx context.Context, studentID string, parentIDs []string, relationship string) error {
	entities := make([]*entities_bob.StudentParent, 0, len(parentIDs))

	for _, parentID := range parentIDs {
		studentParent := &entities_bob.StudentParent{}
		database.AllNullEntity(studentParent)
		err := multierr.Combine(
			studentParent.StudentID.Set(studentID),
			studentParent.ParentID.Set(parentID),
			studentParent.Relationship.Set(relationship),
		)
		if err != nil {
			return err
		}
		entities = append(entities, studentParent)
	}

	if err := (&bob_repo.StudentParentRepo{}).Upsert(ctx, s.BobDBTrace, entities); err != nil {
		return err
	}

	return nil
}

func generateCreateEntryExitRequest(ctx context.Context, entryexit, timeZone string) (*eepb.CreateEntryExitRequest, error) {
	stepState := StepStateFromContext(ctx)

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, err
	}

	req := &eepb.CreateEntryExitRequest{
		EntryExitPayload: &eepb.EntryExitPayload{
			StudentId:     stepState.StudentID,
			EntryDateTime: timestamppb.New(time.Now().Add(-7 * time.Hour).In(location)),
			ExitDateTime:  nil,
			NotifyParents: stepState.NotifyParentRequest,
		},
	}
	if entryexit == "entry and exit" {
		req.EntryExitPayload.ExitDateTime = timestamppb.New(time.Now().In(location))
	}

	return req, nil
}

func generateEntryExitRecord(studentID string, entryAt, exitAt time.Time) (*entities.StudentEntryExitRecords, error) {
	e := &entities.StudentEntryExitRecords{}
	database.AllNullEntity(e)

	err := e.ExitAt.Set(exitAt)
	if exitAt.IsZero() {
		err = e.ExitAt.Set(nil)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot set e.exitAt: %w", err)
	}

	err = multierr.Combine(
		e.StudentID.Set(studentID),
		e.EntryAt.Set(entryAt),
	)

	return e, err
}

func (s *suite) generateExchangeToken(userID, userGroup string, schoolID int64) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, "manabie-local", schoolID, s.ShamirConn)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *suite) entryExitmgmtInternalConfigIs(ctx context.Context, configKey, toggle string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `UPDATE internal_configuration_value SET config_value = $1 
                WHERE config_key = $2  
				AND resource_path = $3 
				AND deleted_at is NULL;`

	_, err := s.MasterMgmtDBTrace.Exec(ctx, stmt, database.Text(toggle), database.Text(EntryExitMgmtConfigKeyPrefix+configKey), database.Text(stepState.ResourcePath))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(b), nil
}

func contextWithToken(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)

	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}

func generateInvalidRequest(ctx context.Context, payload *eepb.EntryExitPayload, invalidArg string) context.Context {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	switch invalidArg {
	case "no entry date":
		payload.EntryDateTime = nil
		payload.ExitDateTime = timestamppb.New(now)
	case "entry date is ahead than exit date":
		payload.EntryDateTime = timestamppb.New(now)
		payload.ExitDateTime = timestamppb.New(now.Add(-24 * time.Hour))
	case "entry time is ahead than exit time":
		payload.EntryDateTime = timestamppb.New(now)
		payload.ExitDateTime = timestamppb.New(now.Add(-1 * time.Hour))
	case "entry time is ahead than current time":
		payload.EntryDateTime = timestamppb.New(now.Add(time.Hour))
		payload.ExitDateTime = nil
	case "entry date is ahead than current date":
		payload.EntryDateTime = timestamppb.New(now.Add(24 * time.Hour))
		payload.ExitDateTime = nil
	case "exit time is ahead than current time":
		payload.EntryDateTime = timestamppb.New(now)
		payload.ExitDateTime = timestamppb.New(now.Add(time.Hour))
	case "exit date is ahead than current date":
		payload.EntryDateTime = timestamppb.New(now)
		payload.ExitDateTime = timestamppb.New(now.Add(24 * time.Hour))
	case "cannot retrieve student id in database":
		payload.StudentId = "not-exist"
		payload.EntryDateTime = timestamppb.New(now)
		payload.ExitDateTime = timestamppb.New(now.Add(time.Hour))
	}
	return StepStateToContext(ctx, stepState)
}

func checkResponseError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func createEntities(ctx context.Context, tx pgx.Tx, user *bob_entities.User, schoolID int64) error {
	userCreator := userCreator{User: user, SchoolID: schoolID}

	switch user.Group.String {
	case constant.UserGroupStudent:
		return userCreator.AsStudent(ctx, tx, user.ResourcePath.String)
	case constant.UserGroupTeacher:
		return userCreator.AsTeacher(ctx, tx, user.ResourcePath.String)
	case constant.UserGroupSchoolAdmin:
		return userCreator.AsSchoolAdmin(ctx, tx, user.ResourcePath.String)
	case constant.UserGroupParent:
		return userCreator.AsParent(ctx, tx, user.ResourcePath.String)
	default:
		return nil
	}
}
