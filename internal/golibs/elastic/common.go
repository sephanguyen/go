package elastic

type (
	SortParam struct {
		ColumnName string
		Ascending  bool
	}

	aggValueCountResponse struct {
		Aggregation aggregationValue `json:"aggregations"`
	}

	aggregationValue struct {
		CountValue countValue `json:"count_value"`
	}

	countValue struct {
		Value uint32 `json:"value"`
	}
)
