package test

import (
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestSubjectNotEmpty(t *testing.T) {
	_, err := emailprovider.MakeSubject("")
	assert.NotNil(t, err)
}

func TestSubjectExceeding(t *testing.T) {
	_, err := emailprovider.MakeSubject(strings.Repeat("h", 80))
	assert.NotNil(t, err)
}

func TestCanMakeSubject(t *testing.T) {
	subject, err := emailprovider.MakeSubject("this is a fine subject")
	assert.NotNil(t, subject)
	assert.Nil(t, err)
}

func TestCorrectlyFormattedEmail(t *testing.T) {
	_, err := emailprovider.MakeEmailAddress("invalid", "invalid")
	assert.NotNil(t, err)
	withName, err := emailprovider.MakeEmailAddress("Morten StarLord", "starlord@example.com")
	assert.NotNil(t, withName)
	assert.Nil(t, err)
	withoutName, err := emailprovider.MakeEmailAddress("", "starlord@example.com")
	assert.NotNil(t, withoutName)
	assert.Nil(t, err)
}
