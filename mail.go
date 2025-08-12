package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type EmailAddress struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type emailMessage struct {
	Sender      EmailAddress   `json:"sender"`
	To          []EmailAddress `json:"to"`
	Subject     string         `json:"subject"`
	HtmlContent string         `json:"htmlContent"`
}

func sendEmailMessage(recipients []string, subject string, body string) error {
	if flag.Lookup("test.v") != nil {
		log.Println("Skip send email message in test.")
		return nil
	}

	emailMessageItem := emailMessage{
		Sender: EmailAddress{
			Name:  os.Getenv("APP_NAME"),
			Email: os.Getenv("APP_SENDER_EMAIL"),
		},
		Subject:     subject,
		HtmlContent: body,
	}
	for _, email := range recipients {
		emailMessageItem.To = append(emailMessageItem.To, EmailAddress{Email: email})
	}

	url := "https://api.sendinblue.com/v3/smtp/email"
	fmt.Println("URL:>", url)

	jsonBody, err := json.Marshal(emailMessageItem)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("api-key", os.Getenv("SENDINBLUE_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	return nil
}
