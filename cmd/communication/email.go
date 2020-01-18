package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Al-un/alun-api/pkg/communication"
)

func sendEmail() {

	recipient := "alun.sng@gmail.com"
	accountUser := "plop@al-un.fr"
	accountPassword := "hM2s02JRX2nIZEUpFGa8"
	accountServer := "ssl0.ovh.net"
	accountPort := 587

	alunAccount := communication.EmailConfiguration{
		Username: accountUser,
		Password: accountPassword,
		Host:     accountServer,
		Port:     accountPort,
	}

	subject := "Testing HTML text mesage"

	templateData := struct {
		Name string
		URL  string
	}{
		Name: "Pouet pouet",
		URL:  "https://duckduckgo.com/?q=it+works!",
	}

	cwd, _ := os.Getwd()
	email, err := communication.NewEmailHTMLMessage(
		"no-reply@al-un.fr",
		[]string{recipient},
		subject,
		// "./test.html",
		filepath.Join(cwd, "/cmd/communication/test.html"),
		templateData,
	)
	if err != nil {
		log.Fatal("Generate email: ", err)
	}
	err = alunAccount.Send(email)
	if err != nil {
		log.Fatal("Send email: ", err)
	}

	// err := alunAccount.Send(communication.NewEmailTextMessage(
	// 	"no-reply@al-un.fr",
	// 	[]string{recipient},
	// 	subject,
	// 	"Testing some text message\r\n"+
	// 		"\r\n"+
	// 		"With an empty line",
	// ))

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Finished!")

}
