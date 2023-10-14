package eurekav2

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/cucumber/godog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		// common
		`^a signed in "([^"]*)"$`:                                  s.takeAnSignedInUser,
		`^returns "([^"]*)" status code$`:                          s.checkStatusCode,
		`^user adds a simple book content$`:                        s.prepareASimpleBookContent,
		`^user adds some learning materials to topic of the book$`: s.someExistingLMInTopic,

		// upsert books
		`^user upsert valid books$`:                    s.upsertBooks,
		`^user creates new "([^"]*)" books$`:           s.createNewBooks,
		`^our system must stores correct books$`:       s.checkUpsertedBooks,
		`^there are books existed$`:                    s.seedBooks,
		`^user updates "([^"]*)" books$`:               s.updateBooks,
		`^our system must update the books correctly$`: s.checkUpdatedBooks,
		`^user has created an empty book$`:             s.createAnEmptyBook,

		// update publish learning material
		`^user updates publish status of learning material to "([^"]*)"$`:            s.sendAnUpdatePublishStatusLearningMaterialsRequest,
		`^our system must update the publish status of learning material correctly$`: s.checkUpdatePublishStatusLearningMaterials,

		// gets book content
		`fake a book content$`:                s.GenerateFakeBookContent,
		`seed faked book content$`:            s.SeedBookContentRecursive,
		`user gets a "([^"]*)" book content$`: s.GetBookContent,
		`returns valid book content$`:         s.ValidateBookContent,

		// upsert courses
		`^user upsert valid courses$`:                    s.upsertCourses,
		`^user creates new "([^"]*)" courses$`:           s.createNewCourses,
		`^there are courses existed$`:                    s.seedCourses,
		`^user updates "([^"]*)" courses$`:               s.updateCourses,
		`^our system must update the courses correctly$`: s.checkUpdatedCourses,
		`^user has created an empty courses$`:            s.createAnEmptyCourse,

		// get a book hierarchy flatten by learning material id
		`user gets a "([^"]*)" book hierarchy flatten`:                      s.GetBookHierarchyFlattenByLmID,
		`^returns correct book hierarchy flatten of that learning material`: s.CheckBookHierarchyFlattenByLmID,
		`^there are study plan created in courses$`:                         s.thereAreStudyPlanCreatedInCourses,
		`^user creates new "([^"]*)" study plan item$`:                      s.userCreatesNewStudyPlanItem,

		// upsert study plan
		`^user creates new "([^"]*)" study plan$`:     s.createNewStudyPlan,
		`^our system must stores correct study plan$`: s.checkUpsertedStudyPlan,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

// A book contains 1 chapter and 1 topic
func (s *suite) prepareASimpleBookContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	bookID, chapterIDs, topicIDs, err := PrepareValidBookContent(ctx, s.EurekaConn)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}
	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.TopicIDs = topicIDs
	return StepStateToContext(ctx, stepState), nil
}

// someExistingLMInTopic for simple this func only ensure each type existed in the topic.
func (s *suite) someExistingLMInTopic(ctx context.Context) (context.Context, error) {
	var err error
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	// not insert parellel to ensure our expectation: I want to get the display_order without retrieve the database.
	for lmType, i := range sspb.LearningMaterialType_value {
		if lmType == sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String() {
			if ctx, err = s.insertASimpleLO(ctx, i); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertASimpleLO(ctx context.Context, val int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	req := &sspb.InsertLearningObjectiveRequest{
		LearningObjective: &sspb.LearningObjectiveBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("learning_objective_%d", val),
			},
		},
	}
	resp, err := sspb.NewLearningObjectiveClient(s.EurekaConn).InsertLearningObjective(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert the learning_objective: %w", err)
	}
	temp := &entity.LearningMaterialPb{
		LearningMaterialBase: req.LearningObjective.Base,
	}
	temp.LearningMaterialBase.LearningMaterialId = resp.GetLearningMaterialId()
	stepState.LearningObjectiveIDs = append(stepState.LearningObjectiveIDs, temp.LearningMaterialId)
	stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, temp.LearningMaterialId)
	return StepStateToContext(ctx, stepState), nil
}

func GenerateBooks(numberOfBooks int, template *epb.UpsertBooksRequest_Book) []*epb.UpsertBooksRequest_Book {
	if template == nil {
		// A valid create book req template
		template = &epb.UpsertBooksRequest_Book{
			Name: "Syllabus" + idutil.ULIDNow(),
		}
	}
	books := make([]*epb.UpsertBooksRequest_Book, 0)
	for i := 0; i < numberOfBooks; i++ {
		books = append(books, proto.Clone(template).(*epb.UpsertBooksRequest_Book))
	}
	return books
}

func GenerateChapters(bookID string, numberOfChapter int, template *cpb.Chapter) []*cpb.Chapter {
	if template == nil {
		template = &cpb.Chapter{
			Info: &cpb.ContentBasicInfo{
				Country:  cpb.Country_COUNTRY_VN,
				SchoolId: constants.ManabieSchool,
				Subject:  cpb.Subject_SUBJECT_BIOLOGY,
				Grade:    1,
				Name:     "Chapter" + idutil.ULIDNow(),
			},
			BookId: bookID,
		}
	}
	chapters := make([]*cpb.Chapter, 0)
	for i := 0; i < numberOfChapter; i++ {
		chapters = append(chapters, proto.Clone(template).(*cpb.Chapter))
	}
	return chapters
}

func GenerateTopic(chapterID string, template *epb.Topic) *epb.Topic {
	if template == nil {
		template = &epb.Topic{
			SchoolId:  constants.ManabieSchool,
			Subject:   epb.Subject_SUBJECT_BIOLOGY,
			Name:      "Topic" + idutil.ULIDNow(),
			ChapterId: chapterID,
			Status:    epb.TopicStatus_TOPIC_STATUS_NONE,
			Type:      epb.TopicType_TOPIC_TYPE_LEARNING,
			TotalLos:  1,
		}
	}

	return proto.Clone(template).(*epb.Topic)
}

func GenerateTopics(chapterID string, numberOfTopics int, template *epb.Topic) []*epb.Topic {
	books := make([]*epb.Topic, 0)
	for i := 0; i < numberOfTopics; i++ {
		books = append(books, GenerateTopic(chapterID, template))
	}
	return books
}

func PrepareValidBookContent(ctx context.Context, eurekaConn *grpc.ClientConn) (bookID string, chapterIDs []string, topicIDs []string, err error) {
	bookResp, err := epb.NewBookModifierServiceClient(eurekaConn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: GenerateBooks(1, nil),
	})
	if err != nil {
		err = fmt.Errorf("NewBookModifierService.UpsertBooks: %w", err)
		return
	}

	bookID = bookResp.BookIds[0]
	chapterResp, err := epb.NewChapterModifierServiceClient(eurekaConn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: GenerateChapters(bookID, 1, nil),
		BookId:   bookID,
	})
	if err != nil {
		err = fmt.Errorf("NewChapterModifierService.UpsertChapters: %w", err)

		return
	}
	chapterIDs = chapterResp.GetChapterIds()
	topicResp, err := epb.NewTopicModifierServiceClient(eurekaConn).Upsert(ctx, &epb.UpsertTopicsRequest{
		Topics: GenerateTopics(chapterIDs[0], 1, nil),
	})
	if err != nil {
		err = fmt.Errorf("NewTopicModifierService.Upsert: %w", err)
		return
	}
	topicIDs = topicResp.GetTopicIds()
	return
}
