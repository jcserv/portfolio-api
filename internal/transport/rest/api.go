package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jcserv/portfolio-api/internal/rag"
	"github.com/jcserv/portfolio-api/internal/transport/rest/httputil"
	v1 "github.com/jcserv/portfolio-api/internal/transport/rest/v1"
)

const (
	HealthCheck = "/health/system"
)

type API struct {
	V1API *v1.API
}

func NewAPI(ragService *rag.Service) *API {
	return &API{
		V1API: v1.NewAPI(ragService),
	}
}

func (a *API) RegisterRoutes() *mux.Router {
	r := mux.NewRouter()
	r.Use(LogIncomingRequests())
	a.V1API.RegisterRoutes(r)
	r.HandleFunc(HealthCheck, a.HealthCheck()).Methods(http.MethodGet)
	return r
}

func (a *API) HealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httputil.OK(w, nil)
	}
}
