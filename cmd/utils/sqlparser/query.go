package sqlparser

type (
	SqlQuery struct {
		Action ActionType
		Entity EntityType
		Name   string
	}

	SqlQueryParser interface {
		Build(text string) (*SqlQuery, error)
	}

	SqlQueryProcessor interface {
		CanProcess(query string) bool
		GetParser() SqlQueryParser
	}
)
