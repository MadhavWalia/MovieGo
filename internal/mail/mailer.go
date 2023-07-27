package mail

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// Embedding the templates directory into the binary

//go:embed templates
var templateFS embed.FS


// Declaring a mailer struct to hold the mailer configuration
// It has a dialer for connecting to the SMTP server and a sender address
type Mailer struct {
	dialer *mail.Dialer
	sender string
}


// Declaring a factory function to create a new mailer
func New(host string, port int, username, password, sender string) Mailer {
	// Initializing a new dialer
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	// Returning a new mailer
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}


// Declaring a method to send a mail
func (m Mailer) Send(recipient, templateFile string, data any) error {
	// Parsing the template file
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}


	// Executing the "subject" template and storing the result in a buffer
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Executing the "plainBody" template and storing the result in a buffer
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// Executing the "htmlBody" template and storing the result in a buffer
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}


	// Initializing a new email message and setting the headers
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())


	// Connecting to the SMTP server and sending the email
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	// Returning nil if no error occurred
	return nil
}