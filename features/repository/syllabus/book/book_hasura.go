package book

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"

	"github.com/hasura/go-graphql-client"
)

//nolint:gosec
func (s *Suite) aUserInsertSomeBooksToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := rand.Intn(5) + 2
	books, err := utils.AUserInsertSomeBooksToDatabase(ctx, s.DB, stepState.DefaultSchoolID, n)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.BookID = books[0].ID.String
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallBooksTitle(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	// no need these steps because we already declared in query_collection and tracked the tables
	// if err := utils.TrackTableForHasuraQuery(
	// 	s.HasuraAdminURL,
	// 	s.HasuraPassword,
	// 	"books",
	// ); err != nil {
	// 	return nil, fmt.Errorf("trackTableForHasuraQuery: %w", err)
	// }

	// if err := utils.CreateSelectPermissionForHasuraQuery(
	// 	s.HasuraAdminURL,
	// 	constant.UserGroupAdmin,
	// 	"books",
	// ); err != nil {
	// 	return nil, fmt.Errorf("createSelectPermissionForHasuraQuery: %w", err)
	// }

	// rawQuery := `query BooksTitle($book_id: String!) {
	// 	books(where: {book_id: {_eq: $book_id}}) {
	// 	  name
	// 	}
	//   }`

	// if err := utils.AddQueryToAllowListForHasuraQuery(s.HasuraAdminURL, s.HasuraPassword, rawQuery); err != nil {
	// 	return StepStateToContext(ctx, stepState), fmt.Errorf("addQueryToAllowListForHasuraQuery(): %v", err)
	// }

	variables := map[string]interface{}{
		"book_id": graphql.String(stepState.BookID),
	}

	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.BookTitleQuery, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnBooksTitleCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.BookTitleQuery.BookTitle) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found any book with ID %s", stepState.BookID)
	}
	actualBookName := stepState.BookTitleQuery.BookTitle[0].Name
	if actualBookName != utils.FormatName(stepState.BookID) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected book title(name): want: %s, acctual: %s", utils.FormatName(stepState.BookID), actualBookName)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
