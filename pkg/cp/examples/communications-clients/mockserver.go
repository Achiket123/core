package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// MockServer provides HTTP endpoints that simulate Slack, Teams, and Discord webhooks
type MockServer struct {
	server   *http.Server
	messages []string
	mu       sync.Mutex
}

// NewMockServer creates a new mock communication server
func NewMockServer(port string) *MockServer {
	ms := &MockServer{
		messages: []string{},
	}

	mux := http.NewServeMux()

	// Slack endpoints
	mux.HandleFunc("/slack/chat.postMessage", ms.handleSlackMessage)
	mux.HandleFunc("/slack/conversations.open", ms.handleSlackConversationsOpen)
	mux.HandleFunc("/slack/conversations.list", ms.handleSlackConversationsList)

	// Teams webhook endpoint
	mux.HandleFunc("/teams/webhook", ms.handleTeamsWebhook)

	// Discord webhook endpoint
	mux.HandleFunc("/discord/webhook", ms.handleDiscordWebhook)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	ms.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return ms
}

func (ms *MockServer) handleSlackMessage(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	channel, _ := payload["channel"].(string)
	text, _ := payload["text"].(string)

	ms.mu.Lock()
	ms.messages = append(ms.messages, fmt.Sprintf("Slack message to %s: %s", channel, text))
	ms.mu.Unlock()

	log.Printf("[Mock Slack] Received message to channel %s: %s", channel, text)

	response := map[string]interface{}{
		"ok":      true,
		"channel": channel,
		"ts":      "1234567890.123456",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) handleSlackConversationsOpen(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	users, _ := payload["users"].(string)
	log.Printf("[Mock Slack] Opening DM conversation with %s", users)

	response := map[string]interface{}{
		"ok": true,
		"channel": map[string]interface{}{
			"id": "D123456",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) handleSlackConversationsList(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Mock Slack] Listing conversations")

	response := map[string]interface{}{
		"ok": true,
		"channels": []map[string]interface{}{
			{"name": "general", "id": "C123456"},
			{"name": "announcements", "id": "C234567"},
			{"name": "support", "id": "C345678"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ms *MockServer) handleTeamsWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	summary, _ := payload["summary"].(string)

	ms.mu.Lock()
	ms.messages = append(ms.messages, fmt.Sprintf("Teams message: %s", summary))
	ms.mu.Unlock()

	log.Printf("[Mock Teams] Received webhook message: %s", summary)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("1"))
}

func (ms *MockServer) handleDiscordWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	content, _ := payload["content"].(string)

	ms.mu.Lock()
	ms.messages = append(ms.messages, fmt.Sprintf("Discord message: %s", content))
	ms.mu.Unlock()

	log.Printf("[Mock Discord] Received webhook message: %s", content)

	w.WriteHeader(http.StatusNoContent)
}

// Start begins listening for HTTP requests
func (ms *MockServer) Start() error {
	log.Printf("Mock communication server starting on %s", ms.server.Addr)
	return ms.server.ListenAndServe()
}

// Stop gracefully shuts down the server
func (ms *MockServer) Stop() error {
	return ms.server.Close()
}

// GetMessages returns all received messages
func (ms *MockServer) GetMessages() []string {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	messages := make([]string, len(ms.messages))
	copy(messages, ms.messages)
	return messages
}
