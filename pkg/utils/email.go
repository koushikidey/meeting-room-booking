package utils

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	from := os.Getenv("EMAIL_USER")
	pass := os.Getenv("EMAIL_PASS")

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, from, pass)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}
	return nil
}