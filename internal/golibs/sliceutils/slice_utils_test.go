package sliceutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

type TStruct struct {
	Name string
	Age  int
}

var entities = []TStruct{
	{
		Name: "name1",
		Age:  20,
	},
	{
		Name: "name2",
		Age:  10,
	},
	{
		Name: "name3",
		Age:  50,
	},
}

var strArr = []string{
	"string 1",
	"string 2",
	"string 3",
}

func TestSliceUtil(t *testing.T) {
	t.Parallel()
	t.Run("has intersect", func(t *testing.T) {
		ret := Intersect([]string{"a", "b", "d"}, []string{"b", "c", "d"})
		assert.ElementsMatch(t, []string{"b", "d"}, ret)
	})
	t.Run("no intersect", func(t *testing.T) {
		ret := Intersect([]string{"a", "b"}, []string{"c", "d"})
		assert.Empty(t, ret)
	})
	t.Run("empty items", func(t *testing.T) {
		ret := Intersect([]string{}, nil)
		assert.Empty(t, ret)
	})
}

func TestContainFunc(t *testing.T) {
	t.Parallel()

	t.Run("contains struct", func(t *testing.T) {
		t.Parallel()
		res := ContainFunc(entities, TStruct{Name: "name1"}, func(t1, t2 TStruct) bool {
			return t1.Name == t2.Name
		})

		res2 := ContainFunc(entities, TStruct{Name: "anyname"}, func(t1, t2 TStruct) bool {
			return t1.Name == t2.Name
		})
		assert.Equal(t, true, res)
		assert.Equal(t, false, res2)
	})

	t.Run("contains comparable", func(t *testing.T) {
		t.Parallel()
		res := ContainFunc(strArr, "string 1", func(t1, t2 string) bool {
			return t1 == t2
		})

		res2 := ContainFunc(strArr, "string any dif", func(t1, t2 string) bool {
			return t1 == t2
		})
		assert.Equal(t, true, res)
		assert.Equal(t, false, res2)
	})
}

func TestFilter(t *testing.T) {
	t.Parallel()

	t.Run("filter struct", func(t *testing.T) {
		t.Parallel()
		res := Filter(entities, func(t1 TStruct) bool {
			return t1.Age < 40
		})
		res2 := Filter(entities, func(t1 TStruct) bool {
			return t1.Age > 100
		})

		assert.Equal(t, entities[0], res[0])
		assert.Equal(t, entities[1], res[1])

		assert.Equal(t, 0, len(res2))

	})

	t.Run("filter comparable", func(t *testing.T) {
		t.Parallel()
		res := Filter(strArr, func(s string) bool {
			return s == "string 1"
		})

		res2 := Filter(strArr, func(s string) bool {
			return s == "any name"
		})

		assert.Equal(t, strArr[0], res[0])
		assert.Equal(t, 0, len(res2))
	})
}

func TestMap(t *testing.T) {
	t.Parallel()

	t.Run("map struct to string", func(t *testing.T) {
		t.Parallel()
		res := Map(entities, func(t1 TStruct) string {
			return t1.Name
		})

		assert.Equal(t, entities[0].Name, res[0])
		assert.Equal(t, entities[1].Name, res[1])
		assert.Equal(t, entities[2].Name, res[2])

		assert.Equal(t, len(entities), len(res))
	})
}

func TestMapSkip(t *testing.T) {
	t.Parallel()

	t.Run("map struct to string, skip the satisfy element", func(t *testing.T) {
		t.Parallel()
		res := MapSkip(entities, func(t1 TStruct) string {
			return t1.Name
		}, func(t1 TStruct) bool {
			return t1.Name == "name2"
		})

		assert.Equal(t, entities[0].Name, res[0])
		assert.Equal(t, entities[2].Name, res[1])

		assert.Equal(t, len(entities)-1, len(res))
	})
}

func TestFilterWithReferenceList(t *testing.T) {
	t.Parallel()

	t.Run("filter a struct using string", func(t *testing.T) {
		t.Parallel()
		filter := []string{"name2", "name3"}
		res := FilterWithReferenceList(filter, entities, func(rl []string, li TStruct) bool {
			return slices.Contains(rl, li.Name)
		})

		assert.Equal(t, filter[0], res[0].Name)
		assert.Equal(t, filter[1], res[1].Name)

		assert.Equal(t, len(filter), len(res))
	})
}

