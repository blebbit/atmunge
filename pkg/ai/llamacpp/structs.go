package llamacpp

import "encoding/json"

// HealthCheckResponse represents the response from the /health endpoint.
type HealthCheckResponse struct {
	Status string    `json:"status"`
	Error  *APIError `json:"error,omitempty"`
}

// APIError represents an error response from the API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// CompletionRequest represents the request to the /completion endpoint.
type CompletionRequest struct {
	Prompt              interface{}     `json:"prompt"`
	Temperature         float64         `json:"temperature,omitempty"`
	DynatempRange       float64         `json:"dynatemp_range,omitempty"`
	DynatempExponent    float64         `json:"dynatemp_exponent,omitempty"`
	TopK                int             `json:"top_k,omitempty"`
	TopP                float64         `json:"top_p,omitempty"`
	MinP                float64         `json:"min_p,omitempty"`
	NPredict            int             `json:"n_predict,omitempty"`
	NIndent             int             `json:"n_indent,omitempty"`
	NKeep               int             `json:"n_keep,omitempty"`
	Stream              bool            `json:"stream,omitempty"`
	Stop                []string        `json:"stop,omitempty"`
	TypicalP            float64         `json:"typical_p,omitempty"`
	RepeatPenalty       float64         `json:"repeat_penalty,omitempty"`
	RepeatLastN         int             `json:"repeat_last_n,omitempty"`
	PresencePenalty     float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty    float64         `json:"frequency_penalty,omitempty"`
	DryMultiplier       float64         `json:"dry_multiplier,omitempty"`
	DryBase             float64         `json:"dry_base,omitempty"`
	DryAllowedLength    int             `json:"dry_allowed_length,omitempty"`
	DryPenaltyLastN     int             `json:"dry_penalty_last_n,omitempty"`
	DrySequenceBreakers []string        `json:"dry_sequence_breakers,omitempty"`
	XTCProbability      float64         `json:"xtc_probability,omitempty"`
	XTCThreshold        float64         `json:"xtc_threshold,omitempty"`
	Mirostat            int             `json:"mirostat,omitempty"`
	MirostatTau         float64         `json:"mirostat_tau,omitempty"`
	MirostatEta         float64         `json:"mirostat_eta,omitempty"`
	Grammar             string          `json:"grammar,omitempty"`
	JSONSchema          string          `json:"json_schema,omitempty"`
	Seed                int             `json:"seed,omitempty"`
	IgnoreEOS           bool            `json:"ignore_eos,omitempty"`
	LogitBias           [][]interface{} `json:"logit_bias,omitempty"`
	NProbs              int             `json:"n_probs,omitempty"`
	MinKeep             int             `json:"min_keep,omitempty"`
	TMaxPredictMs       int             `json:"t_max_predict_ms,omitempty"`
	ImageData           []ImageData     `json:"image_data,omitempty"`
	IDSlot              int             `json:"id_slot,omitempty"`
	CachePrompt         bool            `json:"cache_prompt,omitempty"`
	ReturnTokens        bool            `json:"return_tokens,omitempty"`
	Samplers            []string        `json:"samplers,omitempty"`
	TimingsPerToken     bool            `json:"timings_per_token,omitempty"`
	PostSamplingProbs   bool            `json:"post_sampling_probs,omitempty"`
	ResponseFields      []string        `json:"response_fields,omitempty"`
	LoRA                []LoRAAdapter   `json:"lora,omitempty"`
}

// ImageData represents image data for multimodal models.
type ImageData struct {
	Data string `json:"data"`
	ID   int    `json:"id"`
}

// LoRAAdapter represents a LoRA adapter configuration.
type LoRAAdapter struct {
	ID    int     `json:"id"`
	Scale float64 `json:"scale"`
}

// CompletionResponse represents the response from the /completion endpoint.
type CompletionResponse struct {
	Content                 string                  `json:"content"`
	Tokens                  []int                   `json:"tokens,omitempty"`
	Stop                    bool                    `json:"stop"`
	GenerationSettings      *GenerationSettings     `json:"generation_settings,omitempty"`
	Model                   string                  `json:"model,omitempty"`
	Prompt                  string                  `json:"prompt,omitempty"`
	StopType                string                  `json:"stop_type,omitempty"`
	StoppingWord            string                  `json:"stopping_word,omitempty"`
	Timings                 *Timings                `json:"timings,omitempty"`
	TokensCached            int                     `json:"tokens_cached,omitempty"`
	TokensEvaluated         int                     `json:"tokens_evaluated,omitempty"`
	Truncated               bool                    `json:"truncated,omitempty"`
	CompletionProbabilities []CompletionProbability `json:"completion_probabilities,omitempty"`
}

