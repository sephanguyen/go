package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

// TaggedUserRepo repository
type TaggedUserRepo struct{}

func (r *TaggedUserRepo) FindByTagIDsAndUserIDs(ctx context.Context, db database.QueryExecer, tagIDs, userIDs pgtype.TextArray) ([]*entities.TaggedUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "TaggedUserRepo.FindByTagIDsAndUserIDs")
	defer span.End()

	taggedUsers := &entities.TaggedUsers{}
	taggedUser := &entities.TaggedUser{}
	fields, _ := taggedUser.FieldMap()

	stmt := fmt.Sprintf(`
	SELECT %s FROM %s
	WHERE deleted_at is null
	AND tag_id = ANY($1::TEXT[])
	AND user_id = ANY($2::TEXT[])
	`, strings.Join(fields, ", "), taggedUser.TableName())

	if err := database.Select(ctx, db, stmt, tagIDs, userIDs).ScanAll(taggedUsers); err != nil {
		return nil, err
	}

	return *taggedUsers, nil
}
