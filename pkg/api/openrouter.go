package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/notme6000/go-scrape/pkg/types"
)

const (
	openRouterURL   = "https://openrouter.ai/api/v1/chat/completions"
	openRouterModel = "openrouter/auto"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

func (c *Client) ValidateNamesWithAI(candidates []string) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	candidatesStr := strings.Join(candidates, ", ")

	prompt := fmt.Sprintf(`From the following list of text candidates, identify which ones are actual person names. Return ONLY a comma-separated list of valid names, nothing else.

Candidates: %s

Valid names:`, candidatesStr)

	reqBody := types.OpenRouterRequest{
		Model: openRouterModel,
		Messages: []types.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error marshaling request:", err)
		return candidates
	}

	req, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return candidates
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making API call:", err)
		return candidates
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return candidates
	}

	var response types.OpenRouterResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return candidates
	}

	if len(response.Choices) == 0 {
		return candidates
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)
	validNames := strings.Split(content, ",")

	var cleanNames []string
	for _, name := range validNames {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			cleanNames = append(cleanNames, trimmed)
		}
	}

	return cleanNames
}
