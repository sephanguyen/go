package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	ypb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) someBooksAreExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	total := 12
	books := make([]*entities.Book, 0, total)
	booksRequest := make([]*epb.UpsertBooksRequest_Book, 0, total)
	for i := 1; i <= total; i++ {
		now := time.Now()

		book := &entities.Book{}
		bookID := s.newID()
		database.AllNullEntity(book)
		if err := multierr.Combine(
			book.ID.Set(bookID),
			book.Name.Set(strconv.Itoa(rand.Int())),
			book.Country.Set(cpb.Country_COUNTRY_JP.String()),
			book.SchoolID.Set(constants.ManabieSchool),
			book.Subject.Set(cpb.Subject_SUBJECT_ENGLISH.String()),
			book.Grade.Set(12),
			book.CreatedAt.Set(now),
			book.UpdatedAt.Set(now),
			book.CurrentChapterDisplayOrder.Set(0),
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		booksRequest = append(booksRequest, &epb.UpsertBooksRequest_Book{
			Name:   strconv.Itoa(rand.Int()),
			BookId: bookID,
		})

		books = append(books, book)
	}

	_, err := epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(contextWithToken(s, ctx), &epb.UpsertBooksRequest{
		Books: booksRequest,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = books
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentListBooksByIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.aSignedIn(ctx, constant.RoleStudent); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx = contextWithToken(s, ctx)

	books := stepState.Request.([]*entities.Book)
	bookIDs := make([]string, 0, len(books))
	for _, book := range books {
		bookIDs = append(bookIDs, book.ID.String)
	}

	filter := &cpb.CommonFilter{
		Ids: bookIDs,
	}
	paging := &cpb.Paging{
		Limit: 5,
	}

	for {
		resp, err := epb.NewBookReaderServiceClient(s.Conn).ListBooks(ctx, &epb.ListBooksRequest{
			Filter: filter,
			Paging: paging,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if len(resp.Items) == 0 {
			break
		}
		if len(resp.Items) > int(paging.Limit) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total books: got: %d, want: %d", len(resp.Items), paging.Limit)
		}

		stepState.PaginatedBooks = append(stepState.PaginatedBooks, resp.Items)

		paging = resp.NextPage
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnAListOfBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedBooks := stepState.Request.([]*entities.Book)

	isValidBook := func(book *cpb.Book, books []*entities.Book) bool {
		for _, b := range books {
			if book.Info.Id == b.ID.String {
				return true
			}
		}
		return false
	}

	var (
		total   int
		bookIDs []string
	)
	for _, books := range stepState.PaginatedBooks {
		if !sort.SliceIsSorted(books, func(i, j int) bool {
			return books[i].Info.CreatedAt.AsTime().After(books[j].Info.CreatedAt.AsTime())
		}) {
			return StepStateToContext(ctx, stepState), errors.New("books are not sorted by created_at DESC")
		}

		for _, book := range books {
			if !isValidBook(book, expectedBooks) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected book id: %q", book.Info.Id)
			}
			bookIDs = append(bookIDs, book.Info.Id)
		}

		total += len(books)
	}
	if total != len(expectedBooks) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total books: got %d, want: %d", total, len(expectedBooks))
	}

	fields, _ := (&entities.Book{}).FieldMap()
	stmt := fmt.Sprintf(`SELECT %s FROM books WHERE book_id = ANY($1)`, strings.Join(fields, ","))
	var entBooks entities.Books
	if err := database.Select(ctx, s.DB, stmt, bookIDs).ScanAll(&entBooks); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, entBook := range entBooks {
		if entBook.BookType.Status != pgtype.Present || entBook.BookType.String != cpb.BookType_BOOK_TYPE_GENERAL.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf(" book_type of book (%s): got %s, want: %s", entBook.ID.String, entBook.BookType.String, cpb.BookType_BOOK_TYPE_GENERAL.String())
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedAStudyplanExactMatchWithTheBookContentForStudent(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = s.signedCtx(ctx)

	if resp, err := epb.NewCourseModifierServiceClient(s.Conn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookID},
	}); err != nil || !resp.Successful {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert course book: %w", err)
	}
	req := &epb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              stepState.BookID,
		CourseId:            stepState.CourseID,
	}
	resp, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}

	studyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.DB, database.Text(resp.StudyPlanId))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan items: %w", err)
	}

	stepState.AvailableStudyPlanIDs = nil
	stepState.LoIDs = nil

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for _, item := range studyPlanItems {
		cse := &entities.ContentStructure{}
		err := item.ContentStructure.AssignTo(cse)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Unmarshal ContentStructure: %w", err)
		}

		cs := &epb.ContentStructure{}
		_ = item.ContentStructure.AssignTo(cs)

		stepState.AvailableStudyPlanIDs = append(stepState.AvailableStudyPlanIDs, item.ID.String)
		if len(cse.LoID) != 0 {
			stepState.LoIDs = append(stepState.LoIDs, cse.LoID)
			cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
		} else if len(cse.AssignmentID) != 0 {
			stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
			cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
		}

		upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
			StudyPlanId:             item.StudyPlanID.String,
			StudyPlanItemId:         item.ID.String,
			AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
			AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
			StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
			EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
			ContentStructure:        cs,
			ContentStructureFlatten: item.ContentStructureFlatten.String,
			Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
		})
	}

	_, err = epb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlanItemV2(ctx, upsertSpiReq)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
	}

	stepState.StudyPlanID = resp.StudyPlanId
	stepState.StudyPlanIDs = append(stepState.StudyPlanIDs, resp.StudyPlanId)

	stmt := `SELECT email FROM users WHERE user_id = $1`
	var studentEmail string
	err = s.BobDB.QueryRow(ctx, stmt, stepState.StudentID).Scan(&studentEmail)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	locationID := idutil.ULIDNow()
	e := &bob_entities.Location{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.LocationID.Set(locationID),
		e.Name.Set(fmt.Sprintf("location-%s", locationID)),
		e.IsArchived.Set(false),
		e.CreatedAt.Set(time.Now()),
		e.UpdatedAt.Set(time.Now()),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if _, err := database.Insert(ctx, e, s.BobDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = upb.NewUserModifierServiceClient(s.UsermgmtConn).UpdateStudent(
		ctx,
		&upb.UpdateStudentRequest{
			StudentProfile: &upb.UpdateStudentRequest_StudentProfile{
				Id:               stepState.StudentID,
				Name:             "test-name",
				Grade:            5,
				EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				Email:            studentEmail,
				LocationIds:      []string{locationID},
			},

			SchoolId: stepState.SchoolIDInt,
		},
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student: %w", err)
	}

	if _, err := upb.NewUserModifierServiceClient(s.UsermgmtConn).UpsertStudentCoursePackage(ctx, &upb.UpsertStudentCoursePackageRequest{
		StudentId: stepState.StudentID,
		StudentPackageProfiles: []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
			Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: stepState.CourseID,
			},
			StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
		}},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course package: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedAStudyplanExactMatchWithTheBookEmptyForStudent(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	req := &epb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{3, 4},
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              stepState.BookID,
		CourseId:            stepState.CourseID,
	}
	resp, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}
	stepState.Request = req
	stepState.StudyPlanID = resp.StudyPlanId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedAContentBook(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	ctx, err := s.aValidBookContent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	loReq := s.generateLOsReq(ctx)
	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(ctx, loReq); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable create los: %w", err)
	}

	for _, lo := range loReq.LearningObjectives {
		stepState.LoIDs = append(stepState.LoIDs, lo.Info.Id)
	}

	assignmentID := idutil.ULIDNow()
	stepState.AssignmentID = assignmentID
	stepState.AssignmentIDs = append(stepState.AssignmentIDs, stepState.AssignmentID)
	if _, err := epb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(ctx, &epb.UpsertAssignmentsRequest{
		Assignments: []*epb.Assignment{
			{
				AssignmentId: assignmentID,
				Name:         fmt.Sprintf("assignment-%s", assignmentID),
				Content: &epb.AssignmentContent{
					TopicId: stepState.TopicID,
					LoId:    stepState.LoIDs,
				},
				CheckList: &epb.CheckList{
					Items: []*epb.CheckListItem{
						{
							Content:   "Complete all learning objectives",
							IsChecked: true,
						},
						{
							Content:   "Submitted required videos",
							IsChecked: false,
						},
					},
				},
				Instruction:    "teacher's instruction",
				MaxGrade:       100,
				Attachments:    []string{"media-id-1", "media-id-2"},
				AssignmentType: epb.AssignmentType_ASSIGNMENT_TYPE_LEARNING_OBJECTIVE,
				Setting: &epb.AssignmentSetting{
					AllowLateSubmission: true,
					AllowResubmission:   true,
				},
				RequiredGrade: true,
				DisplayOrder:  0,
			},
		},
	}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable create a assignment %v", err)
	}
	stepState.CourseID = idutil.ULIDNow()
	if _, err := ypb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(ctx, &ypb.UpsertCoursesRequest{
		Courses: []*ypb.UpsertCoursesRequest_Course{
			{
				Id:           stepState.CourseID,
				Name:         fmt.Sprintf("course-name+%s", stepState.CourseID),
				Country:      bob_pb.COUNTRY_VN,
				Subject:      bob_pb.SUBJECT_MATHS,
				Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
				SchoolId:     constants.ManabieSchool,
				DisplayOrder: 1,
			},
		},
	}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}

	if _, err := epb.NewCourseModifierServiceClient(s.Conn).AddBooks(ctx, &epb.AddBooksRequest{
		BookIds:  []string{stepState.BookID},
		CourseId: stepState.CourseID,
	}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to add books: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedAnEmptyBook(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	stepState.BookID = idutil.ULIDNow()

	if _, err := epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: []*epb.UpsertBooksRequest_Book{
			{
				BookId: stepState.BookID,
				Name:   fmt.Sprintf("book-name+%s", stepState.BookID),
			},
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to create a book: %v", err)
	}

	if ctx, err := s.schoolAdminCreateAtopicAndAChapter(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CourseID = idutil.ULIDNow()
	if _, err := ypb.NewCourseServiceClient(s.YasuoConn).UpsertCourses(ctx, &ypb.UpsertCoursesRequest{
		Courses: []*ypb.UpsertCoursesRequest_Course{
			{
				Id:           stepState.CourseID,
				Name:         fmt.Sprintf("course-name+%s", stepState.CourseID),
				Country:      bob_pb.COUNTRY_VN,
				Subject:      bob_pb.SUBJECT_MATHS,
				Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
				SchoolId:     constants.ManabieSchool,
				DisplayOrder: 1,
			},
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}

	if _, err := epb.NewCourseModifierServiceClient(s.Conn).AddBooks(ctx, &epb.AddBooksRequest{
		BookIds:  []string{stepState.BookID},
		CourseId: stepState.CourseID,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to add books: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

type (
	FullChapterQuery struct {
		Chapters []ChapterA
	}
	ChapterA struct {
		ChapterID string
		Name      string
		Topics    []Topic
	}

	Topic struct {
		Name               string
		TopicID            string
		LearningObjectives []LearningObjective
	}

	LearningObjective struct {
		Name         string
		LoID         string
		DisplayOrder int
		Type         string
	}
)

func (s *suite) getChaptersByBookID(ctx context.Context, bookID string) (*FullChapterQuery, error) {
	type (
		Chapter struct {
			BookID    string
			ChapterID string
		}

		GetBookChapterQuery struct {
			Chapters []Chapter
		}
	)

	var bookIDReq pgtype.Text
	bookIDReq.Set(bookID)
	query := `
		select
			chapter_id, book_id
		from
			books_chapters as bc
		where
			bc.book_id = $1
	`

	rows, err := s.DB.Query(ctx, query, &bookIDReq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	c := new(GetBookChapterQuery)
	for rows.Next() {
		var chapter Chapter
		err := rows.Scan(&chapter.ChapterID, &chapter.BookID)
		if err != nil {
			return nil, err
		}
		c.Chapters = append(c.Chapters, chapter)
	}

	var chapterIDs pgtype.TextArray
	cids := make([]string, 0)
	for _, each := range c.Chapters {
		cids = append(cids, each.ChapterID)
	}
	chapterIDs.Set(cids)

	resp := new(FullChapterQuery)

	queryGetChapters := `
		select 
			chapter_id, name
		from chapters as c
		where c.chapter_id = any($1::text[])
	`

	rows1, err := s.DB.Query(ctx, queryGetChapters, &chapterIDs)
	if err != nil {
		return nil, err
	}
	defer rows1.Close()
	for rows1.Next() {
		var chapter ChapterA
		err := rows1.Scan(&chapter.ChapterID, &chapter.Name)
		if err != nil {
			return nil, err
		}

		resp.Chapters = append(resp.Chapters, chapter)
	}

	for i, chapter := range resp.Chapters {
		queryGetTopics := `
			select 
				name, topic_id
			from topics as t
			where t.chapter_id = $1
		`
		var arg pgtype.Text
		arg.Set(chapter.ChapterID)
		rows2, err := s.DB.Query(ctx, queryGetTopics, &arg)
		if err != nil {
			return nil, err
		}
		defer rows2.Close()
		for rows2.Next() {
			var topic Topic
			if err := rows2.Scan(&topic.Name, &topic.TopicID); err != nil {
				return nil, err
			}
			resp.Chapters[i].Topics = append(resp.Chapters[i].Topics, topic)
		}
	}

	for i, chapter := range resp.Chapters {
		for j, topic := range chapter.Topics {
			queryGetLOs := `
				select 
					name, lo_id, display_order, type
				from learning_objectives as lo
				where lo.topic_id = $1
			`
			var arg pgtype.Text
			arg.Set(topic.TopicID)
			rows, err := s.DB.Query(ctx, queryGetLOs, &arg)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			for rows.Next() {
				var lo LearningObjective
				if err := rows.Scan(&lo.Name, &lo.LoID, &lo.DisplayOrder, &lo.Type); err != nil {
					return nil, err
				}
				resp.Chapters[i].Topics[j].LearningObjectives = append(resp.Chapters[i].Topics[j].LearningObjectives, lo)
			}
		}
	}

	return resp, nil
}

type (
	GetAssignmentQuery struct {
		Assignments []Assignment
	}
	Assignment struct {
		Name        string
		AssigmentID string
		Content     pgtype.JSONB
	}
)

func (s *suite) getAssignments(ctx context.Context, topicID []string) (*GetAssignmentQuery, error) {
	var topicIDs pgtype.TextArray
	topicIDs.Set(topicID)

	query := `
		select
			name, assignment_id, content
		from
			assignments a
		where
			a."content"->>'topic_id' = any($1::text[]);
	`

	rows, err := s.DB.Query(ctx, query, &topicIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resp := new(GetAssignmentQuery)
	for rows.Next() {
		var assigment Assignment
		err := rows.Scan(&assigment.Name, &assigment.AssigmentID, &assigment.Content)
		if err != nil {
			return nil, err
		}

		resp.Assignments = append(resp.Assignments, assigment)
	}

	return resp, nil
}
