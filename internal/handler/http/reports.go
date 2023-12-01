package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/emzola/issuetracker/internal/model"
	"github.com/emzola/issuetracker/pkg/validator"
)

type issuesReportService interface {
	GetIssuesStatusReport(ctx context.Context, projectID int64) ([]*model.IssuesStatus, error)
}

func (h *Handler) getIssuesStatusReport(w http.ResponseWriter, r *http.Request) {
	var requestQuery struct {
		ProjectID int64
	}
	v := validator.New()
	qs := r.URL.Query()
	requestQuery.ProjectID = int64(h.readInt(qs, "project_id", 0, v))
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	statuses, err := h.service.GetIssuesStatusReport(ctx, requestQuery.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return
		default:
			h.serverErrorResponse(w, r, err)
		}
		return
	}
	err = h.encodeJSON(w, http.StatusOK, envelop{"report": statuses}, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
