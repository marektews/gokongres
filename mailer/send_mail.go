package mailer

import (
	"os"
)

func SendHtmlMail(to []string, subject string, body string) {
	env := os.Getenv("ENV")
	// log.Println("SendHtmlMail: ENV: ", env)
	if env == "production" {
		Send_By_Postfix(to, subject, body)
	} else {
		Send_By_SMTPMailer(to, subject, body)
	}
}
