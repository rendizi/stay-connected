package mailer

import (
	"net/smtp"
	"os"
)

func Send(email string, bodyText string) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("EMAIL_FROM"),
		os.Getenv("EMAIL_PASSWORD"),
		"smtp.gmail.com",
	)

	from := os.Getenv("EMAIL_FROM")
	subject := "Subject: Daily inst\n"
	to := "To: " + email + "\n"
	bodyText += "\n\n View more on website: somehost"

	msg := []byte(to + "From: " + from + "\n" + subject + "\n" + bodyText)

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		from,
		[]string{email},
		msg,
	)
	if err != nil {
		return err
	}

	return nil
}
