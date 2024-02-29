# Issues Tracker API Backend

Welcome to the Issues Tracker API Backend repository! This API is designed to help you manage and track issues in your organization. It comes with role-based access control, allowing administrators, managers, and employees to have different levels of access and permissions.

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
- **Issues:**
  - `GET /v1/issues` - Retrieve all issues.
  - `GET /v1/issues/:id` - Retrieve a specific issue.
  - `POST /v1/issues` - Create a new issue.
  - `PUT /v1/issues/:id` - Update an issue.
  - `DELETE /v1/issues/:id` - Delete an issue.
  
- **Users:**
  - `GET /v1/users` - Retrieve all users.
  - `GET /v1/users/:id` - Retrieve a specific user.
  - `POST /v1/users` - Create a new user.
  - `PUT /v1/users/:id` - Update a user.
  - `DELETE /v1/users/:id` - Delete a user.

For detailed API documentation and request/response examples, refer to [API Documentation](https://bibliotheca-api-dev-xfnt.4.us-1.fl0.io/api-docs)

## <a id="license"></a>License
This project is licensed under the MIT License.