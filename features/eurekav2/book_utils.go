package eurekav2

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func (s *suite) GenerateFakeBookContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	count := 2
	id := idutil.ULIDNow()

	chapters, err := s.GenerateFakeChapters(id, count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("GenerateFakeBookContent: %w", err)
	}
	book := domain.Book{
		ID:       id,
		Name:     "Faked Book content BDD",
		Chapters: chapters,
		IsV2:     true,
	}

	stepState.BookContent = book
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GenerateFakeChapters(bookID string, count int) (chapters []domain.Chapter, err error) {
	if count < 1 {
		count = 1
	}
	chapters = make([]domain.Chapter, count)
	for i := 0; i < count; i++ {
		id := idutil.ULIDNow()
		topics, err := s.GenerateFakeTopics(id, count)
		if err != nil {
			return nil, err
		}
		c := domain.Chapter{
			BookID:       bookID,
			ID:           id,
			Name:         fmt.Sprintf("Fake chapter BDD %d", i),
			DisplayOrder: i,
			Topics:       topics,
		}
		chapters[i] = c
	}
	return chapters, nil
}

func (s *suite) GenerateFakeTopics(chapterID string, count int) (topics []domain.Topic, err error) {
	if count < 1 {
		count = 1
	}
	topics = make([]domain.Topic, count)
	for i := 0; i < count; i++ {
		id := idutil.ULIDNow()
		materials, err := s.GenerateFakeMaterials(id, count, constants.LearningObjective)
		if err != nil {
			return nil, err
		}
		t := domain.Topic{
			ChapterID:         chapterID,
			ID:                id,
			Name:              fmt.Sprintf("Fake topic BDD %d", i),
			DisplayOrder:      i,
			IconURL:           "https://backoffice.staging.manabie.io/images/manabie.png",
			LearningMaterials: materials,
		}
		topics[i] = t
	}
	return topics, nil
}

func (s *suite) GenerateFakeMaterials(topicID string, count int, lmType constants.LearningMaterialType) (materials []domain.LearningMaterial, err error) {
	if count < 1 {
		count = 1
	}
	materials = make([]domain.LearningMaterial, count)
	for i := 0; i < count; i++ {
		id := idutil.ULIDNow()
		m := domain.LearningMaterial{
			TopicID:      topicID,
			ID:           id,
			Name:         fmt.Sprintf("Fake topic BDD %d", i),
			DisplayOrder: i,
			Type:         lmType,
			Published:    true,
		}
		materials[i] = m
	}
	return materials, nil
}

func (s *suite) SeedBookContentRecursive(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	book := stepState.BookContent
	bookStmt := `INSERT INTO books (book_id, name, is_v2, created_at, updated_at) VALUES($1, $2, TRUE, NOW(), NOW())`
	_, err := s.EurekaDBTrace.Exec(ctx, bookStmt, book.ID, book.Name)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("SeedBookContentRecursive: cannot seed book with `id:%s`, %v", book.ID, err)
	}

	chapterStmt := `INSERT INTO chapters (chapter_id, book_id, name, display_order, created_at, updated_at)
					VALUES($1, $2, $3, $4, NOW(), NOW())`
	var topics []domain.Topic
	for _, c := range book.Chapters {
		_, err := s.EurekaDBTrace.Exec(ctx, chapterStmt, c.ID, c.BookID, c.Name, c.DisplayOrder)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("SeedBookContentRecursive: cannot seed chapter: %v, chapter: %v", err, c)
		}
		topics = append(topics, c.Topics...)
	}

	topicStmt := `INSERT INTO topics (topic_id, chapter_id, name, display_order, icon_url, grade, subject, topic_type, created_at, updated_at)
					VALUES($1, $2, $3, $4, $5, -1, 'SUBJECT_NONE', 'TOPIC_TYPE_NONE', NOW(), NOW())`
	var materials []domain.LearningMaterial
	for _, t := range topics {
		_, err := s.EurekaDBTrace.Exec(ctx, topicStmt, t.ID, t.ChapterID, t.Name, t.DisplayOrder, t.IconURL)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("SeedBookContentRecursive: cannot seed topic: %v, topic: %v", err, t)
		}
		materials = append(materials, t.LearningMaterials...)
	}

	materialStmt := `INSERT INTO learning_material (learning_material_id, topic_id, name, display_order, is_published, type, created_at, updated_at)
					VALUES($1, $2, $3, $4, $5, $6, NOW(), NOW())`
	for _, m := range materials {
		_, err := s.EurekaDBTrace.Exec(ctx, materialStmt, m.ID, m.TopicID, m.Name, m.DisplayOrder, m.Published, m.Type)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("SeedBookContentRecursive: cannot seed LM: %v, material: %v", err, m)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
