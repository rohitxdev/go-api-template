package types

type LogInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignUpRequest struct {
	LogInRequest
	ConfirmPassword string `json:"confirm_password" validate:"required,eqcsfield=Password"`
}
