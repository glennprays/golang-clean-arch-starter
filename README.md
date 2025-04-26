# Golang Clean Code Architecture Starter Template
This is a starter template for building applications in Go using the Clean Architecture principles. It provides a structured way to organize your code, making it easier to maintain and scale.

## Get Started 
### Prerequisites
- Go 1.24 or later
- Docker 
- Make 
- Git 

### Clone the Repository
Clone via SSH:
```bash
git clone git@github.com:glennprays/golang-clean-arch-starter.git
```
Clone via HTTPS:
```bash
git clone https://github.com/glennprays/golang-clean-arch-starter.git 
```

### Change Golang Module Name [IMPORTANT]
This step is crucial. The module name in `go.mod` should be changed to match your project name. This is important for proper dependency management and module resolution.
```bash
make rename NEW_MODULE=github.com/yourname/yourproject
```
After renaming, ensure to check the `go.mod` file to confirm the module name has been updated correctly. Then remove git history and reinitialize the repository (optional): 
```bash 
rm -rf .git 
git init 
git add . 
git commit -m "Initial commit" 
```

## Directory Structure Explanation
Here's a brief overview of the directory structure:
### `cmd/`
This directory contains the main application entry points. Each subdirectory represents a different application or service. 
### `internal/` 
This directory contains the core business logic and domain entities. It is divided into several subdirectories:
- `model/` - Contains the domain models. 
- `repository/` - Contains the repository interfaces and implementations. 
- `service/` - Contains the business logic and service interfaces. 
- `usecase/` - Contains the use case interfaces and implementations.
- `handler/` - Contains the HTTP handlers and controllers. 
- `middleware/` - Contains the middleware functions.
- `router/` - Contains the router setup and configuration.
- `utils/` - Contains utility functions and helpers.
- `worker/` - Contains the worker and job processing logic.
- `httperror/` - Contains custom HTTP error handling and responses for http.
### `pkg/` 
This directory contains shared libraries and packages that can be used across different applications. 
### `misc/`
This directory contains miscellaneous files and configurations, such as Dockerfiles, Makefiles, and other project-related files.
### `templates/`
This directory contains static templates.
### `migrations/`
This directory contains database migration files.
### `docs/`
This directory contains documentation files, in this case, Swagger API documentation.

