package tom

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/elastic"
)

// Ensure new deployment always have correct index settings for Elasticsearch
// dynamic index will make those fields behave incorrectly
// TODO: move this logic somewhere else
func UpsertElasticsearchFields(e *elastic.SearchFactoryImpl) error {
	return e.UpsertFieldDefinition(context.Background(), "conversations", "resource_path", `
{
  "properties": {
    "resource_path": {
      "type": "keyword"
    }
  }
}`)
}
