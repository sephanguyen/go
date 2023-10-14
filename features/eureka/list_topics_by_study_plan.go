package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) aValidatedBookWithChaptersAndTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudyPlanID = idutil.ULIDNow()
	stepState.BookID = idutil.ULIDNow()
	stepState.SchoolIDInt = constants.ManabieSchool
	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)

	bookReq := &pb.UpsertBooksRequest_Book{
		BookId: stepState.BookID,
		Name:   fmt.Sprintf("book-name+%s", stepState.BookID),
	}
	s.aSignedIn(ctx, "school admin")
	if _, err := pb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &pb.UpsertBooksRequest{
		Books: []*pb.UpsertBooksRequest_Book{
			bookReq,
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var (
		chapters []*entities.Chapter
		topics   []*entities.Topic
	)

	mapChapterAndTopics := make(map[string][]*entities.Topic)
	indexTmp := 0

	ctx = auth.InjectFakeJwtToken(ctx, stepState.SchoolID)
	for i := 0; i < 20; i++ {
		chapterID := idutil.ULIDNow()
		chapter := &entities.Chapter{
			ID:                       database.Text(chapterID),
			Name:                     database.Text(fmt.Sprintf("name-%s", chapterID)),
			Country:                  database.Text("Viet Nam"),
			Subject:                  database.Text("Math"),
			Grade:                    database.Int2(int16(rand.Int31n(110000))),
			DisplayOrder:             database.Int2(int16(rand.Int31n(100000))),
			SchoolID:                 database.Int4(constant.ManabieSchool),
			CreatedAt:                database.Timestamptz(time.Now()),
			UpdatedAt:                database.Timestamptz(time.Now()),
			DeletedAt:                pgtype.Timestamptz{Status: pgtype.Null},
			CopiedFrom:               pgtype.Text{Status: pgtype.Null},
			CurrentTopicDisplayOrder: database.Int4(rand.Int31()),
			BookID:                   database.Text(stepState.BookID),
		}
		if _, err := database.Insert(ctx, chapter, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a chapter: %v", err)
		}

		chapters = append(chapters, chapter)

		bookChapter := &entities.BookChapter{
			ChapterID: database.Text(chapterID),
			BookID:    database.Text(stepState.BookID),
			CreatedAt: database.Timestamptz(time.Now()),
			UpdatedAt: database.Timestamptz(time.Now()),
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		}
		if _, err := database.Insert(ctx, bookChapter, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a book_chapter: %v", err)
		}

		var tmpTopics []*entities.Topic
		for j := 0; j < 10; j++ {
			indexTmp += 1
			topicID := idutil.ULIDNow()
			topic := &entities.Topic{
				ID:                    database.Text(topicID),
				Name:                  database.Text(fmt.Sprintf("name-%s", topicID)),
				Country:               database.Text("Viet Nam"),
				Subject:               database.Text("Math"),
				Grade:                 database.Int2(int16(rand.Int31n(100000))),
				DisplayOrder:          database.Int2(int16(rand.Int31n(100000))),
				Status:                database.Text("status"),
				TopicType:             database.Text("topic-type"),
				ChapterID:             database.Text(chapterID),
				SchoolID:              database.Int4(constant.ManabieSchool),
				IconURL:               database.Text("icon-url"),
				TotalLOs:              database.Int4(1),
				PublishedAt:           database.Timestamptz(time.Now()),
				AttachmentNames:       database.TextArray([]string{}),
				AttachmentURLs:        database.TextArray([]string{}),
				Instruction:           database.Text("instruction"),
				CopiedTopicID:         database.Text("copiedTopicID"),
				EssayRequired:         database.Bool(true),
				LODisplayOrderCounter: database.Int4(1),
				CreatedAt:             database.Timestamptz(time.Now().Add(time.Duration(indexTmp * int(time.Second)))),
				UpdatedAt:             database.Timestamptz(time.Now().Add(time.Duration(indexTmp * int(time.Second)))),
				DeletedAt:             pgtype.Timestamptz{Status: pgtype.Null},
			}
			if _, err := database.Insert(ctx, topic, s.DB.Exec); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a topic: %v", err)
			}

			tmpTopics = append(tmpTopics, topic)
		}

		sort.SliceStable(tmpTopics, func(i, j int) bool {
			return tmpTopics[i].DisplayOrder.Int < tmpTopics[j].DisplayOrder.Int
		})
		mapChapterAndTopics[chapterID] = tmpTopics
	}

	sort.SliceStable(chapters, func(i, j int) bool {
		return chapters[i].DisplayOrder.Int < chapters[j].DisplayOrder.Int
	})

	for _, chapter := range chapters {
		topics = append(topics, mapChapterAndTopics[chapter.ID.String]...)
	}

	contentStructureStr := fmt.Sprintf(`{"book_id": "%s"}`, stepState.BookID)

	studyPlan := &entities.StudyPlan{
		ID: database.Text(stepState.StudyPlanID),
		BaseEntity: entities.BaseEntity{
			CreatedAt: database.Timestamptz(time.Now()),
			UpdatedAt: database.Timestamptz(time.Now()),
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		},
		MasterStudyPlan:     pgtype.Text{Status: pgtype.Null},
		Name:                database.Text(fmt.Sprintf("study-plan-%s", stepState.StudyPlanID)),
		StudyPlanType:       database.Text(epb.StudyPlanType_STUDY_PLAN_TYPE_NONE.String()),
		SchoolID:            database.Int4(constant.ManabieSchool),
		CourseID:            database.Text(""),
		BookID:              database.Text(stepState.BookID),
		Status:              database.Text(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()),
		TrackSchoolProgress: database.Bool(true),
		Grades:              database.Int4Array([]int32{1, 2, 3, 4, 5}),
	}
	if _, err := database.Insert(ctx, studyPlan, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a study plan: %v", err)
	}

	studyPlanItem := &entities.StudyPlanItem{
		ID:               database.Text(idutil.ULIDNow()),
		StudyPlanID:      database.Text(stepState.StudyPlanID),
		ContentStructure: database.JSONB(string(contentStructureStr)),
		BaseEntity: entities.BaseEntity{
			CreatedAt: database.Timestamptz(time.Now()),
			UpdatedAt: database.Timestamptz(time.Now()),
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		},
		Status:                  database.Text(epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String()),
		ContentStructureFlatten: database.Text("content_structure_flatten"),
		AvailableFrom:           pgtype.Timestamptz{Status: pgtype.Null},
		AvailableTo:             pgtype.Timestamptz{Status: pgtype.Null},
		StartDate:               pgtype.Timestamptz{Status: pgtype.Null},
		EndDate:                 pgtype.Timestamptz{Status: pgtype.Null},
		CompletedAt:             pgtype.Timestamptz{Status: pgtype.Null},
		DisplayOrder:            database.Int4(1),
		CopyStudyPlanItemID:     pgtype.Text{Status: pgtype.Null},
		SchoolDate:              pgtype.Timestamptz{Status: pgtype.Null},
	}

	stepState.TopicEntities = topics

	if _, err := database.Insert(ctx, studyPlanItem, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a study plan item: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userListTopicsByStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)

	stepState.Response, stepState.ResponseErr = pb.NewCourseReaderServiceClient(s.Conn).ListTopicsByStudyPlan(s.signedCtx(ctx), &pb.ListTopicsByStudyPlanRequest{
		Paging: &cpb.Paging{
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
			Limit: uint32(100),
		},
		StudyPlanId: stepState.StudyPlanID,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) verifyTopicDataAfterListTopicsByStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topicEntities := stepState.TopicEntities
	index := 0

	resp := stepState.Response.(*pb.ListTopicsByStudyPlanResponse)

	for _, topic := range resp.Items {
		if topicEntities[index].ID.String != topic.Info.Id {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong topic_id, expected %s, but got %s", topic.Info.Id, topicEntities[index].ID.String)
		}
		if topicEntities[index].DisplayOrder.Int != int16(topic.Info.DisplayOrder) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong display_order, expected %v, but got %v", topic.Info.DisplayOrder, topicEntities[index].DisplayOrder.Int)
		}
		index++
	}

	return StepStateToContext(ctx, stepState), nil
}
