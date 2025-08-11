package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	DefaultBaseURL = "http://localhost:11434"
)

// Client is a client for the Ollama API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Ollama client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Generate a response for a given prompt with a provided model.
func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	req.Stream = false
	var resp GenerateResponse
	err := c.do(ctx, http.MethodPost, "/api/generate", req, &resp)
	return &resp, err
}

// GenerateStream a response for a given prompt with a provided model.
func (c *Client) GenerateStream(ctx context.Context, req *GenerateRequest, fn func(GenerateResponse)) error {
	req.Stream = true
	return c.doStream(ctx, http.MethodPost, "/api/generate", req, func(data []byte) {
		var resp GenerateResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			fn(resp)
		}
	})
}

// Chat generates the next message in a chat with a provided model.
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	req.Stream = false
	var resp ChatResponse
	err := c.do(ctx, http.MethodPost, "/api/chat", req, &resp)
	return &resp, err
}

// ChatStream generates the next message in a chat with a provided model.
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest, fn func(ChatResponse)) error {
	req.Stream = true
	return c.doStream(ctx, http.MethodPost, "/api/chat", req, func(data []byte) {
		var resp ChatResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			fn(resp)
		}
	})
}

// ListModels lists models that are available locally.
func (c *Client) ListModels(ctx context.Context) (*ListModelsResponse, error) {
	var resp ListModelsResponse
	err := c.do(ctx, http.MethodGet, "/api/tags", nil, &resp)
	return &resp, err
}

// ShowModel shows information about a model.
func (c *Client) ShowModel(ctx context.Context, req *ShowModelRequest) (*ShowModelResponse, error) {
	var resp ShowModelResponse
	err := c.do(ctx, http.MethodPost, "/api/show", req, &resp)
	return &resp, err
}

// CopyModel copies a model.
func (c *Client) CopyModel(ctx context.Context, req *CopyModelRequest) error {
	return c.do(ctx, http.MethodPost, "/api/copy", req, nil)
}

// DeleteModel deletes a model.
func (c *Client) DeleteModel(ctx context.Context, req *DeleteModelRequest) error {
	return c.do(ctx, http.MethodDelete, "/api/delete", req, nil)
}

// PullModel downloads a model from the ollama library.
func (c *Client) PullModel(ctx context.Context, req *PullModelRequest, progressFn func(ProgressResponse)) error {
	req.Stream = true
	return c.doStream(ctx, http.MethodPost, "/api/pull", req, func(data []byte) {
		var resp ProgressResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			progressFn(resp)
		}
	})
}

// PushModel uploads a model to a model library.
func (c *Client) PushModel(ctx context.Context, req *PushModelRequest, progressFn func(ProgressResponse)) error {
	req.Stream = true
	return c.doStream(ctx, http.MethodPost, "/api/push", req, func(data []byte) {
		var resp ProgressResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			progressFn(resp)
		}
	})
}

// CreateModel creates a model from a Modelfile.
func (c *Client) CreateModel(ctx context.Context, req *CreateModelRequest, progressFn func(ProgressResponse)) error {
	req.Stream = true
	return c.doStream(ctx, http.MethodPost, "/api/create", req, func(data []byte) {
		var resp ProgressResponse
		if err := json.Unmarshal(data, &resp); err == nil {
			progressFn(resp)
		}
	})
}

// GenerateEmbeddings generates embeddings from a model.
func (c *Client) GenerateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	var resp EmbeddingsResponse
	err := c.do(ctx, http.MethodPost, "/api/embed", req, &resp)
	return &resp, err
}

// ListRunningModels lists models that are currently loaded into memory.
func (c *Client) ListRunningModels(ctx context.Context) (*ListRunningModelsResponse, error) {
	var resp ListRunningModelsResponse
	err := c.do(ctx, http.MethodGet, "/api/ps", nil, &resp)
	return &resp, err
}

func (c *Client) do(ctx context.Context, method, path string, reqData, respData interface{}) error {
	var body io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error: %s, %s", resp.Status, string(bodyBytes))
	}

	if respData != nil {
		return json.NewDecoder(resp.Body).Decode(respData)
	}

	return nil
}

func (c *Client) doStream(ctx context.Context, method, path string, reqData interface{}, fn func([]byte)) error {
	var body io.Reader
	if reqData != nil {
		jsonData, err := json.Marshal(reqData)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error: %s, %s", resp.Status, string(bodyBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fn(scanner.Bytes())
	}

	return scanner.Err()
}

// CheckBlob checks if a blob exists on the server.
func (c *Client) CheckBlob(ctx context.Context, digest string) (bool, error) {
	path := fmt.Sprintf("/api/blobs/%s", digest)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL+path, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("http error: %s, %s", resp.Status, string(bodyBytes))
	}
}

// CreateBlob creates a blob from a file on the server.
func (c *Client) CreateBlob(ctx context.Context, digest string, data io.Reader) error {
	path := fmt.Sprintf("/api/blobs/%s", digest)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http error: %s, %s", resp.Status, string(bodyBytes))
	}

	return nil
}

// KeepAliveRequest is used to keep a model loaded in memory.
type KeepAliveRequest struct {
	Model    string `json:"model"`
	KeepAlive string `json:"keep_alive"`
}

// LoadModel loads a model into memory.
func (c *Client) LoadModel(ctx context.Context, model string, keepAlive string) error {
	req := KeepAliveRequest{
		Model:    model,
		KeepAlive: keepAlive,
	}
	// Use generate endpoint with empty prompt to load the model
	gReq := GenerateRequest{
		Model:    model,
		Prompt:   "",
		KeepAlive: keepAlive,
	}
	_, err := c.Generate(ctx, &gReq)
	return err
}

// UnloadModel unloads a model from memory.
func (c *Client) UnloadModel(ctx context.Context, model string) error {
	req := KeepAliveRequest{
		Model:    model,
		KeepAlive: "0",
	}
	// Use generate endpoint with empty prompt and keep_alive=0 to unload the model
	gReq := GenerateRequest{
		Model:    model,
		Prompt:   "",
		KeepAlive: "0",
	}
	_, err := c.Generate(ctx, &gReq)
	return err
}
