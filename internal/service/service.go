package service

type RepositoryLayer interface {
	ProjectRepository
}

// Controller defines a new project service controller.
type Service struct {
	repo RepositoryLayer
}

// New creates a project service controller.
func New(repo RepositoryLayer) *Service {
	return &Service{repo}
}
