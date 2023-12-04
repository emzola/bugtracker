package rbac

import (
	"encoding/json"
	"os"

	"github.com/emzola/issuetracker/internal/model"
)

// Resources holds data for URL resources.
type Resources []string

// Actions holds data for HTTP methods.
type Actions map[string]Resources

// Roles holds data for user roles.
type Roles map[string]Actions

// Authorizer defines roles.
type Authorizer struct {
	roles Roles
}

// New creates a new Authorizer instance.
func New(roles Roles) Authorizer {
	return Authorizer{roles: roles}
}

// ActionFromMethod returns role actions from HTTP methods.
func (a Authorizer) ActionFromMethod(httpMethod string) string {
	switch httpMethod {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}

// HasPermission checks whether a user has permissions to access a resource.
func (a Authorizer) HasPermission(user *model.User, action, asset string) bool {
	userRole := user.Role
	role, ok := a.roles[userRole]
	if !ok {
		return false
	}
	resources, ok := role[action]
	if !ok {
		return false
	}
	for _, resource := range resources {
		if resource == asset {
			return true
		}
	}
	return false
}

// LoadRoles loads roles from JSON file.
func LoadRoles(filename string) (Roles, error) {
	var roles Roles
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
