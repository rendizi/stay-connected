package db

import (
	supa "github.com/nedpals/supabase-go"
	"os"
	"stay-connected/internal/encryption"
)

func Insert(username string, password string, email string) ([]User, error) {
	supabase := supa.CreateClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"))
	encryptedPassword, err := encryption.Encrypt(password, []byte(os.Getenv("SECRET_KEY_PASSWORD")))
	if err != nil{
		return nil, err 
	}
	row := User{
		Username: username,
		Password: encryptedPassword,
		Email:    email,
	}

	var results []User
	err = supabase.DB.From("users").Insert(row).Execute(&results)
	if err != nil{
		return nil, err 
	}
	return results, nil 
}

func Get()([]FullUser, error){
	supabase := supa.CreateClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"))
	var results []FullUser
	err := supabase.DB.From("users").Select("*").Execute(&results)
	if err != nil {
		return nil, err 
	}
	return results, nil 
}