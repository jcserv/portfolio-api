package rag

import (
	"context"
	"fmt"

	"github.com/jcserv/portfolio-api/internal/db"
	"github.com/jcserv/portfolio-api/internal/model"
	"github.com/jcserv/portfolio-api/internal/utils/log"
	"github.com/sashabaranov/go-openai"
)

type Service struct {
	db       *db.LibSQL
	embedder *Embedder
}

func NewService(db *db.LibSQL, embedder *Embedder) *Service {
	return &Service{
		db:       db,
		embedder: embedder,
	}
}

func (s *Service) IndexExperience(ctx context.Context, experiences []model.Experience) error {
	for _, exp := range experiences {
		workplace := exp.Workplace
		position := exp.Position
		description := exp.Description

		text := workplace + " - " + position + "\n"
		for _, desc := range description {
			text += "- " + desc + "\n"
		}

		exists, err := s.db.DoesEmbeddingExist(ctx, text)
		if err != nil {
			log.Error(ctx, fmt.Sprintf("unable to check if embedding exists: %v", err))
			continue
		}
		if exists {
			continue
		}

		embedding, err := s.embedder.GetEmbedding(ctx, text)
		if err != nil {
			return err
		}

		if err := s.db.StoreEmbedding(ctx, text, embedding, "experience"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Answer(ctx context.Context, question string) (string, error) {
	questionEmbedding, err := s.embedder.GetEmbedding(ctx, question)
	if err != nil {
		return "", err
	}

	relevant, err := s.db.FindSimilar(ctx, questionEmbedding, 3)
	if err != nil {
		return "", err
	}

	prompt := "Based on the following experiences, answer the question: " + question + "\n\n"
	for _, exp := range relevant {
		prompt += exp + "\n\n"
	}

	completion, err := s.embedder.OpenAIClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant answering questions about Jarrod's professional experience. Answer concisely and accurately based on the provided context.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}
