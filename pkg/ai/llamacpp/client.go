package llamacpp

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

// Client is a client for the Llama.cpp server.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Llama.cpp server client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, buf)
	if err != nil {
		return nil, err
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errRes APIError
		if err := json.NewDecoder(resp.Body).Decode(&errRes); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, errRes.Message)
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return err
		}
	}

	return nil
}

// HealthCheck checks the health of the server.
func (c *Client) HealthCheck(ctx context.Context) (*HealthCheckResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return nil, err
	}

	var resp HealthCheckResponse
	err = c.do(req, &resp)
	return &resp, err
}

// Completion performs text completion.
func (c *Client) Completion(ctx context.Context, completionReq *CompletionRequest) (*CompletionResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/completion", completionReq)
	if err != nil {
		return nil, err
	}

	var resp CompletionResponse
	err = c.do(req, &resp)
	return &resp, err
}

// CompletionStream performs text completion with streaming.
func (c *Client) CompletionStream(ctx context.Context, completionReq *CompletionRequest) (<-chan *CompletionResponse, error) {
	completionReq.Stream = true
	req, err := c.newRequest(ctx, http.MethodPost, "/completion", completionReq)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *CompletionResponse)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				var streamResp CompletionResponse
				if err := json.Unmarshal([]byte(data), &streamResp); err == nil {
					ch <- &streamResp
				}
			}
		}
	}()

	return ch, nil
}

// Tokenize tokenizes text.
func (c *Client) Tokenize(ctx context.Context, tokenizeReq *TokenizeRequest) (*TokenizeResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/tokenize", tokenizeReq)
	if err != nil {
		return nil, err
	}

	var resp TokenizeResponse
	err = c.do(req, &resp)
	return &resp, err
}

// Detokenize detokenizes tokens.
func (c *Client) Detokenize(ctx context.Context, detokenizeReq *DetokenizeRequest) (*DetokenizeResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/detokenize", detokenizeReq)
	if err != nil {
		return nil, err
	}

	var resp DetokenizeResponse
	err = c.do(req, &resp)
	return &resp, err
}

// ApplyTemplate applies a chat template.
func (c *Client) ApplyTemplate(ctx context.Context, applyTemplateReq *ApplyTemplateRequest) (*ApplyTemplateResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/apply-template", applyTemplateReq)
	if err != nil {
		return nil, err
	}

	var resp ApplyTemplateResponse
	err = c.do(req, &resp)
	return &resp, err
}

// Embedding gets text embeddings.
func (c *Client) Embedding(ctx context.Context, embeddingReq *EmbeddingRequest) (*EmbeddingResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/embedding", embeddingReq)
	if err != nil {
		return nil, err
	}

	var resp EmbeddingResponse
	err = c.do(req, &resp)
	return &resp, err
}

// Rerank reranks documents.
func (c *Client) Rerank(ctx context.Context, rerankingReq *RerankingRequest) (*RerankingResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/rerank", rerankingReq)
	if err != nil {
		return nil, err
	}

	var resp RerankingResponse
	err = c.do(req, &resp)
	return &resp, err
}

// Infill performs code infilling.
func (c *Client) Infill(ctx context.Context, infillReq *InfillRequest) (*CompletionResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/infill", infillReq)
	if err != nil {
		return nil, err
	}

	var resp CompletionResponse
	err = c.do(req, &resp)
	return &resp, err
}

// GetProps gets server properties.
func (c *Client) GetProps(ctx context.Context) (*PropsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/props", nil)
	if err != nil {
		return nil, err
	}

	var resp PropsResponse
	err = c.do(req, &resp)
	return &resp, err
}

