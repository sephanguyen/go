package multitenant

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

type HasuraCourse struct {
	CourseID     string `graphql:"course_id"`
	Name         string `graphql:"name"`
	Country      string `graphql:"country"`
	Subject      string `graphql:"subject"`
	Icon         string `graphql:"icon"`
	Grade        int32  `graphql:"grade"`
	DisplayOrder int32  `graphql:"display_order"`
	SchoolId     int    `graphql:"school_id"`
}

func queryCourses(ctx context.Context, courseID string, withAdminPermission bool) ([]*HasuraCourse, error) {
	// Pre-setup for hasura query using admin secret
	if err := trackTableForHasuraQuery("courses"); err != nil {
		return nil, errors.Wrap(err, "trackTableForHasuraQuery()")
	}
	if err := createSelectPermissionForHasuraQuery("courses"); err != nil {
		return nil, errors.Wrap(err, "createSelectPermissionForHasuraQuery()")
	}

	query :=
		`
		query ($course_id: String!) {
			courses(where: {course_id: {_eq: $course_id}}) {
					course_id
					name
					country
					subject
					icon
					grade
					display_order
					school_id
				}
		}
		`

	if err := addQueryToAllowListForHasuraQuery(query); err != nil {
		return nil, errors.Wrap(err, "addQueryToAllowListForHasuraQuery()")
	}

	// Query newly created teacher from hasura
	var profileQuery struct {
		Courses []*HasuraCourse `graphql:"courses(where: {course_id: {_eq: $course_id}})"`
	}

	variables := map[string]interface{}{
		"course_id": graphql.String(courseID),
	}
	err := queryHasura(ctx, &profileQuery, variables, withAdminPermission)
	if err != nil {
		return nil, errors.Wrap(err, "queryHasura")
	}

	return profileQuery.Courses, nil
}
