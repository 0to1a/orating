// Package mailer wraps Resend with a console fallback. New("", from)
// returns a ConsoleMailer that prints to slog — convenient for dev.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string // fallback for clients that block HTML
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

func New(apiKey, from string) Mailer {
	if apiKey == "" {
		slog.Info("mailer: no RESEND_API_KEY, using ConsoleMailer")
		return &ConsoleMailer{From: from}
	}
	return &ResendMailer{APIKey: apiKey, From: from, http: &http.Client{Timeout: 10 * time.Second}}
}

type ConsoleMailer struct {
	From string
}

func (m *ConsoleMailer) Send(_ context.Context, msg Message) error {
	slog.Info("mailer (console)",
		"from", m.From,
		"to", msg.To,
		"subject", msg.Subject,
		"text", msg.Text,
	)
	return nil
}

// ResendMailer kirim via Resend.com REST API.
// https://resend.com/docs/api-reference/emails/send-email
type ResendMailer struct {
	APIKey string
	From   string
	http   *http.Client
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

func (m *ResendMailer) Send(ctx context.Context, msg Message) error {
	body, err := json.Marshal(resendPayload{
		From:    m.From,
		To:      []string{msg.To},
		Subject: msg.Subject,
		HTML:    msg.HTML,
		Text:    msg.Text,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := m.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("resend error: status %d", res.StatusCode)
	}
	return nil
}
