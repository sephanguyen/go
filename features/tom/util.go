package tom

import (
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

const (
	ManabieOrgLocationType = "01FR4M51XJY9E77GSN4QZ1Q9M1"
	JPREPOrgLocationType   = "01FR4M51XJY9E77GSN4QZ1Q9M2"
)

func doRetry(f func() (bool, error)) error {
	return try.Do(func(attempt int) (bool, error) {
		retry, err := f()
		if err != nil {
			if retry {
				time.Sleep(2 * time.Second)
				return attempt < 10, err
			}
			return false, err
		}
		return false, nil
	})
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

func contextWithResourcePath(ctx context.Context, rp string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				UserGroup: "USER_GROUP_SCHOOL_ADMIN",
			},
		}
	}
	claim.Manabie.ResourcePath = rp
	return interceptors.ContextWithJWTClaims(ctx, claim)
}

func getSchoolDefaultLocation(schoolID string) string {
	switch schoolID {
	case strconv.Itoa(constants.ManabieSchool):
		return constants.ManabieOrgLocation
	case strconv.Itoa(constants.JPREPSchool):
		return constants.JPREPOrgLocation
	default:
		panic(fmt.Sprintf("cannot find default location or school %s", schoolID))
	}
}

func getSchoolDefaultLocationType(schoolID string) string {
	switch schoolID {
	case strconv.Itoa(constants.ManabieSchool):
		return ManabieOrgLocationType
	case strconv.Itoa(constants.JPREPSchool):
		return JPREPOrgLocationType
	default:
		panic(fmt.Sprintf("cannot find default location or school %s", schoolID))
	}
}

func int32ResourcePathFromCtx(ctx context.Context) int32 {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		rp := claim.Manabie.ResourcePath
		intrp, err := strconv.ParseInt(rp, 10, 32)
		if err != nil {
			panic(err)
		}
		return int32(intrp)
	}
	panic("ctx has no resource path")
}

func resourcePathFromCtx(ctx context.Context) string {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		return claim.Manabie.ResourcePath
	}
	panic("ctx has no resource path")
}
func setupRls(ctx context.Context, pgdb *pgxpool.Pool) error {
	st := `
CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
RETURNS BOOLEAN 
AS $$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$$  LANGUAGE SQL IMMUTABLE;
	`
	_, err := pgdb.Exec(ctx, st)
	if err != nil {
		return err
	}
	tables := []string{"online_users", "messages", "user_device_tokens", "conversation_lesson", "conversation_locations", "conversation_members",
		"conversation_students", "conversations", "location_types", "locations"}
	creatingPolicies := map[string]string{}
	for _, item := range tables {
		policyname := fmt.Sprintf("rls_%s", item)
		creatingPolicies[policyname] = item
	}

	for policyname, table := range creatingPolicies {
		stmt := fmt.Sprintf(`CREATE POLICY %s ON "%s" using (permission_check(resource_path, '%s')) with check (permission_check(resource_path, '%s'))`,
			policyname, table, table, table)
		_, err = pgdb.Exec(ctx, stmt)
		if err != nil {
			if pgerr, ok := err.(*pgconn.PgError); ok {
				if pgerr.Code != pgerrcode.DuplicateObject {
					return err
				}
			} else {
				return err
			}
		}

		stmt = fmt.Sprintf(`ALTER TABLE %s ENABLE ROW LEVEL security;`, table)
		_, err = pgdb.Exec(ctx, stmt)
		if err != nil {
			return err
		}

		stmt = fmt.Sprintf(`ALTER TABLE %s FORCE ROW LEVEL security;`, table)
		_, err = pgdb.Exec(ctx, stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

func i32ToStr(i int32) string {
	return strconv.Itoa(int(i))
}

func strToI32(str string) int32 {
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		panic(err)
	}
	return int32(i)
}

func EnsureSchoolAdminToken(ctx context.Context, s *common.Suite) (context.Context, error) {
	if !s.ContextHasToken(ctx) {
		ctx2, err := s.ASignedInWithSchool(ctx, "school admin", i32resourcePathFromCtx(ctx))
		if err != nil {
			return ctx, err
		}
		return ctx2, nil
	}
	return ctx, nil
}

func i32resourcePathFromCtx(ctx context.Context) int32 {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		rp := claim.Manabie.ResourcePath
		ret, err := strconv.ParseInt(rp, 10, 32)
		if err != nil {
			panic(fmt.Errorf("ctx has invalid resource path %w", err))
		}
		return int32(ret)
	}
	panic("ctx has no resource path")
}