func TestUnorderedEqual(t *testing.T) {
	t.Parallel()

	t.Run("compare string slices", func(t *testing.T) {
		t.Parallel()
		first := []string{"name2", "name3"}
		second := []string{"name3", "name2"}
		res := UnorderedEqual(first, second)

		assert.True(t, res)

	})

	t.Run("compare struct pointer slices", func(t *testing.T) {
		t.Parallel()
		type Type struct {
			A string
			B int
		}
		first := []*Type{
			{
				A: "a",
				B: 1,
			},
		}
		second := []*Type{
			{
				B: 1,
				A: "a",
			},
		}
		res := UnorderedEqual(first, second)

		assert.True(t, res)

	})

	t.Run("compare struct pointer slices", func(t *testing.T) {
		t.Parallel()
		type Type struct {
			A string
			B int
		}
		first := []*Type{
			{
				A: "a",
				B: 2,
			},
		}
		second := []*Type{
			{
				B: 1,
				A: "a",
			},
		}
		res := UnorderedEqual(first, second)

		assert.False(t, res)

	})

	t.Run("compare struct slices", func(t *testing.T) {
		t.Parallel()
		type Type struct {
			A string
			B int
		}
		first := []Type{
			{
				A: "a",
				B: 1,
			},
		}
		second := []Type{
			{
				B: 1,
				A: "a",
			},
		}
		res := UnorderedEqual(first, second)

		assert.True(t, res)
	})

	t.Run("compare struct slices failed", func(t *testing.T) {
		t.Parallel()
		type Type struct {
			A string
			B int
		}
		first := []Type{
			{
				A: "a",
				B: 2,
			},
		}
		second := []Type{
			{
				B: 1,
				A: "a",
			},
		}
		res := UnorderedEqual(first, second)

		assert.False(t, res)
	})

	t.Run("compare number slices", func(t *testing.T) {
		t.Parallel()

		first := []int{
			1, 0, 3, 5, 1,
		}
		second := []int{
			0, 1, 1, 3, 5,
		}
		res := UnorderedEqual(first, second)

		assert.True(t, res)
	})

	t.Run("compare number slices failed", func(t *testing.T) {
		t.Parallel()

		first := []int{
			1, 0, 3, 5, 2,
		}
		second := []int{
			0, 1, 1, 3, 5,
		}
		res := UnorderedEqual(first, second)

		assert.False(t, res)
	})
}

func TestRemove(t *testing.T) {
	t.Parallel()

	t.Run("remove succeed", func(t *testing.T) {
		t.Parallel()
		data := []TStruct{
			{
				Name: "name1",
				Age:  20,
			},
			{
				Name: "name2",
				Age:  10,
			},
			{
				Name: "name3",
				Age:  50,
			},
		}
		expectedData := []TStruct{
			{
				Name: "name1",
				Age:  20,
			},
			{
				Name: "name3",
				Age:  50,
			},
		}
		res := Remove(data, func(v TStruct) bool {
			return v.Name == "name2"
		})

		assert.Equal(t, expectedData, res)

	})

}

func TestReduce(t *testing.T) {
	t.Parallel()

	t.Run("reduce succeed", func(t *testing.T) {
		t.Parallel()
		data := []TStruct{
			{
				Name: "name1",
				Age:  20,
			},
			{
				Name: "name2",
				Age:  10,
			},
			{
				Name: "name3",
				Age:  50,
			},
		}

		expectedSumAge := 80
		sum, _ := Reduce(data, func(sum int, v TStruct) (int, error) {
			sum += v.Age
			return sum, nil
		}, 0)

		assert.Equal(t, expectedSumAge, sum)

	})

}

func TestChunk(t *testing.T) {
	t.Run("chunk size is not a positive number", func(t *testing.T) {
		// arrange
		t.Parallel()
		slice := []string{"apple", "mastermgmt"}
		chunkSize := -1
		expectedChunks := [][]string{{"apple", "mastermgmt"}}

		// act
		chunks := Chunk(slice, chunkSize)

		// assert
		assert.Equal(t, expectedChunks, chunks)
	})

	t.Run("chunk size is higher than slice length", func(t *testing.T) {
		// arrange
		t.Parallel()
		slice := []string{"apple", "banana", "cherry", "date", "elderberry"}
		chunkSize := 10
		expectedChunks := [][]string{{"apple", "banana", "cherry", "date", "elderberry"}}

		// act
		chunks := Chunk(slice, chunkSize)

		// assert
		assert.Equal(t, expectedChunks, chunks)
	})

	t.Run("chunk size is equal slice length", func(t *testing.T) {
		// arrange
		t.Parallel()
		slice := []string{"apple", "banana", "cherry", "date", "elderberry"}
		chunkSize := 5
		expectedChunks := [][]string{{"apple", "banana", "cherry", "date", "elderberry"}}

		// act
		chunks := Chunk(slice, chunkSize)

		// assert
		assert.Equal(t, expectedChunks, chunks)
	})

	t.Run("slice has a quite higher capacity", func(t *testing.T) {
		// arrange
		t.Parallel()
		slice := make([]string, 0, 20)
		slice = append(slice, []string{"apple", "banana", "kien", "date", "elderberry"}...)
		chunkSize := 3
		expectedChunks := [][]string{{"apple", "banana", "kien"}, {"date", "elderberry"}}

		// act
		chunks := Chunk(slice, chunkSize)

		// assert
		assert.Equal(t, expectedChunks, chunks)
	})

	t.Run("chunk size is smaller than slice length", func(t *testing.T) {
		// arrange
		t.Parallel()
		slice := []string{"apple", "banana", "cherry", "month", "elderberry"}
		chunkSize := 2
		expectedChunks := [][]string{{"apple", "banana"}, {"cherry", "month"}, {"elderberry"}}

		// act
		chunks := Chunk(slice, chunkSize)

		// assert
		assert.Equal(t, expectedChunks, chunks)
	})
}
