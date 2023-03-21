package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v50/github"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

type GithubEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
}

type PullRequest struct {
	User User `json:"user"`
}

type User struct {
	ID int `json:"id"`
}

type MessageRequest struct {
	ID          uuid.UUID `json:"id"`
	GitID       int       `json:"gitId"`
	MessageType string    `json:"messageType"`
	Payload     []byte    `json:"payload"`
}

func validateSignature(signature string, payload []byte) error {
	webhookSecret := os.Getenv("GH_WEBHOOK_SECRET")
	err := github.ValidateSignature(signature, payload, []byte(webhookSecret))
	if err != nil {
		return fmt.Errorf("error validating signature: %w", err)
	}
	return nil
}

func Main(args map[string]interface{}) map[string]interface{} {
	rawReq := args["http"].(map[string]interface{})

	payload := rawReq["body"].([]byte)
	signature := rawReq["headers"].(map[string]string)["x-hub-signature-256"]

	err := validateSignature(signature, payload)
	if err != nil {
		return map[string]interface{}{
			"statusCode": http.StatusUnauthorized,
			"body":       err.Error(),
		}
	}

	// Parse the payload
	var event GithubEvent
	err = json.Unmarshal(payload, &event)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshalling payload")
		return map[string]interface{}{
			"statusCode": http.StatusBadRequest,
			"body":       err.Error(),
		}
	}
	log.Info().Msgf("Received event action: %s", event.Action)

	// Create the message request
	messageRequest := MessageRequest{
		ID:          uuid.New(),
		GitID:       event.PullRequest.User.ID,
		MessageType: "github",
		Payload:     payload,
	}
	data, err := json.Marshal(messageRequest)
	if err != nil {
		return map[string]interface{}{
			"statusCode": http.StatusInternalServerError,
			"body":       err.Error(),
		}
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080",
		bytes.NewBuffer(data),
	)
	if err != nil {
		log.Error().Err(err).Msg("Error creating request")
		return map[string]interface{}{
			"statusCode": http.StatusInternalServerError,
			"body":       err.Error(),
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Require-Whisk-Auth", os.Getenv("WSK_AUTH"))

	log.Info().Msg("Sending message to messenger...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"statusCode": http.StatusInternalServerError,
			"body":       err.Error(),
		}
	}

	if resp.StatusCode != http.StatusCreated {
		return map[string]interface{}{
			"statusCode": resp.StatusCode,
			"body":       "Messenger returned non-200 status code",
		}
	}

	return map[string]interface{}{
		"statusCode": http.StatusCreated,
		"body":       "Created",
	}
}
