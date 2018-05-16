# Email Service written in Go

Basic Email Service for sending emails. This service provides an abstraction
over multiple providers in a fault-tolerant way.

## Design Choices

1. Using Go and dependencies 

   Go is a quite popular language and thought it would be fun to try and learn.
   I therefore tried not to use too many packages, such as mux, to make sure I
   understood the internals. The [Dependencies](#dependencies) are used for
   testing and contacting the email providers. All local packages are put under
   internal, so the project layout may be against best practices, but that is
   what happens on the first project in a new language.
 
2. There is no database

   I do know how to work with databases, but I wanted to keep things simple and
   contained for this project. The authentication is therefore hard-coded into
   the source and the logging is file-based.
   
3. Unit-testing, code-coverage and static types

   I was not aiming for 100 % code-coverage. I am using interfaces and private
   structs to hide the implementation details of email addresses, subjects and
   html bodies. This allow the type-checker to ensure safety, so I do not
   have to test the construction of such.
   
   For most other cases I use unit-testing to ensure the well-behavedness of the
   sender strategy and endpoints. The code for sending to a specific provider is
   not unit-tested, but tested manually.
      
4. Logging and error-messages

   For most validation errors, the error messages are directly relayed to the
   client. However, for internal server errors and provider errors, the error
   messages are hidden and logged instead.

## Send Strategy

The Send strategy used in this project is parameterized by a list providers. The
strategy requires at least one provider, and works as follows. Say we have three
providers and provider 2 is the current one:

* Provider 1
* Provider 2 <-- current <-- last
* Provider 3

Upon a send-attempt, if Provider 2 is successful, we do nothing. Otherwise, we
move the current pointer to the next provider, modulo the length of the list,
while keeping track of the last working provider.

If any provider succeeds, we save the current one as new last. Else if `current
== last`, all providers failed and we report an error to the user.

There is no cost associated with a provider at this point, thus there is no wish
to have a primary or secondary provider, but if we wanted to introduce one, we
could easily change the strategy to _start over_ after a period of time, say 5
minutes.

In this project I chose to use SendGrid and SparkPost as the two email
providers, because both of them provided a go-package for communication with
their api. The packages are only used for convenience, communication could have
been done _manually_ by HTTP-request.

## Api

The api can be found at http://fast-savannah-21734.herokuapp.com and
exposes two endpoints:

#### POST: /send 

/send accepts a POST request and tries to send the email described by the posted
data, if it matches the following json-format and validates:

```go
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
```

The json is parsed and validated. Particularly, the are emails validated by
parsing it through Go's net/mail.ParseAddress, which to my understanding ensures
the emails are valid as specified by RFC 5322 and extended by RFC 6532.

If an error is encountered, the error message will be in the response body,
along with a suitable status code. In the case of no errors, and the email was
properly dispatched to a provider, status code 200 is returned.

The endpoint requires using a specific user, which has to be sent as a Basic
Authorization Header. In a more realistic example, the user would be validated
against a database, to obtain a role/right to send emails.

#### GET: /log

To see what's going on, the api provide the /log endpoint, to see the last 1000
characters of the log. Since logging also logs emails and ip-addresses, the
access to /log has its own user. In a real world example, this would be a
developer-account or the log would be streamed to somewhere else.


## Examples

```bash
curl -H "Authorization: Basic c3RhcmxvcmQ6dWJlcmNoYWxsZW5nZQ=="  -H "Content-Type: application/json" -d '{"from": { "name": "Anders Andersen", "address": "morten@example.com"},"to": [{"name": "Morten", "address": "morten@example.com"},{"name": "Info", "address": "info@example.com"}],"subject": "This is a test","body": "This is the plain text body","html": "This is the html <em>body</em>" }' http://fast-savannah-21734.herokuapp.com/send 
```

```bash
curl -H "Authorization: Basic ZWdvOnViZXJjaGFsbGVuZ2U=" http://fast-savannah-21734.herokuapp.com/log
```

**Note** for the sender, you have to specify a @dotnamics.com email, since the providers required a registered domain for sending. 

## Future Work

1. Adding a database to have multiple users

   The motivation here is pretty obvious.

2. Checking if messages are actually send and received

   Both email providers, in the case of a successful post, accepts the messages
   for further delivery. The are not put through an SMTP-server yet, so the user
   is actually not guaranteed that the emails will be send. Further, it might be
   nice to track if any of the emails bounced.

## Dependencies

The following dependencies are used in the project.

[http://github.com/stretchr/testify](http://github.com/stretchr/testify)

[http://github.com/sendgrid/sendgrid-go](http://github.com/sendgrid/sendgrid-go)

[http://github.com/SparkPost/gosparkpost](http://github.com/SparkPost/gosparkpost)

