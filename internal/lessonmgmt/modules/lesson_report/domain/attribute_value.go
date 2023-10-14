package domain

type AttributeValue struct {
	Int         int
	String      string
	Bool        bool
	IntArray    []int32
	StringArray []string
	IntSet      []int32
	StringSet   []string
}

func (a *AttributeValue) SetInt(v int) {
	a.Int = v
}

func (a *AttributeValue) SetString(v string) {
	a.String = v
}

func (a *AttributeValue) SetBool(v bool) {
	a.Bool = v
}

func (a *AttributeValue) SetIntArray(v []int32) {
	a.IntArray = v
}

func (a *AttributeValue) SetStringArray(v []string) {
	a.StringArray = v
}

func (a *AttributeValue) SetIntSet(v []int32) {
	intMap := make(map[int32]bool)
	a.IntSet = make([]int32, 0, len(v))
	for i := range v {
		if _, ok := intMap[v[i]]; ok {
			continue
		}
		a.IntSet = append(a.IntSet, v[i])
		intMap[v[i]] = true
	}
}

func (a *AttributeValue) SetStringSet(v []string) {
	stringMap := make(map[string]bool)
	a.StringSet = make([]string, 0, len(v))
	for i := range v {
		if _, ok := stringMap[v[i]]; ok {
			continue
		}
		a.StringSet = append(a.StringSet, v[i])
		stringMap[v[i]] = true
	}
}
