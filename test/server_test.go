package test

import (
	"errors"
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
	"github.com/mkj-gram/go_email_service/internal/server"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestStrategy struct {
	sendHandler func(m emailprovider.Email) error
}

func (t TestStrategy) Send(m emailprovider.Email) error {
	return t.sendHandler(m)
}

var testStrategy = new(TestStrategy)

func TestMain(m *testing.M) {
	app := server.ServerApp{testStrategy}
	app.Serve()
	m.Run()
}

func makeAuthorizedRequest(t *testing.T, method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Authorization", server.BasicAuthenticationCode)
	return req
}

func TestSendRequireAuth(t *testing.T) {
	req, err := http.NewRequest("POST", "/send", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

func TestInvalidPathRequest(t *testing.T) {
	req := makeAuthorizedRequest(t, "POST", "/thisisnotapath", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
}

func TestSendInvalidRequestMethod(t *testing.T) {
	req := makeAuthorizedRequest(t, "GET", "/send", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Result().StatusCode)
}

func TestSendInvalidJsonStructure(t *testing.T) {
	req := makeAuthorizedRequest(t, "POST", "/send", strings.NewReader("{from: test,"))
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestSendInvalidJsonData(t *testing.T) {
	req := makeAuthorizedRequest(t, "POST", "/send", strings.NewReader(
		`{
from: test@test.dk,
to: test@test.dk,
subject: 'hello',
body: 'this works'
}`))
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestSendNoneValidatingFromJsonData(t *testing.T) {
	req := makeAuthorizedRequest(t, "POST", "/send", strings.NewReader(
		`{
"from": "test@test@test",
"to": ["test.anden.dk"],
"subject": "hello",
"body": "this works"
}`))
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestSendReportErrorFromStrategy(t *testing.T) {
	testStrategy.sendHandler = func(m emailprovider.Email) error {
		return errors.New("An error occurred")
	}
	req := makeAuthorizedRequest(t, "POST", "/send", strings.NewReader(
		`{
"from": {"name": "anders", "address": "test@test.com"},
"to": [{"name": "thomas", "address": "test@test.dk"}],
"subject": "hello",
"body": "this works"
}`))
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusServiceUnavailable, rr.Result().StatusCode)
}

func TestSendOk(t *testing.T) {
	testStrategy.sendHandler = func(m emailprovider.Email) error {
		assert.Equal(t, "toemail1@example.com", m.To[0].Address())
		assert.Equal(t, "ccemail2@example.com", m.Cc[1].Address())
		assert.Equal(t, "bccemail3@example.com", m.Bcc[2].Address())
		return nil
	}
	req := makeAuthorizedRequest(t, "POST", "/send", strings.NewReader(
		`{
"from": {"name": "anders", "address": "test@test.com"},
"to": [{"name": "morten", "address": "toemail1@example.com"}, {"name": "peter", "address": "toemail2@example.com"}],
"cc": [{"name": "morten1", "address": "ccemail1@example.com"}, {"name": "peter1", "address": "ccemail2@example.com"}, {"name": "Thomas", "address": "ccemail3@example.com"}],
"bcc": [{"name": "morten2", "address": "bccemail1@example.com"}, {"name": "peter2", "address": "bccemail2@example.com"}, {"name": "Thomas1", "address": "bccemail3@example.com"}],
"subject": "hello",
"body": "this works"
}`))
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
}
