package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SlackClient implements MessageClient for Slack using real HTTP API
type SlackClient struct {
	token      string
	teamID     string
	baseURL    string
	httpClient *http.Client
}

func newSlackClient(token, teamID, baseURL string) *SlackClient {
	if baseURL == "" {
		baseURL = "https://slack.com/api"
	}
	return &SlackClient{
		token:   token,
		teamID:  teamID,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SlackClient) SendMessage(ctx context.Context, channel, message string) error {
	payload := map[string]interface{}{
		"channel": channel,
		"text":    message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		return fmt.Errorf("slack API error: %s", errMsg)
	}

	fmt.Printf("[Slack] ✓ Sent message to %s: %s\n", channel, message)
	return nil
}

func (s *SlackClient) SendDirectMessage(ctx context.Context, userID, message string) error {
	// First, open a DM channel
	openPayload := map[string]interface{}{
		"users": userID,
	}

	jsonData, err := json.Marshal(openPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/conversations.open", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var openResult map[string]interface{}
	if err := json.Unmarshal(body, &openResult); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if ok, _ := openResult["ok"].(bool); !ok {
		errMsg, _ := openResult["error"].(string)
		return fmt.Errorf("slack API error: %s", errMsg)
	}

	channel, _ := openResult["channel"].(map[string]interface{})
	channelID, _ := channel["id"].(string)

	// Now send the message
	return s.SendMessage(ctx, channelID, message)
}

func (s *SlackClient) GetChannels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/conversations.list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slack API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		return nil, fmt.Errorf("slack API error: %s", errMsg)
	}

	channels := []string{}
	if channelList, ok := result["channels"].([]interface{}); ok {
		for _, ch := range channelList {
			if channelData, ok := ch.(map[string]interface{}); ok {
				if name, ok := channelData["name"].(string); ok {
					channels = append(channels, "#"+name)
				}
			}
		}
	}

	return channels, nil
}

func (s *SlackClient) Close() error {
	s.httpClient.CloseIdleConnections()
	return nil
}

// TeamsClient implements MessageClient for Microsoft Teams using Incoming Webhooks
type TeamsClient struct {
	webhookURL string
	httpClient *http.Client
}

func newTeamsClient(webhookURL string) *TeamsClient {
	return &TeamsClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *TeamsClient) SendMessage(ctx context.Context, channel, message string) error {
	payload := map[string]interface{}{
		"@type":    "MessageCard",
		"@context": "https://schema.org/extensions",
		"summary":  message,
		"sections": []map[string]interface{}{
			{
				"activityTitle": "Message to " + channel,
				"text":          message,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("teams API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("teams API returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[Teams] ✓ Sent message to %s: %s\n", channel, message)
	return nil
}

func (t *TeamsClient) SendDirectMessage(ctx context.Context, userID, message string) error {
	// For webhooks, we can't send DMs, so we'll send to the default channel
	return t.SendMessage(ctx, "DM to "+userID, message)
}

func (t *TeamsClient) GetChannels(ctx context.Context) ([]string, error) {
	// Webhooks don't support listing channels
	return []string{"General (webhook)"}, nil
}

func (t *TeamsClient) Close() error {
	t.httpClient.CloseIdleConnections()
	return nil
}

// DiscordClient implements MessageClient for Discord using webhooks
type DiscordClient struct {
	webhookURL string
	httpClient *http.Client
}

func newDiscordClient(webhookURL string) *DiscordClient {
	return &DiscordClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *DiscordClient) SendMessage(ctx context.Context, channel, message string) error {
	payload := map[string]interface{}{
		"content": fmt.Sprintf("**%s**: %s", channel, message),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[Discord] ✓ Sent message to #%s: %s\n", channel, message)
	return nil
}

func (d *DiscordClient) SendDirectMessage(ctx context.Context, userID, message string) error {
	// For webhooks, we can't send DMs, so we'll send to the default channel
	return d.SendMessage(ctx, "DM-"+userID, message)
}

func (d *DiscordClient) GetChannels(ctx context.Context) ([]string, error) {
	// Webhooks don't support listing channels
	return []string{"#general (webhook)"}, nil
}

func (d *DiscordClient) Close() error {
	d.httpClient.CloseIdleConnections()
	return nil
}