// CompletionProbability represents the token probabilities in the completion response.
type CompletionProbability struct {
	ID          int          `json:"id"`
	Logprob     float64      `json:"logprob"`
	Token       string       `json:"token"`
	Bytes       []byte       `json:"bytes"`
	TopLogprobs []TopLogprob `json:"top_logprobs"`
	Prob        float64      `json:"prob,omitempty"`
	TopProbs    []TopProb    `json:"top_probs,omitempty"`
}

// TopLogprob represents the top log probabilities for a token.
type TopLogprob struct {
	ID      int     `json:"id"`
	Logprob float64 `json:"logprob"`
	Token   string  `json:"token"`
	Bytes   []byte  `json:"bytes"`
}

// TopProb represents the top probabilities for a token.
type TopProb struct {
	ID    int     `json:"id"`
	Token string  `json:"token"`
	Bytes []byte  `json:"bytes"`
	Prob  float64 `json:"prob"`
}

// GenerationSettings represents the generation settings in the completion response.
type GenerationSettings struct {
	ID           int        `json:"id"`
	IDTask       int        `json:"id_task"`
	NCtx         int        `json:"n_ctx"`
	Speculative  bool       `json:"speculative"`
	IsProcessing bool       `json:"is_processing"`
	Params       *Params    `json:"params"`
	Prompt       string     `json:"prompt"`
	NextToken    *NextToken `json:"next_token"`
}

// Params represents the parameters in the generation settings.
type Params struct {
	NPredict            int      `json:"n_predict"`
	Seed                uint32   `json:"seed"`
	Temperature         float64  `json:"temperature"`
	DynatempRange       float64  `json:"dynatemp_range"`
	DynatempExponent    float64  `json:"dynatemp_exponent"`
	TopK                int      `json:"top_k"`
	TopP                float64  `json:"top_p"`
	MinP                float64  `json:"min_p"`
	XTCProbability      float64  `json:"xtc_probability"`
	XTCThreshold        float64  `json:"xtc_threshold"`
	TypicalP            float64  `json:"typical_p"`
	RepeatLastN         int      `json:"repeat_last_n"`
	RepeatPenalty       float64  `json:"repeat_penalty"`
	PresencePenalty     float64  `json:"presence_penalty"`
	FrequencyPenalty    float64  `json:"frequency_penalty"`
	DryMultiplier       float64  `json:"dry_multiplier"`
	DryBase             float64  `json:"dry_base"`
	DryAllowedLength    int      `json:"dry_allowed_length"`
	DryPenaltyLastN     int      `json:"dry_penalty_last_n"`
	DrySequenceBreakers []string `json:"dry_sequence_breakers"`
	Mirostat            int      `json:"mirostat"`
	MirostatTau         float64  `json:"mirostat_tau"`
	MirostatEta         float64  `json:"mirostat_eta"`
	Stop                []string `json:"stop"`
	MaxTokens           int      `json:"max_tokens"`
	NKeep               int      `json:"n_keep"`
	NDiscard            int      `json:"n_discard"`
	IgnoreEOS           bool     `json:"ignore_eos"`
	Stream              bool     `json:"stream"`
	NProbs              int      `json:"n_probs"`
	MinKeep             int      `json:"min_keep"`
	Grammar             string   `json:"grammar"`
	Samplers            []string `json:"samplers"`
	SpeculativeNMax     int      `json:"speculative.n_max"`
	SpeculativeNMin     int      `json:"speculative.n_min"`
	SpeculativePMin     float64  `json:"speculative.p_min"`
	TimingsPerToken     bool     `json:"timings_per_token"`
}

// NextToken represents the next token information in the generation settings.
type NextToken struct {
	HasNextToken bool   `json:"has_next_token"`
	HasNewLine   bool   `json:"has_new_line"`
	NRemain      int    `json:"n_remain"`
	NDecoded     int    `json:"n_decoded"`
	StoppingWord string `json:"stopping_word"`
}

// Timings represents the timing information in the completion response.
type Timings struct {
	PredictedPerSecond float64 `json:"predicted_per_second"`
}

// TokenizeRequest represents the request to the /tokenize endpoint.
type TokenizeRequest struct {
	Content      string `json:"content"`
	AddSpecial   bool   `json:"add_special,omitempty"`
	ParseSpecial bool   `json:"parse_special,omitempty"`
	WithPieces   bool   `json:"with_pieces,omitempty"`
}

// TokenizeResponse represents the response from the /tokenize endpoint.
type TokenizeResponse struct {
	Tokens []json.RawMessage `json:"tokens"`
}

