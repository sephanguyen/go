package fatima

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
)

type userOption func(u *entity.LegacyUser)

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(s *suite, ctx context.Context) context.Context {
	ctx = contextWithValidVersion(ctx)
	return s.signedCtx(ctx)
}

func withID(id string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.Group.Set(group)
	}
}

func newUserEntity() (*entity.LegacyUser, error) {
	userID := newID()
	now := time.Now()
	user := new(entity.LegacyUser)
	firstName := fmt.Sprintf("user-first-name-%s", userID)
	lastName := fmt.Sprintf("user-last-name-%s", userID)
	fullName := helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)
	database.AllNullEntity(user)
	database.AllNullEntity(&user.AppleUser)
	if err := multierr.Combine(
		user.ID.Set(userID),
		user.Email.Set(fmt.Sprintf("valid-user-%s@email.com", userID)),
		user.Avatar.Set(fmt.Sprintf("http://valid-user-%s", userID)),
		user.IsTester.Set(false),
		user.FacebookID.Set(userID),
		user.PhoneVerified.Set(false),
		user.AllowNotification.Set(true),
		user.EmailVerified.Set(false),
		user.FullName.Set(fullName),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.Country.Set(cpb.Country_COUNTRY_VN.String()),
		user.Group.Set(entity.UserGroupStudent),
		user.Birthday.Set(now),
		user.Gender.Set(pb.Gender_FEMALE.String()),
		user.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
		user.DeletedAt.Set(nil),
	); err != nil {
		return nil, errors.Wrap(err, "set value user")
	}

	user.UserAdditionalInfo = entity.UserAdditionalInfo{
		CustomClaims: map[string]interface{}{
			"external-info": "example-info",
		},
	}
	return user, nil
}

func newID() string {
	return idutil.ULIDNow()
}

func newStudentEntity() (*entity.LegacyStudent, error) {
	id := newID()
	name := "valid-user-import-by-fatima" + id

	student := &entity.LegacyStudent{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)
	database.AllNullEntity(&student.LegacyUser.AppleUser)

	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(name),
		student.FullName.Set(name),
		student.FirstName.Set(name),
		student.LastName.Set(name),
		student.Country.Set(cpb.Country_COUNTRY_VN.String()),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", id)),
		student.Email.Set(fmt.Sprintf("email+%s", id)),
		student.CurrentGrade.Set(12),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.SchoolID.Set(constants.ManabieSchool),
		student.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
	)

	if err != nil {
		return nil, err
	}

	return student, nil
}
