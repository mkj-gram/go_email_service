package sparkpost

import (
	"errors"
	sp "github.com/SparkPost/gosparkpost"
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
	"log"
	"os"
	"strings"
)

type SparkPostProvider struct{}

var client *sp.Client

func (s SparkPostProvider) Init() error {
	cfg := &sp.Config{
		BaseUrl:    "https://api.sparkpost.com",
		ApiKey:     os.Getenv("SPARKPOST_API_KEY"),
		ApiVersion: 1,
	}
	var c sp.Client
	err := c.Init(cfg)
	if err == nil {
		client = &c
	} else {
		log.Println("Could not initialize Spark Post provider.")
	}
	return err
}

func (s SparkPostProvider) Send(m emailprovider.Email) error {
	if client == nil {
		return errors.New("SparkPost provider not initialized correctly")
	}
	log.Printf("Sending through Spark Post: %s\n", m)
	content := sp.Content{
		From:    sp.Address{Name: m.From.Name(), Email: m.From.Address()},
		Subject: m.Subject.String(),
		Text:    m.Body,
		HTML:    m.HtmlBody.String(),
	}
	headerTo := make([]string, 0, len(m.To))
	for _, e := range m.To {
		headerTo = append(headerTo, e.Address())
	}
	headerToValue := strings.Join(headerTo, ",")
	tx := &sp.Transmission{
		Content:    content,
		Recipients: []sp.Recipient{},
	}
	for _, e := range m.To {
		tx.Recipients = append(tx.Recipients.([]sp.Recipient), sp.Recipient{
			Address: sp.Address{Name: e.Name(), Email: e.Address(), HeaderTo: headerToValue},
		})
	}
	if len(m.Cc) > 0 {
		ccTo := make([]string, 0, len(m.Cc))
		for _, e := range m.Cc {
			tx.Recipients = append(tx.Recipients.([]sp.Recipient), sp.Recipient{
				Address: sp.Address{Name: e.Name(), Email: e.Address(), HeaderTo: headerToValue},
			})
			ccTo = append(ccTo, e.Address())
		}
		content.Headers["cc"] = strings.Join(ccTo, ",")
	}
	for _, e := range m.Bcc {
		tx.Recipients = append(tx.Recipients.([]sp.Recipient), sp.Recipient{
			Address: sp.Address{Name: e.Name(), Email: e.Address(), HeaderTo: headerToValue},
		})
	}
	_, response, err := client.Send(tx)
	if err != nil {
		log.Printf("Error sending through Spark Post: %d %s %s\n", response.HTTP.StatusCode, string(response.Body), response.Errors)
	}
	return err
}
