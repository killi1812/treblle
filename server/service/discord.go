package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"treblle/app"

	"go.uber.org/zap"
)

type DiscordService struct {
	logger *zap.SugaredLogger
}

func NewDiscordService() *DiscordService {
	var service *DiscordService
	app.Invoke(func(logger *zap.SugaredLogger) {
		service = &DiscordService{
			logger: logger,
		}
	})
	return service
}

// tokenResponse matches the structure of the JSON response from Discord's API.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

// ExchangeCodeForToken sends the authorization code to Discord to get an access token.
func (DiscordService) ExchangeCodeForToken(code string) (string, error) {
	// API endpoint for token exchange
	tokenURL := "https://discord.com/api/oauth2/token"

	// Prepare the form data for the POST request
	data := url.Values{}
	data.Set("client_id", os.Getenv("VITE_DISCORD_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("DISCORD_CLIENT_SECRET"))
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set the required header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a non-200 status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bad status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the JSON response to retrieve the access_token
	var tokenResponse tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode JSON response: %w", err)
	}

	// Return the access token
	return tokenResponse.AccessToken, nil
}
