package bob

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) retrievesStudentProfile(ctx context.Context, req *pb.GetStudentProfileRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).GetStudentProfile(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userRetrievesStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ids := append(stepState.UnAssignedStudentIDs, stepState.AssignedStudentIDs...)
	ids = append(ids, stepState.OtherStudentIDs...)
	if stepState.CurrentStudentID != "" {
		ids = append(ids, stepState.CurrentStudentID)
	}

	req := &pb.GetStudentProfileRequest{
		StudentIds: ids,
	}
	stepState.Request = req
	return s.retrievesStudentProfile(ctx, req)
}

func (s *suite) anOtherStudentProfileInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	orgtoken := stepState.AuthToken
	if ctx, err := s.aValidStudentInDBV1(ctx, id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = orgtoken

	stepState.OtherStudentIDs = append(stepState.OtherStudentIDs, id)
	stepState.UserID = id

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidStudentWithNameV1(ctx context.Context, id, name string) (context.Context, error) {
	if ctx, err := s.createStudentWithName(ctx, id, name); err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.CurrentUserID = id
	stepState.studentID = id
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidStudentInDBV1(ctx context.Context, id string) (context.Context, error) {
	return s.aValidStudentWithNameV1(ctx, id, "")
}
func (s *suite) createStudentWithName(ctx context.Context, id, name string) (context.Context, error) {
	num := s.newID()
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	if name == "" {
		name = fmt.Sprintf("valid-student-%s", num)
	}

	student := &entities.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	database.AllNullEntity(&student.User.AppleUser)

	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(name),
		student.Country.Set(pb.COUNTRY_VN.String()),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", num)),
		student.Email.Set(fmt.Sprintf("email+%s@example.com", num)),
		student.FullNamePhonetic.Set("phonetic"+name),
		student.CurrentGrade.Set(12),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.SchoolID.Set(stepState.CurrentSchoolID),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := (&repositories.StudentRepo{}).Create(ctx, tx, student); err != nil {
			return errors.Wrap(err, "s.StudentRepo.CreateTx")
		}

		if student.AppleUser.ID.String != "" {
			if err := (&repositories.AppleUserRepo{}).Create(ctx, tx, &student.AppleUser); err != nil {
				return errors.Wrap(err, "s.AppleUserRepo.Create")
			}
		}
		return nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidStudentWithName(ctx context.Context, id, name string) (context.Context, error) {
	if ctx, err := s.createStudentWithName(ctx, id, name); err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.CurrentUserID = id
	stepState.studentID = id
	return StepStateToContext(ctx, stepState), nil
}

// func (s *suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
// 	return s.aValidStudentWithName(ctx, id, "")
// }
func (s *suite) anAssignedStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()

	s.aValidStudentInDB(ctx, id)

	var studentID pgtype.Text
	studentID.Set(id)

	stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, id)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) anUnassignedStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.UnAssignedStudentIDs = append(stepState.UnAssignedStudentIDs, id)
	return StepStateToContext(ctx, stepState), nil
}
func isEmptyString(s string) bool { return s == "" }
func isBasicProfile(p *pb.GetStudentProfileResponse_Data) bool {
	return isEmptyString(p.Profile.Phone) && p.Profile.School == nil && isEmptyString(p.Profile.Email) && isEmptyString(p.Profile.Biography) && isEmptyString(p.Profile.PlanId)
}
func (s *suite) returnsRequestedStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.GetStudentProfileRequest)
	resp := stepState.Response.(*pb.GetStudentProfileResponse)

	for _, studentID := range req.StudentIds {
		found := true
		for _, p := range resp.Datas {
			if p.Profile.Id == studentID {
				found = true
			}
			if p.Profile.Id == stepState.CurrentStudentID {
				continue
			}
			if found && !isBasicProfile(p) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("return profile is not basic profile")
			}
		}

		if !found {
			return StepStateToContext(ctx, stepState), errors.Errorf("expecting return request student profile, requested id: %s", studentID)

		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsEmptyListOfStudentProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.Response.(*pb.GetStudentProfileResponse).Datas) != 0 {
		return StepStateToContext(ctx, stepState), errors.New("expecting empty list of profile")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentRetrievesHisherOwnProfile(ctx context.Context) (context.Context, error) {
	return s.retrievesStudentProfile(ctx, &pb.GetStudentProfileRequest{StudentIds: nil})
	// emtpty return current user profile
}
func (s *suite) returnsHisherOwnProfile(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetStudentProfileResponse)
	if len(resp.Datas) != 1 {
		return StepStateToContext(ctx, stepState), errors.New("expecting only profile")
	}

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	if t.Subject() != resp.Datas[0].Profile.Id {
		return StepStateToContext(ctx, stepState), errors.New("expecting student's profile with same ID")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aFindStudentRequestWithPhone(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := new(pb.FindStudentRequest)

	if arg1 == "valid" {
		studentId := s.newID()
		ctx, err := s.aValidStudentInDB(ctx, studentId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		a := pgtype.TextArray{}
		a.Set([]string{studentId})

		userRepo := &repositories.UserRepo{}
		students, err := userRepo.Retrieve(ctx, s.DB, a)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}

		if len(students) == 1 {
			req.Phone = students[0].PhoneNumber.String
		}
	} else {
		req.Phone = "invalid-phone"
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userRetrievesProfileStudentByPhone(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).FindStudent(s.signedCtx(ctx), stepState.Request.(*pb.FindStudentRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsStudentProfileOwnProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.FindStudentResponse)
	if stepState.Request.(*pb.FindStudentRequest).Phone != resp.Profile.Phone {
		return StepStateToContext(ctx, stepState), errors.New("profile does not match phone number")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentWithOnTrialAndBillingDate(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	studentRepo := &repositories.StudentRepo{
		CreateSchoolFn: nil,
	}

	studentIdPg := pgtype.Text{}
	_ = studentIdPg.Set(t.Subject())

	student, _ := studentRepo.Find(ctx, s.DB, studentIdPg)

	if arg1 == "true" {
		_ = student.OnTrial.Set(true)
	} else {
		_ = student.OnTrial.Set(false)
	}

	if arg2 == "expired" {
		student.BillingDate.Time = time.Now().Add(-2 * 24 * time.Hour)
	} else {
		student.BillingDate.Time = time.Now().Add(2 * 24 * time.Hour)
	}

	_ = studentRepo.Update(ctx, s.DB, student)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userRetrievesStudentProfileWithTokenId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).GetStudentProfile(contextWithToken(s, ctx), &pb.GetStudentProfileRequest{
		StudentIds: []string{t.Subject()},
	})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) statusWithErrorDetailTypeSubject(ctx context.Context, typ, subject string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, d := range status.Convert(stepState.ResponseErr).Details() {
		failure, ok := d.(*errdetails.PreconditionFailure)
		if !ok {
			return StepStateToContext(ctx, stepState), errors.New("expecting PreconditionFailure")
		}

		for _, v := range failure.GetViolations() {
			if v.GetType() == typ && v.GetSubject() == subject {
				return StepStateToContext(ctx, stepState), nil
			}
		}
	}

	return StepStateToContext(ctx, stepState), errors.New("could not found expected error detail")
}
func (s *suite) returnsListOfStudentProfileWithEmpty(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetStudentProfileResponse)
	if len(resp.Datas) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("no data return")
	}

	for _, u := range resp.Datas {
		if strings.Contains(arg1, "email") && u.Profile.Email != "" {
			return StepStateToContext(ctx, stepState), errors.New("email should be empty")
		}

		if strings.Contains(arg1, "phone") && u.Profile.Phone != "" {
			return StepStateToContext(ctx, stepState), errors.New("phone should be empty")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userRetrievesStudentProfileOfClassMembers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.RetrieveClassMemberRequest{
		ClassId: stepState.CurrentClassID,
	}

	ctx, err := s.userRetrieveClassMember(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	ids := []string{}
	resp := stepState.Response.(*pb.RetrieveClassMemberResponse)
	for _, m := range resp.Members {
		if m.UserGroup != pb.USER_GROUP_STUDENT {
			continue
		}

		ids = append(ids, m.UserId)
	}

	req := &pb.GetStudentProfileRequest{
		StudentIds: ids,
	}

	stepState.Request = req
	return s.retrievesStudentProfile(ctx, req)
}
