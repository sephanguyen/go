package dto

type CreateConversationRequest struct {
	OwnerVendorID   string
	MemberVendorIDs []string
}

type CreateConversationResponse struct {
	ConversationID string
}

type AddConversationMembersRequest struct {
	ConversationID  string
	MemberVendorIDs []string
}

type AddConversationMembersResponse struct {
	ConversationID string
}

type RemoveConversationMembersRequest struct {
	ConversationID  string
	MemberVendorIDs []string
}

type RemoveConversationMembersResponse struct {
	ConversationID string
	FailedMembers  []FailedRemoveMember
}

type FailedRemoveMember struct {
	MemberVendorID string
	Reason         string
}
