package v1

import (
	"encoding/json"
	"net/http"

	"github.com/jcserv/portfolio-api/internal/transport/rest/httputil"
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
