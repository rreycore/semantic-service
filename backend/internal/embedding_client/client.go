package embedding_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type EmbeddingRequest struct {
	Input      any    `json:"input"`
	Model      string `json:"model,omitempty"`
	Dimensions *int   `json:"dimensions,omitempty"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  UsageData       `json:"usage"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type UsageData struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type ErrorResponse struct {
	Detail string `json:"detail"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/ping", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute ping request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ping failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read ping response body: %w", err)
	}

	return string(body), nil
}

func (c *Client) CreateSearchEmbedding(ctx context.Context, query string) (*EmbeddingResponse, error) {
	formattedQuery := fmt.Sprintf("task: search result | query: %s", query)

	request := EmbeddingRequest{
		Input: formattedQuery,
	}

	return c.createEmbeddings(ctx, request)
}

func (c *Client) createEmbeddings(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}
	switch req.Input.(type) {
	case string, []string:
	default:
		return nil, fmt.Errorf("input must be a string or a slice of strings")
	}

	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute embeddings request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status code %d and invalid error response", resp.StatusCode)
		}
		return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, errResp.Detail)
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode successful response: %w", err)
	}

	return &embeddingResp, nil
}
