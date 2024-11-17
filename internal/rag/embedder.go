package rag

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Embedder struct {
	client *openai.Client
}

func NewEmbedder(apiKey string) *Embedder {
	return &Embedder{
		client: openai.NewClient(apiKey),
	}
}

func (e *Embedder) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}
