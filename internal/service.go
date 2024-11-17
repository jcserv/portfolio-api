package internal

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/jcserv/portfolio-api/internal/api/openai"
	"github.com/jcserv/portfolio-api/internal/db"
	"github.com/jcserv/portfolio-api/internal/rag"
	"github.com/jcserv/portfolio-api/internal/transport/rest"
	"github.com/jcserv/portfolio-api/internal/utils"
	"github.com/jcserv/portfolio-api/internal/utils/log"
)

type Service struct {
	api *rest.API
	cfg *Configuration
}

func NewService() (*Service, error) {
	cfg, err := NewConfiguration()
	if err != nil {
		return nil, err
	}

	db, err := db.NewLibSQL(context.Background(), cfg.DBPath)
	if err != nil {
		return nil, err
	}

	openAIClient := openai.NewClient(cfg.OpenAIKey)

	ragEmbedder := rag.NewEmbedder(openAIClient)
	ragService := rag.NewService(db, ragEmbedder)

	s := &Service{
		api: rest.NewAPI(ragService),
		cfg: cfg,
	}

	err = s.Init(ragService)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) Init(ragService *rag.Service) error {
	exp, err := utils.ReadExperience()
	if err != nil {
		log.Error(context.Background(), fmt.Sprintf("unable to read experience: %v", err))
		return err
	}

	err = ragService.IndexExperience(context.Background(), exp)
	if err != nil {
		log.Error(context.Background(), fmt.Sprintf("unable to index experience: %v", err))
		return err
	}

	projs, err := utils.ReadProjects()
	if err != nil {
		log.Error(context.Background(), fmt.Sprintf("unable to read projects: %v", err))
		return err
	}

	err = ragService.IndexProjects(context.Background(), projs)
	if err != nil {
		log.Error(context.Background(), fmt.Sprintf("unable to index projects: %v", err))
		return err
	}
	return nil
}

func (s *Service) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		s.StartHTTP(ctx)
	}(ctx)

	wg.Wait()
	return nil
}

func (s *Service) StartHTTP(ctx context.Context) error {
	log.Info(ctx, fmt.Sprintf("starting http server on port %s", s.cfg.HTTPPort))
	r := s.api.RegisterRoutes()
	http.ListenAndServe(fmt.Sprintf(":%s", s.cfg.HTTPPort), r)
	return nil
}
