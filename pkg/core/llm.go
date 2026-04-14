package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type Provider string

const (
	ProviderOpenAI     Provider = "openai"
	ProviderDeepSeek   Provider = "deepseek"
	ProviderQwen       Provider = "qwen"
	ProviderModelScope Provider = "modelscope"
	ProviderKimi       Provider = "kimi"
	ProviderZhipu      Provider = "zhipu"
	ProviderOllama     Provider = "ollama"
	ProviderVLLM       Provider = "vllm"
	ProviderLocal      Provider = "local"
	ProviderAuto       Provider = "auto"
)

var providerBaseURLs = map[Provider]string{
	ProviderOpenAI:     "https://api.openai.com/v1",
	ProviderDeepSeek:   "https://api.deepseek.com",
	ProviderQwen:       "https://dashscope.aliyuncs.com/compatible-mode/v1",
	ProviderModelScope: "https://api-inference.modelscope.cn/v1/",
	ProviderKimi:       "https://api.moonshot.cn/v1",
	ProviderZhipu:      "https://open.bigmodel.cn/api/paas/v4",
	ProviderOllama:     "http://localhost:11434/v1",
	ProviderVLLM:       "http://localhost:8000/v1",
	ProviderLocal:      "http://localhost:8000/v1",
}

var providerDefaultModels = map[Provider]string{
	ProviderOpenAI:     "gpt-3.5-turbo",
	ProviderDeepSeek:   "deepseek-chat",
	ProviderQwen:       "qwen-plus",
	ProviderModelScope: "Qwen/Qwen2.5-72B-Instruct",
	ProviderKimi:       "moonshot-v1-8k",
	ProviderZhipu:      "glm-4",
	ProviderOllama:     "llama3.2",
	ProviderVLLM:       "meta-llama/Llama-2-7b-chat-hf",
	ProviderLocal:      "local-model",
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type LLM struct {
	Model       string
	APIKey      string
	BaseURL     string
	Provider    Provider
	Temperature float64
	MaxTokens   *int
	Timeout     time.Duration
	client      *openai.Client
	mu          sync.RWMutex
}

type LLMOption func(*LLM)

func WithModel(model string) LLMOption {
	return func(l *LLM) { l.Model = model }
}

func WithAPIKey(apiKey string) LLMOption {
	return func(l *LLM) { l.APIKey = apiKey }
}

func WithBaseURL(baseURL string) LLMOption {
	return func(l *LLM) { l.BaseURL = baseURL }
}

func WithProvider(provider Provider) LLMOption {
	return func(l *LLM) { l.Provider = provider }
}

func WithTemperature(temp float64) LLMOption {
	return func(l *LLM) { l.Temperature = temp }
}

func WithMaxTokens(tokens int) LLMOption {
	return func(l *LLM) { l.MaxTokens = &tokens }
}

func WithTimeout(timeout time.Duration) LLMOption {
	return func(l *LLM) { l.Timeout = timeout }
}

func NewLLM(opts ...LLMOption) *LLM {
	l := &LLM{
		Temperature: 0.7,
		Timeout:     60 * time.Second,
	}

	for _, opt := range opts {
		opt(l)
	}

	if l.Model == "" {
		l.Model = l.getDefaultModel()
	}

	if l.Provider == "" {
		l.Provider = l.autoDetectProvider()
	}

	if l.BaseURL == "" {
		l.BaseURL = l.resolveBaseURL()
	}

	if l.APIKey == "" {
		l.APIKey = l.resolveAPIKey()
	}

	l.client = l.createClient()
	return l
}

func (l *LLM) createClient() *openai.Client {
	config := openai.DefaultConfig(l.APIKey)
	config.BaseURL = l.BaseURL
	return openai.NewClientWithConfig(config)
}

func (l *LLM) autoDetectProvider() Provider {
	envProviders := map[string]Provider{
		"OPENAI_API_KEY":     ProviderOpenAI,
		"DEEPSEEK_API_KEY":   ProviderDeepSeek,
		"DASHSCOPE_API_KEY":  ProviderQwen,
		"MODELSCOPE_API_KEY": ProviderModelScope,
		"KIMI_API_KEY":       ProviderKimi,
		"MOONSHOT_API_KEY":   ProviderKimi,
		"ZHIPU_API_KEY":      ProviderZhipu,
		"GLM_API_KEY":        ProviderZhipu,
	}

	for env, provider := range envProviders {
		if os.Getenv(env) != "" {
			return provider
		}
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL != "" {
		lower := strings.ToLower(baseURL)
		switch {
		case strings.Contains(lower, "api.openai.com"):
			return ProviderOpenAI
		case strings.Contains(lower, "api.deepseek.com"):
			return ProviderDeepSeek
		case strings.Contains(lower, "dashscope.aliyuncs.com"):
			return ProviderQwen
		case strings.Contains(lower, "api-inference.modelscope.cn"):
			return ProviderModelScope
		case strings.Contains(lower, "api.moonshot.cn"):
			return ProviderKimi
		case strings.Contains(lower, "open.bigmodel.cn"):
			return ProviderZhipu
		case strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1"):
			if strings.Contains(lower, ":11434") || strings.Contains(lower, "ollama") {
				return ProviderOllama
			}
			if strings.Contains(lower, ":8000") && strings.Contains(lower, "vllm") {
				return ProviderVLLM
			}
			return ProviderLocal
		}
	}

	return ProviderOpenAI
}

func (l *LLM) resolveBaseURL() string {
	if baseURL := os.Getenv("LLM_BASE_URL"); baseURL != "" {
		return baseURL
	}

	if url, ok := providerBaseURLs[l.Provider]; ok {
		return url
	}

	return providerBaseURLs[ProviderOpenAI]
}

func (l *LLM) resolveAPIKey() string {
	envKeys := map[Provider]string{
		ProviderOpenAI:     "OPENAI_API_KEY",
		ProviderDeepSeek:   "DEEPSEEK_API_KEY",
		ProviderQwen:       "DASHSCOPE_API_KEY",
		ProviderModelScope: "MODELSCOPE_API_KEY",
		ProviderKimi:       "KIMI_API_KEY",
		ProviderZhipu:      "ZHIPU_API_KEY",
		ProviderOllama:     "OLLAMA_API_KEY",
		ProviderVLLM:       "VLLM_API_KEY",
	}

	if key, ok := envKeys[l.Provider]; ok {
		if apiKey := os.Getenv(key); apiKey != "" {
			return apiKey
		}
	}

	return os.Getenv("LLM_API_KEY")
}

func (l *LLM) getDefaultModel() string {
	if model := os.Getenv("LLM_MODEL_ID"); model != "" {
		return model
	}

	if model, ok := providerDefaultModels[l.Provider]; ok {
		return model
	}

	return providerDefaultModels[ProviderOpenAI]
}

func (l *LLM) SetModel(model string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Model = model
}

func (l *LLM) SetTemperature(temp float64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Temperature = temp
}

func (l *LLM) SetMaxTokens(tokens int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.MaxTokens = &tokens
}

func (l *LLM) buildMessages(msgs []ChatMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, len(msgs))
	for i, m := range msgs {
		result[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
			Name:    m.Name,
		}
	}
	return result
}

func (l *LLM) Invoke(ctx context.Context, messages []ChatMessage) (string, error) {
	l.mu.RLock()
	model := l.Model
	temperature := float32(l.Temperature)
	maxTokens := l.MaxTokens
	client := l.client
	l.mu.RUnlock()

	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    l.buildMessages(messages),
		Temperature: temperature,
		Stream:      false,
	}
	if maxTokens != nil {
		req.MaxTokens = *maxTokens
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

func (l *LLM) Stream(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error) {
	ch := make(chan string, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		l.mu.RLock()
		model := l.Model
		temperature := float32(l.Temperature)
		maxTokens := l.MaxTokens
		client := l.client
		l.mu.RUnlock()

		req := openai.ChatCompletionRequest{
			Model:       model,
			Messages:    l.buildMessages(messages),
			Temperature: temperature,
			Stream:      true,
		}
		if maxTokens != nil {
			req.MaxTokens = *maxTokens
		}

		stream, err := client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			errCh <- fmt.Errorf("OpenAI API error: %w", err)
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				errCh <- fmt.Errorf("stream error: %w", err)
				return
			}

			if len(resp.Choices) > 0 && resp.Choices[0].Delta.Content != "" {
				ch <- resp.Choices[0].Delta.Content
			}
		}
	}()

	return ch, errCh
}

func (l *LLM) Think(ctx context.Context, messages []ChatMessage) (string, error) {
	fmt.Printf("🤖 Calling %s model...\n", l.Model)

	var buf strings.Builder
	ch, errCh := l.Stream(ctx, messages)

	select {
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				goto done
			}
			fmt.Print(chunk)
			buf.WriteString(chunk)
		case err := <-errCh:
			return "", err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

done:
	fmt.Println()
	return buf.String(), nil
}

func (l *LLM) String() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return fmt.Sprintf("LLM(model=%s, provider=%s)", l.Model, l.Provider)
}
