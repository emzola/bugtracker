package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

// Mailer contains a mail.Dialer instance and sender information.
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// New creates a new Mailer.
func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send sends an email. It accepts a recipient, tempate file and data.
func (m Mailer) Send(recipient, templateFile string, data any) error {
	// Parse template from embedded file system.
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}
	// Execute the named template "plainBody", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}
	// Execute the named template "htmlBody", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}
	// Initialize a new mail.Message instance, then set header, body and alternative parts
	// to the message.
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())
	// Try sending the email up to three times before aborting and returning the final
	// error. Sleep for 5 seconds between each attempt.
	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}
