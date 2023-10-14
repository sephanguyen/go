package bob

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"

	"github.com/manabie-com/backend/internal/bob/constants"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

type genAuthTokenOption func(values url.Values)

// func generateAuthenticationToken(sub, template string, opts ...genAuthTokenOption) (string, error) {
// 	v := url.Values{}
// 	v.Set("template", template)
// 	v.Set("UserID", sub)
// 	for _, opt := range opts {
// 		opt(v)
// 	}
// 	resp, err := http.Get("http://" + firebaseAddr + "/token?" + v.Encode())
// 	if err != nil {
// 		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
// 	}
// 	b, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
// 	}
// 	resp.Body.Close()
// 	return string(b), nil
// }
func generateValidAuthenticationTokenV1(sub string) (string, error) {
	return generateAuthenticationTokenV1(sub, "templates/phone.template")
}

func generateAuthenticationTokenV1(sub, template string, opts ...genAuthTokenOption) (string, error) {
	v := url.Values{}
	v.Set("template", template)
	v.Set("UserID", sub)
	for _, opt := range opts {
		opt(v)
	}
	resp, err := http.Get("http://" + firebaseAddr + "/token?" + v.Encode())
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

func (s *suite) anInvalidAuthenticationToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = "invalid-token-duh"
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidAuthenticationTokenWithIDAlreadyExistInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	/*var id pgtype.Text
	err := s.DB.QueryRow(ctx, "SELECT student_id FROM students s JOIN users_groups ug ON s.student_id = ug.user_id WHERE ug.group_id = 'USER_GROUP_STUDENT' AND ug.status = 'USER_GROUP_STATUS_ACTIVE' AND s.deleted_at IS NULL LIMIT 1").Scan(&id)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "QueryRow")

	}

	var idStr string
	id.AssignTo(&idStr)*/
	var err error
	stepState.AuthToken, err = generateValidAuthenticationTokenV1(stepState.UserID)
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aSchoolNameCountryCityDistrict(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	city := &entities_bob.City{
		Name:    database.Text(arg3),
		Country: database.Text(arg2),
	}
	district := &entities_bob.District{
		Name:    database.Text(arg4),
		Country: database.Text(arg2),
		City:    city,
	}
	school := &entities_bob.School{
		Name:           database.Text(arg1 + stepState.Random),
		Country:        database.Text(arg2),
		City:           city,
		District:       district,
		IsSystemSchool: pgtype.Bool{Bool: true, Status: pgtype.Present},
		Point:          pgtype.Point{Status: pgtype.Null},
	}
	stepState.Schools = append(stepState.Schools, school)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) ASchoolNameCountryCityDistrict(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
	return s.aSchoolNameCountryCityDistrict(ctx, arg1, arg2, arg3, arg4)
}
func (s *suite) adminInsertsSchools(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	r := &repositories.SchoolRepo{}
	if err := r.Import(ctx, s.DB, stepState.Schools); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentSchoolID = constants.ManabieSchool
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) AdminInsertsSchools(ctx context.Context) (context.Context, error) {
	return s.adminInsertsSchools(ctx)
}

func (s *suite) bobMustCreateAnSubscription(ctx context.Context, plan string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	var trialPeriod pgtype.Text
	err = s.DB.QueryRow(
		ctx,
		"SELECT config_value FROM configs WHERE config_key = 'trialPeriod' AND config_group = 'payment' AND country = 'COUNTRY_VN'").Scan(&trialPeriod)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	days, err := strconv.Atoi(trialPeriod.String)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	var endTime pgtype.Timestamptz
	err = s.DB.QueryRow(
		ctx,
		"SELECT end_time "+
			"FROM student_subscriptions "+
			"WHERE plan_id = '"+plan+"' AND student_id = $1", t.Subject()).Scan(&endTime)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	if plan == "Trial" && endTime.Time.Sub(timeutil.Now()) < time.Duration(days)*24*time.Hour-time.Minute {
		return StepStateToContext(ctx, stepState), errors.Errorf("expecting Trial plan in next %d days", days)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentInputsActivationCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch req := stepState.Request.(type) {
	case *pb.RegisterRequest:
		req.ActivationCode = arg1 + stepState.Random
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("stepState.Request must be type *pb.RegisterRequest")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustExpireAllSubscription(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	studentID := pgtype.Text{String: t.Subject(), Status: 2}
	n := time.Now()
	now := pgtype.Timestamptz{Time: n, Status: 2}
	start := pgtype.Timestamptz{Time: n.AddDate(0, -1, 0), Status: 2}
	end := pgtype.Timestamptz{Time: n.AddDate(0, -2, 0), Status: 2}
	planID := pgtype.Text{String: "%" + arg1, Status: 2}

	query := "UPDATE student_subscriptions SET start_time = $1, end_time = $2, updated_at = $3 WHERE student_id = $4 AND plan_id LIKE $5"
	if _, err := s.DB.Exec(ctx, query, &start, &end, &now, &studentID, &planID); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "tx.ExecEx")
	}

	return StepStateToContext(ctx, stepState), nil
}
