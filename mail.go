package main

import (
	"net/smtp"
	"os"
)

func sendMail(tid string, subject string, message string) error {
    auth := smtp.PlainAuth(
        "",
        "aopontan0416@gmail.com", // 送信に使うアカウント
        os.Getenv("SMTP_PASSWORD"), // アカウントのパスワード or アプリケーションパスワード
        "smtp.gmail.com",
    )

    return smtp.SendMail(
        "smtp.gmail.com:587",
        auth,
        "aopontan0416@gmail.com", // 送信元
        []string{"aopontan0416@gmail.com"}, // 送信先
        []byte(
            "To: aopontan0416@gmail.com\r\n" +
            "Subject:" + subject + "\r\n" +
            "\r\n" +
            message + "\r\n" +
            "\r\n" +
            "Tweet URL: https://twitter.com/demo/status/" + tid + "\r\n"),
    )
}
