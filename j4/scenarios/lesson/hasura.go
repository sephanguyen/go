package lesson

import (
	"github.com/manabie-com/backend/j4/serviceutil"
)

var (
	hasuraQueries = []serviceutil.HasuraQuery{
		{
			Name:  "Lesson_CoursesList",
			Query: Lesson_CoursesList,
			VariablesCreator: func() map[string]interface{} {
				return map[string]interface{}{
					"limit":  10,
					"offset": 0,
				}
			},
		},
	}
	Lesson_CoursesList = `
          query Lesson_CoursesList($name: String, $course_id: [String!],
          $course_type: String, $limit: Int = 10, $offset: Int = 0) {
            courses(
              limit: $limit
              offset: $offset
              order_by: {created_at: desc, display_order: asc, name: asc, course_id: asc}
              where: {name: {_ilike: $name}, course_id: {_in: $course_id}, course_type: {_eq: $course_type}}
            ) {
              ...Lesson_CoursesAttrs
            }
            courses_aggregate(
              where: {name: {_ilike: $name}, course_id: {_in: $course_id}, course_type: {_eq: $course_type}}
            ) {
              aggregate {
                count
              }
            }
          }


          fragment Lesson_CoursesAttrs on courses {
            course_id
            name
            icon
            grade
            subject
            country
            school_id
            display_order
            teaching_method
          }`
)
