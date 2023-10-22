package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/emzola/bugtracker/pkg/model"
)

type projectService interface {
	CreateProject(ctx context.Context, name, description string, startDate, endDate time.Time, publicAccess bool, owner, createdBy string) (*model.Project, error)
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name         string    `json:"name"`
		Description  string    `json:"description,omitempty"`
		StartDate    time.Time `json:"start_date"`
		EndDate      time.Time `json:"end_date,omitempty"`
		PublicAccess bool      `json:"public_access"`
	}
	err := h.decodeJSON(w, r, &requestBody)
	if err != nil {
		h.badRequestResponse(w, r, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	project, err := h.service.CreateProject(ctx, requestBody.Name, requestBody.Description, requestBody.StartDate, requestBody.EndDate, requestBody.PublicAccess, "Zoho", "Emzo")
	if err != nil {
		// handle error
	}
	header := make(http.Header)
	header.Set("Location", fmt.Sprintf("/v1/projects/%d", project.ID))
	err = h.encodeJSON(w, http.StatusCreated, envelop{"project": project}, header)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
