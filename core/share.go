package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/psanford/wormhole-william/wormhole"
)

// SharePayload is the JSON envelope sent through the wormhole.
// It carries the tab title, body, and an optional human label for the sender.
type SharePayload struct {
	TabTitle    string `json:"tab_title"`
	Body        string `json:"body"`
	SenderLabel string `json:"sender_label,omitempty"` // e.g. "Alice" — purely informational
	Version     int    `json:"version"`
}

// ShareResult is returned to callers after a successful receive.
type ShareResult struct {
	TabTitle    string
	Body        string
	SenderLabel string // empty string if the sender did not provide one
}

// ShareSend serialises the given tab and opens a Magic Wormhole.
// senderLabel is an optional human-readable name for the sender (e.g. "Alice").
// It is embedded in the encrypted payload and shown to the receiver as context.
// Pass an empty string to omit it.
//
// It returns the human-friendly code (e.g. "7-crossover-alpha") immediately,
// then blocks on the returned wait func until the peer has received the data.
// The caller should pass a cancellable ctx to abort waiting.
func ShareSend(ctx context.Context, tab Tab, senderLabel string) (code string, wait func() error, err error) {
	payload := SharePayload{
		TabTitle:    tab.Title,
		Body:        tab.Body,
		SenderLabel: senderLabel,
		Version:     1,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", nil, fmt.Errorf("share: marshal payload: %w", err)
	}

	var c wormhole.Client
	code, statusCh, err := c.SendText(ctx, string(data))
	if err != nil {
		return "", nil, fmt.Errorf("share: open wormhole: %w", err)
	}

	wait = func() error {
		s, ok := <-statusCh
		if !ok {
			return fmt.Errorf("share: connection closed before transfer")
		}
		if s.Error != nil {
			return fmt.Errorf("share: transfer error: %w", s.Error)
		}
		return nil
	}

	return code, wait, nil
}

// ShareReceive connects to an existing wormhole using the code typed by the user.
// It blocks until the text is fully received, then returns the decoded tab content.
func ShareReceive(ctx context.Context, code string) (ShareResult, error) {
	var c wormhole.Client

	msg, err := c.Receive(ctx, code)
	if err != nil {
		return ShareResult{}, fmt.Errorf("share: receive handshake: %w", err)
	}

	if msg.Type != wormhole.TransferText {
		_ = msg.Reject()
		return ShareResult{}, fmt.Errorf("share: peer sent a file, not text — wrong tool?")
	}

	raw, err := io.ReadAll(msg)
	if err != nil {
		return ShareResult{}, fmt.Errorf("share: read message: %w", err)
	}

	var payload SharePayload
	if jsonErr := json.Unmarshal(raw, &payload); jsonErr != nil {
		// Graceful fallback: treat the raw text as the tab body if it's not
		// our JSON envelope (e.g. shared from the reference Python client).
		return ShareResult{
			TabTitle: "shared note",
			Body:     string(raw),
		}, nil
	}

	title := payload.TabTitle
	if title == "" {
		title = "shared note"
	}
	// Prefix the tab title with the sender's label so the receiver immediately
	// knows where the content came from, e.g. "From Alice · my-notes"
	if payload.SenderLabel != "" {
		title = "From " + payload.SenderLabel + " · " + title
	}
	return ShareResult{
		TabTitle:    title,
		Body:        payload.Body,
		SenderLabel: payload.SenderLabel,
	}, nil
}
