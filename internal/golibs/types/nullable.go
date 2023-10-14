package types

type NullBool struct {
	NotNull bool
	Bool    bool
}

func NewBool(b bool) NullBool {
	return NullBool{
		NotNull: true,
		Bool:    b,
	}
}

type NullStr struct {
	NotNull bool
	Str     string
}
type NullStrArr struct {
	NotNull bool
	StrArr  []string
}

func (n NullStrArr) ToInterfaces() []interface{} {
	var ret = make([]interface{}, 0, len(n.StrArr))
	for _, i := range n.StrArr {
		ret = append(ret, i)
	}
	return ret
}

type NullInt64 struct {
	NotNull bool
	I64     int64
}

func NewInt64(i int64) NullInt64 {
	return NullInt64{
		NotNull: true,
		I64:     i,
	}
}

func NewStr(str string) NullStr {
	return NullStr{
		NotNull: true,
		Str:     str,
	}
}

func NewStrArr(arr []string) NullStrArr {
	if arr == nil {
		return NullStrArr{}
	}
	return NullStrArr{
		NotNull: true,
		StrArr:  arr,
	}
}

type Nullables struct {
	None  bool
	Value []interface{}
}
