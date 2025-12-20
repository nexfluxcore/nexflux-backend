package dto

// RegisterRequest - request body untuk registrasi user baru
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Username string `json:"username"`
}

// LoginRequest - request body untuk login user
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse - response berisi token dan info user
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserAuthInfo `json:"user"`
}

// UserAuthInfo - informasi user yang dikembalikan saat auth
type UserAuthInfo struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	Username   string  `json:"username"`
	Avatar     string  `json:"avatar,omitempty"`
	Provider   string  `json:"provider,omitempty"`
	IsVerified bool    `json:"is_verified"`
}

// OAuthCallbackRequest - request dari OAuth callback
type OAuthCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// OAuthUserInfo - informasi user dari OAuth provider
type OAuthUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
}

// OAuthURLResponse - response berisi URL untuk OAuth
type OAuthURLResponse struct {
	URL   string `json:"url"`
	State string `json:"state"`
}
