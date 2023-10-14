package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type StudentTopicOverdueRepo struct{}

type TopicDueDate struct {
	Topic   *entities.Topic
	Duedate pgtype.Timestamptz
}

// func (r *StudentTopicOverdueRepo)
func (r *StudentTopicOverdueRepo) RemoveStudentTopicOverdue(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, topicIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentTopicOverdueRepo.RemoveStudenTopicOverdue")
	defer span.End()

	t := &entities.StudentTopicOverdue{}
	query := fmt.Sprintf("DELETE FROM %s WHERE student_id = $1 AND topic_id=ANY($2)", t.TableName())
	_, e := db.Exec(ctx, query, &studentID, &topicIDs)
	if e != nil {
		return errors.Wrap(e, "AssignOverdueTopic.Remove")
	}
	return nil
}

func (r *StudentTopicOverdueRepo) RetrieveStudentTopicOverdue(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*TopicDueDate, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentTopicOverdueRepo.RetrieveStudentTopicOverdue")
	defer span.End()

	topic := &entities.Topic{}
	topicOverdue := &entities.StudentTopicOverdue{}
	topicFields := database.GetFieldNames(topic)

	stmt := "SELECT t.%s, std.due_date FROM %s AS t JOIN %s AS std ON t.topic_id = std.topic_id WHERE std.student_id = $1 AND std.due_date<NOW() AND t.deleted_at IS NULL ORDER BY std.due_date desc"
	query := fmt.Sprintf(stmt, strings.Join(topicFields, ", t."), topic.TableName(), topicOverdue.TableName())

	rows, err := db.Query(ctx, query, &studentID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	var topics []*TopicDueDate

	for rows.Next() {
		t := &entities.Topic{}
		var dueDate pgtype.Timestamptz
		if err := rows.Scan(append(database.GetScanFields(t, topicFields), &dueDate)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		topicDueDate := &TopicDueDate{
			Topic:   t,
			Duedate: dueDate,
		}
		topics = append(topics, topicDueDate)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return topics, nil
}