// DetokenizeRequest represents the request to the /detokenize endpoint.
type DetokenizeRequest struct {
	Tokens []int `json:"tokens"`
}

// DetokenizeResponse represents the response from the /detokenize endpoint.
type DetokenizeResponse struct {
	Content string `json:"content"`
}

// ApplyTemplateRequest represents the request to the /apply-template endpoint.
type ApplyTemplateRequest struct {
	Messages []ChatMessage `json:"messages"`
}

// ApplyTemplateResponse represents the response from the /apply-template endpoint.
type ApplyTemplateResponse struct {
	Prompt string `json:"prompt"`
}

// EmbeddingRequest represents the request to the /embedding endpoint.
type EmbeddingRequest struct {
	Content       string      `json:"content"`
	ImageData     []ImageData `json:"image_data,omitempty"`
	EmbdNormalize int         `json:"embd_normalize,omitempty"`
}

// EmbeddingResponse represents the response from the /embedding endpoint.
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// RerankingRequest represents the request to the /rerank endpoint.
type RerankingRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n,omitempty"`
	Model     string   `json:"model,omitempty"`
}

// RerankingResponse represents the response from the /rerank endpoint.
type RerankingResponse struct {
	Results []RerankResult `json:"results"`
}

// RerankResult represents a single rerank result.
type RerankResult struct {
	Document       string  `json:"document"`
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

// InfillRequest represents the request to the /infill endpoint.
type InfillRequest struct {
	InputPrefix string       `json:"input_prefix"`
	InputSuffix string       `json:"input_suffix"`
	InputExtra  []InputExtra `json:"input_extra,omitempty"`
	Prompt      string       `json:"prompt,omitempty"`
	CompletionRequest
}

// InputExtra represents the extra input for the infill endpoint.
type InputExtra struct {
	Filename string `json:"filename"`
	Text     string `json:"text"`
}

// PropsResponse represents the response from the /props endpoint.
type PropsResponse struct {
	DefaultGenerationSettings *GenerationSettings `json:"default_generation_settings"`
	TotalSlots                int                 `json:"total_slots"`
	ModelPath                 string              `json:"model_path"`
	ChatTemplate              string              `json:"chat_template"`
	Modalities                *Modalities         `json:"modalities"`
	BuildInfo                 string              `json:"build_info"`
}

// Modalities represents the modalities in the props response.
type Modalities struct {
	Vision bool `json:"vision"`
}

// PostPropsRequest represents the request to the POST /props endpoint.
// The documentation currently states there are no options.
type PostPropsRequest struct{}

// EmbeddingsResponse represents the response from the non-OAI /embeddings endpoint.
type EmbeddingsResponse []struct {
	Index     int         `json:"index"`
	Embedding [][]float64 `json:"embedding"`
}

// SlotsResponse represents the response from the /slots endpoint.
type SlotsResponse []GenerationSettings

// SlotSaveRequest represents the request to save a slot.
type SlotSaveRequest struct {
	Filename string `json:"filename"`
}

// SlotSaveResponse represents the response from saving a slot.
type SlotSaveResponse struct {
	IDSlot   int    `json:"id_slot"`
	Filename string `json:"filename"`
	NSaved   int    `json:"n_saved"`
	NWritten int    `json:"n_written"`
	Timings  struct {
		SaveMs float64 `json:"save_ms"`
	} `json:"timings"`
}

// SlotRestoreRequest represents the request to restore a slot.
type SlotRestoreRequest struct {
	Filename string `json:"filename"`
}

// SlotRestoreResponse represents the response from restoring a slot.
type SlotRestoreResponse struct {
	IDSlot    int    `json:"id_slot"`
	Filename  string `json:"filename"`
	NRestored int    `json:"n_restored"`
	NRead     int    `json:"n_read"`
	Timings   struct {
		RestoreMs float64 `json:"restore_ms"`
	} `json:"timings"`
}

// SlotEraseResponse represents the response from erasing a slot.
type SlotEraseResponse struct {
	IDSlot  int `json:"id_slot"`
	NErased int `json:"n_erased"`
}

// GetLoRAAdaptersResponse represents the response from GET /lora-adapters.
type GetLoRAAdaptersResponse []LoRAAdapterInfo

// LoRAAdapterInfo represents information about a LoRA adapter.
type LoRAAdapterInfo struct {
	ID    int     `json:"id"`
	Path  string  `json:"path"`
	Scale float64 `json:"scale"`
}

// PostLoRAAdaptersRequest represents the request for POST /lora-adapters.
type PostLoRAAdaptersRequest []LoRAAdapter

// OpenAIModelsResponse represents the response from the /v1/models endpoint.
type OpenAIModelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// ModelInfo represents information about a model.
type ModelInfo struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Created int64      `json:"created"`
	OwnedBy string     `json:"owned_by"`
	Meta    *ModelMeta `json:"meta,omitempty"`
}

