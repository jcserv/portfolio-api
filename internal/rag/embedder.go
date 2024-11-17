package rag

import (
	"context"

	oAI "github.com/jcserv/portfolio-api/internal/api/openai"
	"github.com/sashabaranov/go-openai"
)

type Embedder struct {
	OpenAIClient *oAI.Client
}

func NewEmbedder(OpenAIClient *oAI.Client) *Embedder {
	return &Embedder{
		OpenAIClient: OpenAIClient,
	}
}

func (e *Embedder) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := e.OpenAIClient.CreateEmbedding(ctx, openai.EmbeddingRequest{
		Model: openai.SmallEmbedding3,
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}
