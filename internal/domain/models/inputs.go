package models

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
	AppID    int    `validate:"required,gt=0"`
}

type RegisterInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}
