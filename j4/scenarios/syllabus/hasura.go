package syllabus

import (
	"github.com/manabie-com/backend/j4/serviceutil"
)

var (
	hasuraQueries = []serviceutil.HasuraQuery{
		{
			Name:  "Syllabus_BooksListV2",
			Query: Syllabus_BooksListV2,
			VariablesCreator: func() map[string]interface{} {
				return map[string]interface{}{
					"limit":  10,
					"offset": 0,
				}
			},
		},
	}
	Syllabus_BooksListV2 = `
          query Syllabus_BooksListV2($name: String, $limit: Int = 10, $offset: Int = 0, $type: String = "BOOK_TYPE_GENERAL") {
            books(limit: $limit, offset: $offset, order_by: {created_at: desc, name: asc, book_id: asc}, where: {name: {_ilike: $name}, book_type: {_eq: $type}}) {
              book_id
              name
            }
            books_aggregate(where: {name: {_ilike: $name}, book_type: {_eq: $type}}) {
              aggregate {
                count
              }
            }
          }`
)
