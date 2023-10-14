package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
)

func GetEventFromSGEvent(ev dto.SGEmailEvent) constants.EmailEvent {
	// ev.Type only exist when it's bounce
	if ev.Type != "" {
		return constants.EmailEventBySGEvent[constants.SGEmailEvent(ev.Type)]
	}
	return constants.EmailEventBySGEvent[constants.SGEmailEvent(ev.Event)]
}

func GetEventTypeFromSGEvent(ev constants.SGEmailEvent) constants.EmailEventType {
	return constants.EventEventTypeBySGEvent[ev]
}

func GetSGEventFromEvent(ev constants.EmailEvent) constants.SGEmailEvent {
	return constants.SGEventByEmailEvent[ev]
}

func GetEventTypeFromEvent(ev constants.EmailEvent) constants.EmailEventType {
	return constants.EmailEventTypeByEvent[ev]
}

func GetEventIdentifyInfo(ev dto.SGEmailEvent) string {
	return fmt.Sprintf("%s-%s", ev.EmailRecipientID, GetEventFromSGEvent(ev))
}

func FromEventIdentifyInfo(info string) (emailRecipientID, event string) {
	splitArrInfos := strings.Split(info, "-")
	if len(splitArrInfos) < 2 {
		return "", ""
	}

	return splitArrInfos[0], splitArrInfos[1]
}

func NewMockRequest(method string, bodyContent interface{}, headers map[string][]string) (*http.Request, io.ReadCloser) {
	byteData, _ := json.Marshal(bodyContent)
	header := http.Header{}
	for k, vs := range headers {
		for _, v := range vs {
			header.Add(k, v)
		}
	}
	expectedBody := io.NopCloser(bytes.NewBuffer(byteData))
	r := &http.Request{
		Method: method,
		Body:   io.NopCloser(bytes.NewBuffer(byteData)),
		Header: header,
	}
	return r, expectedBody
}
