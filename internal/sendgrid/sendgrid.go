package sendgrid

import (
	"errors"
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log"
	"os"
)

type SendGridProvider struct{}

func (s SendGridProvider) Init() error {
	// Send Grid needs no initialization
	return nil
}

func (s SendGridProvider) Send(m emailprovider.Email) error {
	log.Printf("Sending through Send Grid: %s\n", m)
	message := mail.NewV3Mail()
	message.SetFrom(mail.NewEmail(m.From.Name(), m.From.Address()))
	message.Subject = m.Subject.String()
	if len(m.Body) > 0 {
		message.AddContent(mail.NewContent("text/plain", m.Body))
	}
	if m.HtmlBody != nil {
		message.AddContent(mail.NewContent("text/html", m.HtmlBody.String()))
	}
	p := mail.NewPersonalization()
	for _, to := range m.To {
		p.AddTos(mail.NewEmail(to.Name(), to.Address()))
	}
	for _, cc := range m.Cc {
		p.AddCCs(mail.NewEmail(cc.Name(), cc.Address()))
	}
	for _, bcc := range m.Bcc {
		p.AddBCCs(mail.NewEmail(bcc.Name(), bcc.Address()))
	}
	message.AddPersonalizations(p)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		log.Printf("Error sending through Send Grid: %s\n", err)
		return err
	}
	if response.StatusCode != 200 && response.StatusCode != 202 {
		log.Printf("Error sending through Send Grid: %d %s\n", response.StatusCode, response.Body)
		return errors.New("Could not deliver email.")
	}
	return nil
}
