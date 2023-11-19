package service

type RepositoryLayer interface {
	projectRepository
	userRepository
}

// Controller defines a new project service controller.
type Service struct {
	repo RepositoryLayer
}

// New creates a project service controller.
func New(repo RepositoryLayer) *Service {
	return &Service{repo}
}
