package eureka

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) anAddBooksRequest(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	request := &epb.AddBooksRequest{
		BookIds:  stepState.BookIDs,
		CourseId: stepState.CourseID,
	}
	switch validity {
	case "non-existed bookIDs":
		for i := 0; i < rand.Intn(10)+5; i++ {
			request.BookIds = append(request.BookIds, idutil.ULIDNow())
		}
	case "empty bookIDs":
		request.BookIds = make([]string, 0)
	case "empty courseID":
		request.CourseId = ""
	}
	stepState.Request = request
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userTryToAddBooksToCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = epb.NewCourseModifierServiceClient(s.Conn).AddBooks(ctx, stepState.Request.(*epb.AddBooksRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustAddsBooksToCourseCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := entities.CoursesBooks{}
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE book_id = ANY($1::TEXT[]) AND course_id = $2::TEXT`, e.TableName())
	var count int
	if err := s.DB.QueryRow(ctx, query, database.TextArray(stepState.BookIDs), database.Text(stepState.CourseID)).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to query courses_books count: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
