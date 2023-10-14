package application

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type LessonSearchIndexer struct {
	Logger      *zap.Logger
	DB          database.Ext
	SearchRepo  infrastructure.SearchRepo
	LessonRepo  infrastructure.LessonRepo
	StudentRepo infrastructure.StudentRepo
	UserRepo    infrastructure.UserRepo
}

func (l *LessonSearchIndexer) SyncLessonIndex(ctx context.Context, lessonIDs []string) error {
	return l.indexLessonDocument(ctx, lessonIDs)
}

func (l *LessonSearchIndexer) indexLessonDocument(ctx context.Context, lessonIDs []string) error {
	lessons, err := l.LessonRepo.GetLessonByIDs(ctx, l.DB, lessonIDs)
	if err != nil {
		return fmt.Errorf("cannot get lessons: %w", err)
	}
	lessonDocs := domain.LessonSearchs{}
	g, groupCtx := errgroup.WithContext(ctx)
	lessonSearchChan := make(chan *domain.LessonSearch)
	for _, lesson := range lessons {
		lessonDoc := &domain.LessonSearch{
			LessonID:       lesson.LessonID,
			LocationID:     lesson.LocationID,
			TeachingMedium: string(lesson.TeachingMedium),
			TeachingMethod: string(lesson.TeachingMethod),
			CreatedAt:      lesson.CreatedAt,
			UpdatedAt:      lesson.UpdatedAt,
			StartTime:      lesson.StartTime,
			EndTime:        lesson.EndTime,
			DeletedAt:      lesson.DeletedAt,
			LessonTeacher:  lesson.Teachers.GetIDs(),
		}

		if lesson.TeachingMethod == domain.LessonTeachingMethodGroup {
			lessonDoc.ClassID = lesson.ClassID
			lessonDoc.CourseID = lesson.CourseID
		}

		lessonMemberInfos := make([]LessonMemberInfo, 0, len(lesson.Learners))
		for _, ls := range lesson.Learners {
			lessonMemberInfos = append(lessonMemberInfos, LessonMemberInfo{
				userId:   ls.LearnerID,
				courseId: ls.CourseID,
			})
		}
		g.Go(func() error {
			lessonMembers, err := l.fetchLessonMembers(groupCtx, lessonMemberInfos)
			if err != nil {
				return fmt.Errorf("failed to get lesson member data: %w", err)
			}
			lessonDoc.AddLessonMembers(lessonMembers)
			lessonSearchChan <- lessonDoc
			return nil
		})
	}
	go func() {
		err := g.Wait()
		if err != nil {
			return
		}
		close(lessonSearchChan)
	}()
	for lessonSearch := range lessonSearchChan {
		lessonDocs = append(lessonDocs, lessonSearch)
	}
	if err := g.Wait(); err != nil {
		return err
	}

	_, err = l.SearchRepo.BulkUpsert(ctx, lessonDocs)
	if err != nil {
		return fmt.Errorf("cannot upsert lesson document: %w", err)
	}
	return nil
}

type LessonMemberInfo struct {
	userId   string
	courseId string
}

func (l *LessonSearchIndexer) fetchLessonMembers(ctx context.Context, lessonMemberInfo []LessonMemberInfo) ([]*domain.LessonMemberEs, error) {
	userIDs := []string{}
	lessonMemberMap := make(map[string]*domain.LessonMemberEs)
	for _, lm := range lessonMemberInfo {
		userIDs = append(userIDs, lm.userId)
		lessonMemberMap[lm.userId] = &domain.LessonMemberEs{
			ID:       lm.userId,
			CourseID: lm.courseId,
		}
	}
	if len(lessonMemberInfo) != 0 {
		students, err := l.StudentRepo.FindStudentProfilesByIDs(ctx, l.DB, database.TextArray(userIDs))
		if err != nil {
			return nil, fmt.Errorf("l.StudentRepo.FindStudentProfilesByIDs: %w", err)
		}
		for _, st := range students {
			if lm, ok := lessonMemberMap[st.ID.String]; ok {
				lm.CurrentGrade = int(st.CurrentGrade.Int)
			}
		}
		users, err := l.UserRepo.Retrieve(ctx, l.DB, database.TextArray(userIDs))
		if err != nil {
			return nil, fmt.Errorf("l.UserRepo.Retrieve: %w", err)
		}
		for _, user := range users {
			if lm, ok := lessonMemberMap[user.ID.String]; ok {
				lm.Name = user.LastName.String
			}
		}
	}
	lessonMember := []*domain.LessonMemberEs{}
	for _, lm := range lessonMemberMap {
		lessonMember = append(lessonMember, lm)
	}
	return lessonMember, nil
}
