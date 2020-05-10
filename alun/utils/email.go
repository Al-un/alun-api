package utils

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/Al-un/alun-api/pkg/communication"
	"github.com/joho/godotenv"
)

// AlunEmailSender is a convenient interface to send an email from a specific
// no-reply Alun email
type AlunEmailSender interface {
	SendNoReplyEmail(to []string, subject string, templateName string, emailData interface{}) error
}

// AlunEmail is the default production implementation of AlunEmailSender
type AlunEmail struct {
	Account        communication.EmailConfiguration
	Sender         string
	TemplateFolder string
}

// DummyEmail prevents from sending real email and does nothing
type DummyEmail struct {
}

const (
	defaultSender = "Al-un.fr <no-reply@al-un.fr>"
	// EmailTemplateUserRegistration when sending email for new user
	EmailTemplateUserRegistration = "user_registration"
	// EmailTemplateUserPwdReset when user is requesting a password reset
	EmailTemplateUserPwdReset = "user_pwd-reset"
)

var (
	alunEmail *AlunEmail
)

// SendNoReplyEmail sends an email from a no-reply account.
//
// The returned error is only for telling the calling method that something went
// wrong. Parent method is not expected to tell the error content to the client
// and error handling must be done by checking the logs.
//
// If templateName does not end up with `.html`, it is automatically appended
func (ae AlunEmail) SendNoReplyEmail(to []string, subject string, templateName string, emailData interface{}) error {
	// Automatic appending of `.html`
	if templateName[len(templateName)-5:] != ".html" {
		templateName = templateName + ".html"
	}

	// Build HTML email
	email, err := communication.NewEmailHTMLMessage(
		ae.Sender,
		to,
		subject,
		filepath.Join(ae.TemplateFolder, templateName),
		emailData,
	)
	if err != nil {
		utilsLogger.Info("Error when loading template %s: %v", templateName, err)
		return err
	}

	// Send
	err = ae.Account.Send(email)
	if err != nil {
		utilsLogger.Info("Error when sending email of template %s to %v: %v", templateName, to, err)
		return err
	}

	// All good
	return nil
}

// SendNoReplyEmail does nothing
func (de DummyEmail) SendNoReplyEmail(to []string, subject string, templateName string, emailData interface{}) error {
	// Do nothing
	return nil
}

// GetAlunEmail loads the AlunEmail singleton
func GetAlunEmail() *AlunEmail {
	err := godotenv.Load()
	if err != nil {
		// log.Fatal("Error when loading .env: ", err)
	}

	// Email account configuration
	if alunEmail == nil {
		accountUser := os.Getenv(EnvVarEmailUsername)
		accountPassword := os.Getenv(EnvVarEmailPassword)
		accountServer := os.Getenv(EnvVarEmailHost)
		accountPortText := os.Getenv(EnvVarEmailPort)
		accountPort, err := strconv.Atoi(accountPortText)
		if err != nil {
			utilsLogger.Fatal(2, "Error when parsing EmailServerPort <%s>: %v",
				accountPortText, err.Error())
		}

		// Extra email configuration
		sender := os.Getenv(EnvVarEmailSender)
		if sender == "" {
			sender = defaultSender
		}

		// Template configuration
		cwd, _ := os.Getwd()
		templateFolder := filepath.Join(cwd, "alun/utils/email_templates/")

		alunEmail = &AlunEmail{
			Account: communication.EmailConfiguration{
				Username: accountUser,
				Password: accountPassword,
				Host:     accountServer,
				Port:     accountPort,
			},
			Sender:         sender,
			TemplateFolder: templateFolder,
		}
	}

	return alunEmail
}

// GetDummyEmail generates a DummyEmail and keep the GetXXX singleton syntax
// to align with LoadAlunEmail
func GetDummyEmail() *DummyEmail {
	return &DummyEmail{}
}
