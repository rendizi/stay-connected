package db

import (
	"errors"
	"fmt"
	supa "github.com/nedpals/supabase-go"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"stay-connected/internal/encryption"
	"strings"
)

var supabase *supa.Client

func InitSupabase() {
	log.Println("Connecting to db...")
	log.Println(os.Getenv("SUPABASE_URL"))
	supabase = supa.CreateClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_KEY"))
}

func Register(email string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	username := strings.ReplaceAll(email, ".", "")
	row := InsertUser{
		Email:    email,
		Password: string(hashedPassword),
		Username: username,
	}

	var inserted []map[string]interface{}

	err = supabase.DB.From("users").Insert(row).Execute(&inserted)
	if err != nil {
		return err
	}
	log.Println(inserted)
	anotherRow := Usage{
		Used:   0,
		Limit:  100,
		UserId: int8(inserted[0]["id"].(float64)),
	}
	var data []map[string]interface{}
	err = supabase.DB.From("usage").Insert(anotherRow).Execute(&data)
	return err
}

func Login(email string, password string) (int8, error) {
	username := strings.ReplaceAll(email, ".", "")
	var results []map[string]interface{}
	query := supabase.DB.From("users").Select("*").Eq("username", username)

	err := query.Execute(&results)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %v", err)
	}

	if len(results) == 0 {
		return 0, errors.New("user not found")
	}

	storedHashedPassword, ok := results[0]["password"].(string)
	if !ok {
		return 0, errors.New("invalid stored password format")
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(password))
	if err != nil {
		return 0, errors.New("incorrect password")
	}

	return int8(results[0]["id"].(float64)), nil
}

func LinkInstagram(username string, password string, userId int8) error {
	encryptedPassword, err := encryption.Encrypt(password, []byte(os.Getenv("SECRET_KEY_PASSWORD")))
	if err != nil {
		return err
	}

	row := Instagram{
		Username: username,
		Password: encryptedPassword,
		UserId:   userId,
	}
	var data []map[string]interface{}
	err = supabase.DB.From("instagram").Insert(row).Execute(&data)
	return err
}

func UnlinkInstagram(userId int8) error {
	var data []map[string]interface{}
	err := supabase.DB.From("instagram").Delete().Eq("user_id", fmt.Sprintf("%d", userId)).Execute(&data)
	return err
}

func LinkTelegram(userId int8, telegramId int64) error {
	updateData := map[string]interface{}{
		"telegram": telegramId,
	}
	var data []map[string]interface{}
	err := supabase.DB.From("users").Update(updateData).Eq("id", fmt.Sprintf("%d", userId)).Execute(&data)
	return err
}

func GetUsers() ([]GetUser, error) {
	var data []GetUser
	err := supabase.DB.From("users").Select("*").Execute(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetInstas() ([]Instagram, error) {
	var data []Instagram
	err := supabase.DB.From("instagram").Select("*").Execute(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetEmail(id int8) (string, int8, error) {
	var data []map[string]interface{}
	err := supabase.DB.From("users").Select("*").Eq("id", fmt.Sprintf("%d", id)).Execute(&data)
	if err != nil {
		return "", 0, err
	}
	return data[0]["email"].(string), data[0]["telegram"].(int8), nil
}

func LeftToReactLimit(id int8) (int8, error) {
	var data []map[string]interface{}
	err := supabase.DB.From("usage").Select("*").Eq("user_id", fmt.Sprintf("%d", id)).Execute(&data)
	if err != nil {
		return 0, err
	}
	return int8(data[0]["limit"].(float64) - data[0]["used"].(float64)), nil
}

func Used(id int8, count int8) error {
	// Create a Supabase client

	// Fetch the current "used" and "limit" values for the given "id"
	var data []map[string]interface{}
	err := supabase.DB.From("usage").
		Select("used, limit").
		Eq("user_id", fmt.Sprintf("%d", id)).
		Execute(&data)
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("no record found for id: %d", id)
	}

	used := int8(data[0]["used"].(float64)) // Type assertion from float64 to int8

	updateData := map[string]interface{}{
		"used": used + count,
	}
	err = supabase.DB.From("usage").
		Update(updateData).
		Eq("user_id", fmt.Sprintf("%d", id)).
		Execute(nil)
	if err != nil {
		return fmt.Errorf("error updating data: %w", err)
	}

	return nil
}

func Delete(username string) error {
	var data map[string]interface{}
	err := supabase.DB.From("users").Delete().Eq("username", username).Execute(&data)
	return err
}
