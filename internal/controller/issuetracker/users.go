package issuetracker

import (
	"context"
	"errors"
	"time"

	"github.com/emzola/issuetracker/internal/repository"
	"github.com/emzola/issuetracker/pkg/model"

	"github.com/emzola/issuetracker/pkg/validator"
)

type userRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetAllUsers(ctx context.Context, name, email, role string, filters model.Filters) ([]*model.User, model.Metadata, error)
	CreateToken(ctx context.Context, userID int64, ttl time.Duration, scope string) (*model.Token, error)
	GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id int64) error
	AssignUserToProject(ctx context.Context, userID, projectID int64) error
	GetAllProjectsForUser(ctx context.Context, userID int64, filters model.Filters) ([]*model.Project, model.Metadata, error)
}

func (c *Controller) CreateUser(ctx context.Context, name, email, password, role, createdBy, modifiedBy string) (*model.User, error) {
	user := &model.User{
		Name:       name,
		Email:      email,
		Role:       role,
		Activated:  false,
		CreatedBy:  createdBy,
		ModifiedBy: modifiedBy,
	}
	err := user.Password.Set(password)
	if err != nil {
		return nil, err
	}
	v := validator.New()
	if user.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = c.repo.CreateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			v.AddError("email", "a user with this email already exists")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	// Generate an activation token.
	token, err := c.repo.CreateToken(ctx, user.ID, 3*24*time.Hour, model.ScopeActivation)
	if err != nil {
		return nil, err
	}
	// Send welcome email with activation token in a background goroutine.
	data := map[string]string{
		"activationToken": token.Plaintext,
		"name":            user.Name,
	}
	c.SendEmail(data, user.Email, "user_welcome.tmpl")
	return user, nil
}

func (c *Controller) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	v := validator.New()
	if model.ValidateEmail(v, email); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	user, err := c.repo.GetUserByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			v.AddError("email", "no matching email address found")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	return user, nil
}

func (c *Controller) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := c.repo.GetUserByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (c *Controller) GetAllUsers(ctx context.Context, name, email, role string, filters model.Filters, v *validator.Validator) ([]*model.User, model.Metadata, error) {
	if filters.Validate(v); !v.Valid() {
		return nil, model.Metadata{}, failedValidationErr(v.Errors)
	}
	users, metadata, err := c.repo.GetAllUsers(ctx, name, email, role, filters)
	if err != nil {
		return nil, model.Metadata{}, err
	}
	return users, metadata, nil
}

func (c *Controller) GetUserForToken(ctx context.Context, tokenScope, tokenPlaintext string) (*model.User, error) {
	v := validator.New()
	if model.ValidateTokenPlaintext(v, tokenPlaintext); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	user, err := c.repo.GetUserForToken(ctx, model.ScopeActivation, tokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			v.AddError("token", "invalid or expired activation token")
			return nil, failedValidationErr(v.Errors)
		default:
			return nil, err
		}
	}
	return user, nil
}

func (c *Controller) ActivateUser(ctx context.Context, user *model.User, modifiedBy string) error {
	// Update user.
	user.Activated = true
	user.ModifiedBy = modifiedBy
	err := c.repo.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEditConflict):
			return ErrEditConflict
		default:
			return err
		}
	}
	// Delete all activation tokens for user.
	err = c.repo.DeleteAllTokensForUser(ctx, model.ScopeActivation, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) UpdateUser(ctx context.Context, id int64, name, email, role *string, modifiedBy string) (*model.User, error) {
	user, err := c.repo.GetUserByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	if name != nil {
		user.Name = *name
	}
	if email != nil {
		user.Email = *email
	}
	if role != nil {
		user.Role = *role
	}
	user.ModifiedBy = modifiedBy
	v := validator.New()
	if user.Validate(v); !v.Valid() {
		return nil, failedValidationErr(v.Errors)
	}
	err = c.repo.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			v.AddError("email", "a user with this email already exists")
			return nil, failedValidationErr(v.Errors)
		case errors.Is(err, repository.ErrEditConflict):
			return nil, ErrEditConflict
		default:
			return nil, err
		}
	}
	return user, nil
}

func (c *Controller) DeleteUser(ctx context.Context, id int64) error {
	err := c.repo.DeleteUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

// AssignUserToProject assigns a user to a project.
func (c *Controller) AssignUserToProject(ctx context.Context, userID, projectID int64) error {
	v := validator.New()
	user, err := c.repo.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return ErrNotFound
		default:
			return err
		}
	}
	project, err := c.repo.GetProject(ctx, projectID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return ErrNotFound
		default:
			return err
		}
	}
	if user.Role != "member" {
		return ErrInvalidRole
	}
	err = c.repo.AssignUserToProject(ctx, user.ID, project.ID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateKey):
			v.AddError("user", "already assigned to project")
			return failedValidationErr(v.Errors)
		default:
			return err
		}
	}
	// Send email notification to assigned user.
	// data := map[string]string{
	// 	"name":        user.Name,
	// 	"projectID":   strconv.Itoa(int(project.ID)),
	// 	"projectName": project.Name,
	// }
	// // Send email notification to assignee only if assignee is project lead.
	// if assignee.Role == "lead" {
	// 	s.SendEmail(data, assignee.Email, "project_assign.tmpl")
	// }
	return nil
}

func (c *Controller) GetAllProjectsForUser(ctx context.Context, userID int64, filters model.Filters, v *validator.Validator) ([]*model.Project, model.Metadata, error) {
	if filters.Validate(v); !v.Valid() {
		return nil, model.Metadata{}, failedValidationErr(v.Errors)
	}
	projects, metadata, err := c.repo.GetAllProjectsForUser(ctx, userID, filters)
	if err != nil {
		return nil, model.Metadata{}, err
	}
	return projects, metadata, nil
}
