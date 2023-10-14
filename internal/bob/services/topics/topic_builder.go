package topics

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	coursesRepo "github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/bob/services/topics/repo"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

// TopicBuilder implement builder for topic
type TopicBuilder struct {
	db         database.Ext
	courseRepo coursesRepo.CourseRepo
	topicRepo  repo.TopicRepo
}

func NewTopicBuilder(db database.Ext, courseRepo coursesRepo.CourseRepo, topicRepo repo.TopicRepo) *TopicBuilder {
	return &TopicBuilder{
		db:         db,
		courseRepo: courseRepo,
		topicRepo:  topicRepo,
	}
}

// CreateTopicsByLiveLesson will create list topics with every topics
// in this case will hold relationship about "data" between lesson and every
// courses of this lesson. An include lesson's name, course's country, course's grade,
// course's subject and course's schoolID with topic type is TOPIC_TYPE_LIVE_LESSON
// and status is TOPIC_STATUS_PUBLISHED
func (t *TopicBuilder) CreateTopicsByLiveLesson(ctx context.Context, lesson *entities.Lesson) (map[pgtype.Text]*entities.Topic, error) {
	courseIDs := lesson.CourseIDs.CourseIDs
	if len(courseIDs.Elements) == 0 && len(lesson.CourseID.String) != 0 {
		courseIDs = database.TextArrayVariadic(lesson.CourseID.String)
	}

	if len(courseIDs.Elements) == 0 {
		return nil, nil
	}

	courses, err := t.courseRepo.FindByIDs(ctx, t.db, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("courseRepo.FindByIDs: %s", err)
	}
	if len(courses) != len(courseIDs.Elements) {
		return nil, fmt.Errorf("expected %d courses from database, got %d", len(courseIDs.Elements), len(courses))
	}

	return t.CreateTopicByLiveLessonAndCourses(ctx, *lesson, courses)
}

func (t *TopicBuilder) CreateTopicByLiveLessonAndCourses(ctx context.Context, lesson entities.Lesson, courses map[pgtype.Text]*entities.Course) (map[pgtype.Text]*entities.Topic, error) {
	if len(courses) == 0 {
		return nil, nil
	}

	now := time.Now()
	topics := make([]*entities.Topic, 0, len(courses))
	topicByCourseID := make(map[pgtype.Text]*entities.Topic)
	for _, course := range courses {
		e := &entities.Topic{}
		database.AllNullEntity(e)
		err := multierr.Combine(
			e.ID.Set(idutil.ULIDNow()),
			e.Name.Set(lesson.Name),
			e.Country.Set(course.Country.String),
			e.Grade.Set(course.Grade.Int),
			e.Subject.Set(course.Subject.String),
			e.SchoolID.Set(course.SchoolID.Int),
			e.TopicType.Set(entities.TopicTypeLiveLesson),
			e.Status.Set(entities.TopicStatusPublished),
			// e.AttachmentNames.Set(attachmentNames),
			// e.AttachmentURLs.Set(attachmentURLs),
			e.DisplayOrder.Set(1),
			e.PublishedAt.Set(now),
			e.TotalLOs.Set(0),
			e.ChapterID.Set(nil),
			e.IconURL.Set(nil),
			e.DeletedAt.Set(nil),
			e.EssayRequired.Set(false),
		)
		if err != nil {
			return nil, fmt.Errorf("entities.Topic: %s", err)
		}
		topics = append(topics, e)
		topicByCourseID[course.ID] = e
	}

	err := t.topicRepo.Create(ctx, t.db, topics)
	if err != nil {
		return nil, fmt.Errorf("topicRepo.Create: %s", err)
	}
	return topicByCourseID, nil
}
