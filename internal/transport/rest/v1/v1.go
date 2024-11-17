package v1

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jcserv/portfolio-api/internal/rag"
)

const (
	APIV1URLPath = "/api/v1/"
)

type API struct {
	ragService *rag.Service
}

func NewAPI(ragService *rag.Service) *API {
	return &API{ragService: ragService}
}

func (a *API) RegisterRoutes(r *mux.Router) {
	r.HandleFunc(APIV1URLPath+"ask", a.Ask()).Methods(http.MethodPost)
}
