package mailer

import (
	"log"
	"net/smtp"
	"os"
	"strings"
)

func Send_By_Postfix(to []string, subject string, body string) {
	client, err := smtp.Dial("localhost:25")
	if err != nil {
		log.Println("Dial to postfix error:", err)
		return
	}
	defer client.Close()

	smtpFrom := os.Getenv("SMTP_FROM")
	err = client.Mail(smtpFrom)
	if err != nil {
		log.Println("Postfix client mail from error:", err)
		return
	}

	// dodaj każdego odbiorcę
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			log.Println("Postfix client recp to error:", err)
			return
		}
	}

	w, err := client.Data()
	if err != nil {
		log.Println("Postfix client data error:", err)
		return
	}
	defer w.Close()

	hdr := []byte(
		"From: " + smtpFrom + "\r\n" +
			"To: " + strings.Join(to, ", ") + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
			"\r\n",
	)

	msg := append(hdr, []byte(body+"\r\n")...)

	_, err = w.Write(msg)
	if err != nil {
		log.Printf("Postfix sender error: %v", err)
		return
	}
}
