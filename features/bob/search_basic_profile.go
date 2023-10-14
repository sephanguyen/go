package bob

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func (s *suite) aListUserValidInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.NumberOfId = 5
	stepState.StudentIds = make([]string, 5)
	for i := 0; i < stepState.NumberOfId; i++ {
		id := s.newID()
		if ctx, err = s.aValidStudentInDB(ctx, id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.StudentIds[i] = id
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getUserIds(ctx context.Context, limit int) ([]string, error) {
	ids := make([]string, 0, limit)
	query := fmt.Sprintf(`SELECT user_id from "users" order by user_id DESC limit %d`, limit)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return ids, nil
}
func (s *suite) updateAStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SearchText = "new-name"
	ids := stepState.StudentIds
	var userId string = ids[0]
	stmt := `UPDATE "users" SET name = $1 WHERE user_id =$2`
	_, err := db.Exec(ctx, stmt, stepState.SearchText+userId, userId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.ExpectedStudentID = userId
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) searchBasicProfile(ctx context.Context, args string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedInStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ids := stepState.StudentIds
	search_text := &wrapperspb.StringValue{Value: stepState.SearchText}

	if args == "ids" {
		search_text = nil
	}
	resp, err := pb.NewUserReaderServiceClient(s.Conn).SearchBasicProfile(s.signedCtx(ctx), &pb.SearchBasicProfileRequest{
		UserIds:    ids,
		SearchText: search_text,
		Paging:     &cpb.Paging{Limit: 10},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) searchBasicProfileMustReturnCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*pb.SearchBasicProfileResponse)
	if len(response.Profiles) != stepState.NumberOfId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the length of response have to equal with the student ids, got %d, expected %d",
			len(response.GetProfiles()), len(stepState.StudentIds))
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) searchBasicProfileMustReturnCorrectlyWithSearchText(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*pb.SearchBasicProfileResponse)
	if len(response.Profiles) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("got none profile")
	}
	if response.Profiles[0].UserId != stepState.ExpectedStudentID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the length of response have to equal with the current student ids, got %s, expected %s", response.Profiles[0].UserId, stepState.ExpectedStudentID)
	}
	return StepStateToContext(ctx, stepState), nil
}
