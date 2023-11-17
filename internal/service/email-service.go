package service

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"

	"gopkg.in/gomail.v2"

	"github.com/rohitxdev/go-api-template/internal/env"
)

var smtpPort, _ = strconv.ParseUint(env.SMTP_PORT, 10, 16)
var smtpDialer = gomail.NewDialer(env.SMTP_HOST, int(smtpPort), env.SMTP_USERNAME, env.SMTP_PASSWORD)

var passwordResetTemplate *template.Template

func init() {
	var err error
	passwordResetTemplate, err = template.New("password-reset.tmpl").ParseFiles(env.PROJECT_ROOT + "/templates/emails/password-reset.tmpl")
	if err != nil {
		panic(err)
	}
}

type Email struct {
	ToAddresses []string
	Cc          []string
	Bcc         []string
	Subject     string
	ContentType string
	Body        string
	FromAddress string
	FromName    string
}

func SendEmail(email *Email) error {
	msg := gomail.NewMessage()
	msg.SetHeaders(map[string][]string{
		"From":    {msg.FormatAddress(email.FromAddress, email.FromName)},
		"Subject": {email.Subject},
		"To":      email.ToAddresses,
		"Cc":      email.Cc,
		"Bcc":     email.Bcc,
	})
	msg.SetBody(email.ContentType, email.Body)
	if err := smtpDialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}
	return nil
}

func SendPasswordResetLink(urlWithToken string, toEmail string) {
	var buf bytes.Buffer
	err := passwordResetTemplate.Execute(&buf, map[string]any{"URL": urlWithToken})
	if err != nil {
		panic(err)
	}
	SendEmail(&Email{ToAddresses: []string{toEmail}, Subject: "Password Reset", Body: buf.String(), ContentType: "text/html", FromAddress: "rohitreddy.gangwar@gmail.com"})
}
