package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type VirtualLessonPollingRepo struct{}

func (v *VirtualLessonPollingRepo) Create(ctx context.Context, db database.Ext, poll *VirtualLessonPolling) (*VirtualLessonPolling, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualLessonPollingRepo.Create")
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
