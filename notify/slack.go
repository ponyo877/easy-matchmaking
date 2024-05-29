package notify

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Slack struct {
	webhookEndpoint string
}

func NewSlack(webhookEndpoint string) *Slack {
	return &Slack{webhookEndpoint}

}

func (s Slack) Notify(text string) error {
	message := struct {
		Text string `json:"text"`
	}{
		Text: text,
	}
	jsonStr, _ := json.Marshal(message)
	req, err := http.NewRequest(
		http.MethodPost,
		s.webhookEndpoint,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
