package telegram

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"stay-connected/internal/services/db"
	"strings"
	"time"

	tb "gopkg.in/telebot.v3"
)

// Bot token from the environment variable
var botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
var bot *tb.Bot

// Initialize the Telegram bot
func InitTelegram() error {
	log.Println(botToken)
	pref := tb.Settings{
		Token:  botToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}
	var err error

	bot, err = tb.NewBot(pref)
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}

	bot.Handle("/start", HandleStart)
	bot.Handle(tb.OnText, HandleMessages)
	SendMessage(939659614, "Bot is running")

	bot.Start()

	return nil
}

// State management
var userStates = make(map[int64]string)
var userData = make(map[int64]*UserInfo)

// UserInfo stores user's email and password
type UserInfo struct {
	Email    string
	Password string
}

// Handler for the /start command
func HandleStart(c tb.Context) error {
	user := c.Sender()
	userStates[user.ID] = "awaiting_email" // Set state to awaiting email
	return c.Send("What is your email?")
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Function to handle the interaction and get email and password
func HandleMessages(c tb.Context) error {
	user := c.Sender()
	text := c.Message().Text

	// Retrieve the current state of the user
	state := userStates[user.ID]

	switch state {
	case "awaiting_email":
		if !emailRegex.MatchString(strings.TrimSpace(text)) {
			return c.Send("The email format is invalid. Please enter a valid email address.")
		}
		userData[user.ID] = &UserInfo{Email: text}
		userStates[user.ID] = "awaiting_password"
		return c.Send("Got it. Now, what is your password?")
	case "awaiting_password":
		userData[user.ID].Password = text
		id, err := db.Login(userData[user.ID].Email, userData[user.ID].Password)
		if err != nil {
			userData[user.ID].Password = "awaiting_email"
			return c.Send(err.Error() + ". Awaiting for email")
		}
		err = db.LinkTelegram(id, user.ID)
		if err != nil {
			userData[user.ID].Password = "awaiting_email"
			return c.Send(err.Error() + ". Awaiting for email")
		}
		userStates[user.ID] = ""
		return c.Send("Thanks! Your email and password have been saved.")
	default:
		return c.Send("Please use the /start command to begin.")
	}
}

func SendMessage(telegramId int, text string) error {
	recipient := &tb.User{ID: int64(telegramId)}
	_, err := bot.Send(recipient, text)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}
