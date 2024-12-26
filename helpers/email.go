package helpers

import (
	"fmt"

	"github.com/ankush-web-eng/contest-backend/config"
	"gopkg.in/gomail.v2"
)

type EmailDetails struct {
	From    string
	To      string
	Subject string
	Body    string
}

func SendEmail(details EmailDetails) error {
	smtpConfig := config.LoadSMTPConfig()

	m := gomail.NewMessage()
	m.SetHeader("From", details.From)
	m.SetHeader("To", details.To)
	m.SetHeader("Subject", details.Subject)
	m.SetBody("text/plain", details.Body)

	d := gomail.NewDialer(smtpConfig.Host, smtpConfig.Port, smtpConfig.Username, smtpConfig.Password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}
	return nil
}
