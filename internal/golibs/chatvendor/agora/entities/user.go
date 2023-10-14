package entities

type User struct {
	NickName  string `json:"nickname"`
	Type      string `json:"type"`
	UUID      string `json:"uuid"`
	UserName  string `json:"username"`
	Activated bool   `json:"activated"`
	Created   uint64 `json:"created"`
	Modified  uint64 `json:"modified"`
}
