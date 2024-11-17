package openai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jcserv/portfolio-api/internal/utils/log"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

type Client struct {
	client     *openai.Client
	rateLimits openai.RateLimitHeaders
	mu         sync.RWMutex
	maxRetries int
}

func NewClient(apiKey string) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
		rateLimits: openai.RateLimitHeaders{
			LimitRequests:     60,     // Default
			LimitTokens:       150000, // Default
			RemainingRequests: 60,
			RemainingTokens:   150000,
		},
		maxRetries: 5,
	}
}

func (c *Client) updateRateLimits(headers openai.RateLimitHeaders) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if headers.LimitRequests == 0 && headers.LimitTokens == 0 && headers.RemainingRequests == 0 && headers.RemainingTokens == 0 {
		return
	}

	c.rateLimits = headers
}

func (c *Client) waitForCapacity(ctx context.Context, tokensNeeded int) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.rateLimits.RemainingRequests <= 0 || c.rateLimits.RemainingTokens < tokensNeeded {
		var waitTime time.Time

		// If we're out of requests, wait until request reset
		if c.rateLimits.RemainingRequests <= 0 {
			requestReset := c.rateLimits.ResetRequests.Time()
			waitTime = requestReset
		}

		// If we're out of tokens, wait until token reset
		if c.rateLimits.RemainingTokens < tokensNeeded {
			tokenReset := c.rateLimits.ResetTokens.Time()
			if waitTime.IsZero() || tokenReset.After(waitTime) {
				waitTime = tokenReset
			}
		}

		// Add a small buffer to ensure the reset has occurred
		waitTime = waitTime.Add(100 * time.Millisecond)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Until(waitTime)):
			return nil
		}
	}
	return nil
}

func (c *Client) CreateEmbedding(ctx context.Context, request openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	var response openai.EmbeddingResponse

	// Estimate tokens needed (rough estimation)
	tokensNeeded := 0
	if texts, ok := request.Input.([]string); ok {
		for _, text := range texts {
			tokensNeeded += len(text) / 4 // Rough estimate: 4 chars per token
		}
	}
	log.Info(ctx, fmt.Sprintf("tokens needed: %d", tokensNeeded))

	operation := func() error {
		if err := c.waitForCapacity(ctx, tokensNeeded); err != nil {
			return backoff.Permanent(err)
		}
		log.Info(ctx, "creating embeddings")
		log.Info(ctx, fmt.Sprintf("rate limits: %v", c.rateLimits))
		resp, err := c.client.CreateEmbeddings(ctx, request)
		if err != nil {
			if apiErr, ok := err.(*openai.APIError); ok {
				headers := resp.GetRateLimitHeaders()
				log.Info(ctx, fmt.Sprintf("apiErr: %v", apiErr))
				c.updateRateLimits(headers)
				if apiErr.HTTPStatusCode == 429 {
					return err
				}
				return backoff.Permanent(err)
			}
			return backoff.Permanent(err)
		}

		// Update rate limits from successful response
		headers := resp.GetRateLimitHeaders()
		log.Info(ctx, fmt.Sprintf("headers: %v", headers))
		c.updateRateLimits(headers)
		response = resp
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 2 * time.Second
	b.MaxInterval = 60 * time.Second
	b.MaxElapsedTime = time.Duration(c.maxRetries) * 60 * time.Second

	err := backoff.Retry(operation, b)
	if err != nil {
		return openai.EmbeddingResponse{}, errors.Wrap(err, "failed to create embedding")
	}

	return response, nil
}

func (c *Client) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	var response openai.ChatCompletionResponse

	// Estimate tokens needed from messages
	tokensNeeded := 0
	for _, msg := range request.Messages {
		tokensNeeded += len(msg.Content) / 4
	}

	if request.MaxTokens > 0 {
		tokensNeeded += request.MaxTokens
	} else {
		tokensNeeded *= 2 // Default buffer if max_tokens not specified
	}

	operation := func() error {
		if err := c.waitForCapacity(ctx, tokensNeeded); err != nil {
			return backoff.Permanent(err)
		}

		log.Info(ctx, "prompting for chat completion")
		resp, err := c.client.CreateChatCompletion(ctx, request)
		if err != nil {
			if apiErr, ok := err.(*openai.APIError); ok {
				headers := resp.GetRateLimitHeaders()
				c.updateRateLimits(headers)
				if apiErr.HTTPStatusCode == 429 {
					return err
				}
				return backoff.Permanent(err)
			}
			return backoff.Permanent(err)
		}

		headers := resp.GetRateLimitHeaders()
		c.updateRateLimits(headers)
		response = resp
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 2 * time.Second
	b.MaxInterval = 60 * time.Second
	b.MaxElapsedTime = time.Duration(c.maxRetries) * 60 * time.Second

	err := backoff.Retry(operation, b)
	if err != nil {
		return openai.ChatCompletionResponse{}, errors.Wrap(err, "failed to create chat completion")
	}

	return response, nil
}
