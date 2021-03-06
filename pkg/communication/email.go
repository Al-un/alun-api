package communication

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

/**
Golang documentation		: https://golang.org/pkg/net/smtp/#SendMail
							  https://golang.org/pkg/html/template/#Template.Execute
RFC-822						: https://tools.ietf.org/html/rfc822
Tutorial					: https://blog.mailtrap.io/golang-send-email/
MimeType					: https://stackoverflow.com/a/9951508/4906586
OVH documentation			: https://docs.ovh.com/fr/emails/guide-configuration-mail-de-mac-mavericks-et-yosemite/#rappel-des-parametres-pop-imap_1
							  https://docs.ovh.com/fr/domains/mail-mutualise-guide-de-configuration-mx-avec-zone-dns-ovh/
HTML email with template	: https://medium.com/@dhanushgopinath/sending-html-emails-using-templates-in-golang-9e953ca32f3d
Template relative path		: https://stackoverflow.com/a/20417010/4906586
Some example				: https://github.com/tangingw/go_smtp/blob/master/send_mail.go

Actions required:
- Update MX records in Cloudflare: domain is registered with OVH but DNS servers are Cloudflare
- Adding "From" in message headers to get accepted by Gmail
**/

// EmailSender qualifies an entity which is capable to send an `EmailMessage`
//
// A given email sender does not have to cover all email format (HTML, text)
type EmailSender interface {
	Send(emailMsg EmailMessage) error
}

// EmailConfiguration configures an email sender account.
//
// There is no default email configuration so email credentials are required.
type EmailConfiguration struct {
	Username string
	Password string
	Host     string
	Port     int
}

// EmailMessage harmonises email message format which will be sent by an EmailConfiguration
//
// Jan-2020: Multipart emails are not supported
type EmailMessage struct {
	// email sender is not necessarily the same account used in the email configuration, e.g.: no-reply account
	From    string
	To      []string
	Subject string
	Body    string
	// text/plain or text/html
	ContentType string
	// Additional headers out of "From", "To", "Subject"
	AdditionalHeaders map[string]string
}

// Send sends the provided message given an email configuration
func (emailCfg EmailConfiguration) Send(emailMsg EmailMessage) error {

	// Build email to send
	msgHeaders := make(map[string]string)
	for headerKey, headerValue := range emailMsg.AdditionalHeaders {
		msgHeaders[headerKey] = headerValue
	}
	// Overrides reserved fields if defined in emailMsg.AdditionalHeaders
	msgHeaders["From"] = emailMsg.From
	msgHeaders["To"] = strings.Join(emailMsg.To, ",")
	msgHeaders["Subject"] = emailMsg.Subject
	msgHeaders["MIME-Version"] = "1.0" // As non-ASCII might be expected: https://stackoverflow.com/a/3569363/4906586
	msgHeaders["Content-Type"] = emailMsg.ContentType

	// Build body message with RFC-822 headers
	msg := ""
	for headerKey, headerValue := range msgHeaders {
		msg += fmt.Sprintf("%s: %s\r\n", headerKey, headerValue)
	}
	msg += "\r\n" + emailMsg.Body

	// Configure auth and send
	emailAuth := smtp.PlainAuth("", emailCfg.Username, emailCfg.Password, emailCfg.Host)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", emailCfg.Host, emailCfg.Port),
		emailAuth,
		emailMsg.From,
		emailMsg.To,
		[]byte(msg),
	)

	return err
}

// NewEmailTextMessage generate a basic text EmailMessage
func NewEmailTextMessage(from string, to []string, subject string, message string) EmailMessage {
	return EmailMessage{
		From:        from,
		To:          to,
		Subject:     subject,
		ContentType: "text/plain",
		Body:        message,
	}
}

// NewEmailHTMLMessage generate a HTML EmailMessage from a template path
func NewEmailHTMLMessage(from string, to []string, subject string, templatePath string, emailData interface{}) (EmailMessage, error) {
	// Fetch template
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return EmailMessage{}, err
	}

	// Fill the template with provided data, if applicable
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, emailData); err != nil {
		return EmailMessage{}, err
	}

	return EmailMessage{
		From:        from,
		To:          to,
		Subject:     subject,
		ContentType: "text/html",
		Body:        buf.String(),
	}, nil
}
