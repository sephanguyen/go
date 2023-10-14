package common

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var courses = [][]interface{}{{"course-1", "Course 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2", "Course 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-3", "Course 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-1", "Course teacher 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-2", "Course teacher 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-3", "Course teacher 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-1-JP", "Course 1 JP name", "COUNTRY_JP", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2-JP", "Course 2 JP name", "COUNTRY_JP", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-1-SG", "Course 1 SG name", "COUNTRY_SG", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2-SG", "Course 2 SG name", "COUNTRY_SG", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-1-ID", "Course 1 ID name", "COUNTRY_ID", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-2-ID", "Course 2 ID name", "COUNTRY_ID", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-live-1", "Course live 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_LIVE", "course-live-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-2", "Course live 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-2-plan", "2020-07-02", "2025-07-02"}, {"course-live-3", "Course live 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_LIVE", "course-live-3-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-1", "Course live teacher 1 name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 12, 1, nil, "COURSE_TYPE_LIVE", "course-live-teacher-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-2", "Course live teacher 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-2-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-3", "Course live teacher 3 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, "2000-05-10T17:55:56z", "COURSE_TYPE_LIVE", "course-live-teacher-3-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-4", "Course live teacher 4 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-4-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-5", "Course live teacher 5 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-5-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-6", "Course live teacher 6 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-6-plan", "2020-07-02", "2025-07-02"}, {"course-live-teacher-7", "Course live teacher 7 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-teacher-7-plan", "2020-07-02", "2025-07-02"}, {"course-live-dont-have-lesson-1", "Course live dont have lesson 1 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-dont-have-lesson-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-dont-have-lesson-2", "Course live dont have lesson 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-dont-have-lesson-1-plan", "2020-07-02", "2025-07-02"}, {"course-live-complete-lesson-1", "Course live complete lesson 1 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-complete-lesson-1-plan", "2020-07-02", "2020-07-03"}, {"course-live-complete-lesson-2", "Course live complete lesson 2 name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_LIVE", "course-live-complete-lesson-2-plan", "2019-07-02", "2019-07-03"}, {"course-dont-have-chapter-1-JP", "Course dont have chapter 1 JP name", "COUNTRY_JP", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-dont-have-chapter-1-VN", "Course dont have chapter 1 VN name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-have-chapter-dont-exist-1-VN", "Course have chapter dont exist 1 VN name", "COUNTRY_VN", "SUBJECT_BIOLOGY", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-have-book-1-JP", "Course have book 1 JP name", "COUNTRY_JP", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-2-VN", "Course teacher have book 2 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-4-VN", "Course teacher have book 3 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-5-VN", "Course teacher have book 4 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}, {"course-teacher-have-book-missing-subject-grade", "Course teacher have book 6 VN name", "COUNTRY_VN", "SUBJECT_MATHS", 10, 2, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-02", "2025-07-02"}}
var JPREPWhitelistedCourses = [][]interface{}{
	{"JPREP_COURSE_000000162", "JPREP Whitelisted Course Name 1", "COUNTRY_JP", "SUBJECT_ENGLISH", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-05T00:00:00Z", "2020-07-07T00:00:00Z", "-2147483647", nil},
	{"JPREP_COURSE_000000218", "JPREP Whitelisted Course Name 2", "COUNTRY_JP", "SUBJECT_ENGLISH", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-05T00:00:00Z", "2020-07-07T00:00:00Z", "-2147483647", nil},
	{"JPREP_COURSE_000000163", "JPREP Whitelisted Course Name 3", "COUNTRY_JP", "SUBJECT_ENGLISH", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-07-05T00:00:00Z", "2020-07-07T00:00:00Z", "-2147483647", nil},
}
var JPREPBlacklistedCourses = [][]interface{}{
	{"JPREP_COURSE_000000312", "JPREP Blacklisted Course Name 1", "COUNTRY_JP", "SUBJECT_ENGLISH", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-03-10 14:30:00.000 +0700", "2022-05-23 20:43:22.764 +0700", "-2147483647", nil},
	{"JPREP_COURSE_000000313", "JPREP Blacklisted Course Name 2", "COUNTRY_JP", "SUBJECT_ENGLISH", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-03-10 14:30:00.000 +0700", "2022-05-23 20:43:22.764 +0700", "-2147483647", nil},
	{"JPREP_COURSE_000000314", "JPREP Blacklisted Course Name 3", "COUNTRY_JP", "SUBJECT_ENGLISH", 12, 1, nil, "COURSE_TYPE_CONTENT", nil, "2020-03-10 14:30:00.000 +0700", "2022-05-23 20:43:22.764 +0700", "-2147483647", nil},
}

const (
	Name            = "name"
	Country         = "country"
	SchoolID        = "schoolID"
	Subject         = "subject"
	Grade           = "grade"
	CountryAndGrade = "country and grade"
	All             = "all"
	None            = "none"
)

func (s *suite) newID() string {
	return idutil.ULIDNow()
}

func (s *suite) ARandomNumber(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Random = strconv.Itoa(rand.Int())
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) SignedCtx(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}

func (s *suite) CtxWithAuthToken(ctx context.Context, authtoken string) context.Context {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = authtoken
	return StepStateToContext(ctx, stepState)
}

func contextWithToken(s *suite, ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
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
func intResourcePathFromCtx(ctx context.Context) int64 {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		rp := claim.Manabie.ResourcePath
		intrp, err := strconv.ParseInt(rp, 10, 64)
		if err != nil {
			panic(err)
		}
		return intrp
	}
	panic("ctx has no resource path")
}

func hasResourcePath(ctx context.Context) bool {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		return claim.Manabie.ResourcePath != ""
	}
	return false
}
func resourcePathFromCtx(ctx context.Context) string {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim != nil && claim.Manabie != nil {
		return claim.Manabie.ResourcePath
	}
	panic("ctx has no resource path")
}

func (s *suite) ReturnsStatusCodeWithExamInfo(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	return s.ReturnsStatusCode(ctx, arg1)
}

func (s *suite) ReturnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}
	return ctx, nil
}

func (s *suite) generateAnAdminToken(ctx context.Context) (context.Context, string) {
	id := s.newID()
	var err error
	ctx, _ = s.aValidUser(ctx, withID(id), withRole(bob_entities.UserGroupAdmin))
	token, err := s.GenerateExchangeToken(id, bob_entities.UserGroupAdmin)
	if err != nil {
		return ctx, ""
	}
	return ctx, token
}

func UpdateResourcePath(db *pgxpool.Pool) error {
	ctx := context.Background()
	query := `UPDATE school_configs SET resource_path = '1';
	UPDATE schools SET resource_path = '1';
	UPDATE configs SET resource_path = '1';
	UPDATE cities SET resource_path = '1';
	UPDATE districts SET resource_path = '1';`
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "1",
			DefaultRole:  bob_entities.UserGroupAdmin,
			UserGroup:    bob_entities.UserGroupAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	_, err := db.Exec(ctx, query)
	return err
}

func ValidContext(ctx context.Context, orgID int, userID string, token string) context.Context {
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: strconv.Itoa(orgID),
			UserID:       userID,
		},
	})
	return ContextWithTokenV2(ctx, token)
}

func ContextWithTokenV2(ctx context.Context, token string) context.Context {
	return helper.GRPCContext(ctx, "token", token)
}

func (s *suite) ContextHasToken(ctx context.Context) bool {
	st := StepStateFromContext(ctx)
	return st.AuthToken != ""
}

func (s *suite) ContextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}
