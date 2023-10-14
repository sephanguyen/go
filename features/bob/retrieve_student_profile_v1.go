package bob

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *suite) retrievesStudentProfileV1(ctx context.Context, req *bpb.RetrieveStudentProfileRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpb.NewStudentReaderServiceClient(s.Conn).RetrieveStudentProfile(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRetrievesStudentProfileV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ids := append(stepState.UnAssignedStudentIDs, stepState.AssignedStudentIDs...)
	ids = append(ids, stepState.OtherStudentIDs...)
	if stepState.CurrentStudentID != "" {
		ids = append(ids, stepState.CurrentStudentID)
	}

	req := &bpb.RetrieveStudentProfileRequest{
		StudentIds: ids,
	}
	stepState.Request = req
	return s.retrievesStudentProfileV1(ctx, req)
}

func (s *suite) returnsRequestedStudentProfileV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*bpb.RetrieveStudentProfileRequest)
	resp := stepState.Response.(*bpb.RetrieveStudentProfileResponse)
	id := make([]string, 0, len(req.StudentIds))
	for _, p := range resp.Items {
		id = append(id, p.Profile.Id)
	}
	for _, studentID := range req.StudentIds {
		found := false
		for _, p := range resp.Items {
			if p.Profile.Id == studentID {
				stmt :=
					`SELECT last_login_date, full_name_phonetic FROM users WHERE user_id = $1`

				lastLoginDate := &time.Time{}
				var fullNamePhonetic pgtype.Text
				err := s.DB.QueryRow(ctx, stmt, studentID).Scan(&lastLoginDate, &fullNamePhonetic)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				if (lastLoginDate != nil && lastLoginDate.Equal(p.Profile.LastLoginDate.AsTime())) ||
					(lastLoginDate == nil && p.Profile.LastLoginDate == nil) {
					found = true
				}

				if fullNamePhonetic.String != p.Profile.FullNamePhonetic {
					found = false
				}

			}
			if p.Profile.Id == stepState.CurrentStudentID {
				continue
			}
			//if found && !isBasicProfileV1(p) {
			//  return StepStateToContext(ctx, stepState), fmt.Errorf("return profile is not basic profile")
			//}
		}

		if !found {
			return StepStateToContext(ctx, stepState), errors.Errorf("expecting return request student profile, requested id: %s", studentID)

		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func isBasicProfileV1(p *bpb.RetrieveStudentProfileResponse_Data) bool {
	return isEmptyString(p.Profile.Phone) && p.Profile.School == nil && isEmptyString(p.Profile.Email) && isEmptyString(p.Profile.Biography) && isEmptyString(p.Profile.PlanId)
}

func (s *suite) userRetrievesStudentProfileOfClassMembersV1(ctx context.Context) (context.Context, error) {
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

	req := &bpb.RetrieveStudentProfileRequest{
		StudentIds: ids,
	}

	stepState.Request = req
	return s.retrievesStudentProfileV1(ctx, req)
}

func (s *suite) teacherRetrievesAStudentProfileV1(ctx context.Context, kindOfStudent string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch kindOfStudent {
	case "has signed in before":
		query := "UPDATE users SET last_login_date = $1 WHERE user_id = $2"
		if _, err := s.DB.Exec(ctx, query, time.Now().UTC().Add(-time.Hour), &id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "newly created":
		break
	default:
		return StepStateToContext(ctx, stepState), errors.New("not supported scenario step")
	}

	req := &bpb.RetrieveStudentProfileRequest{
		StudentIds: []string{id},
	}
	stepState.Request = req
	return s.retrievesStudentProfileV1(ctx, req)
}
