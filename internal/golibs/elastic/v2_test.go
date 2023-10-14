package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func getfilecontent(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
func Test_BulkIndexWithResourcePath(t *testing.T) {
	t.Parallel()
	t.Run("all error", func(t *testing.T) {
		resp, err := getfilecontent("./resp/bulk_error.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		rp := "manabie"
		ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
		})
		totalSuccess, err := mockElas.BulkIndexWithResourcePath(ctx, map[string]Doc{
			"1": NewDoc(""),
			"2": NewDoc(""),
		}, "index")
		assert.Equal(t, 0, totalSuccess)
		assert.Error(t, err)
	})
	t.Run("all success", func(t *testing.T) {
		resp, err := getfilecontent("./resp/bulk_success.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		rp := "manabie"
		ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
		})
		totalSuccess, err := mockElas.BulkIndexWithResourcePath(ctx, map[string]Doc{
			"1": NewDoc(""),
			"2": NewDoc(""),
		}, "index")
		assert.Equal(t, 2, totalSuccess)
		assert.NoError(t, err)
	})
}
func Test_SearchJwt(t *testing.T) {
	t.Parallel()
	t.Run("error no token in ctx", func(t *testing.T) {
		resp, err := getfilecontent("./resp/search_success.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		source := NewSearchSource()
		rp := "manabie"
		ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
		})
		_, err = DoSearchFromSourceUsingJwtToken(ctx, mockElas, "chat", source)
		assert.Equal(t, fmt.Errorf("SearchUsingJwtToken: ctx has no incoming grpc metadata"), err)
	})
	t.Run("success", func(t *testing.T) {
		resp, err := getfilecontent("./resp/search_success.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		rp := "manabie"
		ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
		})
		ctx = metadata.NewIncomingContext(ctx, metadata.MD{"token": []string{"sometoken"}})
		source := NewSearchSource()
		elasresp, err := DoSearchFromSourceUsingJwtToken(ctx, mockElas, "chat", source)
		assert.NoError(t, err)
		err = ParseSearchResponse(elasresp.Body, func(h *SearchHit) error {
			return nil
		})
		assert.NoError(t, err)
	})

}
func Test_Search(t *testing.T) {
	t.Parallel()
	t.Run("erorr parsing resp", func(t *testing.T) {
		resp, err := getfilecontent("./resp/search_success.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		// rp := "manabie"
		source := NewSearchSource()
		elasresp, err := DoSearchFromSource(context.Background(), mockElas, "chat", source)
		assert.NoError(t, err)
		err = ParseSearchResponse(elasresp.Body, func(h *SearchHit) error {
			return fmt.Errorf("dummy")
		})
		assert.Error(t, err)
	})
	t.Run("success", func(t *testing.T) {
		resp, err := getfilecontent("./resp/search_success.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		// rp := "manabie"
		source := NewSearchSource()
		elasresp, err := DoSearchFromSource(context.Background(), mockElas, "chat", source)
		assert.NoError(t, err)
		err = ParseSearchResponse(elasresp.Body, func(h *SearchHit) error {
			return nil
		})
		assert.NoError(t, err)
	})

}
func Test_BulkIndex(t *testing.T) {
	t.Parallel()
	t.Run("all error", func(t *testing.T) {
		resp, err := getfilecontent("./resp/bulk_error.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		rp := "manabie"
		ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
		})
		totalSuccess, err := mockElas.BulkIndex(ctx, map[string][]byte{
			"1": []byte(""),
			"2": []byte(""),
		}, "index", "chat")
		assert.Equal(t, 0, totalSuccess)
		assert.Error(t, err)
	})
	t.Run("all success", func(t *testing.T) {
		resp, err := getfilecontent("./resp/bulk_success.json")
		assert.NoError(t, err)

		mockElas, close := NewMockSearchFactory(resp)
		defer close()
		rp := "manabie"
		ctx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{ResourcePath: rp},
		})
		totalSuccess, err := mockElas.BulkIndex(ctx, map[string][]byte{
			"1": []byte(""),
			"2": []byte(""),
		}, "index", "chat")
		assert.Equal(t, 2, totalSuccess)
		assert.NoError(t, err)
	})
}

func Test_Doc(t *testing.T) {
	t.Parallel()
	{
		type SomeT struct {
			A string `json:"a"`
		}
		a := SomeT{
			A: "some_val",
		}
		d := NewDoc(a)
		AssertDocIsValid(t, d)
		bs, _ := json.Marshal(d)
		assert.Equal(t, `{"resource_path":"","a":"some_val"}`, string(bs))
	}
	{
		type SomeT struct {
			A string `json:"a"`
		}
		a := SomeT{
			A: "some_val",
		}
		d := NewDoc(a)
		AssertDocIsValid(t, d)
		d._mandatory.ResourcePath = "some school"
		bs, _ := json.Marshal(d)
		assert.Equal(t, `{"resource_path":"some school","a":"some_val"}`, string(bs))
	}
	{
		type SomeT struct {
			a string
		}
		a := SomeT{
			a: "don't show me",
		}
		d := NewDoc(a)
		AssertDocIsValid(t, d)
		d._mandatory.ResourcePath = "some school"

		bs, err := json.Marshal(d)
		assert.NoError(t, err)
		assert.Equal(t, `{"resource_path":"some school"}`, string(bs))
	}
}
