package memory

import (
	"os"
	"sync"

	tfidf "github.com/rioloc/tfidf-go"
	"github.com/rioloc/tfidf-go/token"
)

type EmbeddingModel interface {
	Encode(texts []string) [][]float64
	Dimension() int
}

type embeddingConfig struct {
	ModelType string
	ModelName string
	APIKey    string
	BaseURL   string
}

func getEmbeddingConfig() embeddingConfig {
	return embeddingConfig{
		ModelType: getEnvOrDefault("EMBED_MODEL_TYPE", "dashscope"),
		ModelName: getEnvOrDefault("EMBED_MODEL_NAME", "text-embedding-v3"),
		APIKey:    os.Getenv("EMBED_API_KEY"),
		BaseURL:   os.Getenv("EMBED_BASE_URL"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	embedder     EmbeddingModel
	embedderOnce sync.Once
	embedderLock sync.Mutex
)

func GetTextEmbedder() EmbeddingModel {
	embedderLock.Lock()
	defer embedderLock.Unlock()

	if embedder != nil {
		return embedder
	}

	embedderOnce.Do(func() {
		cfg := getEmbeddingConfig()
		embedder = createEmbeddingModel(cfg.ModelType, cfg.ModelName, cfg.APIKey, cfg.BaseURL)
	})

	return embedder
}

func createEmbeddingModel(modelType, modelName, apiKey, baseURL string) EmbeddingModel {
	switch modelType {
	case "dashscope":
		return &dashscopeEmbedder{modelName: modelName, apiKey: apiKey, baseURL: baseURL}
	case "local":
		return &localEmbedder{modelName: modelName}
	case "tfidf":
		return &tfidfEmbedder{maxFeatures: 1000}
	default:
		return &dashscopeEmbedder{modelName: modelName, apiKey: apiKey, baseURL: baseURL}
	}
}

type dashscopeEmbedder struct {
	modelName string
	apiKey    string
	baseURL   string
	dimension int
}

func (e *dashscopeEmbedder) Encode(texts []string) [][]float64 {
	return make([][]float64, len(texts))
}

func (e *dashscopeEmbedder) Dimension() int {
	return 384
}

type localEmbedder struct {
	modelName string
	dimension int
}

func (e *localEmbedder) Encode(texts []string) [][]float64 {
	return make([][]float64, len(texts))
}

func (e *localEmbedder) Dimension() int {
	return 384
}

type tfidfEmbedder struct {
	maxFeatures int
	dimension   int
	vocabulary  []string
	tfIDFMatrix [][]float64
	vectorizer  *tfidf.TfIdfVectorizer
	documents   []string
	mu          sync.Mutex
}

func (e *tfidfEmbedder) Fit(documents []string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(documents) == 0 {
		return
	}

	e.documents = documents
	e.vectorizer = tfidf.NewTfIdfVectorizer()

	tokenizer := token.NewTokenizer()
	vocabulary, tokens, _ := tokenizer.Tokenize(documents)

	e.vocabulary = vocabulary
	if e.maxFeatures > 0 && len(vocabulary) > e.maxFeatures {
		e.vocabulary = vocabulary[:e.maxFeatures]
	}
	e.dimension = len(e.vocabulary)

	tfMatrix := tfidf.Tf(e.vocabulary, tokens)
	idfVector := tfidf.Idf(e.vocabulary, tokens, true)

	var err error
	e.tfIDFMatrix, err = e.vectorizer.TfIdf(tfMatrix, idfVector)
	if err != nil {
		e.tfIDFMatrix = make([][]float64, len(documents))
	}
}

func (e *tfidfEmbedder) Encode(texts []string) [][]float64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.vectorizer == nil || len(e.tfIDFMatrix) == 0 {
		return make([][]float64, len(texts))
	}

	tokenizer := token.NewTokenizer()
	_, tokens, _ := tokenizer.Tokenize(texts)

	tfMatrix := tfidf.Tf(e.vocabulary, tokens)
	result, err := e.vectorizer.TfIdf(tfMatrix, tfidf.Idf(e.vocabulary, tokens, true))
	if err != nil {
		return make([][]float64, len(texts))
	}

	for i := range result {
		if len(result[i]) > e.dimension {
			result[i] = result[i][:e.dimension]
		}
	}
	return result
}

func (e *tfidfEmbedder) Dimension() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.dimension
}

func GetEmbeddingDimension() int {
	return 384
}

func RefreshEmbedder() EmbeddingModel {
	embedderLock.Lock()
	defer embedderLock.Unlock()

	cfg := getEmbeddingConfig()
	embedder = createEmbeddingModel(cfg.ModelType, cfg.ModelName, cfg.APIKey, cfg.BaseURL)
	return embedder
}
