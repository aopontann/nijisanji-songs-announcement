package nsa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Result struct {
	Token string `json:"token"`
	Song  int    `json:"song"`
	Word  string `json:"word"`
	Time  string `json:"time"`
}

type D1Response struct {
	Result []struct {
		Results []Result `json:"results"`
		Success bool `json:"success"`
		Meta    struct {
			ServedBy    string  `json:"served_by"`
			Duration    float64 `json:"duration"`
			Changes     int     `json:"changes"`
			LastRowID   int     `json:"last_row_id"`
			ChangedDb   bool    `json:"changed_db"`
			SizeAfter   int     `json:"size_after"`
			RowsRead    int     `json:"rows_read"`
			RowsWritten int     `json:"rows_written"`
		} `json:"meta"`
	} `json:"result"`
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
}

func GetUserTokenList() ([]Result, error) {

	url := os.Getenv("D1_URL")
	method := "POST"
	token := os.Getenv("D1_TOKEN")

	payload := strings.NewReader(`
  {
    "sql": "SELECT * FROM users;"
  }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	s := &D1Response{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	json.Unmarshal(body, s)
	// fmt.Println(string(body))
	return s.Result[0].Results, nil
}