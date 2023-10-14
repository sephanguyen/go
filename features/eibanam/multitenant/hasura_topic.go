package multitenant

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

type HasuraTopic struct {
	ID       string `graphql:"topic_id"`
	Name     string `graphql:"name"`
	SchoolID int    `graphql:"school_id"`
}

func queryTopics(ctx context.Context, topicID string, withAdminPermission bool) ([]*HasuraTopic, error) {
	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery("topics"); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery("topics"); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($topicID: String!){
			topics(where: {topic_id: {_eq: $topicID}}) {
			  	topic_id
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
		Topics []*HasuraTopic `graphql:"topics(where: {topic_id: {_eq: $topicID}})"`
	}

	variables := map[string]interface{}{
		"topicID": graphql.String(topicID),
	}

	err := queryHasura(ctx, &profileQuery, variables, withAdminPermission)
	if err != nil {
		return nil, errors.Wrap(err, "queryHasura")
	}

	return profileQuery.Topics, nil
}
