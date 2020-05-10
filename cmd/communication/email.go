package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Al-un/alun-api/alun/utils"
	"github.com/Al-un/alun-api/pkg/communication"
	"github.com/joho/godotenv"
)

// ----------------------------------------------------------------------------
var (
	alunAccount communication.EmailConfiguration
	alunEmail   utils.AlunEmailSender
	sender      string
	recipients  []string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error when loading .env: ", err)
	}

	// Account config
	accountUser := os.Getenv(utils.EnvVarEmailUsername)
	accountPassword := os.Getenv(utils.EnvVarEmailPassword)
	accountServer := os.Getenv(utils.EnvVarEmailHost)
	accountPortText := os.Getenv(utils.EnvVarEmailPort)
	accountPort, err := strconv.Atoi(accountPortText)
	if err != nil {
		log.Fatalf("Error when parsing Email PORT <%s>: %v\n ", accountPortText, err)
	}
	// log.Printf("Loading configuration %s / %s / %s / %s\n",
	// 	accountUser, accountPassword, accountServer, accountPortText)

	alunAccount = communication.EmailConfiguration{
		Username: accountUser,
		Password: accountPassword,
		Host:     accountServer,
		Port:     accountPort,
	}

	alunEmail = utils.GetAlunEmail()

	// Misc
	sender = "Al-un.fr <no-reply@al-un.fr>"
	recipients = []string{"alun.sng@gmail.com"}
}

// ----------------------------------------------------------------------------
// sendTextEmail checks plain text email
func sendTextEmail() {
	err := alunAccount.Send(communication.NewEmailTextMessage(
		sender,
		recipients,
		"Testing some plain and inline text email",
		"Testing some text message\r\n"+
			"\r\n"+
			"With an empty line",
	))

	if err != nil {
		log.Fatal(err)
	}
	log.Println("TextEmail sending Finished!")
}

// sendHTMLEmail checks HTML email from template in the current folder
func sendHTMLEmail() {

	templateData := struct {
		Name string
		URL  string
	}{
		Name: "Pouet pouet",
		URL:  "https://duckduckgo.com/?q=it+works!",
	}

	cwd, _ := os.Getwd()
	email, err := communication.NewEmailHTMLMessage(
		sender,
		recipients,
		"Testing HTML email",
		filepath.Join(cwd, "/cmd/communication/test.html"),
		templateData,
	)
	if err != nil {
		log.Fatal("Generate email error: ", err)
	}
	err = alunAccount.Send(email)
	if err != nil {
		log.Fatal("Send email error: ", err)
	}

	log.Println("HTML email sending Finished!")

}

func sendNoReplyEmail() {
	alunEmail.SendNoReplyEmail(
		recipients,
		"Testing no-reply email",
		"user_registration",
		struct{ URL string }{URL: "https://youtube.com"},
	)
}
