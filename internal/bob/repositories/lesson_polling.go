package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type LessonPollingRepo struct{}

func (l *LessonPollingRepo) Create(ctx context.Context, db database.Ext, poll *entities.LessonPolling) (*entities.LessonPolling, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonPollingRepo.Create")
	defer span.End()
	fieldNames, args := poll.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		poll.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)
	if _, err := db.Exec(ctx, query, args...); err != nil {
		return nil, err
	}

	return poll, nil
}
