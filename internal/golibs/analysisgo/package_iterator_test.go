package analysisgo

import (
	"go/constant"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageIterator_GetNext(t *testing.T) {
	path, err := os.Getwd()
	require.NoError(t, err)
	path = path + "/test"
	path, err = filepath.Abs(path)
	require.NoError(t, err)

	// cases with sign comment filter
	t.Run("constant and variable with sign comment filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

// @signComment => will be collected
// this is description
const b = 2

// append
func main() {
        // @signComment => will get an error
        fmt.Println("Hello, world")

		// this is description
        // @signComment => will be collected
        main := "hi"

        // @signComment => is a func calling, ignore without error
        print(main)
        // @signComment => empty, ignore without error
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(pkgs[0], ObjWithSignComments(map[string]bool{
			"@signComment": true,
		}))
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		assert.Equal(t, 1, errorCount)
		require.Len(t, res, 2)

		// validate result
		o := res[0]
		assert.Equal(t, "@signComment", o.sign)
		assert.Equal(t, "b", o.Name())
		assert.Equal(t, "2", o.Value().String())
		assert.NotNil(t, o.IsConst())
		assert.Equal(t, constant.Int, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o = res[1]
		assert.Equal(t, "@signComment", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Equal(t, "\"hi\"", o.Value().String())
		assert.NotNil(t, o.IsVar())
		assert.Equal(t, constant.String, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})

	t.Run("constant and variable with multiple sign comments filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

// @signComment1 => will be collected
// this is description
const b = 2

// append
func main() {
        // @signComment1 => will get an error
        fmt.Println("Hello, world")

		// this is description
        // @signComment2 => will be collected
        main := "hi"

        // @signComment2 => is a func calling, ignore without error
        print(main)
        // @signComment1 => empty, ignore without error
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(pkgs[0], ObjWithSignComments(map[string]bool{
			"@signComment1": true,
			"@signComment2": true,
		}))
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		assert.Equal(t, 1, errorCount)
		assert.Len(t, res, 2)

		// validate result
		o := res[0]
		assert.Equal(t, "@signComment1", o.sign)
		assert.Equal(t, "b", o.Name())
		assert.Equal(t, "2", o.Value().String())
		assert.NotNil(t, o.IsConst())
		assert.Equal(t, constant.Int, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o = res[1]
		assert.Equal(t, "@signComment2", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Equal(t, "\"hi\"", o.Value().String())
		assert.NotNil(t, o.IsVar())
		assert.Equal(t, constant.String, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})

	t.Run("constant and variable reassign value with multiple sign comments filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

// @signComment1 => will be collected
// this is description
const b = 2

// append
func main() {
        // @signComment1 => will get an error
        fmt.Println("Hello, world")

		// this is description
        // @signComment2 => will be collected
        main := "hi"

        // @signComment2 => will get an error
        main = "hello"
        // @signComment1 => is a func calling, ignore without error
        print(main)
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(pkgs[0], ObjWithSignComments(map[string]bool{
			"@signComment1": true,
			"@signComment2": true,
		}))
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		assert.Equal(t, 2, errorCount)
		assert.Len(t, res, 2)

		// validate result
		o := res[0]
		assert.Equal(t, "@signComment1", o.sign)
		assert.Equal(t, "b", o.Name())
		assert.Equal(t, "2", o.Value().String())
		assert.NotNil(t, o.IsConst())
		assert.Equal(t, constant.Int, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o = res[1]
		assert.Equal(t, "@signComment2", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Equal(t, "\"hi\"", o.Value().String())
		assert.NotNil(t, o.IsVar())
		assert.Equal(t, constant.String, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})

	t.Run("func and method with multiple sign comments filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

// @signComment-ignore
const b = 2
type C struct {}
// @signComment1 => will be collected
func (cc C) main() {
        // @signComment1 => will get an error
        fmt.Println("Hello, world")

        // @signComment-ignore
        main := 1

        // @signComment1 => is a func calling, ignore without error
        print(main)
        // @signComment1 => empty, ignore without error
}

// @signComment2 => will be collected
func main() {
        // @signComment1 => will get an error
        fmt.Println("Hello, world")

        // @signComment-ignore
        main := 1

        // @signComment1 => is a func calling, ignore without error
        print(main)
        // @signComment1 => empty, ignore without error
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(pkgs[0], ObjWithSignComments(map[string]bool{
			"@signComment1": true,
			"@signComment2": true,
		}))
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		assert.Equal(t, 2, errorCount)
		// TODO: check why duplicated
		assert.Len(t, res, 3)

		// validate result
		o := res[0]
		assert.Equal(t, "@signComment1", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Nil(t, o.Value())
		assert.NotNil(t, o.IsFunc())
		assert.NotNil(t, o.Owner)
		assert.Equal(t, "C", o.Owner.Name())

		o = res[1]
		assert.Equal(t, "@signComment2", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Nil(t, o.Value())
		assert.NotNil(t, o.IsFunc())
		assert.Nil(t, o.Owner)

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})

	// case with name filter
	t.Run("constant and variable with name filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

// this is description
// will be collected
const b = 2

// append
func main() {
        // will get an error
        fmt.Println("Hello, world")

		// this is description
        // will be collected
        variable := "hi"

        // is a func calling, ignore without error
        print(variable)
        // empty, ignore without error
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(pkgs[0], ObjWithName(func(name string) bool {
			if name != "b" && name != "variable" && name != "fmt" {
				return false
			}
			return true
		}))
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		assert.Equal(t, 1, errorCount)
		require.Len(t, res, 2)

		// validate result
		o := res[0]
		assert.Empty(t, o.sign)
		assert.Equal(t, "b", o.Name())
		assert.Equal(t, "2", o.Value().String())
		assert.NotNil(t, o.IsConst())
		assert.Equal(t, constant.Int, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o = res[1]
		assert.Empty(t, o.sign)
		assert.Equal(t, "variable", o.Name())
		assert.Equal(t, "\"hi\"", o.Value().String())
		assert.NotNil(t, o.IsVar())
		assert.Equal(t, constant.String, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})

	t.Run("constant and variable reassign value with multiple name filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

// this is description
// will be collected
const b = 2

// append
func main() {
        // will get an error
        fmt.Println("Hello, world")

		// this is description
        // will be collected
        variable := "hi"

        // will get an error
        variable = "hello"
        // is a func calling, ignore without error
        print(variable)
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(pkgs[0], ObjWithName(func(name string) bool {
			if name != "b" && name != "variable" && name != "fmt" {
				return false
			}
			return true
		}))
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		assert.Equal(t, 2, errorCount)
		assert.Len(t, res, 2)

		// validate result
		o := res[0]
		assert.Empty(t, o.sign)
		assert.Equal(t, "b", o.Name())
		assert.Equal(t, "2", o.Value().String())
		assert.NotNil(t, o.IsConst())
		assert.Equal(t, constant.Int, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o = res[1]
		assert.Empty(t, o.sign)
		assert.Equal(t, "variable", o.Name())
		assert.Equal(t, "\"hi\"", o.Value().String())
		assert.NotNil(t, o.IsVar())
		assert.Equal(t, constant.String, o.Value().Kind())
		assert.Nil(t, o.Owner)

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})

	t.Run("func and method with multiple name filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

const b = 2
type C struct {}
// this is description
// will be collected
func (cc C) main() {
        // will get an error
        fmt.Println("Hello, world")

        variable := 1

        // is a func calling, ignore without error
        print(variable)
        // empty, ignore without error	
}

// will be collected
func main() {
        // will get an error
        fmt.Println("Hello, world")

        variable := 1

        // is a func calling, ignore without error
        print(variable)
        // empty, ignore without error
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(pkgs[0], ObjWithName(func(name string) bool {
			if name != "main" && name != "fmt" {
				return false
			}
			return true
		}))
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		// TODO: check why there is no any error
		assert.Equal(t, 0, errorCount)
		// TODO: check why duplicated
		assert.Len(t, res, 3)

		// validate result
		o := res[0]
		assert.Empty(t, o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Nil(t, o.Value())
		assert.NotNil(t, o.IsFunc())
		assert.NotNil(t, o.Owner)
		assert.Equal(t, "C", o.Owner.Name())

		o = res[1]
		assert.Empty(t, o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Nil(t, o.Value())
		assert.NotNil(t, o.IsFunc())
		assert.Nil(t, o.Owner)

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})

	t.Run("constant, variable, func and method with both comment and name filter", func(t *testing.T) {
		const hello = `
package test

import "fmt"

// @signComment1 => will be collected
// this is description
const b = 2

type C struct {}
// @signComment1 => will be collected
func (cc C) main() {
        // @signComment1 => will get an error
        fmt.Println("Hello, world")

		// this is description
        // @signComment1 => will be collected
        main := "hi"

        // @signComment2 => is a func calling, ignore without error
        print(main)
        // @signComment1 => empty, ignore without error
}

// @signComment-ignore
func main() {
        // @signComment2 => will get an error
        fmt.Println("Hello, world")


		// @signComment2 => not match name
        variable := 2
		_ = variable

		// this is description
        // @signComment2 => will be collected
        main := "hi"

        // @signComment2 => is a func calling, ignore without error
        print(main)
        // @signComment2 => empty, ignore without error
}
// x
`

		pkgs, err := NewPackages(
			WithDir(path),
			WithFileContents(map[string][]byte{path + "/hello.go": []byte(hello)}),
		)
		require.NoError(t, err)
		assert.Len(t, pkgs, 1)
		assert.Equal(t, "test", pkgs[0].Name)

		pkgIter, err := NewPackageIterator(
			pkgs[0],
			ObjWithName(func(name string) bool {
				if name != "b" && name != "main" && name != "fmt" {
					return false
				}
				return true
			}),
			ObjWithSignComments(map[string]bool{
				"@signComment1": true,
				"@signComment2": true,
			}),
		)
		require.NoError(t, err)

		fileIter := pkgIter.GetNext()
		assert.NotNil(t, fileIter)

		var res []*Object
		errorCount := 0
		for {
			o, err := fileIter.GetNext()
			if err != nil {
				errorCount++
				continue
			}
			if o == nil {
				break
			}
			res = append(res, o)
		}
		assert.Equal(t, 2, errorCount)
		assert.Len(t, res, 4)

		// validate result
		o := res[0]
		assert.Equal(t, "@signComment1", o.sign)
		assert.Equal(t, "b", o.Name())
		assert.Equal(t, "2", o.Value().String())
		assert.NotNil(t, o.IsConst())
		assert.Equal(t, constant.Int, o.Value().Kind())

		o = res[1]
		assert.Equal(t, "@signComment1", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Nil(t, o.Value())
		assert.NotNil(t, o.IsFunc())
		assert.NotNil(t, o.Owner)
		assert.Equal(t, "C", o.Owner.Name())

		o = res[2]
		assert.Equal(t, "@signComment1", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Equal(t, "\"hi\"", o.Value().String())
		assert.NotNil(t, o.IsVar())
		assert.Equal(t, constant.String, o.Value().Kind())

		o = res[3]
		assert.Equal(t, "@signComment2", o.sign)
		assert.Equal(t, "main", o.Name())
		assert.Equal(t, "\"hi\"", o.Value().String())
		assert.NotNil(t, o.IsVar())
		assert.Equal(t, constant.String, o.Value().Kind())

		o, err = fileIter.GetNext()
		require.NoError(t, err)
		assert.Nil(t, o)

		fileIter = pkgIter.GetNext()
		assert.Nil(t, fileIter)
	})
}
