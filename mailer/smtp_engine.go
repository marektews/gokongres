package mailer

import (
	"crypto/tls"
	"log"
	"net/smtp"
	"os"
	"strings"
)

func Send_By_SMTPMailer(to []string, subject string, body string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpFrom := os.Getenv("SMTP_FROM")
	smtpPass := os.Getenv("SMTP_PASS")

	smtpServerName := os.Getenv("SMTP_SERVERNAME")
	if smtpServerName == "" {
		smtpServerName = smtpHost
	}

	server := smtpHost + ":" + smtpPort

	tlsConfig := &tls.Config{ServerName: smtpServerName}

	conn, err := tls.Dial("tcp", server, tlsConfig)
	if err != nil {
		log.Printf("TLS dial error: %v", err)
		return
	}

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		log.Printf("TLS new client error: %v", err)
		return
	}
	defer func() {
		if err := client.Quit(); err != nil {
			client.Close() // wymuszony fallback
		}
	}()

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err = client.Auth(auth)
	if err != nil {
		log.Printf("SMTP auth error: %v", err)
		return
	}

	err = client.Mail(smtpFrom)
	if err != nil {
		log.Printf("Init mail from error: %v", err)
		return
	}

	for _, addr := range to {
		err = client.Rcpt(addr)
		if err != nil {
			log.Printf("Init mail rcpt error: %v", err)
			return
		}
	}

	w, err := client.Data()
	if err != nil {
		log.Printf("Init SMTP data error: %v", err)
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
		log.Printf("SMTP sender error: %v", err)
	}
}