// ModelMeta represents metadata about a model.
type ModelMeta struct {
	VocabType int   `json:"vocab_type"`
	NVocab    int   `json:"n_vocab"`
	NCtxTrain int   `json:"n_ctx_train"`
	NEmbd     int   `json:"n_embd"`
	NParams   int64 `json:"n_params"`
	Size      int64 `json:"size"`
}

// OpenAICompletionsRequest represents the request to the /v1/completions endpoint.
type OpenAICompletionsRequest struct {
	Model            string          `json:"model"`
	Prompt           interface{}     `json:"prompt"`
	MaxTokens        int             `json:"max_tokens,omitempty"`
	Temperature      float64         `json:"temperature,omitempty"`
	TopP             float64         `json:"top_p,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	Stop             []string        `json:"stop,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	LogitBias        [][]interface{} `json:"logit_bias,omitempty"`
	Seed             int             `json:"seed,omitempty"`
	Mirostat         int             `json:"mirostat,omitempty"`
	MirostatTau      float64         `json:"mirostat_tau,omitempty"`
	MirostatEta      float64         `json:"mirostat_eta,omitempty"`
	Grammar          string          `json:"grammar,omitempty"`
	JSONSchema       string          `json:"json_schema,omitempty"`
	NProbs           int             `json:"n_probs,omitempty"`
	ImageData        []ImageData     `json:"image_data,omitempty"`
	LoRA             []LoRAAdapter   `json:"lora,omitempty"`
}

// OpenAICompletionsResponse represents the response from the /v1/completions endpoint.
type OpenAICompletionsResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// Choice represents a single choice in the completions response.
type Choice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

// OpenAIChatCompletionsRequest represents the request to the /v1/chat/completions endpoint.
type OpenAIChatCompletionsRequest struct {
	Model              string                 `json:"model"`
	Messages           []ChatMessage          `json:"messages"`
	Temperature        float64                `json:"temperature,omitempty"`
	TopP               float64                `json:"top_p,omitempty"`
	Stream             bool                   `json:"stream,omitempty"`
	Stop               []string               `json:"stop,omitempty"`
	MaxTokens          int                    `json:"max_tokens,omitempty"`
	PresencePenalty    float64                `json:"presence_penalty,omitempty"`
	FrequencyPenalty   float64                `json:"frequency_penalty,omitempty"`
	LogitBias          [][]interface{}        `json:"logit_bias,omitempty"`
	Seed               int                    `json:"seed,omitempty"`
	ResponseFormat     *ResponseFormat        `json:"response_format,omitempty"`
	Tools              interface{}            `json:"tools,omitempty"`
	ToolChoice         interface{}            `json:"tool_choice,omitempty"`
	ChatTemplateKwargs map[string]interface{} `json:"chat_template_kwargs,omitempty"`
	ReasoningFormat    string                 `json:"reasoning_format,omitempty"`
	ThinkingForcedOpen bool                   `json:"thinking_forced_open,omitempty"`
	ParseToolCalls     bool                   `json:"parse_tool_calls,omitempty"`
	Mirostat           int                    `json:"mirostat,omitempty"`
	MirostatTau        float64                `json:"mirostat_tau,omitempty"`
	MirostatEta        float64                `json:"mirostat_eta,omitempty"`
	Grammar            string                 `json:"grammar,omitempty"`
	JSONSchema         string                 `json:"json_schema,omitempty"`
	NProbs             int                    `json:"n_probs,omitempty"`
	ImageData          []ImageData            `json:"image_data,omitempty"`
	LoRA               []LoRAAdapter          `json:"lora,omitempty"`
}

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat represents the response format for chat completions.
type ResponseFormat struct {
	Type   string      `json:"type"`
	Schema interface{} `json:"schema,omitempty"`
}

// OpenAIChatCompletionsResponse represents the response from the /v1/chat/completions endpoint.
type OpenAIChatCompletionsResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
}

// ChatChoice represents a single choice in the chat completions response.
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// OpenAIEmbeddingsRequest represents the request to the /v1/embeddings endpoint.
type OpenAIEmbeddingsRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
}

// OpenAIEmbeddingsResponse represents the response from the /v1/embeddings endpoint.
type OpenAIEmbeddingsResponse struct {
	Object string            `json:"object"`
	Data   []EmbeddingObject `json:"data"`
	Model  string            `json:"model"`
}

// EmbeddingObject represents a single embedding object.
type EmbeddingObject struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}
