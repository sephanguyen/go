package tom

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	constants_lib "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gogo/protobuf/proto"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	googleproto "google.golang.org/protobuf/proto"
)

func (s *suite) aValidUserDeviceTokenMessage(ctx context.Context) (context.Context, error) {
	rand.Seed(time.Now().UnixNano())

	// nolint:gosec
	userID := "user-id" + strconv.Itoa(rand.Int())
	if s.studentID != "" {
		userID = s.studentID
	}
	s.Request = &pb.EvtUserInfo{
		UserId:            "someuserid",
		DeviceToken:       "token",
		AllowNotification: true,
		Name:              fmt.Sprintf("user-name-%s", userID),
	}
	return ctx, nil
}
func (s *suite) bobSendEventUpsertUserDeviceToken(ctx context.Context) (context.Context, error) {
	data, err := proto.Marshal(s.Request.(*pb.EvtUserInfo))
	if err != nil {
		return ctx, err
	}

	_, err = s.JSM.TracedPublish(ctx, "gandalf:bobSendEventUpsertUserDeviceToken", constants_lib.SubjectUserDeviceTokenUpdated, data)
	return ctx, err
}
func (s *suite) tomMustRecordDeviceTokenMessage(ctx context.Context) (context.Context, error) {
	// sleep to wait for the NATS to deliver the msg
	time.Sleep(5 * time.Second)

	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var userID pgtype.Text
	_ = userID.Set(s.Request.(*pb.EvtUserInfo).UserId)

	row := s.DB.QueryRow(ctx2, "SELECT COUNT(*) FROM user_device_tokens WHERE user_id = $1", &userID)

	var count int
	if err := row.Scan(&count); err != nil {
		return ctx, err
	}
	if count == 0 {
		return ctx, errors.New("cannot upsert user device tokens")
	}
	return ctx, nil
}
func (s *suite) bobSendEventUpsertUserDeviceTokenWithNewTokenAndNewName(ctx context.Context) (context.Context, error) {
	req := s.Request.(*pb.EvtUserInfo)
	msg := &pb.EvtUserInfo{
		UserId:            req.UserId,
		DeviceToken:       "new-token",
		AllowNotification: true,
		Name:              fmt.Sprintf("user-name-%s", req.UserId),
	}
	s.Request = msg

	data, err := proto.Marshal(msg)
	if err != nil {
		return ctx, err
	}
	_, err = s.JSM.TracedPublish(ctx, "bobSendEventUpsertUserDeviceTokenWithNewTokenAndNewName", constants_lib.SubjectUserDeviceTokenUpdated, data)
	return ctx, err
}
func (s *suite) tomMustUpdateTheUserDeviceToken(ctx context.Context) (context.Context, error) {
	// sleep to wait for the NATS to deliver the msg
	time.Sleep(5 * time.Second)

	req := s.Request.(*pb.EvtUserInfo)

	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var userID pgtype.Text
	_ = userID.Set(req.UserId)

	var (
		token             string
		allowNotification bool
	)
	row := s.DB.QueryRow(ctx2, "SELECT token,allow_notification FROM user_device_tokens WHERE user_id = $1", &userID)
	if err := row.Scan(&token, &allowNotification); err != nil {
		return ctx, err
	}
	if token != req.DeviceToken {
		return ctx, fmt.Errorf("got: %q, want: %q", token, req.DeviceToken)
	}

	if allowNotification != req.AllowNotification {
		return ctx, fmt.Errorf("got: %v, want: %v", allowNotification, req.AllowNotification)
	}
	return ctx, nil
}

func (s *suite) tomMustUpdateConversationLocationCorrectlyForEvent(ctx context.Context, eventType string) (context.Context, error) {
	var (
		name, user       string
		updatedLocations []string
	)
	switch eventType {
	case "UserDeviceToken":
		req := s.Request.(*upb.EvtUserInfo)
		name, user = req.GetName(), req.GetUserId()
		updatedLocations = req.GetLocationIds()
	case "UpdateStudent":
		req := s.Request.(*upb.EvtUser).GetUpdateStudent()
		// nolint:staticcheck
		name, user = req.GetName(), req.GetStudentId()
		updatedLocations = req.GetLocationIds()
	default:
		return ctx, fmt.Errorf("unknown event type %s", eventType)
	}
	err := doRetry(func() (bool, error) {
		ctx2, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		var userID, conversationName pgtype.Text
		_ = userID.Set(user)
		_ = conversationName.Set(name)

		query := `SELECT array_remove(array_agg(cl.location_id),null) as locations FROM conversations c JOIN conversation_students cs ON c.conversation_id = cs.conversation_id
	 LEFT JOIN conversation_locations cl ON c.conversation_id = cl.conversation_id AND cl.deleted_at IS NULL
	 WHERE cs.student_id = $1 AND c.name = $2 GROUP BY c.conversation_id `
		rows, err := s.DB.Query(ctx2, query, userID, conversationName)
		if err != nil {
			return false, err
		}
		defer rows.Close()
		totalRow := 0
		for rows.Next() {
			var totalLoc pgtype.TextArray
			err := rows.Scan(&totalLoc)
			if err != nil {
				return false, err
			}
			locs := database.FromTextArray(totalLoc)
			if !stringutil.SliceElementsMatch(locs, updatedLocations) {
				return false, fmt.Errorf("want %v currently has %v locations", updatedLocations, locs)
			}
			totalRow++
		}
		if totalRow == 0 {
			return true, fmt.Errorf("select count return 0 row")
		}
		return false, nil
	})

	return ctx, err
}

