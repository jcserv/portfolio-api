package v1

import (
	"encoding/json"
	"fmt"
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

		var req AskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error(ctx, fmt.Sprintf("unable to decode request body: %v", err))
			httputil.BadRequest(w)
			return
		}

		if req.Question == "" {
			httputil.BadRequest(w)
			return
		}

		answer, err := a.ragService.Answer(ctx, req.Question)
		if err != nil {
			log.Error(ctx, fmt.Sprintf("unable to answer question: %v, err: %v", req.Question, err))
			httputil.InternalServerError(ctx, w, err)
			return
		}
		log.Info(ctx, fmt.Sprintf("answered question: %s with answer: %s", req.Question, answer))
		httputil.OK(w, AskResponse{Answer: answer})
	}
}
