package user

// OAuth user operations Input structs
type CreateInput struct {
	Email     string
	Name      string
	AvatarURL string
}

type UpdateInput struct {
	UserID string
	Role   string
}
