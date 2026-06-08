package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/psanford/wormhole-william/wormhole"
)

// SharePayload is the JSON envelope sent through the wormhole.
// It carries the tab title and its full body text.
type SharePayload struct {
	TabTitle string `json:"tab_title"`
	Body     string `json:"body"`
	Version  int    `json:"version"`
}

// ShareResult is returned to callers after a successful receive.
type ShareResult struct {
	TabTitle string
	Body     string
}

// ShareSend serialises the given tab and opens a Magic Wormhole.
// It returns the human-friendly code (e.g. "7-crossover-alpha") immediately,
// then blocks on the returned wait func until the peer has received the data.
//
// The caller should pass a cancellable ctx to abort waiting.
func ShareSend(ctx context.Context, tab Tab) (code string, wait func() error, err error) {
	payload := SharePayload{
		TabTitle: tab.Title,
		Body:     tab.Body,
		Version:  1,
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
	return ShareResult{TabTitle: title, Body: payload.Body}, nil
}
