package sendgrid

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// TODO: implement logic for mocks
type Mock struct {
}

func NewSendGridMock() SendGridClient {
	return &Mock{}
}

func (*Mock) Send(_ *mail.SGMailV3) (string, error) {
	id := idutil.ULIDNow()
	return id, nil
}

func (*Mock) SendWithContext(_ context.Context, _ *mail.SGMailV3) (string, error) {
	id := idutil.ULIDNow()
	return id, nil
}

func (*Mock) AuthenticateHTTPRequest(header http.Header, _ []byte) (bool, error) {
	code := header.Get("code")
	switch {
	case strings.HasPrefix(code, "4"): // client errors
		return false, nil
	case strings.HasPrefix(code, "5"): // server errors
		return false, fmt.Errorf("unable to authenticate")
	case strings.HasPrefix(code, "2"): // okela
		return true, nil
	default:
		return true, nil
	}
}
