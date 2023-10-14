package timesheet

import (
	"time"

	"github.com/manabie-com/backend/j4/serviceutil"
)

var (
	hasuraQueries = []serviceutil.HasuraQuery{
		{
			Name:  "Timesheet_CountTimesheetAdmin",
			Query: Timesheet_CountTimesheetAdmin,
			VariablesCreator: func() map[string]interface{} {
				y, m, d := time.Now().Date()
				from := time.Date(y, m, d, 1, 0, 0, 0, time.Now().Location())
				to := from.Add(31 * time.Second * 24)

				return map[string]interface{}{
					"from_date": from.Format(time.RFC3339),
					"to_date":   to.Format(time.RFC3339),
				}
			},
		},
		{
			Name:  "Timesheet_TimesheetListAdmin",
			Query: Timesheet_TimesheetListAdmin,
			VariablesCreator: func() map[string]interface{} {
				y, m, d := time.Now().Date()
				from := time.Date(y, m, d, 1, 0, 0, 0, time.Now().Location())
				to := from.Add(31 * time.Second * 24)

				return map[string]interface{}{
					"from_date": from.Format(time.RFC3339),
					"to_date":   to.Format(time.RFC3339),
					"offset":    0,
					"limit":     10,
				}
			},
		},
	}
	Timesheet_TimesheetListAdmin = `
          query Timesheet_TimesheetListAdmin($from_date: timestamptz!, $to_date:
          timestamptz!, $timesheet_status: String, $location_id: String,
          $keyword: String, $limit: Int = 10, $offset: Int = 0) {
            timesheet(
              limit: $limit
              offset: $offset
              where: {timesheet_date: {_gte: $from_date, _lte: $to_date}, location_id: {_eq: $location_id}, timesheet_status: {_eq: $timesheet_status}, users: {name: {_ilike: $keyword}}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
              order_by: {timesheet_date: desc, timesheet_id: desc}
              distinct_on: [timesheet_date, timesheet_id]
            ) {
              timesheet_date
              timesheet_id
              timesheet_status
              user: users {
                user_id
                name
                email
              }
              timesheet_lesson_hours {
                lesson: lessons {
                  start_time
                  end_time
                  scheduling_status
                }
              }
              other_working_hours {
                total_hour
              }
              location {
                location_id
                name
              }
              transportation_expenses {
                cost_amount
                round_trip
              }
            }
            timesheet_aggregate(
              where: {timesheet_date: {_gte: $from_date, _lte: $to_date}, location_id: {_eq: $location_id}, timesheet_status: {_eq: $timesheet_status}, users: {name: {_ilike: $keyword}}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
            ) {
              aggregate {
                count
              }
            }
          }`
	Timesheet_CountTimesheetAdmin = `
          query Timesheet_CountTimesheetAdmin($from_date: timestamptz!,
          $to_date: timestamptz!, $keyword: String, $location_id: String) {
            all_count: timesheet_aggregate(
              where: {location_id: {_eq: $location_id}, timesheet_date: {_gte: $from_date, _lte: $to_date}, users: {name: {_ilike: $keyword}}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
            ) {
              aggregate {
                count
              }
            }
            draff_count: timesheet_aggregate(
              where: {location_id: {_eq: $location_id}, users: {name: {_ilike: $keyword}}, timesheet_date: {_gte: $from_date, _lte: $to_date}, timesheet_status: {_eq: "TIMESHEET_STATUS_DRAFT"}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
            ) {
              aggregate {
                count
              }
            }
            submitted_count: timesheet_aggregate(
              where: {location_id: {_eq: $location_id}, users: {name: {_ilike: $keyword}}, timesheet_date: {_gte: $from_date, _lte: $to_date}, timesheet_status: {_eq: "TIMESHEET_STATUS_SUBMITTED"}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
            ) {
              aggregate {
                count
              }
            }
            approved_count: timesheet_aggregate(
              where: {location_id: {_eq: $location_id}, users: {name: {_ilike: $keyword}}, timesheet_date: {_gte: $from_date, _lte: $to_date}, timesheet_status: {_eq: "TIMESHEET_STATUS_APPROVED"}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
            ) {
              aggregate {
                count
              }
            }
            confirmed_count: timesheet_aggregate(
              where: {location_id: {_eq: $location_id}, users: {name: {_ilike: $keyword}}, timesheet_date: {_gte: $from_date, _lte: $to_date}, timesheet_status: {_eq: "TIMESHEET_STATUS_CONFIRMED"}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
            ) {
              aggregate {
                count
              }
            }
            rejected_count: timesheet_aggregate(
              where: {location_id: {_eq: $location_id}, users: {name: {_ilike: $keyword}}, timesheet_date: {_gte: $from_date, _lte: $to_date}, timesheet_status: {_eq: "TIMESHEET_STATUS_REJECTED"}, _or: [{other_working_hours: {}}, {timesheet_lesson_hours: {}}, {transportation_expenses: {}}, {_and: [{remark: {_is_null: false}}, {remark: {_neq: ""}}]}]}
            ) {
              aggregate {
                count
              }
            }
          }`
)
