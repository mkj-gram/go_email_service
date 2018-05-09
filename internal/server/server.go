package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
	"github.com/mkj-gram/go_email_service/internal/emailsender"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// These should not be constants, but put into a database somewhere.
const basicAuthenticationCode = "Basic c3RhcmxvcmQ6dWJlcmNoYWxsZW5nZQ=="
const debugAuthenticationCode = "Basic ZWdvOnViZXJjaGFsbGVuZ2U="

type ServerApp struct {
	Strategy emailsender.Strategy
	LogFile  string
}

type handler func(w http.ResponseWriter, r *http.Request)

type EmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
type Email struct {
	From    EmailAddress   `json:"from"`
	To      []EmailAddress `json:"to"`
	Cc      []EmailAddress `json:"cc"`
	Bcc     []EmailAddress `json:"bcc"`
	Subject string         `json:"subject"`
	Body    string         `json:"body"`
	Html    string         `json:"html"`
}

// parseEmails is a utility function for converting posted json emails to
// emailprovider.Email.
func parseEmails(emailStrings []EmailAddress, errors []error) []emailprovider.EmailAddress {
	emails := make([]emailprovider.EmailAddress, 0, len(emailStrings))
	for _, e := range emailStrings {
		email, err := emailprovider.MakeEmailAddress(e.Name, e.Address)
		if err == nil {
			emails = append(emails, email)
		} else {
			errors = append(errors, err)
		}
	}
	return emails
}

// joinErrors is a utility function for combining errors into a single string.
func joinErrors(errors []error) string {
	var buffer bytes.Buffer
	for _, e := range errors {
		buffer.WriteString(e.Error())
		buffer.WriteString("\n")
	}
	return fmt.Sprint(buffer)
}

// logRequestHandler is a higher-order handler for logging requests.
func logRequestHandler(subHandler handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		subHandler(w, r)
	}
}

// securityHandler is a higher-order handler for rejecting unauthorized
// requests.
func securityHandler(code string, subHandler handler) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != code {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		subHandler(w, r)
	}
}

// logHandler is the
func logHandler(a ServerApp) handler {
	return securityHandler(debugAuthenticationCode, func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("log")
		if err != nil {
			http.Error(w, "could not open log", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		file.Seek(-1000, 2)
		logData, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "could not read log", http.StatusInternalServerError)
			return
		}
		w.Write(logData)
	})
}

// sendHandler is the main handler for sending data. It is wrapping a
// securityHandler to make sure only authenticated requests are allowed to send
// emails. It then decodes the posted JSON, validates it, and calls the strategy
// for delivery.
func sendHandler(a ServerApp) handler {
	return securityHandler(basicAuthenticationCode, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "invalid request method",
				http.StatusMethodNotAllowed)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "whoops", http.StatusInternalServerError)
			return
		}
		var dto Email
		if json.Unmarshal(body, &dto) != nil {
			http.Error(w, "invalid json structure", http.StatusBadRequest)
			return
		}
		// Validate all email addresses and subjects
		errs := make([]error, 0, len(dto.To)+len(dto.Cc)+len(dto.Bcc)+3)
		from, err := emailprovider.MakeEmailAddress(dto.From.Name, dto.From.Address)
		if err != nil {
			errs = append(errs, err)
		}
		subject, err := emailprovider.MakeSubject(dto.Subject)
		if err != nil {
			errs = append(errs, err)
		}
		// If there are no errors, check if at least one to-address has been specified
		to := parseEmails(dto.To, errs)
		if len(errs)+len(to) == 0 {
			errs = append(errs, errors.New("provide at one correct recipient in the to-field"))
		}
		email := emailprovider.Email{
			From:     from,
			To:       to,
			Cc:       parseEmails(dto.Cc, errs),
			Bcc:      parseEmails(dto.Bcc, errs),
			Subject:  subject,
			Body:     dto.Body,
			HtmlBody: emailprovider.MakeHtmlBody(dto.Html),
		}
		if len(errs) > 0 {
			http.Error(w, joinErrors(errs), http.StatusBadRequest)
			return
		}
		if err := a.Strategy.Send(email); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func (a ServerApp) Serve() {
	http.HandleFunc("/send", logRequestHandler(sendHandler(a)))
	http.HandleFunc("/log", logHandler(a))
}
