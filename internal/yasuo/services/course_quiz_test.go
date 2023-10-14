package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_services "github.com/manabie-com/backend/mock/yasuo/services"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func md5String(s string) string {
	h := md5.New()
	io.WriteString(h, s)

	return fmt.Sprintf("%x", h.Sum(nil))
}

func mockForUpload(uploader *mock_services.Uploader, content, bucket string, err error) {
	uploader.On("UploadWithContext", mock.AnythingOfType("*context.cancelCtx"), &s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("/content/" + md5String(content) + ".html"),
		Body:        strings.NewReader(content),
		ACL:         aws.String("public-read"),
		ContentType: aws.String("text/html"),
	}).Once().Return(nil, err)
}

func TestInTextArray(t *testing.T) {
	t.Parallel()
	ta := database.TextArray([]string{"t1", "t2", "t3"})
	t.Run("found", func(t *testing.T) {
		t.Parallel()
		s := database.Text("t1")
		found := inTextArray(ta, s)
		assert.True(t, found)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		s := database.Text("notFound")
		found := inTextArray(ta, s)
		assert.False(t, found)
	})
}

func TestCourseService_AssignLosToQuiz(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	s := CourseService{
		EurekaDBTrace: db,
		DBTrace:       db,
		QuizSetRepo:   quizSetRepo,
	}
	t.Run("happy case", func(t *testing.T) {
		quizSets := entities.QuizSets{}
		expect := database.TextArray([]string{"lo1", "lo2", "lo3"})
		for _, e := range expect.Elements {
			quizSets = append(quizSets, &entities.QuizSet{LoID: e})
		}
		quiz := &entities.Quiz{ExternalID: database.Text("quiz_id_1")}
		quizSetRepo.On("GetQuizSetsContainQuiz", ctx, s.DBTrace, quiz.ExternalID).Once().Return(quizSets, nil)
		err := s.AssignLosToQuizV1(ctx, quiz)
		assert.Nil(t, err)
		assert.Equal(t, expect, quiz.LoIDs)
	})
	t.Run("can not find quiz sets", func(t *testing.T) {
		quizSets := entities.QuizSets{}
		quiz := &entities.Quiz{ExternalID: database.Text("quiz_id_1")}
		quizSetRepo.On("GetQuizSetsContainQuiz", ctx, s.DBTrace, quiz.ExternalID).Once().Return(quizSets, pgx.ErrNoRows)
		err := s.AssignLosToQuizV1(ctx, quiz)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
	t.Run("duplicate los", func(t *testing.T) {
		quizSets := entities.QuizSets{}
		losTextArray := database.TextArray([]string{"lo1", "lo1", "lo2"})
		expect := database.TextArray([]string{"lo1", "lo2"})
		for _, e := range losTextArray.Elements {
			quizSets = append(quizSets, &entities.QuizSet{LoID: e})
		}
		quiz := &entities.Quiz{ExternalID: database.Text("quiz_id_1")}
		quizSetRepo.On("GetQuizSetsContainQuiz", ctx, s.DBTrace, quiz.ExternalID).Once().Return(quizSets, nil)
		err := s.AssignLosToQuizV1(ctx, quiz)
		assert.Nil(t, err)
		assert.Equal(t, expect, quiz.LoIDs)
	})
}
