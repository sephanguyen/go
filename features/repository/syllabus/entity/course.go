package entity

type GraphqlCourseStudentsListByCourseIdsQuery struct {
	CourseStudentsListByCourseIds []struct {
		StudentID string `graphql:"student_id"`
		CourseID  string `graphql:"course_id"`
	} `graphql:"course_students(order_by: {created_at: desc}, where: {course_id: {_in: $course_ids}})"`

	CourseStudentsAggregate struct {
		Aggregate struct {
			Count int `graphql:"count"`
		} `graphql:"aggregate"`
	} ` graphql:"  course_students_aggregate(where: {course_id: {_in: $course_ids}})"`
}

type GraphqlCourseStudentsListQuery struct {
	CourseStudentsList []struct {
		StudentID string `graphql:"student_id"`
		CourseID  string `graphql:"course_id"`
	} `graphql:"course_students(order_by: {created_at: desc}, where: {course_id: {_eq: $course_id}})"`
}

type GraphqlCourseStudentsListV2Query struct {
	CourseStudentsListV2 []struct {
		StudentID string `graphql:"student_id"`
		CourseID  string `graphql:"course_id"`
	} `graphql:"course_students(order_by: { created_at: desc }, where: { course_id: { _eq: $course_id } }, limit: $limit, offset: $offset)"`
	CourseStudentsAggregate struct {
		Aggregate struct {
			Count int `graphql:"count"`
		} `graphql:"aggregate"`
	} ` graphql:"course_students_aggregate(where: { course_id: { _eq: $course_id } })"`
}
