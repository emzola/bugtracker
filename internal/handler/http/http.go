package http

type ServiceLayer interface {
	projectService
	userService
}

// Handler defines the app's HTTP handler.
type Handler struct {
	service ServiceLayer
}

// New creates a new HTTP handler.
func New(service ServiceLayer) *Handler {
	return &Handler{service}
}
