package tools

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// This benchmark shows that using strings.SplitN is greatly faster than using regexp.
// $ go test -bench=BenchmarkGetServiceName -benchmem
// goos: linux
// goarch: amd64
// pkg: github.com/manabie-com/backend/internal/golibs/tools
// cpu: 11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz
// BenchmarkGetServiceNameUsingSplit0-16           20413395                55.99 ns/op           48 B/op          1 allocs/op
// BenchmarkGetServiceNameUsingRegexp0-16           7075076               162.5 ns/op            32 B/op          1 allocs/op
// BenchmarkGetServiceNameUsingSplit1-16           19493385                59.55 ns/op           48 B/op          1 allocs/op
// BenchmarkGetServiceNameUsingRegexp1-16           6659203               172.3 ns/op            32 B/op          1 allocs/op
// BenchmarkGetServiceNameUsingSplit2-16           19895595                55.41 ns/op           48 B/op          1 allocs/op
// BenchmarkGetServiceNameUsingRegexp2-16           5738907               202.5 ns/op            32 B/op          1 allocs/op
// BenchmarkGetServiceNameUsingSplit3-16           19945081                57.45 ns/op           48 B/op          1 allocs/op
// BenchmarkGetServiceNameUsingRegexp3-16           2229043               546.2 ns/op            32 B/op          1 allocs/op
// PASS

func getServiceNameUsingSplit(path string) string {
	if !strings.HasPrefix(path, "internal/") {
		panic(fmt.Errorf("path %q is malformed, expected \"internal/...\"", path))
	}
	return strings.SplitN(path, "/", 3)[1]
}

var testGetServiceNameRe = regexp.MustCompile(`^internal\/(\w+)(?:\/.*)?$`)

func getServiceNameUsingRegexp(path string) string {
	m := testGetServiceNameRe.FindStringSubmatch(path)
	if len(m) != 2 {
		panic(fmt.Errorf("path %q is malformed, expected \"internal/...\"", path))
	}
	return m[1]
}

func BenchmarkGetServiceNameUsingSplit0(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingSplit(testPaths[0])
	}
}

func BenchmarkGetServiceNameUsingRegexp0(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingRegexp(testPaths[0])
	}
}

func BenchmarkGetServiceNameUsingSplit1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingSplit(testPaths[1])
	}
}

func BenchmarkGetServiceNameUsingRegexp1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingRegexp(testPaths[1])
	}
}

func BenchmarkGetServiceNameUsingSplit2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingSplit(testPaths[2])
	}
}

func BenchmarkGetServiceNameUsingRegexp2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingRegexp(testPaths[2])
	}
}

func BenchmarkGetServiceNameUsingSplit3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingSplit(testPaths[3])
	}
}

func BenchmarkGetServiceNameUsingRegexp3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getServiceNameUsingRegexp(testPaths[3])
	}
}
