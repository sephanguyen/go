package support

type SliceOfSlice struct {
	Data         [][]string
	IndexCompare []int
}

func (a SliceOfSlice) Len() int { return len(a.Data) }
func (a SliceOfSlice) Less(i, j int) bool {
	for y := 0; y < len(a.IndexCompare); y++ {
		if a.Data[i][a.IndexCompare[y]] != a.Data[j][a.IndexCompare[y]] {
			return a.Data[i][a.IndexCompare[y]] < a.Data[j][a.IndexCompare[y]]
		}
	}
	return false
}
func (a SliceOfSlice) Swap(i, j int) { a.Data[i], a.Data[j] = a.Data[j], a.Data[i] }
