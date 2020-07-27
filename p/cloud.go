// Package p contains a Pub/Sub Cloud Function.
package p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type Config struct {
	TransferwiseAPIKey string `envconfig:"TRANSFERWISE_API_KEY" required:"true"`
	SlackWebhookURL    string `envconfig:"SLACK_WEBHOOK_URL" required:"true"`
}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(ctx context.Context, m PubSubMessage) error {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return fmt.Errorf("failed to process env vars: %w", err)
	}

	rate, err := getExchangeRate(config.TransferwiseAPIKey)
	if err != nil {
		return err
	}

	if err := sendToSlack(config.SlackWebhookURL, rate); err != nil {
		return err
	}

	return nil
}

type exchangeRateResponse struct {
}

type exchangeRate struct {
	Rate   float64 `json:"rate"`
	Source string  `json:"source"`
	Target string  `json:"target"`
	// Time   time.Time `json:"time"`
}

func getExchangeRate(apiKey string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.transferwise.com/v1/rates?source=USD&target=AUD", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create exchange rate request: %w", err)
	}

	req.Header.Set("User-Agent", "github.com/porty/transferwise-exchange-rate")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send exchange rate request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status code from exchange rate API: %s", resp.Status)
	}

	var rates []exchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return 0, fmt.Errorf("failed to unmarshal exchange rate API response: %w", err)
	}

	if len(rates) != 1 {
		return 0, fmt.Errorf("expected one exchange rate, recieved %d", len(rates))
	}

	return rates[0].Rate, nil
}

type slackMessage struct {
	// Text is what is said
	Text string `json:"text"`
	// Username is an optional username for the bot message
	Username string `json:"username,omitempty"`
	// IconURL is an optional URL for the bot avatar
	IconURL string `json:"icon_url,omitempty"`
	// IconEmoji is an optional emoji for the bot avatar, i.e. `:ghost"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	// Channel is an optional channel for the message
	Channel string `json:"channel,omitempty"`
	// Fallback is an optional text summary of attachments
	Fallback string `json:"fallback,omitempty"`
	// Pretext optionally appears above formatted data
	Pretext string `json:"pretext,omitempty"`
	// Color is the color for the attachment, i.e. `#36a64f`, `good`, `warning`, `danger`
	Color string `json:"color,omitempty"`
	// Fields are displayed in a table on the message
	Fields []slackMessageField `json:"fields,omitempty"`
}

type slackMessageField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func sendToSlack(slackURL string, rate float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := slackMessage{
		Text:      fmt.Sprintf("The exchange rate is %.5f", rate),
		IconEmoji: ":moneybag:",
		Username:  "TransferwiseBot",
		Channel:   "transferwise",
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to JSON marshal Slack message: %w", err)
	}

	body := bytes.NewReader(b)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackURL, body)
	if err != nil {
		return fmt.Errorf("failed to create Slack webhook request: %w", err)
	}

	req.Header.Set("User-Agent", "github.com/porty/transferwise-exchange-rate")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code from Slack webhook: %s", resp.Status)
	}
	return nil
}
