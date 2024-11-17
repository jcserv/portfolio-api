package v1

import (
	"encoding/json"
	"net/http"

	"github.com/jcserv/portfolio-api/internal/transport/rest/httputil"
	"github.com/jcserv/portfolio-api/internal/utils/log"
)

type AskRequest struct {
	Question string `json:"question"`
}

type AskResponse struct {
	Answer string `json:"answer"`
}

func (a *API) Ask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Info(ctx, "received request")

		var req AskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Info(ctx, "unable to decode request body")
			httputil.BadRequest(w)
			return
		}

		if req.Question == "" {
			log.Info(ctx, "empty question")
			httputil.BadRequest(w)
			return
		}

		answer, err := a.ragService.Answer(ctx, req.Question)
		if err != nil {
			httputil.InternalServerError(ctx, w, err)
			return
		}

		httputil.OK(w, AskResponse{Answer: answer})
	}
}
