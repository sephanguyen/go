package multitenant

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

type HasuraChapter struct {
	ID       string `graphql:"chapter_id"`
	Name     string `graphql:"name"`
	SchoolID int    `graphql:"school_id"`
}

func queryChapter(ctx context.Context, chapterID string, withAdminPermission bool) ([]*HasuraChapter, error) {
	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery("chapters"); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery("chapters"); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($chapter_id: String!){
			chapters(where: {chapter_id: {_eq: $chapter_id}}) {
			  	chapter_id
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
		Chapters []*HasuraChapter `graphql:"chapters(where: {chapter_id: {_eq: $chapter_id}})"`
	}

	variables := map[string]interface{}{
		"chapter_id": graphql.String(chapterID),
	}

	err := queryHasura(ctx, &profileQuery, variables, withAdminPermission)
	if err != nil {
		return nil, errors.Wrap(err, "queryHasura")
	}

	return profileQuery.Chapters, nil
}
