package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveMemberFromChatGroupResponse_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name                string
		Data                []byte
		Expect              interface{}
		ExpectRemovedMember []RemoveMember
	}{
		{
			Name: "remove single member success",
			Data: []byte(`
			{
				"action": "delete",
				"application": "application",
				"applicationName": "123456",
				"data": {
				  "result": true,
				  "action": "remove_member",
				  "user": "user_id",
				  "groupid": "group_id"
				},
				"organization": "manabie",
				"properties": {},
				"timestamp": 1693124263876,
				"uri": "https://a61.chat.agora.io"
			  }
			`),
			Expect: &RemoveMemberFromChatGroupResponse{
				Action:          "delete",
				Application:     "application",
				ApplicationName: "123456",
				Organization:    "manabie",
				URI:             "https://a61.chat.agora.io",
				Timestamp:       1693124263876,
				Data: RemoveMember{
					Result:  true,
					Action:  "remove_member",
					GroupID: "group_id",
					User:    "user_id",
				},
			},
			ExpectRemovedMember: []RemoveMember{
				{
					Result:  true,
					Action:  "remove_member",
					GroupID: "group_id",
					User:    "user_id",
				},
			},
		},
		{
			Name: "remove multiple member success",
			Data: []byte(`
			{
				"action": "delete",
				"application": "application",
				"applicationName": "123456",
				"data": [
					{
						"result": true,
						"action": "remove_member",
						"user": "user_id1",
						"groupid": "group_id"
					},
					{
						"result": true,
						"action": "remove_member",
						"user": "user_id2",
						"groupid": "group_id"
					}
				],
				"organization": "manabie",
				"properties": {},
				"timestamp": 1693124263876,
				"uri": "https://a61.chat.agora.io"
			}
			`),
			Expect: &RemoveMemberFromChatGroupResponse{
				Action:          "delete",
				Application:     "application",
				ApplicationName: "123456",
				Organization:    "manabie",
				URI:             "https://a61.chat.agora.io",
				Timestamp:       1693124263876,
				Data: []RemoveMember{
					{
						Result:  true,
						Action:  "remove_member",
						GroupID: "group_id",
						User:    "user_id1",
					},
					{
						Result:  true,
						Action:  "remove_member",
						GroupID: "group_id",
						User:    "user_id2",
					},
				},
			},
			ExpectRemovedMember: []RemoveMember{
				{
					Result:  true,
					Action:  "remove_member",
					GroupID: "group_id",
					User:    "user_id1",
				},
				{
					Result:  true,
					Action:  "remove_member",
					GroupID: "group_id",
					User:    "user_id2",
				},
			},
		},
		{
			Name: "remove multiple member with some fails",
			Data: []byte(`
			{
				"action": "delete",
				"application": "application",
				"applicationName": "123456",
				"data": [
					{
						"result": true,
						"action": "remove_member",
						"user": "user_id1",
						"groupid": "group_id"
					},
					{
						"result": false,
						"action": "remove_member",
						"user": "user_id3",
						"groupid": "group_id",
						"reason": "some reason"
					},
					{
						"result": true,
						"action": "remove_member",
						"user": "user_id2",
						"groupid": "group_id"
					}
				],
				"organization": "manabie",
				"properties": {},
				"timestamp": 1693124263876,
				"uri": "https://a61.chat.agora.io"
			}
			`),
			Expect: &RemoveMemberFromChatGroupResponse{
				Action:          "delete",
				Application:     "application",
				ApplicationName: "123456",
				Organization:    "manabie",
				URI:             "https://a61.chat.agora.io",
				Timestamp:       1693124263876,
				Data: []RemoveMember{
					{
						Result:  true,
						Action:  "remove_member",
						GroupID: "group_id",
						User:    "user_id1",
					},
					{
						Result:  false,
						Action:  "remove_member",
						GroupID: "group_id",
						User:    "user_id3",
						Reason:  "some reason",
					},
					{
						Result:  true,
						Action:  "remove_member",
						GroupID: "group_id",
						User:    "user_id2",
					},
				},
			},
			ExpectRemovedMember: []RemoveMember{
				{
					Result:  true,
					Action:  "remove_member",
					GroupID: "group_id",
					User:    "user_id1",
				},
				{
					Result:  false,
					Action:  "remove_member",
					GroupID: "group_id",
					User:    "user_id3",
					Reason:  "some reason",
				},
				{
					Result:  true,
					Action:  "remove_member",
					GroupID: "group_id",
					User:    "user_id2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp := &RemoveMemberFromChatGroupResponse{}

			err := resp.UnmarshalJSON(tc.Data)
			assert.Nil(t, err)
			assert.Equal(t, tc.Expect, resp)

			removedMembers, err := resp.GetRemoveMembers()
			assert.Nil(t, err)
			assert.Equal(t, tc.ExpectRemovedMember, removedMembers)
		})
	}
}
