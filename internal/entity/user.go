package entity

type User struct {
	Username string `json:"username" validate:"min=5,max=30"`
	Password string `json:"password" validate:"min=5,max=30"`
	Email    string `json:"email" validate:"email"`
	IsAdmin  bool   `json:"-"`
}

type UserLogin struct {
	Username string `json:"username" validate:"min=5,max=30"`
	Password string `json:"password" validate:"min=5,max=30"`
}

type UserResponse struct {
	Token string `json:"token"`
}
