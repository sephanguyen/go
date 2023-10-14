package hasura

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/yudai/gojsondiff"
)

// Response represents a response from a Hasura query.
type Response struct {
	data []byte
}

// Data returns the raw data as byte slice.
func (r Response) Data() []byte {
	return r.data
}

// Compare returns true if r is identical to o.
func (r Response) Compare(o *Response) (bool, error) {
	diff, err := gojsondiff.New().Compare(r.Data(), o.Data())
	if err != nil {
		return false, fmt.Errorf("gojsondiff.Differ.Compare(): %s", err)
	}
	return !diff.Modified(), nil
}

// EavesdropHTTPResponse sends HTTP request req to the handler h, then extracts
// the response while also writing that response back to wr.
//
// This is useful to sniff responses while serving HTTP requests.
func EavesdropHTTPResponse(h http.Handler, wr http.ResponseWriter, req *http.Request) (*Response, error) {
	buf := &bytes.Buffer{}
	teeWriter := newTeeHTTPResponseWriter(wr, buf)
	h.ServeHTTP(teeWriter, req)
	return parseResponse(buf, teeWriter.Header().Get("Content-Encoding"))
}

func parseResponse(in *bytes.Buffer, encoding string) (*Response, error) {
	switch encoding {
	case "gzip":
		gzreader, err := gzip.NewReader(in)
		if err != nil {
			return nil, fmt.Errorf("failed to create new gzip reader: %s", err)
		}
		out, err := io.ReadAll(gzreader)
		if err != nil {
			return nil, fmt.Errorf("failed to read data using gzip reader: %s", err)
		}
		return &Response{data: out}, nil
	case "":
		return &Response{data: in.Bytes()}, nil
	default:
		return nil, fmt.Errorf("invalid encoding: %q", encoding)
	}
}

// teeHTTPResponseWriter is used to safely eavesdrops the response from http.ResponseWriter
// without affecting the actual request-response.
type teeHTTPResponseWriter struct {
	httpWriter http.ResponseWriter
	tee        io.Writer
}

func newTeeHTTPResponseWriter(wr http.ResponseWriter, tee io.Writer) *teeHTTPResponseWriter {
	return &teeHTTPResponseWriter{httpWriter: wr, tee: tee}
}

func (wr *teeHTTPResponseWriter) Header() http.Header {
	return wr.httpWriter.Header()
}

func (wr *teeHTTPResponseWriter) WriteHeader(statusCode int) {
	wr.httpWriter.WriteHeader(statusCode)
}

func (wr *teeHTTPResponseWriter) Write(b []byte) (int, error) {
	return wr.tee.Write(b)
}
