package main_test

import (
	"testing"

	"github.com/aopontann/nijisanji-songs-announcement"
)

func TestSendMail(t *testing.T) {
	err := main.SendMail("subject", "title")
	if err != nil {
		t.Errorf("sql.Open")
	}
}
