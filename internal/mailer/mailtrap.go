package mailer

import (
	"bytes"
	"errors"
	"text/template"

	gomail "gopkg.in/mail.v2"
)

type mailtrapClient struct {
	fromEmail string
	apiKey     string
}

func NewMailtrapClient(apiKey, fromEmail string) (mailtrapClient, error) {
	if apiKey == "" {
		return mailtrapClient{}, errors.New("mailtrap api key is required")
	}

	return mailtrapClient{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}, nil
}

func (m mailtrapClient) Send (templateFile, username, email string, data any, isSandbox bool) (int, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}

	message := gomail.NewMessage()
	message.SetHeader("To", email)
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("Subject", subject.String())
	message.SetBody("text/html", body.String())

	dailer := gomail.NewDialer("smtp.gmail.com", 587,"md.hasibuzzaman001@gmail.com", "ftcr rwcc uxbw woil")
	if err := dailer.DialAndSend(message); err != nil {
		return -1, err
	}

	return 200, nil
}