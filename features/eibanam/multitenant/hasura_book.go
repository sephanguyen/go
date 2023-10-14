package multitenant

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

type HasuraBook struct {
	ID       string `graphql:"book_id"`
	Name     string `graphql:"name"`
	SchoolID int    `graphql:"school_id"`
}

func queryBooks(ctx context.Context, bookID string, withAdminPermission bool) ([]*HasuraBook, error) {
	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery("books"); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery("books"); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($book_id: String!) {
			books(where: { book_id: { _eq: $book_id } }) {
				book_id
				name
				school_id
			}
		}
		`

	if err := addQueryToAllowListForHasuraQuery(query); err != nil {
		return nil, errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly created teacher from hasura
	var profileQuery struct {
		Books []*HasuraBook `graphql:"books(where: {book_id: {_eq: $book_id}})"`
	}

	variables := map[string]interface{}{
		"book_id": graphql.String(bookID),
	}
	err := queryHasura(ctx, &profileQuery, variables, withAdminPermission)
	if err != nil {
		return nil, errors.Wrap(err, "queryHasura")
	}

	return profileQuery.Books, nil
}
