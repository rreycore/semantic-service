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

// --- Модели запросов и ответов (соответствуют Pydantic-моделям) ---

// EmbeddingRequest представляет тело запроса к эндпоинту /embeddings.
type EmbeddingRequest struct {
	// Input может быть как строкой (string), так и срезом строк ([]string).
	// Использование `any` позволяет обрабатывать оба случая.
	Input any `json:"input"`
	// Model - необязательное поле, omitempty не будет включать его в JSON, если оно пустое.
	Model string `json:"model,omitempty"`
	// Dimensions - необязательное поле, используем указатель, чтобы можно было передать 0.
	Dimensions *int `json:"dimensions,omitempty"`
}

// EmbeddingResponse представляет успешный ответ от эндпоинта /embeddings.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  UsageData       `json:"usage"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"` // float в Python обычно соответствует float64 в Go
	Index     int       `json:"index"`
}

type UsageData struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ErrorResponse представляет тело ответа при ошибке от сервера.
type ErrorResponse struct {
	Detail string `json:"detail"`
}

// --- Клиент для взаимодействия с сервисом ---

// Client - это клиент для API сервиса эмбеддингов.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создает новый экземпляр клиента.
// baseURL - базовый URL сервиса, например "http://localhost:8000".
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"), // Убираем слэш в конце, если он есть
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Ping проверяет доступность сервиса.
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

// CreateEmbeddings запрашивает векторные представления для одного или нескольких текстов.
func (c *Client) CreateEmbeddings(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	// Валидация входных данных
	if req.Input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}
	switch req.Input.(type) {
	case string, []string:
		// all good
	default:
		return nil, fmt.Errorf("input must be a string or a slice of strings")
	}

	// Кодируем тело запроса в JSON
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Создаем HTTP-запрос
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute embeddings request: %w", err)
	}
	defer resp.Body.Close()

	// Обрабатываем ответ
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			// Если не удалось распарсить JSON ошибки, возвращаем просто статус
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
