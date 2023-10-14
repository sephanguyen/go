package eureka

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

/*
lo1---topic1---chapter1---book1---studyplan1

	       \
		\
	         \

lo2---topic2---chapter2---book2---studyplan2

	  /                   /
	 /                   /
	/                   /

lo3---topic3---chapter3---book3---studyplan3
*/
func (s *suite) anLearningObjectivesCreatedEvent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx = contextWithToken(s, ctx)

	books := []*entities.Book{newBook(), newBook(), newBook()}
	for _, book := range books {
		if _, err := database.Insert(ctx, book, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	book1, book2, book3 := books[0], books[1], books[2]

	chapters := []*entities.Chapter{newChapter(), newChapter(), newChapter()}
	for _, c := range chapters {
		if _, err := database.Insert(ctx, c, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	chapter1, chapter2, chapter3 := chapters[0], chapters[1], chapters[2]

	bookChapters := []*entities.BookChapter{
		newBookChapter(book1, chapter1),
		newBookChapter(book2, chapter1),
		newBookChapter(book2, chapter2),
		newBookChapter(book2, chapter3),
		newBookChapter(book3, chapter3),
	}
	for _, bc := range bookChapters {
		if _, err := database.Insert(ctx, bc, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	topic1 := newTopic()
	topic1.ChapterID = chapter1.ID
	if _, err := database.Insert(ctx, topic1, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	topic2 := newTopic()
	topic2.ChapterID = chapter2.ID
	if _, err := database.Insert(ctx, topic2, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	topic3 := newTopic()
	topic3.ChapterID = chapter3.ID
	if _, err := database.Insert(ctx, topic3, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	sp1 := newStudyPlan(book1.ID.String, stepState.CourseID)
	if _, err := database.Insert(ctx, sp1, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	copiedSP1 := newStudyPlan(idutil.ULIDNow(), stepState.CourseID)
	copiedSP1.MasterStudyPlan = sp1.ID
	if _, err := database.Insert(ctx, copiedSP1, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	sp2 := newStudyPlan(book2.ID.String, stepState.CourseID)
	if _, err := database.Insert(ctx, sp2, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	copiedSP2 := newStudyPlan(idutil.ULIDNow(), stepState.CourseID)
	copiedSP2.MasterStudyPlan = sp2.ID
	if _, err := database.Insert(ctx, copiedSP2, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	sp3 := newStudyPlan(book3.ID.String, stepState.CourseID)
	if _, err := database.Insert(ctx, sp3, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	copiedSP3 := newStudyPlan(idutil.ULIDNow(), stepState.CourseID)
	copiedSP3.MasterStudyPlan = sp3.ID
	if _, err := database.Insert(ctx, copiedSP3, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var items []*epb.StudyPlanItem

	ctx, item := s.generateStudyPlanItem(ctx, "", sp1.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book1.ID.String,
		ChapterId: chapter1.ID.String,
		TopicId:   topic1.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp2.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book2.ID.String,
		ChapterId: chapter1.ID.String,
		TopicId:   topic1.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp2.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book2.ID.String,
		ChapterId: chapter2.ID.String,
		TopicId:   topic2.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp2.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book2.ID.String,
		ChapterId: chapter3.ID.String,
		TopicId:   topic3.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp3.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book3.ID.String,
		ChapterId: chapter3.ID.String,
		TopicId:   topic3.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	s.insertCourseBookIntoBobWithArgs(ctx, book1.ID.String, stepState.CourseID)
	s.insertCourseBookIntoBobWithArgs(ctx, book2.ID.String, stepState.CourseID)
	s.insertCourseBookIntoBobWithArgs(ctx, book3.ID.String, stepState.CourseID)

	s.insertCourseStudyPlanIntoBobWithArgs(ctx, stepState.CourseID, sp1.ID.String)
	s.insertCourseStudyPlanIntoBobWithArgs(ctx, stepState.CourseID, sp2.ID.String)
	s.insertCourseStudyPlanIntoBobWithArgs(ctx, stepState.CourseID, sp3.ID.String)

	if _, err := epb.NewAssignmentModifierServiceClient(s.Conn).UpsertStudyPlanItem(
		contextWithValidVersion(ctx),
		&epb.UpsertStudyPlanItemRequest{
			StudyPlanItems: items,
		},
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = map[string][]entities.ContentStructure{
		topic1.ID.String: {
			{
				BookID:    book1.ID.String,
				ChapterID: chapter1.ID.String,
				TopicID:   topic1.ID.String,
				CourseID:  stepState.CourseID,
			},
			{
				BookID:    book2.ID.String,
				ChapterID: chapter1.ID.String,
				TopicID:   topic1.ID.String,
				CourseID:  stepState.CourseID,
			},
		},
		topic2.ID.String: {
			{
				BookID:    book2.ID.String,
				ChapterID: chapter2.ID.String,
				TopicID:   topic2.ID.String,
				CourseID:  stepState.CourseID,
			},
		},
		topic3.ID.String: {
			{
				BookID:    book2.ID.String,
				ChapterID: chapter3.ID.String,
				TopicID:   topic3.ID.String,
				CourseID:  stepState.CourseID,
			},
			{
				BookID:    book3.ID.String,
				ChapterID: chapter3.ID.String,
				TopicID:   topic3.ID.String,
				CourseID:  stepState.CourseID,
			},
		},
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemReceivesLearningObjectivesCreatedEvent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var topicIDs []string
	contentStructuresByTopic := stepState.Request.(map[string][]entities.ContentStructure)
	for topicID := range contentStructuresByTopic {
		topicIDs = append(topicIDs, topicID)
	}

	var (
		topic1, topic2, topic3 = topicIDs[0], topicIDs[1], topicIDs[2]
		contentStructuresByLO  = make(map[string][]entities.ContentStructure) // lo id => content structures
	)

	// add new LO for topic1
	lo1 := newLOPb()
	lo1.TopicId = topic1
	lo1.Info.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByLO[lo1.Info.Id] = contentStructuresByTopic[lo1.TopicId]
	if ctx, err := s.upsertLOs(ctx, []*cpb.LearningObjective{lo1}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// add new LO for topic2
	lo2 := newLOPb()
	lo2.TopicId = topic2
	lo2.Info.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByLO[lo2.Info.Id] = contentStructuresByTopic[lo2.TopicId]
	if ctx, err := s.upsertLOs(ctx, []*cpb.LearningObjective{lo2}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	tmpLO := newLOPb()
	// add same LO for both topic2 and topic3
	// lo3 and lo4 is same LO but with different topic id
	lo3 := *tmpLO
	lo3.TopicId = topic2

	lo4 := *tmpLO
	lo4.TopicId = topic3
	if ctx, err := s.upsertLOs(ctx, []*cpb.LearningObjective{&lo3, &lo4}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	contentStructuresByLO[lo4.Info.Id] = contentStructuresByTopic[lo4.TopicId]
	contentStructuresByLO[lo4.Info.Id] = append(
		contentStructuresByLO[lo4.Info.Id],
		contentStructuresByTopic[lo3.TopicId]...,
	)

	lo5 := newLOPb()
	lo5.TopicId = topic1
	lo5.Info.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByLO[lo5.Info.Id] = contentStructuresByTopic[lo5.TopicId]

	lo6 := newLOPb()
	lo6.TopicId = topic2
	lo6.Info.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByLO[lo6.Info.Id] = contentStructuresByTopic[lo6.TopicId]

	lo7 := newLOPb()
	lo7.TopicId = topic3
	lo7.Info.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByLO[lo7.Info.Id] = contentStructuresByTopic[lo7.TopicId]

	if ctx, err := s.upsertLOs(ctx, []*cpb.LearningObjective{lo5, lo6, lo7}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = contentStructuresByLO

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateStudyPlanItemsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for event handler to be run
	time.Sleep(2 * time.Second)

	contentStructuresByLO := stepState.Request.(map[string][]entities.ContentStructure)

	inCS := func(bookID, chapterID, topicID string, cs []entities.ContentStructure) bool {
		for _, c := range cs {
			if c.BookID == bookID && c.ChapterID == chapterID && c.TopicID == topicID {
				return true
			}
		}
		return false
	}

	equalCS := func(cs1, cs2 []entities.ContentStructure) bool {
		if len(cs1) != len(cs2) {
			return false
		}

		for _, c1 := range cs1 {
			if !inCS(c1.BookID, c1.ChapterID, c1.TopicID, cs2) {
				return false
			}
		}
		return true
	}

	verifyItemDisplayOrder := func(loID string) error {
		ctx, _, err := s.getLODisplayOrder(ctx, loID)
		if err != nil {
			return fmt.Errorf("s.getLODisplayOrder: %v", err)
		}

		ctx, items, err := s.getStudyPlanItems(ctx, loID)
		if err != nil {
			return err
		}
		for _, item := range items {
			// content_structure_flatten must contain loID
			if !strings.Contains(item.ContentStructureFlatten.String, loID) {
				return fmt.Errorf("content_structure_flatten of item %q doesn't have loID: %q", item.ID.String, loID)
			}
		}

		return nil
	}

	for loID, expectedCS := range contentStructuresByLO {
		ctx, cs, err := s.getContentStructuresByLO(ctx, loID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if !equalCS(cs, expectedCS) {
			return StepStateToContext(ctx, stepState), fmt.Errorf(`content structure for lo: %q is wrong
				got: %+v,
				expected: %+v`,
				loID, cs, expectedCS,
			)
		}

		if err := verifyItemDisplayOrder(loID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// upsert lo's display order
		existedLO := newLOPb()
		existedLO.Info.Id = loID
		existedLO.TopicId = expectedCS[0].TopicID
		existedLO.Info.DisplayOrder = rand.Int31n(math.MaxInt16)

		// create new los
		newLO1 := newLOPb()
		newLO1.TopicId = expectedCS[0].TopicID
		newLO1.Info.DisplayOrder = rand.Int31n(math.MaxInt16)

		newLO2 := newLOPb()
		newLO2.TopicId = expectedCS[0].TopicID
		newLO2.Info.DisplayOrder = rand.Int31n(math.MaxInt16)

		if ctx, err := s.upsertLOs(ctx, []*cpb.LearningObjective{existedLO, newLO1, newLO2}); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		time.Sleep(2 * time.Second)

		if err := verifyItemDisplayOrder(loID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if err := verifyItemDisplayOrder(newLO1.Info.Id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if err := verifyItemDisplayOrder(newLO2.Info.Id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertLOs(ctx context.Context, los []*cpb.LearningObjective) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(
		contextWithToken(s, ctx),
		&epb.UpsertLOsRequest{
			LearningObjectives: los,
		},
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getContentStructuresByLO(ctx context.Context, loID string) (context.Context, []entities.ContentStructure, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT study_plan_item_id FROM lo_study_plan_items WHERE lo_id = $1"
	rows, err := s.DB.Query(ctx, query, loID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		items = append(items, itemID)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	query = "SELECT DISTINCT(content_structure) FROM study_plan_items WHERE study_plan_item_id = ANY($1) AND copy_study_plan_item_id IS NULL"
	rows, err = s.DB.Query(ctx, query, database.TextArray(items))
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var ret []entities.ContentStructure
	for rows.Next() {
		var cs entities.ContentStructure
		if err := rows.Scan(&cs); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		ret = append(ret, cs)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), ret, nil
}

func (s *suite) getStudyPlanItems(ctx context.Context, loID string) (context.Context, []entities.StudyPlanItem, error) {
	stepState := StepStateFromContext(ctx)

	query := `
		SELECT spi.study_plan_item_id, spi.display_order, content_structure_flatten
		FROM lo_study_plan_items lspi
		INNER JOIN study_plan_items spi ON spi.study_plan_item_id = lspi.study_plan_item_id
		WHERE lo_id = $1
	`
	rows, err := s.DB.Query(ctx, query, loID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var items []entities.StudyPlanItem
	for rows.Next() {
		item := entities.StudyPlanItem{}
		if err := rows.Scan(&item.ID, &item.DisplayOrder, &item.ContentStructureFlatten); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), items, nil
}

func (s *suite) getLODisplayOrder(ctx context.Context, loID string) (context.Context, int, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT display_order FROM learning_objectives WHERE lo_id = $1"
	var displayOrder int
	if err := s.DB.QueryRow(ctx, query, loID).Scan(&displayOrder); err != nil {
		return StepStateToContext(ctx, stepState), 0, err
	}
	return StepStateToContext(ctx, stepState), displayOrder, nil
}

func newChapter() *entities.Chapter {
	now := time.Now()
	e := &entities.Chapter{}
	database.AllNullEntity(e)
	e.ID.Set(idutil.ULIDNow())
	e.Name = e.ID
	e.Country.Set(pb.COUNTRY_VN.String())
	e.Grade.Set(1)
	e.Subject.Set(pb.SUBJECT_BIOLOGY.String())
	e.DisplayOrder.Set(1)
	e.SchoolID.Set(constants.ManabieSchool)
	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)
	e.CurrentTopicDisplayOrder.Set(0)
	return e
}

func newTopic() *entities.Topic {
	now := time.Now()
	e := &entities.Topic{}
	database.AllNullEntity(e)
	e.ID.Set(idutil.ULIDNow())
	e.Name = e.ID
	e.Country.Set(pb.COUNTRY_VN.String())
	e.Grade.Set(12)
	e.Subject.Set(pb.SUBJECT_MATHS.String())
	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)
	e.Status.Set(pb.TOPIC_STATUS_NONE.String())
	e.DisplayOrder.Set(1)
	e.PublishedAt.Set(now)
	e.SchoolID.Set(constant.ManabieSchool)
	e.TopicType.Set(pb.TOPIC_TYPE_NONE.String())
	e.TotalLOs.Set(1)
	e.EssayRequired.Set(false)
	return e
}

func newLOPb() *cpb.LearningObjective {
	id := idutil.ULIDNow()
	return &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:           id,
			Name:         id,
			Country:      cpb.Country_COUNTRY_VN,
			Grade:        12,
			Subject:      cpb.Subject_SUBJECT_MATHS,
			MasterId:     "",
			DisplayOrder: 1,
			SchoolId:     constants.ManabieSchool,
		},
		TopicId: "",
		// VideoScript: "script",
		Prerequisites: []string{},
		// StudyGuide:    "https://guides/1/master",
		// Video:         "https://videos/1/master",
	}
}

func newBook() *entities.Book {
	now := time.Now()
	id := idutil.ULIDNow()
	e := new(entities.Book)
	e.ID.Set(id)
	e.Name.Set(id)
	e.Country.Set(pb.COUNTRY_VN.String())
	e.Subject.Set(pb.SUBJECT_BIOLOGY.String())
	e.Grade.Set(12)
	e.SchoolID.Set(constants.ManabieSchool)
	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)
	e.DeletedAt.Set(nil)
	e.CopiedFrom.Set("")
	e.CurrentChapterDisplayOrder.Set(0)
	e.BookType.Set(cpb.BookType_BOOK_TYPE_GENERAL.String())
	return e
}

func newBookChapter(b *entities.Book, c *entities.Chapter) *entities.BookChapter {
	now := time.Now()
	e := &entities.BookChapter{}
	e.BookID = b.ID
	e.ChapterID = c.ID
	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)
	e.DeletedAt.Set(nil)
	return e
}

func newStudyPlan(id, courseID string) *entities.StudyPlan {
	now := time.Now()
	e := new(entities.StudyPlan)
	e.ID.Set(id)
	e.MasterStudyPlan.Set(nil)
	e.Name.Set(id)
	e.StudyPlanType.Set("STUDY_PLAN_TYPE_COURSE")
	e.SchoolID.Set(constants.ManabieSchool)
	e.CourseID.Set(courseID)
	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)
	e.DeletedAt.Set(nil)
	e.Grades.Set(nil)
	e.Status.Set("STUDY_PLAN_STATUS_ACTIVE")
	e.TrackSchoolProgress.Set(false)
	e.BookID.Set(nil)
	return e
}

func toContentStructureFlatten(cs *epb.ContentStructure, loID string) string {
	// contentStructureFlatten format:
	return fmt.Sprintf("book::%stopic::%schapter::%scourse::%slo::%s", cs.BookId, cs.TopicId, cs.ChapterId, cs.CourseId, loID)
}
