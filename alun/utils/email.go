package utils

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Al-un/alun-api/pkg/communication"
	"github.com/joho/godotenv"
)

const (
	defaultSender = "Al-un.fr <no-reply@al-un.fr>"
)

var (
	templateFolder string
	sender         string
	alunAccount    communication.EmailConfiguration
)

// Email utilities assume that the program is executed at the root of the
// project.
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error when loading .env: ", err)
	}

	// Email account configuration
	accountUser := os.Getenv(EnvVarEmailUsername)
	accountPassword := os.Getenv(EnvVarEmailPassword)
	accountServer := os.Getenv(EnvVarEmailHost)
	accountPortText := os.Getenv(EnvVarEmailPort)
	accountPort, err := strconv.Atoi(accountPortText)
	if err != nil {
		utilsLogger.Fatal(2, "Error when parsing EmailServerPort <%s>: %v",
			accountPortText, err)
	}
	alunAccount = communication.EmailConfiguration{
		Username: accountUser,
		Password: accountPassword,
		Host:     accountServer,
		Port:     accountPort,
	}

	// Extra email configuration
	sender = os.Getenv(EnvVarEmailSender)
	if sender == "" {
		sender = defaultSender
	}

	// Template configuration
	cwd, _ := os.Getwd()
	templateFolder = filepath.Join(cwd, "alun/utils/email_templates/")
}

// SendNoReplyEmail sends an email from a no-reply account.
//
// The returned error is only for telling the calling method that something went
// wrong. Parent method is not expected to tell the error content to the client
// and error handling must be done by checking the logs.
//
// If templateName does not end up with `.html`, it is automatically appended
func SendNoReplyEmail(to []string, subject string, templateName string, emailData interface{}) error {
	// Automatic appending of `.html`
	if templateName[len(templateName)-5:] != ".html" {
		templateName = templateName + ".html"
	}

	// Build HTML email
	email, err := communication.NewEmailHTMLMessage(
		sender,
		to,
		subject,
		filepath.Join(templateFolder, templateName),
		emailData,
	)
	if err != nil {
		utilsLogger.Info("Error when loading template %s: %v", templateName, err)
		return err
	}

	// Send
	err = alunAccount.Send(email)
	if err != nil {
		utilsLogger.Info("Error when sending email of template %s to %v: %v", templateName, to, err)
		return err
	}

	// All good
	return nil
}
