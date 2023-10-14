package search

type (
	PagingParam struct {
		FromIdx    int64
		NumberRows uint32
	}

	SortParam struct {
		ColumnName string
		Ascending  bool
	}

	InsertionContent struct {
		ID   string
		Data interface{}
	}
)
