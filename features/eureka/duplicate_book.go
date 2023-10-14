package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *suite) insertBookIntoBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	stepState.BookID = idutil.ULIDNow()
	book := &entities.Book{}
	database.AllNullEntity(book)
	bookName := "book-name-course-id_" + stepState.BookID
	err := multierr.Combine(
		book.Country.Set("COUNTRY_VN"),
		book.SchoolID.Set(constants.ManabieSchool),
		book.Subject.Set("SUBJECT_MATH"),
		book.Grade.Set(12),
		book.Name.Set(bookName),
		book.ID.Set(stepState.BookID),
		book.CreatedAt.Set(now),
		book.UpdatedAt.Set(now),
		book.CurrentChapterDisplayOrder.Set(0),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, book, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertBookIntoBobWithArgs(ctx context.Context, bookID, bookName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	if bookID == "" {
		bookID = idutil.ULIDNow()
	}
	stepState.BookID = bookID

	if bookName == "" {
		bookName = "book-name-course-id_" + stepState.BookID
	}

	book := &entities.Book{}
	database.AllNullEntity(book)
	err := multierr.Combine(
		book.Country.Set("COUNTRY_VN"),
		book.SchoolID.Set(constants.ManabieSchool),
		book.Subject.Set("SUBJECT_MATH"),
		book.Grade.Set(12),
		book.Name.Set(bookName),
		book.ID.Set(bookID),
		book.CreatedAt.Set(now),
		book.UpdatedAt.Set(now),
		book.CurrentChapterDisplayOrder.Set(0),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, book, s.DB.Exec)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertCourseBookWithArgs(ctx context.Context, bookID, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	courseBook := &entities.CoursesBooks{}
	database.AllNullEntity(courseBook)
	err := multierr.Combine(
		courseBook.CourseID.Set(courseID),
		courseBook.BookID.Set(bookID),
		courseBook.CreatedAt.Set(now),
		courseBook.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, courseBook, s.DB.Exec)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertCourseBookIntoBobWithArgs(ctx context.Context, bookID, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	courseBook := &entities.CoursesBooks{}
	database.AllNullEntity(courseBook)
	err := multierr.Combine(
		courseBook.CourseID.Set(courseID),
		courseBook.BookID.Set(bookID),
		courseBook.CreatedAt.Set(now),
		courseBook.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, courseBook, s.DB.Exec)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertCourseStudyPlanIntoBobWithArgs(ctx context.Context, courseID, studyPlanID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	courseStudyPlan := &entities.CourseStudyPlan{}
	database.AllNullEntity(courseStudyPlan)
	err := multierr.Combine(
		courseStudyPlan.CourseID.Set(courseID),
		courseStudyPlan.StudyPlanID.Set(studyPlanID),
		courseStudyPlan.CreatedAt.Set(now),
		courseStudyPlan.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, courseStudyPlan, s.DB.Exec)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertChapterIntoBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.insertChapterIntoBobWithArgs(ctx, "", "")
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertChapterIntoBobWithArgs(ctx context.Context, chapterID, chapterName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	if chapterID == "" {
		chapterID = idutil.ULIDNow()
	}
	stepState.ChapterID = chapterID

	if chapterName == "" {
		chapterName = "book-name-course-id_" + stepState.ChapterID
	}

	chapter := &entities.Chapter{}
	database.AllNullEntity(chapter)
	err := multierr.Combine(
		chapter.ID.Set(chapterID),
		chapter.Name.Set(chapterName),
		chapter.Grade.Set(12),
		chapter.Country.Set("COUNTRY_VN"),
		chapter.DisplayOrder.Set(1),
		chapter.Subject.Set("SUBJECT_MATH"),
		chapter.CreatedAt.Set(now),
		chapter.UpdatedAt.Set(now),
		chapter.SchoolID.Set(constants.ManabieSchool),
		chapter.CurrentTopicDisplayOrder.Set(0),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}

	_, err = database.Insert(ctx, chapter, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertBookChapterIntoBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.insertBookChapterIntoBobWithArgs(ctx, "", "")
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertBookChapterIntoBobWithArgs(ctx context.Context, bookID, chapterID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	if bookID == "" {
		bookID = stepState.BookID
	}
	if chapterID == "" {
		chapterID = stepState.ChapterID
	}

	bookChapter := &entities.BookChapter{}
	database.AllNullEntity(bookChapter)
	err := multierr.Combine(
		bookChapter.ChapterID.Set(chapterID),
		bookChapter.BookID.Set(bookID),
		bookChapter.CreatedAt.Set(now),
		bookChapter.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, bookChapter, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertTopicIntoBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.insertTopicIntoBobWithArgs(ctx, "", "", "")
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) insertTopicIntoBobWithArgs(ctx context.Context, topicID, topicName, chapterID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	if topicID == "" {
		topicID = idutil.ULIDNow()
	}
	stepState.TopicID = topicID

	if topicName == "" {
		topicName = "topic-name " + stepState.TopicID
	}

	if chapterID == "" {
		chapterID = stepState.ChapterID
	}

	topic := &entities.Topic{}
	database.AllNullEntity(topic)
	err := multierr.Combine(
		topic.ID.Set(topicID),
		topic.ChapterID.Set(chapterID),
		topic.Name.Set(topicName),
		topic.Subject.Set("SUBJECT_MATHS"),
		topic.CreatedAt.Set(now),
		topic.UpdatedAt.Set(now),
		topic.Grade.Set(12),
		topic.TopicType.Set(epb.TopicType_TOPIC_TYPE_LEARNING.String()),
		topic.TotalLOs.Set(5),
		topic.LODisplayOrderCounter.Set(0),
		topic.SchoolID.Set(constants.ManabieSchool),
		topic.EssayRequired.Set(true),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	_, err = database.Insert(ctx, topic, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) validBookInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.insertBookIntoBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.insertChapterIntoBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.insertBookChapterIntoBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.insertTopicIntoBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDuplicateBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolAdminID := idutil.ULIDNow()
	token, err := generateValidAuthenticationToken(schoolAdminID, "USER_GROUP_SCHOOL_ADMIN")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token
	req := &pb.DuplicateBookRequest{
		BookId:   stepState.BookID,
		BookName: fmt.Sprintf("randome name: %s", stepState.BookID),
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseModifierServiceClient(s.Conn).DuplicateBook(contextWithToken(s, ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validAssignmentInCurrentBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	assignmentContent := entities.AssignmentContent{
		TopicID: stepState.TopicID,
		LoIDs:   []string{},
	}

	for i := 0; i < 5; i++ {
		assignment := &entities.Assignment{}
		database.AllNullEntity(assignment)
		err := multierr.Combine(
			assignment.ID.Set(idutil.ULIDNow()),
			assignment.CreatedAt.Set(now),
			assignment.UpdatedAt.Set(now),
			assignment.DisplayOrder.Set(genRand(100)),
			assignment.Content.Set(assignmentContent),
			assignment.Name.Set("assignment-name"),
			assignment.Type.Set("assignment"),
			assignment.OriginalTopic.Set(stepState.TopicID),
			assignment.TopicID.Set(stepState.TopicID),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil
		}

		if _, err = database.Insert(ctx, assignment, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		topicAssignment := &entities.TopicsAssignments{
			TopicID:      database.Text(stepState.TopicID),
			AssignmentID: assignment.ID,
			DisplayOrder: database.Int2(int16(assignment.DisplayOrder.Int)),
			CreatedAt:    database.Timestamptz(now),
			UpdatedAt:    database.Timestamptz(now),
			DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
		}
		if _, err = database.Insert(ctx, topicAssignment, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustDuplicateAllAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT topic_id FROM topics WHERE copied_topic_id = $1 and deleted_at IS NULL"
	var newTopicID string
	if err := s.DB.QueryRow(ctx, query, stepState.TopicID).Scan(&newTopicID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var orgAssignmentCount, newAssignmentCount int64
	assignmentQueryCount := "SELECT count(*) FROM assignments asm WHERE asm.content ->> 'topic_id' = ANY($1) and deleted_at IS NULL"
	if err := s.DB.QueryRow(ctx, assignmentQueryCount, database.TextArray([]string{stepState.TopicID})).Scan(&orgAssignmentCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err := s.DB.QueryRow(ctx, assignmentQueryCount, database.TextArray([]string{newTopicID})).Scan(&newAssignmentCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if orgAssignmentCount != newAssignmentCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not all assignment is copied")
	}

	var orgTopicAssignmentsCount, newTopicAssignmentsCount int64
	topicAssignmentQueryCount := "SELECT count(*) FROM topics_assignments WHERE topic_id = ANY($1) AND deleted_at IS NULL"
	if err := s.DB.QueryRow(ctx, topicAssignmentQueryCount, database.TextArray([]string{stepState.TopicID})).Scan(&orgTopicAssignmentsCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err := s.DB.QueryRow(ctx, topicAssignmentQueryCount, database.TextArray([]string{newTopicID})).Scan(&newTopicAssignmentsCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if orgTopicAssignmentsCount != newTopicAssignmentsCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not all topics_assignments is copied")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aBookWithoutContentInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.insertBookIntoBob(ctx)
	return StepStateToContext(ctx, stepState), err
}

// end create book
func (s *suite) aValidBookInDB(ctx context.Context) (context.Context, error) {
	return s.hasCreatedABook(ctx)
}

func (s *suite) userSendDuplicateBookRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.DuplicateBookRequest{
		BookId: stepState.BookID,
	}
	stepState.Response, stepState.ResponseErr = pb.NewCourseModifierServiceClient(s.Conn).DuplicateBook(s.signedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkChapterCopied(ctx context.Context, newBookID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT chapter_id FROM books_chapters WHERE book_id = $1`
	rows, err := s.DB.Query(ctx, query, &newBookID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	var chapterIDs []string
	for rows.Next() {
		var classID string
		if err := rows.Scan(&classID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Err: %w", err)
		}
		chapterIDs = append(chapterIDs, classID)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Err: %w", err)
	}
	countChapterQuery := `SELECT count(*) FROM chapters WHERE chapter_id = ANY($1) AND copied_from = ANY($2)`
	var count int64
	if err = s.DB.QueryRow(ctx, countChapterQuery, &chapterIDs, &stepState.ChapterIDs).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if int(count) != len(stepState.ChapterIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not all chapter is copied: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) orgTopicMustBeCorrect(ctx context.Context, orgTopic []string, orgBookID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count (*) FROM books_chapters bc JOIN topics t ON bc.chapter_id = t.chapter_id AND bc.book_id = $1
	AND t.topic_id = ANY($2)`
	var count int64
	if err := s.DB.QueryRow(ctx, query, &orgBookID, &orgTopic).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if int(count) != len(orgTopic) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("original topic ids not valid")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) copiedTopicMustBeCorrect(ctx context.Context, newTopicID []string, orgTopicID []string, newBookID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count (*) FROM books_chapters bc JOIN topics t ON bc.chapter_id = t.chapter_id AND bc.book_id = $1
	AND t.topic_id = ANY($2) AND t.copied_topic_id = ANY($3)`
	var count int64
	if err := s.DB.QueryRow(ctx, query, &newBookID, &newTopicID, &orgTopicID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if int(count) != len(newTopicID) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("new topic ids not valid")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) copiedLearningObjectivesMustBeCorrect(ctx context.Context, newTopicIDs []string, orgTopicIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `SELECT count (*) FROM learning_objectives lo WHERE topic_id = ANY($1) AND deleted_at IS NULL`
	var countLOsByNewTopicIDs, countLOsByOldTopicIDs int64
	if err := s.DB.QueryRow(ctx, query, newTopicIDs).Scan(&countLOsByNewTopicIDs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := s.DB.QueryRow(ctx, query, orgTopicIDs).Scan(&countLOsByOldTopicIDs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if countLOsByNewTopicIDs != countLOsByOldTopicIDs {
		return StepStateToContext(ctx, stepState), fmt.Errorf("new learning objectives not valid")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eurekaMustReturnCopiedTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.DuplicateBookResponse)
	ctx, err1 := s.checkChapterCopied(ctx, rsp.NewBookID)
	ctx, err2 := s.orgTopicMustBeCorrect(ctx, rsp.OrgTopicId, stepState.BookID)
	ctx, err3 := s.copiedTopicMustBeCorrect(ctx, rsp.NewTopicId, rsp.OrgTopicId, rsp.NewBookID)
	ctx, err4 := s.copiedLearningObjectivesMustBeCorrect(ctx, rsp.NewTopicId, rsp.OrgTopicId)

	err := multierr.Combine(err1, err2, err3, err4)
	return StepStateToContext(ctx, stepState), err
}
