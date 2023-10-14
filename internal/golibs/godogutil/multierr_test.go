package godogutil

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MultierrChan(t *testing.T) {
	t.Parallel()
	str := &testStruct{}
	aval := "a"
	bval := "b"
	t.Run("one context step, one non context step", func(t *testing.T) {
		ctx, err := MultiErrChain(
			context.Background(),
			str.setACtx, aval,
			str.setB, bval,
		)
		assert.NoError(t, err)
		assert.Equal(t, aval, str.a)
		assert.Equal(t, aval, ctx.Value(setACtxCalled{}))
		assert.Equal(t, bval, str.b)
	})
	t.Run("2 non context step", func(t *testing.T) {
		_, err := MultiErrChain(
			context.Background(),
			str.setA, aval,
			str.setB, bval,
		)
		assert.NoError(t, err)
		assert.Equal(t, aval, str.a)
		assert.Equal(t, bval, str.b)
	})
	t.Run("normal steps, one error step", func(t *testing.T) {
		_, err := MultiErrChain(
			context.Background(),
			str.setACtx, aval,
			str.setB, bval,
			str.hasErr,
		)
		assert.ErrorIs(t, err, dummyerr)
	})
	t.Run("normal steps, one error with context step", func(t *testing.T) {
		_, err := MultiErrChain(
			context.Background(),
			str.setACtx, aval,
			str.setB, bval,
			str.hasErrCtx,
		)
		assert.ErrorIs(t, err, dummyerr)
	})

}

type testStruct struct {
	a string
	b string
}

type setBCtxCalled struct{}
type setACtxCalled struct{}

var dummyerr = errors.New("dummy err")

func (t *testStruct) hasErr() error {
	return dummyerr
}

func (t *testStruct) hasErrCtx(ctx context.Context) (context.Context, error) {
	return ctx, dummyerr
}

func (t *testStruct) setBCtx(ctx context.Context, val string) (context.Context, error) {
	t.b = val
	type setBCtxCalled struct{}
	return context.WithValue(ctx, setBCtxCalled{}, val), nil
}

func (t *testStruct) setB(val string) error {
	t.b = val
	return nil
}

func (t *testStruct) setACtx(ctx context.Context, val string) (context.Context, error) {
	t.a = val
	return context.WithValue(ctx, setACtxCalled{}, val), nil
}

func (t *testStruct) setA(val string) error {
	t.a = val
	return nil
}
