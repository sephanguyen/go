package support

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
)

func stringIn(input string, sl []string) bool {
	for _, item := range sl {
		if input == item {
			return true
		}
	}
	return false

}

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	customCtx    func(context.Context) context.Context
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
	env          string
}

func randomDBText() pgtype.Text {
	return database.Text(idutil.ULIDNow())
}
func dbNow() pgtype.Timestamptz {
	return database.Timestamptz(time.Now())
}
func dbText(str string) pgtype.Text {
	return database.Text(str)
}
func customStudentCtx(ctxt context.Context) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			DefaultRole: cpb.UserGroup_USER_GROUP_STUDENT.String(),
		},
	}
	return interceptors.ContextWithJWTClaims(ctxt, claims)
}

func customParentCtx(ctxt context.Context) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			DefaultRole: cpb.UserGroup_USER_GROUP_PARENT.String(),
		},
	}
	return interceptors.ContextWithJWTClaims(ctxt, claims)
}
