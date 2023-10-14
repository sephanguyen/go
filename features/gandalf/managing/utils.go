package managing

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/gandalf"
	bobEnt "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userOption func(u *bobEnt.User)

func withID(id string) userOption { return func(u *bobEnt.User) { u.ID = database.Text(id) } }
func withRole(group string) userOption {
	return func(u *bobEnt.User) { u.Group = database.Text(group) }
}

type genAuthTokenOption func(values url.Values)

func generateValidAuthenticationToken(endpoint, sub string) (string, error) {
	return generateAuthenticationToken(endpoint, sub, "templates/phone.template")
}

func generateAuthenticationToken(endpoint, sub, template string, opts ...genAuthTokenOption) (string, error) {
	v := url.Values{}
	v.Set("template", template)
	v.Set("UserID", sub)
	for _, opt := range opts {
		opt(v)
	}
	resp, err := http.Get(endpoint + v.Encode())
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot generate new user token, err: %v", err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()
	return string(b), nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func (s *suite) signedInAs(ctx context.Context, role string) (string, string, error) {
	if role == "" {
		return "", "", fmt.Errorf("role must be not nil")
	}
	id := idutil.ULIDNow()
	authToken, err := generateValidAuthenticationToken("http://"+firebaseAddr+"/token?", id)
	if err != nil {
		return "", "", err
	}
	return id, authToken, s.validUser(ctx, withID(id), withRole(role))
}

func (s *suite) validUser(ctx context.Context, opts ...userOption) error {
	stepState := GandalfStepStateFromContext(ctx)
	num := rand.Int()
	u := &bobEnt.User{}
	database.AllNullEntity(u)
	err := multierr.Combine(u.ID.Set(idutil.ULIDNow()), u.LastName.Set(fmt.Sprintf("valid-user-%d", num)), u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)), u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)), u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)), u.Country.Set(pb.COUNTRY_VN.String()), u.Group.Set(bobEnt.UserGroupStudent), u.DeviceToken.Set(nil), u.AllowNotification.Set(true), u.CreatedAt.Set(time.Now()), u.UpdatedAt.Set(time.Now()), u.IsTester.Set(nil))
	if err != nil {
		return err

	}
	for _, opt := range opts {
		opt(u)
	}
	cmdTag, err := database.Insert(ctx, u, s.bobDB.Exec)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot insert user for testing")
	}

	if u.Group.String == constant.UserGroupTeacher {
		schoolID := int64(stepState.GandalfStateSchool.ID.Int)
		if schoolID == 0 {
			schoolID = 1
		}
		teacher := &bobEnt.Teacher{}
		database.AllNullEntity(teacher)
		err = multierr.Combine(teacher.ID.Set(u.ID.String),
			teacher.SchoolIDs.Set([]int64{schoolID}),
			teacher.UpdatedAt.Set(time.Now()),
			teacher.CreatedAt.Set(time.Now()))
		if err != nil {
			return err

		}
		_, err = database.Insert(ctx, teacher, s.bobDB.Exec)
		if err != nil {
			return err
		}
	}
	if u.Group.String == constant.UserGroupSchoolAdmin {
		schoolID := int64(stepState.GandalfStateSchool.ID.Int)
		if schoolID == 0 {
			schoolID = 1
		}
		schoolAdminAccount := &bobEnt.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		err := multierr.Combine(schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
			schoolAdminAccount.SchoolID.Set(schoolID),
			schoolAdminAccount.UpdatedAt.Set(time.Now()),
			schoolAdminAccount.CreatedAt.Set(time.Now()))
		if err != nil {
			return err

		}
		_, err = database.Insert(ctx, schoolAdminAccount, s.bobDB.Exec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *suite) signedCtx(ctx context.Context) context.Context {
	stepState := GandalfStepStateFromContext(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.GandalfStateAuthToken)
}

func (s *suite) ExecuteWithRetry(process func() error, waitTime time.Duration, retryTime int) error {
	var count int
	var err error
	for count <= retryTime {
		err = process()
		if err == nil {
			return err
		}
		time.Sleep(waitTime)
		count++
	}
	return err
}

func (s *suite) returnsStatusCode(ctx context.Context, code string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	stt, ok := status.FromError(stepState.GandalfStateResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.GandalfStateResponseErr.Error())
	}
	if stt.Code().String() != code {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", code, stt.Code().String(), stt.Message())
	}
	return ctx, nil
}

func (s *suite) bobMustPushMsgSubjectToNats(ctx context.Context, msg, subject string) error {
	mainProcess := func() error {
		_, err := s.bobSuite.BobMustPushMsgSubjectToNats(ctx, msg, subject)
		return err
	}
	return gandalf.Execute(mainProcess, gandalf.DefaultOption...)
}

func (s *suite) bobMustPublishEventToUser_device_tokenChannel(ctx context.Context) error {
	mainProcess := func() error {
		_, err := s.bobSuite.BobMustPublishEventToUser_device_tokenChannel(ctx)
		return err
	}
	return gandalf.Execute(mainProcess, gandalf.DefaultOption...)
}

func (s *suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, err := s.bobSuite.ASignedInStudent(ctx)
	stepState.BobStepState.CurrentStudentId = bob.StepStateFromContext(ctx).CurrentUserID

	return GandalfStepStateToContext(ctx, stepState), err
}

func (s *suite) tomMustStoreMessage(ctx context.Context, messageContent string, messageType string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := "select count(message_id) from messages where conversation_id = $1 and message = $2 and type = $3"
		rows, err := s.tomDB.Query(ctx, query, stepState.GandalfStateConversationID, messageContent, messageType)
		if err != nil {
			return err

		}
		defer rows.Close()
		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err

			}
		}
		if count == 0 {
			return errors.New(fmt.Sprintf("tom must create message with conversation_id = %v, message =%v, type = %v", stepState.GandalfStateConversationID, messageContent, messageType))
		}
		return nil
	}
	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}
