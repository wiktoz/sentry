package scripts

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"

	"github.com/wiktoz/sentry/db"
)

func SendEmail(subject, body string) error {
	cfg, err := db.GetConfig(db.DB)
	if err != nil {
		return err
	}

	to := cfg.Email
	from := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASS")

	// Serwer SMTP WP.pl
	smtpHost := "smtp.wp.pl"
	smtpPort := 465

	// Komponowanie wiadomości
	msg := "From: Scan Report <" + from + ">\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body

	// Połączenie TLS
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", smtpHost, smtpPort), &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	})
	if err != nil {
		return fmt.Errorf("TLS connection error: %v", err)
	}

	// Nowa sesja SMTP z użyciem TLS
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("SMTP client error: %v", err)
	}
	defer client.Quit()

	// Autoryzacja SMTP
	auth := smtp.PlainAuth("", from, password, smtpHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth error: %v", err)
	}

	// Wysyłka maila
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM error: %v", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO error: %v", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA error: %v", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	writer.Close()

	return nil
}
