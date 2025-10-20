# Contributing to CodeValdCortex

Thank you for your interest in contributing to CodeValdCortex! This document provides guidelines and information for contributors.

## 🤝 Code of Conduct

We are committed to providing a welcoming and inclusive experience for everyone. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## 🎯 How to Contribute

### Reporting Issues

1. **Search existing issues** first to avoid duplicates
2. **Use issue templates** when creating new issues
3. **Provide detailed information**:
   - Environment details (OS, Go version, Kubernetes version)
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs or error messages

### Suggesting Features

1. **Check existing feature requests** in issues and discussions
2. **Create a detailed feature request** including:
   - Use case and business value
   - Proposed implementation approach
   - Potential alternatives considered
   - Impact on existing functionality

### Contributing Code

#### Prerequisites

- Go 1.21 or later
- Git
- Make
- Docker and Docker Compose
- golangci-lint
- Basic knowledge of Kubernetes

#### Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/CodeValdCortex.git
   cd CodeValdCortex
   ```

2. **Set up development environment**:
   ```bash
   make dev-setup
   ```

3. **Start local services**:
   ```bash
   docker-compose up -d
   ```

4. **Verify setup**:
   ```bash
   make check
   make build
   ```

#### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/MVP-XXX_description
   # or
   git checkout -b fix/issue-description
   ```

2. **Make your changes**:
   - Follow the [coding standards](#coding-standards)
   - Add tests for new functionality
   - Update documentation as needed
   - Ensure all checks pass: `make check`

3. **Test your changes**:
   ```bash
   make test
   make test-coverage
   make lint
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add agent lifecycle management
   
   - Implement basic agent CRUD operations
   - Add health monitoring capabilities
   - Include unit tests with 90% coverage
   
   Closes #123"
   ```

5. **Push and create pull request**:
   ```bash
   git push origin feature/MVP-XXX_description
   ```

#### Pull Request Guidelines

- **Use descriptive titles** following conventional commit format
- **Fill out the PR template** completely
- **Link related issues** using "Closes #123" or "Addresses #123"
- **Include tests** for new functionality
- **Update documentation** if needed
- **Ensure all CI checks pass**
- **Keep PRs focused** - one feature or fix per PR
- **Rebase before merging** to maintain clean history

## 📋 Coding Standards

### Go Code Style

- **Follow Go conventions**: Use `gofmt`, `goimports`, and `go vet`
- **Use meaningful names**: Variables, functions, and packages should be descriptive
- **Write documentation**: All exported functions should have godoc comments
- **Handle errors properly**: Always check and handle errors appropriately
- **Use interfaces wisely**: Define interfaces where they make sense for testing and modularity

### Code Organization

```
internal/
├── agent/          # Agent management logic
├── api/           # API handlers and routing
├── config/        # Configuration management
├── db/            # Database abstractions
├── messaging/     # Inter-agent communication
└── orchestrator/  # Workflow orchestration

pkg/
├── client/        # Public client library
├── types/         # Shared types and interfaces
└── utils/         # Utility functions

cmd/
└── main.go        # Application entry point
```

### Testing Standards

- **Write unit tests** for all new functionality
- **Aim for >80% coverage** on new code
- **Use table-driven tests** for multiple test cases
- **Mock external dependencies** using interfaces
- **Include integration tests** for API endpoints
- **Test error conditions** not just happy paths

Example test structure:
```go
func TestAgentManager_CreateAgent(t *testing.T) {
    tests := []struct {
        name     string
        input    AgentRequest
        want     *Agent
        wantErr  bool
    }{
        {
            name: "valid agent creation",
            input: AgentRequest{Name: "test-agent"},
            want: &Agent{ID: "123", Name: "test-agent"},
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Documentation Standards

- **Godoc comments** for all exported types and functions
- **README updates** for new features
- **API documentation** using OpenAPI/Swagger
- **Architecture decisions** documented in `/docs`
- **Configuration examples** for new options

## 🏗️ Architecture Guidelines

### Design Principles

1. **Cloud Native**: Design for Kubernetes-first deployment
2. **Microservices**: Loosely coupled, independently deployable components
3. **Observability**: Built-in metrics, logging, and tracing
4. **Security**: Secure by default, zero-trust principles
5. **Performance**: Optimized for high concurrency and low latency

### Component Guidelines

- **Use dependency injection** for testability
- **Define clear interfaces** between components
- **Implement graceful shutdown** for all services
- **Support configuration via environment variables**
- **Include health checks** for all services

## 🚀 Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in relevant files
- [ ] Docker images built and tested
- [ ] Security scan completed
- [ ] Performance benchmarks run

## 🔍 Code Review Process

### For Reviewers

- **Focus on correctness** and design
- **Check test coverage** and quality
- **Verify documentation** is updated
- **Ensure style consistency**
- **Consider performance implications**
- **Be respectful and constructive**

### Review Checklist

- [ ] Code follows project conventions
- [ ] Tests are comprehensive and pass
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced
- [ ] Performance impact considered
- [ ] Error handling is appropriate

## 🛠️ Development Tools

### Required Tools

```bash
# Install development tools
make install-tools

# This installs:
# - golangci-lint (linting)
# - air (hot reload)
# - Additional Go tools
```

### Recommended IDE Setup

#### VS Code
```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true
}
```

#### GoLand/IntelliJ
- Enable Go modules support
- Configure golangci-lint integration
- Set up test coverage visualization

### Git Hooks

We recommend setting up pre-commit hooks:

```bash
# Setup pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
make check
EOF

chmod +x .git/hooks/pre-commit
```

## 📝 Commit Message Format

We follow [Conventional Commits](https://conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

### Examples
```
feat(agent): add lifecycle management capabilities

Implement basic CRUD operations for agent management
including creation, deletion, and status monitoring.

Closes #123

fix: resolve memory leak in message broker

The message broker was not properly cleaning up closed
connections, leading to memory leaks under high load.

test(api): add integration tests for agent endpoints

Increases test coverage for agent management API endpoints
from 60% to 85%.
```

## 🆘 Getting Help

- **Documentation**: Check the `/docs` directory
- **Issues**: Search existing GitHub issues
- **Discussions**: Use GitHub Discussions for questions
- **Discord**: Join our Discord server (link in README)
- **Email**: Contact maintainers directly for security issues

## 🙏 Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Annual contributor appreciation

Thank you for contributing to CodeValdCortex! 🚀