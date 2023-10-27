package nsa

import (
	"net/smtp"
	"os"

	"github.com/rs/zerolog/log"
)

func SendMail(subject string, message string) error {
	auth := smtp.PlainAuth(
		"",
		"aopontan0416@gmail.com",   // 送信に使うアカウント
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
		"aopontan0416@gmail.com",           // 送信元
		[]string{"aopontan0416@gmail.com"}, // 送信先
		[]byte(
			"To: aopontan0416@gmail.com\r\n"+
				"Subject:"+subject+"\r\n"+
				"\r\n"+
				message+"\r\n"),
	)
}
