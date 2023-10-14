package entity

type GraphqlBookTitleQuery struct {
	BookTitle []struct {
		Name string `graphql:"name"`
	} `graphql:"books(where: {book_id: {_eq: $book_id}})"`
}
