package support

type Offset interface {
	int | string
}

type Paging[T Offset] struct {
	Limit  int
	Offset T
}
