package nsa

import (
	"testing"
)

func TestSendMail(t *testing.T) {
	err := SendMail("subject", "title")
	if err != nil {
		t.Errorf("SendMail failed with error: %v", err)
	}
}
