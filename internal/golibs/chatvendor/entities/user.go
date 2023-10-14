package entities

type User struct {
	UserID       string
	VendorUserID string
	Activated    bool
	CreatedAt    uint64
	UpdatedAt    uint64
}
