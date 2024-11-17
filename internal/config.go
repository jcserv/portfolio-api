package internal

import (
	"fmt"

	"github.com/jcserv/portfolio-api/internal/utils/env"
)

type Configuration struct {
	Region      string
	Environment string
	HTTPPort    string
	DBPath      string
	OpenAIKey   string
}

func NewConfiguration() (*Configuration, error) {
	cfg := &Configuration{}
	cfg.Region = env.GetString("REGION", "us-east-1")
	cfg.Environment = env.GetString("ENVIRONMENT", "prod")
	cfg.HTTPPort = env.GetString("HTTP_PORT", "8080")
	cfg.DBPath = env.GetString("DB_PATH", "./internal/db/portfolio-api.db")
	cfg.OpenAIKey = env.GetString("OPENAI_API_KEY", "")

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Configuration) Validate() error {
	variables := []string{
		c.Region,
		c.Environment,
		c.HTTPPort,
		c.DBPath,
		c.OpenAIKey,
	}
	for i, v := range variables {
		if v == "" {
			return fmt.Errorf("missing required variable: %d", i)
		}
	}
	return nil
}
