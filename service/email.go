package service

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"

	"gopkg.in/gomail.v2"

	"github.com/rohitxdev/go-api-template/embedded"
	"github.com/rohitxdev/go-api-template/env"
)

var smtpPort = func() uint {
	port, err := strconv.ParseUint(env.SMTP_PORT, 10, 16)
	if err != nil {
		panic("could not parse SMTP port: " + err.Error())
	}
	return uint(port)
}()

var smtpDialer = gomail.NewDialer(env.SMTP_HOST, int(smtpPort), env.SMTP_USERNAME, env.SMTP_PASSWORD)

var passwordResetTemplate = func() *template.Template {
	tmpl, err := template.New("password-reset.tmpl").ParseFS(embedded.FS, "templates/emails/password-reset.tmpl")
	if err != nil {
		panic("could not parse password reset template: " + err.Error())
	}
	return tmpl
}()

/*----------------------------------- Send Email ----------------------------------- */

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

/*----------------------------------- Send Password Reset Link ----------------------------------- */

func SendPasswordResetLink(urlWithToken string, toEmail string) error {
	var buf bytes.Buffer
	err := passwordResetTemplate.Execute(&buf, map[string]any{"URL": urlWithToken})
	if err != nil {
		return fmt.Errorf("could not execute password reset template: %w", err)
	}
	if err := SendEmail(&Email{
		ToAddresses: []string{toEmail},
		Subject:     "Password Reset",
		Body:        buf.String(),
		ContentType: "text/html",
		FromAddress: "rohitreddy.gangwar@gmail.com",
	}); err != nil {
		return fmt.Errorf("could not send password reset link: %w", err)
	}
	return nil
}
