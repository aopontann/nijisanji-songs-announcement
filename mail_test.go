package main

import (
	"testing"
)

func TestSendMail(t *testing.T) {
	err := SendMail("subject", "title")
	if err != nil {
		t.Errorf("sql.Open")
	}
}
