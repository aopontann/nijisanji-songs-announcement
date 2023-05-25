package mail_test

import (
	"testing"

	"github.com/aopontann/nijisanji-songs-announcement/cmd/mail"
)

func TestSendMail(t *testing.T) {
	err := mail.Send("subject", "title")
	if err != nil {
		t.Errorf("sql.Open")
	}
}