// PostProps changes server global properties.
func (c *Client) PostProps(ctx context.Context, propsReq *PostPropsRequest) error {
	req, err := c.newRequest(ctx, http.MethodPost, "/props", propsReq)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// NonOAIEmbeddings gets non-OpenAI-compatible embeddings.
func (c *Client) NonOAIEmbeddings(ctx context.Context, embeddingsReq *OpenAIEmbeddingsRequest) (*EmbeddingsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/embeddings", embeddingsReq)
	if err != nil {
		return nil, err
	}

	var resp EmbeddingsResponse
	err = c.do(req, &resp)
	return &resp, err
}

// GetSlots gets the state of processing slots.
func (c *Client) GetSlots(ctx context.Context) (*SlotsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/slots", nil)
	if err != nil {
		return nil, err
	}

	var resp SlotsResponse
	err = c.do(req, &resp)
	return &resp, err
}

// SaveSlot saves the prompt cache of a slot.
func (c *Client) SaveSlot(ctx context.Context, idSlot int, saveReq *SlotSaveRequest) (*SlotSaveResponse, error) {
	path := fmt.Sprintf("/slots/%d?action=save", idSlot)
	req, err := c.newRequest(ctx, http.MethodPost, path, saveReq)
	if err != nil {
		return nil, err
	}

	var resp SlotSaveResponse
	err = c.do(req, &resp)
	return &resp, err
}

// RestoreSlot restores the prompt cache of a slot.
func (c *Client) RestoreSlot(ctx context.Context, idSlot int, restoreReq *SlotRestoreRequest) (*SlotRestoreResponse, error) {
	path := fmt.Sprintf("/slots/%d?action=restore", idSlot)
	req, err := c.newRequest(ctx, http.MethodPost, path, restoreReq)
	if err != nil {
		return nil, err
	}

	var resp SlotRestoreResponse
	err = c.do(req, &resp)
	return &resp, err
}

// EraseSlot erases the prompt cache of a slot.
func (c *Client) EraseSlot(ctx context.Context, idSlot int) (*SlotEraseResponse, error) {
	path := fmt.Sprintf("/slots/%d?action=erase", idSlot)
	req, err := c.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var resp SlotEraseResponse
	err = c.do(req, &resp)
	return &resp, err
}

// GetLoRAAdapters gets the list of LoRA adapters.
func (c *Client) GetLoRAAdapters(ctx context.Context) (*GetLoRAAdaptersResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/lora-adapters", nil)
	if err != nil {
		return nil, err
	}

	var resp GetLoRAAdaptersResponse
	err = c.do(req, &resp)
	return &resp, err
}

// SetLoRAAdapters sets the LoRA adapters.
func (c *Client) SetLoRAAdapters(ctx context.Context, loraAdaptersReq *PostLoRAAdaptersRequest) error {
	req, err := c.newRequest(ctx, http.MethodPost, "/lora-adapters", loraAdaptersReq)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// GetOpenAIModels gets OpenAI-compatible models.
func (c *Client) GetOpenAIModels(ctx context.Context) (*OpenAIModelsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/v1/models", nil)
	if err != nil {
		return nil, err
	}

	var resp OpenAIModelsResponse
	err = c.do(req, &resp)
	return &resp, err
}

// OpenAICompletions performs OpenAI-compatible completions.
func (c *Client) OpenAICompletions(ctx context.Context, completionsReq *OpenAICompletionsRequest) (*OpenAICompletionsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/v1/completions", completionsReq)
	if err != nil {
		return nil, err
	}

	var resp OpenAICompletionsResponse
	err = c.do(req, &resp)
	return &resp, err
}

// OpenAIChatCompletions performs OpenAI-compatible chat completions.
func (c *Client) OpenAIChatCompletions(ctx context.Context, chatCompletionsReq *OpenAIChatCompletionsRequest) (*OpenAIChatCompletionsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/v1/chat/completions", chatCompletionsReq)
	if err != nil {
		return nil, err
	}

	var resp OpenAIChatCompletionsResponse
	err = c.do(req, &resp)
	return &resp, err
}

// OpenAIEmbeddings gets OpenAI-compatible embeddings.
func (c *Client) OpenAIEmbeddings(ctx context.Context, embeddingsReq *OpenAIEmbeddingsRequest) (*OpenAIEmbeddingsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/v1/embeddings", embeddingsReq)
	if err != nil {
		return nil, err
	}

	var resp OpenAIEmbeddingsResponse
	err = c.do(req, &resp)
	return &resp, err
}
