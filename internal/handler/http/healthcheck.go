package http

import "net/http"

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	data := envelop{
		"status": "available",
		"system_info": map[string]string{
			"environment": h.Config.Env,
		},
	}
	err := h.encodeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		h.serverErrorResponse(w, r, err)
	}
}
