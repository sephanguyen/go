package fatima

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
)

func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("generateAuthenticationToken: cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("generateAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(b), nil
}

func (s *suite) aSignedIn(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var userGroup string
	switch user {
	case "teacher":
		userGroup = entity.UserGroupTeacher
	case "school admin":
		userGroup = entity.UserGroupSchoolAdmin
	}

	stepState.UserID = newID()
	ctx, err := s.aValidToken(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidToken: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidToken(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidUserInDB(ctx, withID(stepState.UserID), withRole(userGroup))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUserInDB: %w", err)
	}

	if err := checkUserExisted(ctx, s.DB, stepState.UserID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("user isn't existed: %w", err)
	}
	token, err := s.generateExchangeToken(stepState.UserID, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken: %w", err)
	}

	s.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentPackageUpsertedV2Subscribe() error {
	s.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{nats.StartTime(time.Now()), nats.ManualAck(), nats.AckWait(2 * time.Second)},
	}
	handlerStudentPackageUpserted := func(ctx context.Context, data []byte) (bool, error) {
		eventStudentPackageV2 := &npb.EventStudentPackageV2{}
		if err := proto.Unmarshal(data, eventStudentPackageV2); err != nil {
			return false, err
		}
		switch req := s.Request.(type) {
		case *pb.AddStudentPackageCourseRequest:
			locationIDs := make([]string, 0)
			classIDs := make([]string, 0)
			for _, studentPackage := range req.StudentPackageExtra {
				locationIDs = append(locationIDs, studentPackage.LocationId)
				classIDs = append(classIDs, studentPackage.ClassId)
			}
			if req.StudentId == s.StudentID && golibs.InArrayString(eventStudentPackageV2.StudentPackage.Package.LocationId, locationIDs) && golibs.InArrayString(eventStudentPackageV2.StudentPackage.Package.ClassId, classIDs) {
				s.FoundChanForJetStream <- eventStudentPackageV2
				return true, nil
			}
		case *pb.EditTimeStudentPackageRequest:
			locationIDs := make([]string, 0)
			classIDs := make([]string, 0)
			for _, studentPackage := range req.StudentPackageExtra {
				locationIDs = append(locationIDs, studentPackage.LocationId)
				classIDs = append(classIDs, studentPackage.ClassId)
			}
			if golibs.InArrayString(eventStudentPackageV2.StudentPackage.Package.LocationId, locationIDs) && golibs.InArrayString(eventStudentPackageV2.StudentPackage.Package.ClassId, classIDs) {
				s.FoundChanForJetStream <- eventStudentPackageV2
				return true, nil
			}
		}
		return false, nil
	}
	subs, err := s.JSM.Subscribe(constants.SubjectStudentPackageV2EventNats, opts, handlerStudentPackageUpserted)
	if err != nil {
		return fmt.Errorf("createStudentPackageUpsertedV2Subscribe: s.JSM.Subscribe: %w", err)
	}
	s.Subs = append(s.Subs, subs.JetStreamSub)
	return nil
}

func (s *suite) generateExchangeToken(userID, userGroup string) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", fmt.Errorf("error when create generateValidAuthenticationToken: %v", err)
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.ApplicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return "", fmt.Errorf("error when create exchange token: %v", err)
	}

	return token, nil
}

func (s *suite) aValidUserInDB(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}
	ctx = golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

	user, err := newUserEntity()
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newUserEntity")
	}

	for _, opt := range opts {
		opt(user)
	}

	err = database.ExecInTx(ctx, s.BobDB, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := repository.UserRepo{}
		if err := userRepo.Create(ctx, tx, user); err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}

		switch user.Group.String {
		case constant.UserGroupTeacher:
			teacherRepo := repository.TeacherRepo{}
			teacher := &entity.Teacher{}
			database.AllNullEntity(teacher)
			teacher.ID = user.ID
			err := multierr.Combine(
				teacher.SchoolIDs.Set([]int64{schoolID}),
				teacher.ResourcePath.Set(fmt.Sprint(schoolID)),
			)
			if err != nil {
				return err
			}

			err = teacherRepo.CreateMultiple(ctx, tx, []*entity.Teacher{teacher})
			if err != nil {
				return fmt.Errorf("cannot create teacher: %w", err)
			}

		case constant.UserGroupSchoolAdmin:
			schoolAdminRepo := repository.SchoolAdminRepo{}
			schoolAdmin := &entity.SchoolAdmin{}
			database.AllNullEntity(schoolAdmin)

			err := multierr.Combine(
				schoolAdmin.SchoolAdminID.Set(user.ID.String),
				schoolAdmin.SchoolID.Set(schoolID),
				schoolAdmin.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
			)
			if err != nil {
				return fmt.Errorf("cannot create school admin: %w", err)
			}
			if err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entity.SchoolAdmin{schoolAdmin}); err != nil {
				return err
			}
		}

		userGroup := &entity.UserGroup{}
		database.AllNullEntity(userGroup)

		err = multierr.Combine(
			userGroup.GroupID.Set(user.Group.String),
			userGroup.UserID.Set(user.ID.String),
			userGroup.IsOrigin.Set(true),
			userGroup.Status.Set(entity.UserGroupStatusActive),
			userGroup.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
		)
		if err != nil {
			return err
		}

		userGroupRepo := &repository.UserGroupRepo{}
		if err = userGroupRepo.Upsert(ctx, tx, userGroup); err != nil {
			return fmt.Errorf("userGroupRepo.Upsert: %w %s", err, user.Group.String)
		}

		return nil
	})

	return StepStateToContext(ctx, stepState), err
}

func checkUserExisted(ctx context.Context, db database.QueryExecer, userID string) error {
	retryTimes := 3
	err := try.Do(func(attempt int) (retry bool, err error) {
		if _, err = (&repository.UserRepo{}).FindByIDUnscope(ctx, db, database.Text(userID)); err == nil {
			return false, nil
		}
		if attempt < retryTimes {
			time.Sleep(time.Second)
			return false, nil
		}
		return true, err
	})

	return err
}
