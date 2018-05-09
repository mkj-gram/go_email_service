package emailprovider

import (
	"errors"
	"net/mail"
)

// Enhanced string types to validate and enforce static checks of email arguments.
type EmailAddress interface {
	Name() string
	Address() string
}
type emailAddress struct {
	address string
	name    string
}

func (e emailAddress) Name() string {
	return e.name
}
func (e emailAddress) Address() string {
	return e.address
}

func MakeEmailAddress(name, address string) (EmailAddress, error) {
	_, err := mail.ParseAddress(address)
	if err != nil {
		return nil, err
	}
	return emailAddress{
		address: address,
		name:    name,
	}, nil
}

type Subject interface {
	String() string
}
type subject struct{ string }

func (s subject) String() string {
	return s.string
}
func MakeSubject(sub string) (Subject, error) {
	if sub == "" {
		return nil, errors.New("Subject must not be empty")
	}
	if len(sub) > 78 {
		return nil, errors.New("Subject must not be longer than 78 characters")
	}
	return subject{sub}, nil
}

type HtmlBody interface {
	String() string
}
type htmlBody struct {
	string
}

func (h htmlBody) String() string {
	return h.string
}
func MakeHtmlBody(body string) HtmlBody {
	return htmlBody{body}
}

type Email struct {
	To       []EmailAddress
	Cc       []EmailAddress
	Bcc      []EmailAddress
	From     EmailAddress
	Subject  Subject
	Body     string
	HtmlBody HtmlBody
}

type Provider interface {
	Send(m Email) error
	Init() error
}
