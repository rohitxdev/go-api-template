package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"

	"gopkg.in/gomail.v2"
)

type Email struct {
	To          []string
	Subject     string
	Body        string
	ContentType string
}

func SendEmail(e *Email) error {
	from := "rohitreddy.gangwar@gmail.com"
	password := "zxkz vsgr smqb muwy"

	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", e.To...)
	msg.SetHeader("Subject", e.Subject)
	msg.SetBody(e.ContentType, e.Body)
	if err := d.DialAndSend(msg); err != nil {
		return fmt.Errorf("could not send email: %s", err.Error())
	}
	for _, v := range e.To {
		fmt.Printf("Sent email to %s\n", v)
	}
	fmt.Println("Sent email successfully!")
	return nil
}

func SendOTP(code string, toEmail string) {
	t := template.New("2fa.templ")
	templ, err := t.ParseFiles("./templates/2fa.templ")
	if err != nil {
		log.Fatalln("could not parse HTML:", err.Error())
	}
	var buf bytes.Buffer
	err = templ.Execute(&buf, map[string]any{"Code": code, "Name": toEmail})
	if err != nil {
		log.Fatalln("could not execute template:", err.Error())
	}
	SendEmail(&Email{To: []string{toEmail}, ContentType: "text/html", Subject: "Your One-Time Password (OTP) for login", Body: buf.String()})
}
