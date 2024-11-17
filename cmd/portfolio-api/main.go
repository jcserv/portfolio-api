package main

import (
	"context"

	"github.com/jcserv/portfolio-api/internal"
	"github.com/jcserv/portfolio-api/internal/utils/log"
	"go.uber.org/zap"
)

func main() {
	logger := log.GetLogger(context.Background())
	defer logger.Sync()
	log.Info(context.Background(), "starting service")

	service, err := internal.NewService()
	if err != nil {
		logger.Fatal("could not create service", zap.Error(err))
	}

	if err := service.Run(); err != nil {
		logger.Fatal("could not start service", zap.Error(err))
	}
}
