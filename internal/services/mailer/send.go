package mailer

import (
	"net/smtp"
	"os"
)

func Send(email string, bodyHTML string) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("EMAIL_FROM"),
		os.Getenv("EMAIL_PASSWORD"),
		"smtp.gmail.com",
	)

	from := os.Getenv("EMAIL_FROM")
	subject := "Subject: Daily inst\n"
	to := "To: " + email + "\n"

	// Add the MIME header for HTML content
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"

	// Combine headers and body into the final message
	msg := []byte(to + "From: " + from + "\n" + subject + mime + "\n" + bodyHTML)

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
