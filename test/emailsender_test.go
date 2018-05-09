package test

import (
	"errors"
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
	"github.com/mkj-gram/go_email_service/internal/emailsender"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestProvider struct {
	send func(m emailprovider.Email) error
}

func (t TestProvider) Send(m emailprovider.Email) error {
	return t.send(m)
}

type SuccessProvider struct{}

func (s SuccessProvider) Send(m emailprovider.Email) error {
	return nil
}

type FailProvider struct{}

func (f FailProvider) Send(m emailprovider.Email) error {
	return errors.New("Some error here")
}

func testProviderGenerator(index *int, err error) TestProvider {
	return TestProvider{send: func(m emailprovider.Email) error {
		*index += 1
		return err
	}}
}

func makeSimpleEmail() emailprovider.Email {
	from, _ := emailprovider.MakeEmailAddress("Morten", "morten@example.com")
	to, _ := emailprovider.MakeEmailAddress("Morten", "morten@example.com")
	subject, _ := emailprovider.MakeSubject("this is a subject")
	return emailprovider.Email{
		From:    from,
		To:      []emailprovider.EmailAddress{to},
		Subject: subject,
		Body:    "this is a body"}
}

func TestSendChecksForProviders(t *testing.T) {
	sender := emailsender.RoundRobinSender{
		Providers: []emailprovider.Provider{},
	}
	err := sender.Send(emailprovider.Email{})
	assert.NotNil(t, err)
}

func TestWillCallFirstProvider(t *testing.T) {
	called := 0
	sender := emailsender.RoundRobinSender{
		Providers: []emailprovider.Provider{
			testProviderGenerator(&called, nil),
		},
	}
	err := sender.Send(makeSimpleEmail())
	assert.Equal(t, 1, called)
	assert.Nil(t, err)
}

func TestWillNotCallAfterSuccessProvider(t *testing.T) {
	called := 0
	sender := emailsender.RoundRobinSender{
		Providers: []emailprovider.Provider{
			SuccessProvider{},
			testProviderGenerator(&called, nil),
		},
	}
	sender.Send(makeSimpleEmail())
	assert.Equal(t, 0, called, "Called second provider after success")
}

func TestWillCallAfterFailProvider(t *testing.T) {
	called := 0
	providers := []emailprovider.Provider{
		FailProvider{},
		testProviderGenerator(&called, nil),
	}
	sender := emailsender.RoundRobinSender{Providers: providers}
	sender.Send(makeSimpleEmail())
	assert.Equal(t, 1, called, "Not called second provider after fail")
}

func TestWillNotLoopOnFailingProviders(t *testing.T) {
	providers := []emailprovider.Provider{
		FailProvider{},
		FailProvider{},
		FailProvider{},
	}
	sender := emailsender.RoundRobinSender{Providers: providers}
	// This will keep looping forever, if not implemented correctly
	sender.Send(makeSimpleEmail())
}

func TestWillContinueWithLastSuccess(t *testing.T) {
	counters := []int{0, 0, 0}
	providers := []emailprovider.Provider{
		testProviderGenerator(&counters[0], errors.New("This will fail")),
		testProviderGenerator(&counters[1], nil),
		testProviderGenerator(&counters[2], nil),
	}
	sender := emailsender.RoundRobinSender{Providers: providers}
	sender.Send(makeSimpleEmail())
	sender.Send(makeSimpleEmail())
	assert.Equal(t, 1, counters[0], "Calling first provider again")
	assert.Equal(t, 2, counters[1], "Not starting with last successful provider")
	assert.Equal(t, 0, counters[2], "Not starting with last successful provider")
}