func (s *suite) tomMustUpdateConversationCorrectly(ctx context.Context) (context.Context, error) {
	req := s.Request.(*pb.EvtUserInfo)

	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var userID, conversationName pgtype.Text
	_ = userID.Set(req.UserId)
	_ = conversationName.Set(req.Name)

	query := `SELECT count(*) FROM conversations c JOIN conversation_students cs ON c.conversation_id = cs.conversation_id WHERE 
		cs.student_id = $1 AND c.name = $2`
	var count int64
	err := s.DB.QueryRow(ctx2, query, &userID, &conversationName).Scan(&count)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) usermgmtSendEventWithNewTokenAndLocationInDB(ctx context.Context, eventType string, locDesc string) (context.Context, error) {
	manabieRp := strconv.Itoa(constants_lib.ManabieSchool)
	ctx = contextWithResourcePath(ctx, manabieRp)
	var newLocs []string
	switch locDesc {
	case "new":
		newloc, _, err := s.CommonSuite.CreateLocationWithDB(ctx, manabieRp, "center", constants_lib.ManabieOrgLocation, ManabieOrgLocationType)
		if err != nil {
			return ctx, err
		}
		newLocs = []string{newloc}
	case "no":
	default:
		return ctx, fmt.Errorf("unknown locatin description %s", locDesc)
	}

	// nolint:gosec
	userID := "user-id" + strconv.Itoa(rand.Int())
	if s.studentID != "" {
		userID = s.studentID
	}
	var (
		data []byte
		subj string
	)
	switch eventType {
	case "UserDeviceToken":
		subj = constants_lib.SubjectUserDeviceTokenUpdated
		msg := &upb.EvtUserInfo{
			UserId:            userID,
			DeviceToken:       "new-token",
			AllowNotification: true,
			Name:              fmt.Sprintf("user-name-%s", userID),
			LocationIds:       newLocs,
		}
		s.Request = msg
		data2, err := googleproto.Marshal(msg)
		if err != nil {
			return ctx, err
		}
		data = data2
	case "UpdateStudent":
		subj = constants_lib.SubjectUserUpdated
		msg := &upb.EvtUser{
			Message: &upb.EvtUser_UpdateStudent_{
				UpdateStudent: &upb.EvtUser_UpdateStudent{
					StudentId:         userID,
					DeviceToken:       "new-token",
					AllowNotification: true,
					Name:              fmt.Sprintf("user-name-%s", userID),
					LocationIds:       newLocs,
				},
			},
		}
		s.Request = msg
		data2, err := googleproto.Marshal(msg)
		if err != nil {
			return ctx, err
		}
		data = data2
	default:
		return ctx, fmt.Errorf("unknown event type %s", eventType)
	}

	_, err := s.JSM.TracedPublish(ctx, "usermgmtSendEventWithNewTokenAndNewLocationInDb", subj, data)
	return ctx, err
}

func (s *suite) createUserGroupWithRoleNamesAndLocations(ctx context.Context, roleName string, locationType string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	manabieRp := strconv.Itoa(constants_lib.ManabieSchool)
	ctx = contextWithResourcePath(ctx, manabieRp)
	roleNames := []string{roleName}

	locationID, _, err := s.CommonSuite.CreateLocationWithDB(ctx, manabieRp, locationType, constants_lib.ManabieOrgLocation, ManabieOrgLocationType)
	if err != nil {
		return ctx, err
	}
	grantedLocations := []string{locationID}
	s.CommonSuite.LocationIDs = grantedLocations

	req := &upb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
	}

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.CommonSuite.BobDBTrace.Query(ctx, stmt, roleNames, len(roleNames))
	if err != nil {
		return ctx, err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return ctx, fmt.Errorf("rows.Err: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return ctx, fmt.Errorf("rows.Err: %w", err)
	}

	for _, roleID := range roleIDs {
		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&upb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: grantedLocations,
			},
		)
	}

	resp, err := upb.NewUserGroupMgmtServiceClient(s.CommonSuite.UserMgmtConn).CreateUserGroup(contextWithToken(ctx2, s.schoolAdminToken), req)
	if err != nil {
		return ctx, fmt.Errorf("CreateUserGroup: %w", err)
	}

	s.userGroupIDs = []string{}
	s.userGroupIDs = append(s.userGroupIDs, resp.GetUserGroupId())

	return ctx, nil
}

func (s *suite) checkTeachersDeactivated(ctx context.Context) (context.Context, error) {
	err := doRetry(func() (bool, error) {
		ctx2, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		var count pgtype.Int8
		query := `SELECT COUNT(*) FROM conversation_members where conversation_id = $1
			AND user_id = ANY($2)
			AND status = 'CONVERSATION_STATUS_INACTIVE'`
		if err := s.DB.QueryRow(ctx2, query, database.Text(s.conversationID), database.TextArray(s.teachersInConversation)).Scan(&count); err != nil {
			return false, err
		}

		if int(count.Int) != len(s.teachersInConversation) {
			return true, fmt.Errorf("expected there are %d inactive teachers, got %d", len(s.teachersInConversation), int(count.Int))
		}
		return false, nil
	})

	return ctx, err
}
