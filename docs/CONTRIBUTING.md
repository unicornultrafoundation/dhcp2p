# Contributing to DHCP2P

Thank you for your interest in contributing to DHCP2P! This guide will help you get started with contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors. By participating in this project, you agree to:

- Be respectful and inclusive
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards other community members

### Unacceptable Behavior

The following behaviors are considered unacceptable:

- Harassment, discrimination, or inappropriate language
- Personal attacks or trolling
- Public or private harassment
- Publishing private information without permission
- Any conduct that would be inappropriate in a professional setting

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.25+**: [Download and install Go](https://golang.org/dl/)
- **Docker**: [Install Docker](https://www.docker.com/get-started)
- **Git**: [Install Git](https://git-scm.com/downloads)
- **Make**: Optional but recommended for convenience commands

### Fork and Clone

1. **Fork the repository** on GitHub
2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/dhcp2p.git
   cd dhcp2p
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/unicornultrafoundation/dhcp2p.git
   ```

4. **Verify setup**:
   ```bash
   git remote -v
   # Should show:
   # origin    https://github.com/YOUR_USERNAME/dhcp2p.git (fetch)
   # origin    https://github.com/YOUR_USERNAME/dhcp2p.git (push)
   # upstream  https://github.com/unicornultrafoundation/dhcp2p.git (fetch)
   # upstream  https://github.com/unicornultrafoundation/dhcp2p.git (push)
   ```

## Development Setup

### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 2. Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
# Update .env with your local settings
```

### 3. Start Development Services

```bash
# Start PostgreSQL and Redis
make docker-up

# Verify services
make docker-health
```

### 4. Run Database Setup

```bash
# Apply migrations
make migrate

# Generate SQL code
make sqlc
```

### 5. Run Tests

```bash
# Run all tests
make test

# Run specific test categories
make test-unit
make test-integration
make test-e2e
```

## Contributing Process

### 1. Choose an Issue

- **Good First Issues**: Look for issues labeled `good first issue`
- **Bug Reports**: Issues labeled `bug`
- **Feature Requests**: Issues labeled `enhancement`
- **Documentation**: Issues labeled `documentation`

### 2. Create a Branch

```bash
# Update your fork
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
# or
git checkout -b docs/your-documentation-update
```

### 3. Make Changes

- **Follow coding standards** (see [Code Standards](#code-standards))
- **Write tests** for new functionality
- **Update documentation** as needed
- **Keep commits focused** and atomic

### 4. Test Your Changes

```bash
# Run all tests
make test

# Run linter
golangci-lint run

# Test your specific changes
go test -v ./path/to/your/changes
```

### 5. Commit Changes

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add new feature description"

# Push to your fork
git push origin feature/your-feature-name
```

## Code Standards

### Go Code Style

Follow standard Go conventions:

- **Formatting**: Use `gofmt` and `goimports`
- **Linting**: Use `golangci-lint`
- **Documentation**: Document all public functions and types
- **Naming**: Use clear, descriptive names

### Project-Specific Standards

#### Error Handling

```go
// Use domain errors
if err != nil {
    return nil, errors.ErrLeaseNotFound
}

// Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to allocate lease: %w", err)
}
```

#### Dependency Injection

```go
// Use constructor functions
func NewLeaseService(config *config.AppConfig, repo ports.LeaseRepository, logger *zap.Logger) *LeaseService {
    return &LeaseService{
        repo:   repo,
        logger: logger,
        config: config,
    }
}
```

#### Interface Design

```go
// Keep interfaces small and focused
type LeaseRepository interface {
    AllocateNewLease(ctx context.Context, peerID string) (*models.Lease, error)
    GetLeaseByPeerID(ctx context.Context, peerID string) (*models.Lease, error)
}
```

### Code Organization

- **One package per file**: Keep related functionality together
- **Clear separation**: Domain, application, adapters, infrastructure
- **Interface segregation**: Small, focused interfaces
- **Dependency inversion**: Depend on abstractions, not concretions

### Commit Message Format

Use conventional commit format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

#### Examples

```bash
feat(auth): add nonce expiration validation
fix(lease): handle concurrent lease allocation
docs(api): update authentication flow documentation
test(services): add unit tests for lease service
refactor(repositories): extract common database logic
```

## Testing Requirements

### Test Coverage

- **Unit Tests**: 80%+ coverage for new code
- **Integration Tests**: Test critical paths
- **End-to-End Tests**: Test user workflows

### Writing Tests

#### Unit Test Example

```go
func TestLeaseService_AllocateIP(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mocks.NewMockLeaseRepository(ctrl)
    service := services.NewLeaseService(&config.AppConfig{}, mockRepo, zap.NewNop())
    
    expectedLease := &models.Lease{
        TokenID: 12345,
        PeerID:  "peer123",
    }
    
    mockRepo.EXPECT().GetLeaseByPeerID(gomock.Any(), "peer123").Return(nil, nil)
    mockRepo.EXPECT().AllocateNewLease(gomock.Any(), "peer123").Return(expectedLease, nil)
    
    result, err := service.AllocateIP(context.Background(), "peer123")
    
    assert.NoError(t, err)
    assert.Equal(t, expectedLease, result)
}
```

#### Integration Test Example

```go
func TestLeaseRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    ctx := context.Background()
    
    // Start PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx, ...)
    require.NoError(t, err)
    defer postgresContainer.Terminate(ctx)
    
    // Test with real database
    repo, err := postgres.NewLeaseRepository(connStr)
    require.NoError(t, err)
    
    lease, err := repo.AllocateNewLease(ctx, "peer123")
    assert.NoError(t, err)
    assert.NotNil(t, lease)
}
```

### Test Commands

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test -v ./tests/unit/application/services/lease_service_test.go

# Run tests with race detection
go test -race ./tests/unit/...
```

## Pull Request Process

### 1. Create Pull Request

1. **Push your branch** to your fork
2. **Create pull request** on GitHub
3. **Fill out the template** completely
4. **Link related issues** using `Fixes #123` or `Closes #123`

### 2. Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] End-to-end tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No breaking changes (or documented)

## Related Issues
Fixes #123
```

### 3. Review Process

- **Automated Checks**: CI/CD pipeline runs tests and linting
- **Code Review**: At least one maintainer reviews the code
- **Testing**: Reviewer may test changes locally
- **Feedback**: Address all review comments

### 4. Merge Process

- **Squash and Merge**: Preferred for feature branches
- **Merge Commit**: For complex changes with multiple commits
- **Rebase and Merge**: For clean commit history

## Issue Reporting

### Bug Reports

Use the bug report template:

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., Ubuntu 20.04]
- Go Version: [e.g., 1.25.0]
- Docker Version: [e.g., 20.10.0]

## Additional Context
Any additional information
```

### Feature Requests

Use the feature request template:

```markdown
## Feature Description
Clear description of the feature

## Use Case
Why is this feature needed?

## Proposed Solution
How should this feature work?

## Alternatives
Alternative solutions considered

## Additional Context
Any additional information
```

## Documentation

### Documentation Standards

- **Clear and Concise**: Write for your audience
- **Examples**: Include code examples where appropriate
- **Up-to-date**: Keep documentation current with code
- **Consistent**: Follow project documentation style

### Types of Documentation

- **API Documentation**: Update API docs for new endpoints
- **Architecture Documentation**: Update for architectural changes
- **User Guides**: Update for new features
- **Developer Guides**: Update for new development processes

### Documentation Process

1. **Update relevant docs** with your changes
2. **Test documentation** for accuracy
3. **Include in PR** documentation updates
4. **Review documentation** during code review

## Release Process

### Version Numbering

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps

1. **Update version** in relevant files
2. **Update changelog** with new features/fixes
3. **Create release branch** from main
4. **Run full test suite** on release branch
5. **Create release tag** on GitHub
6. **Build and publish** Docker images
7. **Update documentation** for new release

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] Changelog updated
- [ ] Version numbers updated
- [ ] Release notes prepared
- [ ] Docker images built and tested
- [ ] Security scan completed

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Pull Request Comments**: Code-specific questions

### Resources

- **Documentation**: [docs/](docs/) directory
- **Code Examples**: Look at existing tests and examples
- **Architecture Guide**: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- **Development Guide**: [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)

### Mentorship

- **Good First Issues**: Look for issues labeled `good first issue`
- **Help Wanted**: Issues where maintainers need help
- **Mentorship**: Experienced contributors can mentor newcomers

## Recognition

### Contributors

We recognize contributors in several ways:

- **Contributors List**: GitHub automatically tracks contributors
- **Release Notes**: Contributors mentioned in release notes
- **Documentation**: Contributors acknowledged in documentation

### Contribution Types

We welcome various types of contributions:

- **Code**: Bug fixes, new features, refactoring
- **Documentation**: Guides, API docs, examples
- **Testing**: Unit tests, integration tests, e2e tests
- **Design**: UI/UX improvements, architecture discussions
- **Community**: Helping other contributors, answering questions

Thank you for contributing to DHCP2P! Your contributions help make the project better for everyone.
