package handler 


type SignCredentials struct{
	Email string `json:"email"`
	Password string `json:"password"`
}

type InstagramCredentials struct{
	Username string `json:"username"`
	Password string `json:"password"`
	UserId int8 `json:"user_id"`
}

type Usage struct{
	Used int8 `json:"used"`
	Limit int8 `json:"limit"`
	UserId int8 `json:"user_id"`
}