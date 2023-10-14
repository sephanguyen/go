package usermgmt

import (
	"github.com/manabie-com/backend/j4/serviceutil"
)

var (
	hasuraQueries = []serviceutil.HasuraQuery{
		{
			Name:  "User_StaffListV4",
			Query: User_StaffListV4,
			VariablesCreator: func() map[string]interface{} {
				return map[string]interface{}{
					"offset": 0,
					"limit":  10,
				}
			},
		},
		{
			Name:             "User_CountStudentWithLocationsFilterV5",
			Query:            User_CountStudentWithLocationsFilterV5,
			VariablesCreator: func() map[string]interface{} { return nil },
		},
		{
			Name:  "User_GetManyStudentLocationsFiltersV5",
			Query: User_GetManyStudentLocationsFiltersV5,
			VariablesCreator: func() map[string]interface{} {
				return map[string]interface{}{
					"offset": 0,
					"limit":  10,
				}
			},
		},
		{
			Name:  "User_UserGroupListV2",
			Query: User_UserGroupListV2,
			VariablesCreator: func() map[string]interface{} {
				return map[string]interface{}{
					"offset": 0,
					"limit":  10,
				}
			},
		},
	}
	User_StaffListV4 = `
          query User_StaffListV4($limit: Int = 10, $offset: Int = 0, $user_name:
          String) {
            staff(
              limit: $limit
              offset: $offset
              order_by: {created_at: desc}
              where: {user: {name: {_ilike: $user_name}}}
            ) {
              staff_id
              user {
                email
                name
                resource_path
                user_group_members {
                  user_group {
                    user_group_id
                    user_group_name
                  }
                }
              }
            }
            staff_aggregate(where: {user: {name: {_ilike: $user_name}}}) {
              aggregate {
                count
              }
            }
          }`
	User_UserGroupListV2 = `
         query User_UserGroupListV2($limit: Int = 10, $offset: Int = 0,
          $is_system: Boolean = false) {
            user_group(
              limit: $limit
              offset: $offset
              where: {is_system: {_eq: $is_system}}
              order_by: {created_at: desc}
            ) {
              user_group_id
              user_group_name
            }
            user_group_aggregate(where: {is_system: {_eq: $is_system}}) {
              aggregate {
                count
              }
            }
          }`
	User_CountStudentWithLocationsFilterV5 = `
          query User_CountStudentWithLocationsFilterV5($keyword: String,
          $grades: [smallint!], $grade_ids: [String!], $student_ids: [String!],
          $enrollment_status: String, $last_login_date: Boolean, $location_ids:
          [String!], $student_ids_by_phone_number: [String!] = []) {
            users_aggregate(
              where: {_or: [{name: {_ilike: $keyword}}, {full_name_phonetic: {_ilike: $keyword}}, {user_id: {_in: $student_ids_by_phone_number}}], user_id: {_in: $student_ids}, user_group: {_eq: "USER_GROUP_STUDENT"}, last_login_date: {_is_null: $last_login_date}, student: {current_grade: {_in: $grades}, grade_id: {_in: $grade_ids}, enrollment_status: {_eq: $enrollment_status}}, user_access_paths: {location_id: {_in: $location_ids}}}
            ) {
              aggregate {
                count
              }
            }
          }`
	User_GetManyStudentLocationsFiltersV5 = `
query User_GetManyStudentLocationsFiltersV5($keyword: String, $limit:
          Int = 10, $offset: Int = 0, $order_by: users_order_by! = {created_at:
          desc}, $student_ids: [String!], $grades: [smallint!], $grade_ids:
          [String!], $enrollment_status: String, $last_login_date: Boolean,
          $location_ids: [String!], $student_ids_by_phone_number: [String!] =
          []) {
            users(
              where: {_or: [{name: {_ilike: $keyword}}, {full_name_phonetic: {_ilike: $keyword}}, {user_id: {_in: $student_ids_by_phone_number}}], user_id: {_in: $student_ids}, user_group: {_eq: "USER_GROUP_STUDENT"}, last_login_date: {_is_null: $last_login_date}, student: {current_grade: {_in: $grades}, grade_id: {_in: $grade_ids}, enrollment_status: {_eq: $enrollment_status}}, user_access_paths: {location_id: {_in: $location_ids}}}
              limit: $limit
              offset: $offset
              order_by: [$order_by]
            ) {
              user_id
              name
              full_name_phonetic
              email
              phone_number
              country
              last_login_date
              resource_path
              student {
                contact_preference
              }
            }
          }`
)
