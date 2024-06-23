package db 

type User struct{
	Username string `json:"username"`
	Password string `json:"password"`
	Email string `json:"email"`
}

type FullUser struct{
	Username string `json:"username"`
	Password string `json:"password"`
	Email string `json:"email"`
	Usage int8 `json:"usage"`
	Limit int8 `json:"limit"`
}