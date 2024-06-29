package db

type InsertUser struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type GetUser struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Id       int8   `json:"id"`
	Username string `json:"username"`
	Telegram string `json:"telegram"`
}

type Usage struct {
	Used   int8 `json:"used"`
	Limit  int8 `json:"limit"`
	UserId int8 `json:"user_id"`
}

type Instagram struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserId   int8   `json:"user_id"`
}
