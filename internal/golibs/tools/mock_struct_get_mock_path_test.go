package tools

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// $ go test -bench=BenchmarkGetMockPathFrom -benchmem
// goos: linux
// goarch: amd64
// pkg: github.com/manabie-com/backend/internal/golibs/tools
// cpu: 11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz
// BenchmarkGetMockPathFromUsingReplace0-16        26315986                41.27 ns/op            8 B/op          1 allocs/op
// BenchmarkGetMockPathFromUsingRegexp0-16          6040974               185.6 ns/op            32 B/op          3 allocs/op
// BenchmarkGetMockPathFromUsingReplace1-16        23513272                45.05 ns/op           16 B/op          1 allocs/op
// BenchmarkGetMockPathFromUsingRegexp1-16          5533228               217.1 ns/op            56 B/op          4 allocs/op
// BenchmarkGetMockPathFromUsingReplace2-16        23201341                44.98 ns/op           16 B/op          1 allocs/op
// BenchmarkGetMockPathFromUsingRegexp2-16          5295254               212.4 ns/op            56 B/op          4 allocs/op
// BenchmarkGetMockPathFromUsingReplace3-16        18582847                60.65 ns/op           48 B/op          1 allocs/op
// BenchmarkGetMockPathFromUsingRegexp3-16          4602608               240.0 ns/op           121 B/op          4 allocs/op
// PASS

func getMockPathFromUsingReplace(path string) string {
	if !strings.HasPrefix(path, "internal/") {
		panic(fmt.Errorf("path %q is malformed, expected \"internal/...\"", path))
	}
	return strings.Replace(path, "internal/", "mock/", 1)
}

var testGetMockPathRe = regexp.MustCompile(`^internal\/`)

func getMockPathFromUsingRegexp(path string) string {
	if !strings.HasPrefix(path, "internal/") {
		panic(fmt.Errorf("path %q is malformed, expected \"internal/...\"", path))
	}
	return testGetMockPathRe.ReplaceAllLiteralString(path, "mock/")
}

func BenchmarkGetMockPathFromUsingReplace0(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingReplace(testPaths[0])
	}
}

func BenchmarkGetMockPathFromUsingRegexp0(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingRegexp(testPaths[0])
	}
}

func BenchmarkGetMockPathFromUsingReplace1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingReplace(testPaths[1])
	}
}

func BenchmarkGetMockPathFromUsingRegexp1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingRegexp(testPaths[1])
	}
}

func BenchmarkGetMockPathFromUsingReplace2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingReplace(testPaths[2])
	}
}

func BenchmarkGetMockPathFromUsingRegexp2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingRegexp(testPaths[2])
	}
}

func BenchmarkGetMockPathFromUsingReplace3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingReplace(testPaths[3])
	}
}

func BenchmarkGetMockPathFromUsingRegexp3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getMockPathFromUsingRegexp(testPaths[3])
	}
}
