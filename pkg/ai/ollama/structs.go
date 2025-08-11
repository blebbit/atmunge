package ollama

import (
	"time"
)

type GenerateRequest struct {
	Model    string   `json:"model"`
	Prompt   string   `json:"prompt"`
	Suffix   string   `json:"suffix,omitempty"`
	Images   []string `json:"images,omitempty"`
	Format   string   `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string   `json:"system,omitempty"`
	Template string   `json:"template,omitempty"`
	Context  []int    `json:"context,omitempty"`
	Stream   bool     `json:"stream"`
	Raw      bool     `json:"raw,omitempty"`
	KeepAlive string  `json:"keep_alive,omitempty"`
}

type GenerateResponse struct {
	Model           string    `json:"model"`
	CreatedAt       time.Time `json:"created_at"`
	Response        string    `json:"response"`
	Done            bool      `json:"done"`
	Context         []int     `json:"context,omitempty"`
	TotalDuration   int64     `json:"total_duration,omitempty"`
	LoadDuration    int64     `json:"load_duration,omitempty"`
	PromptEvalCount int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount       int       `json:"eval_count,omitempty"`
	EvalDuration    int64     `json:"eval_duration,omitempty"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Format   string    `json:"format,omitempty"`
	KeepAlive string   `json:"keep_alive,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type ChatResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
	TotalDuration   int64     `json:"total_duration,omitempty"`
	LoadDuration    int64     `json:"load_duration,omitempty"`
	PromptEvalCount int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount       int       `json:"eval_count,omitempty"`
	EvalDuration    int64     `json:"eval_duration,omitempty"`
}

type ListModelsResponse struct {
	Models []Model `json:"models"`
}

type Model struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    ModelDetails `json:"details"`
}

type ModelDetails struct {
	Format           string   `json:"format"`
	Family           string   `json:"family"`
	Families         []string `json:"families"`
	ParameterSize    string   `json:"parameter_size"`
	QuantizationLevel string  `json:"quantization_level"`
}

type ShowModelRequest struct {
	Model string `json:"model"`
	Verbose bool `json:"verbose,omitempty"`
}

type ShowModelResponse struct {
	Modelfile  string                 `json:"modelfile"`
	Parameters string                 `json:"parameters"`
	Template   string                 `json:"template"`
	Details    ModelDetails           `json:"details"`
	ModelInfo  map[string]interface{} `json:"model_info"`
}

type CopyModelRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type DeleteModelRequest struct {
	Model string `json:"model"`
}

type PullModelRequest struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream"`
}

type PushModelRequest struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream"`
}

type ProgressResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type CreateModelRequest struct {
	Model     string `json:"model"`
	Modelfile string `json:"modelfile,omitempty"`
	Stream    bool   `json:"stream"`
	Path      string `json:"path,omitempty"`
	Quantize  string `json:"quantize,omitempty"`
}

type EmbeddingsRequest struct {
	Model  string   `json:"model"`
	Input  []string `json:"input"`
	Truncate bool    `json:"truncate,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
	KeepAlive string `json:"keep_alive,omitempty"`
}

type EmbeddingsResponse struct {
	Model       string      `json:"model"`
	Embeddings  [][]float64 `json:"embeddings"`
	TotalDuration int64     `json:"total_duration,omitempty"`
	LoadDuration  int64     `json:"load_duration,omitempty"`
	PromptEvalCount int     `json:"prompt_eval_count,omitempty"`
}

type ListRunningModelsResponse struct {
	Models []RunningModel `json:"models"`
}

type RunningModel struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
	ExpiresAt  time.Time    `json:"expires_at"`
	SizeVRAM   int64        `json:"size_vram"`
}
