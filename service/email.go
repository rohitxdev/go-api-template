package service

import (
	"bytes"
	"fmt"
	"html/template"

	"gopkg.in/gomail.v2"

	"github.com/rohitxdev/go-api-template/embedded"
)

func NewSMTPClient(host string, port int, username string, password string) *gomail.Dialer {
	return gomail.NewDialer(host, port, username, password)
}

var passwordResetTemplate = func() *template.Template {
	tmpl, err := template.New("password-reset.tmpl").ParseFS(embedded.FS, "templates/emails/password-reset.tmpl")
	if err != nil {
		panic("could not parse password reset template: " + err.Error())
	}
	return tmpl
}()

type EmailClient struct {
	dialer *gomail.Dialer
}

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

func (e *EmailClient) SendEmail(email *Email) error {
	msg := gomail.NewMessage()
	msg.SetHeaders(map[string][]string{
		"From":    {msg.FormatAddress(email.FromAddress, email.FromName)},
		"Subject": {email.Subject},
		"To":      email.ToAddresses,
		"Cc":      email.Cc,
		"Bcc":     email.Bcc,
	})
	msg.SetBody(email.ContentType, email.Body)
	if err := e.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}
	return nil
}

/*----------------------------------- Send Password Reset Link ----------------------------------- */

func (e *EmailClient) SendPasswordResetLink(urlWithToken string, toEmail string) error {
	var buf bytes.Buffer
	err := passwordResetTemplate.Execute(&buf, map[string]any{"URL": urlWithToken})
	if err != nil {
		return fmt.Errorf("could not execute password reset template: %w", err)
	}
	if err := e.SendEmail(&Email{
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
