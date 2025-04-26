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
