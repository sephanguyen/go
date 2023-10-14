package book

import (
	"github.com/manabie-com/backend/features/repository/syllabus/entity"
	"github.com/manabie-com/backend/internal/eureka/entities"
)

type StepState struct {
	DefaultSchoolID int32
	BookID          string
	BookIDs         []string
	Book            *entities.Book
	BookTitleQuery  entity.GraphqlBookTitleQuery
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^a user insert a book to database$`:       s.aUserInsertABookToDatabase,
		`^our system return the book correctly$`:   s.ourSystemReturnTheBookCorrectly,
		`^user get book by call FindByID$`:         s.userGetABookByID,
		`^a user insert some books to database$`:   s.aUserInsertSomeBooksToDatabase,
		`^our system return BooksTitle correctly$`: s.ourSystemReturnBooksTitleCorrectly,
		`^user call BooksTitle$`:                   s.userCallBooksTitle,
	}

	return steps
}
