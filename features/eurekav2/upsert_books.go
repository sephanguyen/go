package eurekav2

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"google.golang.org/protobuf/proto"
)

func (s *suite) generateBooks(numberOfBooks int, template *epb.UpsertBooksRequest_Book) []*epb.UpsertBooksRequest_Book {
	if template == nil {
		// A valid create book req template
		template = &epb.UpsertBooksRequest_Book{
			Name: "Book 1",
		}
	}
	books := make([]*epb.UpsertBooksRequest_Book, 0)
	for i := 0; i < numberOfBooks; i++ {
		books = append(books, proto.Clone(template).(*epb.UpsertBooksRequest_Book))
	}
	return books
}

func (s *suite) createAnEmptyBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	reqBooks := s.generateBooks(1, nil)
	resp, err := epb.NewBookServiceClient(s.Connections.EurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: reqBooks,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not create book: %w", err)
	}
	stepState.BookID = resp.BookIds[0]
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createNewBooks(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfBooks := rand.Intn(20) + 1 //nolint:gosec
	var reqBooks []*epb.UpsertBooksRequest_Book
	switch validity {
	case "valid":
		reqBooks = s.generateBooks(numberOfBooks, nil)
	case "invalid":
		reqBooks = s.generateBooks(numberOfBooks, &epb.UpsertBooksRequest_Book{
			Name: "",
		})
	}

	stepState.Response, stepState.ResponseErr = epb.NewBookServiceClient(s.EurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: reqBooks,
	})
	if stepState.ResponseErr == nil {
		stepState.BookIDs = stepState.Response.(*epb.UpsertBooksResponse).BookIds
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpsertedBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	bookIDs := stepState.Response.(*epb.UpsertBooksResponse).BookIds
	query := "SELECT count(*) FROM books WHERE book_id = ANY($1)"
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, bookIDs).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != len(bookIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("eureka doesn't store correct books: expected %d, got %d", len(bookIDs), count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) seedBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfBooks := rand.Intn(20) + 1 //nolint:gosec
	reqBooks := s.generateBooks(numberOfBooks, nil)

	res, err := epb.NewBookServiceClient(s.EurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: reqBooks,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to seeding some books: %w", err)
	}
	stepState.BookIDs = res.BookIds

	query := "SELECT book_id, updated_at FROM books WHERE book_id = ANY($1)"
	rows, err := s.EurekaDB.Query(ctx, query, stepState.BookIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query book: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		book := new(domain.Book)
		if err := rows.Scan(&book.ID, &book.UpdatedAt); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to scan book: %w", err)
		}
		stepState.Books = append(stepState.Books, *book)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateBooks(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	// Take first <randomly> books to update, 1 at least
	numberOfUpdates := rand.Intn(len(stepState.BookIDs)) //nolint:gosec
	if numberOfUpdates == 0 {
		numberOfUpdates = 1
	}

	stepState.UpdatedBookIDs = append(stepState.UpdatedBookIDs, stepState.BookIDs[0:numberOfUpdates]...)
	reqBooks := make([]*epb.UpsertBooksRequest_Book, 0)
	// Update data by each case.
	switch validity {
	case "valid":
		for _, bookID := range stepState.UpdatedBookIDs {
			reqBooks = append(reqBooks, &epb.UpsertBooksRequest_Book{
				BookId: bookID,
				Name:   "Book 1",
			})
		}
	case "invalid":
		for _, bookID := range stepState.UpdatedBookIDs {
			reqBooks = append(reqBooks, &epb.UpsertBooksRequest_Book{
				BookId: bookID,
				Name:   "",
			})
		}
	}

	stepState.Response, stepState.ResponseErr = epb.NewBookServiceClient(s.Connections.EurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: reqBooks,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpdatedBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	existedBookMap := make(map[string]domain.Book)
	for _, book := range stepState.Books {
		existedBookMap[book.ID] = book
	}

	query := "SELECT book_id, updated_at FROM books WHERE book_id = ANY($1)"
	rows, err := s.EurekaDB.Query(ctx, query, stepState.UpdatedBookIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	for rows.Next() {
		book := new(domain.Book)
		if err := rows.Scan(&book.ID, &book.UpdatedAt); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query book: %w", err)
		}
		existedBook, ok := existedBookMap[book.ID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("book is missing book_id: %s", book.ID)
		}
		if existedBook.UpdatedAt.Equal(book.UpdatedAt) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("book was not updated book_id: %s", book.ID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfBooks := rand.Intn(20) + 1 //nolint:gosec
	reqBooks := s.generateBooks(numberOfBooks, nil)
	stepState.Response, stepState.ResponseErr = epb.NewBookServiceClient(s.Connections.EurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: reqBooks,
	})
	return StepStateToContext(ctx, stepState), nil
}
