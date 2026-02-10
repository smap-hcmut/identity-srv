package repository

// OAuth user operations Options structs
type UpsertOptions struct {
	Email     string
	Name      string
	AvatarURL string
}

type UpdateOptions struct {
	UserID string
	Role   string
}

type DetailOptions struct {
	UserID string
}
