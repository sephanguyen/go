package dto

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// POST /chatgroups
type CreateChatGroupRequest struct {
	GroupName string   `json:"groupname"`
	Desc      string   `json:"desc"`
	Public    bool     `json:"public"`
	MaxUsers  int      `json:"maxusers"`
	Owner     string   `json:"owner"`
	Members   []string `json:"members"`
}

type CreateChatGroupResponse struct {
	Action          string `json:"action"`
	Application     string `json:"application"`
	ApplicationName string `json:"applicationName"`
	Organization    string `json:"organization"`
	URI             string `json:"uri"`
	Timestamp       uint64 `json:"timestamp"`
	Data            struct {
		GroupID string `json:"groupid"`
	} `json:"data"`
}

type AddMemberToChatGroupRequest struct {
	Usernames []string `json:"usernames"`
}

type AddMemberToChatGroupResponse struct {
	Action          string `json:"action"`
	Application     string `json:"application"`
	ApplicationName string `json:"applicationName"`
	Organization    string `json:"organization"`
	URI             string `json:"uri"`
	Timestamp       uint64 `json:"timestamp"`
	Data            struct {
		NewMembers []string `json:"newmembers"`
		GroupID    string   `json:"groupid"`
		Action     string   `json:"action"`
	} `json:"data"`
}

type RemoveMemberFromChatGroupResponse struct {
	Action          string      `json:"action"`
	Application     string      `json:"application"`
	ApplicationName string      `json:"applicationName"`
	Organization    string      `json:"organization"`
	URI             string      `json:"uri"`
	Timestamp       uint64      `json:"timestamp"`
	Data            interface{} `json:"data"`
}

type RemoveMember struct {
	Result  bool   `json:"result"`
	GroupID string `json:"groupid"`
	Reason  string `json:"reason"`
	Action  string `json:"action"`
	User    string `json:"user"`
}

func (i *RemoveMemberFromChatGroupResponse) UnmarshalJSON(d []byte) error {
	var x struct {
		RemoveMemberFromChatGroupResponse
		UnmarshalJSON struct{}
	}
	if err := json.Unmarshal(d, &x); err != nil {
		return err
	}

	var y map[string]interface{}
	_ = json.Unmarshal(d, &y)

	// delete all props except for the "data" property
	for key := range y {
		if key != "data" {
			delete(y, key)
		}
	}

	*i = x.RemoveMemberFromChatGroupResponse

	switch reflect.TypeOf(y["data"]).Kind() {
	case reflect.Slice:
		data := reflect.ValueOf(y["data"]).Interface().([]interface{})
		bytes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		members := []RemoveMember{}

		err = json.Unmarshal(bytes, &members)
		if err != nil {
			return err
		}

		i.Data = members
	case reflect.Map:
		data := reflect.ValueOf(y["data"]).Interface().(map[string]interface{})
		bytes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		member := RemoveMember{}

		err = json.Unmarshal(bytes, &member)
		if err != nil {
			return err
		}

		i.Data = member
	}
	return nil
}

func (i *RemoveMemberFromChatGroupResponse) GetRemoveMembers() ([]RemoveMember, error) {
	switch reflect.TypeOf(i.Data).Kind() {
	case reflect.Slice:
		data := i.Data.([]RemoveMember)
		return data, nil
	case reflect.Struct:
		data := i.Data.(RemoveMember)
		return []RemoveMember{data}, nil
	}
	return nil, fmt.Errorf("unhandled data type")
}
