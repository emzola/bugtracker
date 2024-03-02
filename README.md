# Issues Tracker API Backend

The Issues Tracker API Backend is designed to help you manage and track issues in your organization. It comes with role-based access control, allowing administrators, managers, and employees to have different levels of access and permissions.

## Table of Contents
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Usage](#usage)
  - [Authentication](#authentication)
  - [Roles and Permissions](#roles-and-permissions)
  - [Endpoints](#endpoints)
  - [Swagger API Documentation](#swagger-doc)
- [License](#license)

## <a id="features"></a>Features
- **Role-Based Access Control (RBAC):** Three user roles - administrators, managers, and employees - each with different access levels.
- **Issue Management:** Create, update, delete, and retrieve issues.
- **User Management:** Add, remove, and update users with different roles.
- **Authentication:** Secure access with token-based authentication.
- **Reporting:** Keep track of issue reports according to project.

## <a id="getting-started"></a>Getting Started

### <a id="prerequisites"></a>Prerequisites
- Go installed on your machine.
- PostgreSQL instance for data storage.

### <a id="installation"></a>Installation
1. Clone this repository: `git clone https://github.com/emzola/issuetracker.git`
2. Set up environment variables (see [Configuration](#configuration)).
3. Run database migration: `go run db/migrations/up`

### <a id="configuration"></a>Configuration
Create a `.envrc` file in the project root and configure the following variables:
```.envrc
export DSN=postgres://YourUserName:YourPassword@YourHostname/YourDatabaseName?sslmode=disable
export JWT_SECRET=YourJWTSecret
export SMTP_HOST=YourSMTPHost
export SMTP_USERNAME=YourSMTPUsername
export SMTP_PASSWORD=YourSMTPPassword
```

## <a id="usage"></a>Usage

### <a id="authentication"></a>Authentication
1. Create a new user account by making a POST request to `/v1/users`.
2. Obtain an access token by making a POST request to `/v1/tokens/authentication` with valid credentials. Include the token in the headers of subsequent requests.

### <a id="roles-and-permissions"></a>Roles and Permissions
- **Administrator:** Full access to all endpoints.
- **Manager:** Limited access (e.g., cannot add or remove users).
- **Employee:** Restricted access (e.g., can only view and update their own issues).

### <a id="endpoints"></a>Endpoints
- **Projects:**
  - `GET /v1/projects` - Retrieve all projects.
  - `GET /v1/projects/:id` - Retrieve a specific project.
  - `GET /v1/projects/:id/users` - Retrieve all users for a project.
  - `POST /v1/projects` - Create a new project.
  - `PUT /v1/projects/:id` - Update a project.
  - `DELETE /v1/projects/:id` - Delete a project.

- **Issues:**
  - `GET /v1/issues` - Retrieve all issues.
  - `GET /v1/issues/:id` - Retrieve a specific issue.
  - `POST /v1/issues` - Create a new issue.
  - `PUT /v1/issues/:id` - Update an issue.
  - `DELETE /v1/issues/:id` - Delete an issue.

- **Reports:**
  - `GET /v1/issuesreport/status` - Retrieve report for issues statuses.
  - `GET /v1/issuesreport/assignee` - Retrieve report for issues assignees.
  - `GET /v1/issuesreport/reporter` - Retrieve report for issues reporters.
  - `GET /v1/issuesreport/priority` - Retrieve report for issues priorities.
  - `GET /v1/issuesreport/date` - Retrieve report for issues target dates.
  
- **Users:**
  - `GET /v1/users` - Retrieve all users.
  - `GET /v1/users/:id` - Retrieve a specific user.
  - `POST /v1/users` - Create a new user.
  - `PUT /v1/users/:id` - Update a user.
  - `DELETE /v1/users/:id` - Delete a user.
  - `PUT /v1/users/activated` - Activate a new user.
  - `GET /v1/users/:id/projects` - Retrieve all projects for a user.
  - `POST /v1/users/:id/projects` - Assign user to project.

- **Tokens:**
  - `POST /v1/tokens/activation` - Create user activation token.
  - `POST /v1/tokens/authentication` - Create user authentication token.

### <a id="swagger-doc"></a>Swagger API Documentation

Swagger API documentation and request/response examples can be found on [http://localhost:8080/docs] when you run the API locally.

## <a id="license"></a>License
This project is licensed under the MIT License.