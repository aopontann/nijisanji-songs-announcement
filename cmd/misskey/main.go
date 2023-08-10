package misskey

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const url = "https://@aopontan@misskey.io/api/notes/create"

type Misskey struct {
	token string
}

type ReqBody struct {
	I      string `json:"i"`
	Text   string `json:"text"`
	Detail bool   `json:"detail"`
}

func New(token string) *Misskey {
	return &Misskey{token: token}
}

func (m *Misskey) Post(id string, title string) error {
	content := fmt.Sprintf(`
	【5分後に公開】
	%s
	https://www.youtube.com/watch?v=%s
	`, title, id)

	resb := ReqBody{
		I:      m.token,
		Text:   content,
		Detail: false,
	}

	payload, err := json.Marshal(resb)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
