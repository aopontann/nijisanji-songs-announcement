package nsa

import (
	"net/smtp"
	"os"

	"github.com/rs/zerolog/log"
)

func SendMail(subject string, message string) error {
	addr := os.Getenv("MAIL_ADDRESS")
	auth := smtp.PlainAuth(
		"",
		addr,   // 送信に使うアカウント
		os.Getenv("SMTP_PASSWORD"), // アカウントのパスワード or アプリケーションパスワード
		"smtp.gmail.com",
	)

	log.Info().
		Str("severity", "INFO").
		Str("service", "send-mail").
		Str("subject", subject).
		Str("message", message).
		Send()

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		addr,           // 送信元
		[]string{addr}, // 送信先
		[]byte(
			"To: "+addr+"\r\n"+
				"Subject:"+subject+"\r\n"+
				"\r\n"+
				message+"\r\n"),
	)
}
