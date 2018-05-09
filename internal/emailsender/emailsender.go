package emailsender

import (
	"errors"
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
)

type Strategy interface {
	Send(m emailprovider.Email) error
}

type RoundRobinSender struct {
	Providers []emailprovider.Provider
	lastIndex int
}

func (s *RoundRobinSender) Send(m emailprovider.Email) error {
	if len(s.Providers) == 0 {
		return errors.New("Empty list of providers. It seems impossible to send an email through a provider if no email providers are provided.")
	}
	currentIndex := s.lastIndex
	for do := true; do; do = currentIndex != s.lastIndex {
		current := s.Providers[currentIndex]
		err := current.Send(m)
		if err == nil {
			s.lastIndex = currentIndex
			return nil
		}
		currentIndex = (currentIndex + 1) % len(s.Providers)
	}
	return errors.New("All providers reported an error while attempting to send.")
}
