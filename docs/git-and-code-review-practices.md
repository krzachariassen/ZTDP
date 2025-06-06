# Git and Code Review Practices

## Overview

This document outlines the Git workflow, branching strategy, commit standards, and code review practices for the ZTDP AI-native platform. These practices ensure consistent, high-quality development while maintaining clean architecture principles and supporting the AI-enhanced development workflow.

## Table of Contents

1. [Git Workflow Strategy](#git-workflow-strategy)
2. [Branching Model](#branching-model)
3. [Commit Standards](#commit-standards)
4. [Pull Request Process](#pull-request-process)
5. [Code Review Guidelines](#code-review-guidelines)
6. [Development Environment](#development-environment)
7. [CI/CD Integration](#cicd-integration)
8. [Release Management](#release-management)
9. [Hotfix Procedures](#hotfix-procedures)
10. [Best Practices](#best-practices)

---

## Git Workflow Strategy

### Core Principles

1. **Main Branch Protection**: The `main` branch always contains deployable code
2. **Feature-Driven Development**: All changes go through feature branches
3. **Review-Required**: No direct commits to main branch
4. **Clean History**: Use squash merges to maintain readable commit history
5. **AI-Enhanced Reviews**: Leverage AI tools for code analysis and review assistance

### Workflow Overview

```mermaid
gitGraph
    commit id: "Initial"
    branch feature/deployment-ai
    commit id: "Add AI provider"
    commit id: "Implement service"
    commit id: "Add tests"
    checkout main
    merge feature/deployment-ai
    commit id: "Release v1.1"
    branch hotfix/policy-bug
    commit id: "Fix policy validation"
    checkout main
    merge hotfix/policy-bug
    commit id: "Hotfix v1.1.1"
```

---

## Branching Model

### Branch Types

#### Main Branch
- **Purpose**: Production-ready code
- **Protection**: Protected branch requiring PR and reviews
- **Deployment**: Automatically deployed to production
- **Naming**: `main`

#### Feature Branches
- **Purpose**: New features and enhancements
- **Lifetime**: Created from main, merged back to main
- **Naming**: `feature/domain-description`
- **Examples**:
  - `feature/deployment-ai-integration`
  - `feature/policy-graph-validation`
  - `feature/event-streaming-websockets`

#### Hotfix Branches
- **Purpose**: Critical production bug fixes
- **Lifetime**: Created from main, merged to main
- **Naming**: `hotfix/issue-description`
- **Examples**:
  - `hotfix/deployment-timeout-fix`
  - `hotfix/policy-validation-edge-case`

#### Release Branches
- **Purpose**: Release preparation and stabilization
- **Lifetime**: Created from main, merged to main with tag
- **Naming**: `release/version`
- **Examples**:
  - `release/v1.2.0`
  - `release/v2.0.0-beta.1`

### Branch Creation Commands

```bash
# Create feature branch
git checkout main
git pull origin main
git checkout -b feature/deployment-ai-integration

# Create hotfix branch
git checkout main
git pull origin main
git checkout -b hotfix/policy-validation-fix

# Create release branch
git checkout main
git pull origin main
git checkout -b release/v1.2.0
```

---

## Commit Standards

### Commit Message Format

#### Structure
```
type(scope): description

[optional body]

[optional footer]
```

#### Types
- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring without changing functionality
- **test**: Adding or updating tests
- **chore**: Maintenance tasks, dependency updates
- **perf**: Performance improvements
- **ci**: CI/CD pipeline changes

#### Scopes
- **deployment**: Deployment-related changes
- **policy**: Policy engine and validation
- **ai**: AI provider and integration
- **api**: API handlers and routing
- **graph**: Graph database and operations
- **events**: Event system and messaging
- **auth**: Authentication and authorization
- **docs**: Documentation

#### Examples

```bash
# Feature commits
feat(deployment): integrate AI for deployment plan generation
feat(policy): add graph-based policy validation engine
feat(events): implement real-time WebSocket event streaming

# Bug fix commits
fix(deployment): resolve timeout issue in deployment execution
fix(policy): handle edge case in nested policy evaluation
fix(api): correct error response format for validation failures

# Documentation commits
docs(architecture): update AI integration guidelines
docs(api): add Swagger documentation for deployment endpoints
docs(testing): document TDD practices for domain services

# Refactoring commits
refactor(handlers): extract common error handling patterns
refactor(ai): simplify AI provider interface
refactor(deployment): improve service method organization

# Test commits
test(deployment): add comprehensive deployment service tests
test(policy): add integration tests for policy evaluation
test(ai): add mock AI provider for unit testing
```

### Commit Best Practices

#### Atomic Commits
```bash
# ✅ Good: Single logical change
git commit -m "feat(deployment): add AI-enhanced plan generation"

# ❌ Bad: Multiple unrelated changes
git commit -m "feat(deployment): add AI planning and fix policy bug and update docs"
```

#### Descriptive Messages
```bash
# ✅ Good: Clear and specific
git commit -m "fix(policy): resolve null pointer in nested policy evaluation"

# ❌ Bad: Vague and unclear
git commit -m "fix: bug fix"
```

#### Include Context When Needed
```bash
# Complex change with context
git commit -m "refactor(ai): extract AI provider interface

Separates business logic from AI infrastructure concerns.
- Move deployment logic to deployment service
- Keep AI provider as pure infrastructure
- Add comprehensive interface for future providers

Resolves #123"
```

---

## Pull Request Process

### PR Creation Workflow

#### 1. Prepare Branch
```bash
# Ensure latest main branch
git checkout main
git pull origin main

# Rebase feature branch
git checkout feature/deployment-ai-integration
git rebase main

# Run tests locally
go test ./...
golangci-lint run

# Push changes
git push origin feature/deployment-ai-integration
```

#### 2. Create Pull Request

**Title Format**: Same as commit message format
```
feat(deployment): integrate AI for deployment plan generation
```

**Description Template**:
```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Architecture Impact
- [ ] Follows clean architecture principles
- [ ] Business logic in appropriate domain service
- [ ] AI integration follows AI-as-infrastructure pattern
- [ ] Events emitted for all significant operations

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] All existing tests pass
- [ ] Test coverage maintained/improved

## Documentation
- [ ] Code comments updated
- [ ] API documentation updated
- [ ] Architecture documentation updated (if needed)

## Checklist
- [ ] Self-review completed
- [ ] Code follows style guidelines
- [ ] No console.log or debug statements
- [ ] Error handling follows established patterns
- [ ] Event emission follows standards

## Screenshots/Examples
Include relevant screenshots, API examples, or demo output.
```

### PR Size Guidelines

#### Optimal Size
- **Lines Changed**: 200-400 lines preferred
- **Files Changed**: 5-10 files maximum
- **Duration**: Should be reviewable in 30-60 minutes

#### Large PR Guidelines
If PR exceeds size guidelines:
1. **Justify Size**: Explain why it can't be broken down
2. **Extra Documentation**: Provide detailed explanation
3. **Multiple Reviewers**: Assign additional reviewers
4. **Extended Review Time**: Allow more time for review

#### Breaking Down Large Changes
```bash
# Example: Break down large AI integration
# 1. First PR: Add AI provider interface
feat(ai): add AI provider interface and base implementation

# 2. Second PR: Integrate with deployment service
feat(deployment): integrate AI provider for plan generation

# 3. Third PR: Add comprehensive tests
test(deployment): add AI integration tests

# 4. Fourth PR: Update API handlers
feat(api): update deployment handlers to use AI-enhanced service
```

---

## Code Review Guidelines

### Review Responsibilities

#### Author Responsibilities
1. **Self-Review**: Review your own PR before requesting review
2. **Complete Description**: Provide thorough PR description
3. **Tests Included**: Ensure comprehensive test coverage
4. **Documentation Updated**: Update relevant documentation
5. **Response Timeliness**: Respond to review feedback promptly

#### Reviewer Responsibilities
1. **Timely Review**: Review within 24 hours of assignment
2. **Thorough Analysis**: Check code, tests, and documentation
3. **Constructive Feedback**: Provide helpful, actionable feedback
4. **Architecture Alignment**: Ensure changes follow architectural principles
5. **Knowledge Sharing**: Share relevant knowledge and alternatives

### Review Criteria

#### Architecture Compliance
```go
// ✅ Good: Business logic in domain service
func (s *DeploymentService) PlanDeployment(ctx context.Context, app string) (*Plan, error) {
    // Validation
    if err := s.validateApplication(app); err != nil {
        return nil, err
    }
    
    // Use AI as infrastructure tool
    if s.aiProvider != nil {
        // AI logic here
    }
    
    // Business logic
    return s.generatePlan(app), nil
}

// ❌ Bad: Business logic in API handler
func (h *Handler) PlanDeployment(w http.ResponseWriter, r *http.Request) {
    // Don't put business logic here
    if isProductionApp(app) {
        // Policy validation in handler - WRONG
    }
}
```

#### Error Handling
```go
// ✅ Good: Proper error wrapping and context
if err := s.validateApplication(app); err != nil {
    return nil, fmt.Errorf("application validation failed for %s: %w", app, err)
}

// ❌ Bad: Generic error without context
if err := s.validateApplication(app); err != nil {
    return nil, err
}
```

#### Event Emission
```go
// ✅ Good: Emit events for operations
s.eventBus.Emit("deployment.planning.started", map[string]interface{}{
    "app": app,
    "timestamp": time.Now(),
})

// Process deployment planning...

s.eventBus.Emit("deployment.planning.completed", map[string]interface{}{
    "app": app,
    "plan": plan,
    "duration": time.Since(start),
})
```

#### AI Integration
```go
// ✅ Good: AI as infrastructure tool with fallback
func (s *DeploymentService) generatePlan(ctx context.Context, app string) (*Plan, error) {
    // Try AI-enhanced planning
    if s.aiProvider != nil {
        if plan, err := s.aiEnhancedPlan(ctx, app); err == nil {
            return plan, nil
        }
    }
    
    // Fallback to basic planning
    return s.basicPlan(app), nil
}

// ❌ Bad: No fallback, AI dependency required
func (s *DeploymentService) generatePlan(ctx context.Context, app string) (*Plan, error) {
    return s.aiProvider.GeneratePlan(ctx, app) // What if AI fails?
}
```

### Review Process

#### 1. Initial Review
- **Code Structure**: Check overall organization and clarity
- **Architecture**: Verify clean architecture compliance
- **Business Logic**: Ensure proper domain separation
- **Dependencies**: Check dependency direction and interfaces

#### 2. Detailed Review
- **Implementation**: Check algorithm correctness and efficiency
- **Error Handling**: Verify comprehensive error scenarios
- **Security**: Look for potential security vulnerabilities
- **Performance**: Consider performance implications

#### 3. Testing Review
- **Coverage**: Ensure comprehensive test coverage
- **Test Quality**: Check test clarity and maintainability
- **Edge Cases**: Verify edge cases are tested
- **Integration**: Check integration test scenarios

#### 4. Documentation Review
- **Code Comments**: Check for clear, helpful comments
- **API Documentation**: Verify API docs are updated
- **Architecture Docs**: Check if architecture docs need updates

### Review Comments

#### Constructive Feedback Examples

```markdown
# ✅ Good feedback: Specific, helpful, actionable
**Suggestion**: Consider extracting this validation logic into a separate method to improve readability and reusability.

```go
func (s *DeploymentService) validateDeploymentRequest(req *Request) error {
    // validation logic here
}
```

**Architecture**: This business logic should be in the domain service rather than the API handler. Could you move it to `DeploymentService.PlanDeployment()`?

**Performance**: This could potentially cause N+1 query issues. Consider using a batch operation or caching.

# ❌ Poor feedback: Vague, not helpful
This looks wrong.
Fix this.
Bad code.
```

#### Review Approval

```markdown
# Complete review approval
✅ **LGTM** - Code follows clean architecture principles, comprehensive tests included, and documentation is updated. Great work on the AI integration fallback pattern!

# Conditional approval
✅ **LGTM with minor suggestions** - Approved pending the suggested refactoring of the validation logic. Non-blocking changes.

# Request changes
❌ **Changes Requested** - Need to address the business logic placement in API handlers before this can be merged. See specific comments for details.
```

---

## Development Environment

### Local Setup

#### Prerequisites
```bash
# Install Go 1.21+
go version

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install testing tools
go install gotest.tools/gotestsum@latest
go install github.com/boumenot/gocover-cobertura@latest
```

#### Git Configuration
```bash
# Configure Git user
git config --global user.name "Your Name"
git config --global user.email "your.email@domain.com"

# Configure Git hooks
git config --global core.hooksPath .githooks

# Set up commit message template
git config --global commit.template .gitmessage

# Configure diff and merge tools
git config --global merge.tool vscode
git config --global diff.tool vscode
```

#### Git Hooks Setup

**.githooks/pre-commit**
```bash
#!/bin/bash
# Pre-commit hook for ZTDP

echo "Running pre-commit checks..."

# Run linting
golangci-lint run
if [ $? -ne 0 ]; then
    echo "❌ Linting failed. Please fix issues before committing."
    exit 1
fi

# Run tests
go test ./...
if [ $? -ne 0 ]; then
    echo "❌ Tests failed. Please fix tests before committing."
    exit 1
fi

# Check formatting
gofmt -l . | grep -v vendor/
if [ $? -eq 0 ]; then
    echo "❌ Code not formatted. Please run 'gofmt -w .'"
    exit 1
fi

echo "✅ Pre-commit checks passed!"
```

**.githooks/commit-msg**
```bash
#!/bin/bash
# Commit message validation hook

commit_regex='^(feat|fix|docs|style|refactor|test|chore|perf|ci)(\(.+\))?: .{1,50}'

if ! grep -qE "$commit_regex" "$1"; then
    echo "❌ Invalid commit message format!"
    echo "Format: type(scope): description"
    echo "Example: feat(deployment): add AI integration"
    exit 1
fi

echo "✅ Commit message format valid!"
```

### IDE Configuration

#### VS Code Settings
```json
{
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "go.testOnSave": true,
    "go.formatTool": "goimports",
    "go.generateTestsFlags": ["-exported"],
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    },
    "files.exclude": {
        "**/vendor/": true,
        "**/node_modules/": true
    }
}
```

#### VS Code Extensions
```json
{
    "recommendations": [
        "golang.go",
        "ms-vscode.vscode-go",
        "github.copilot",
        "github.copilot-chat",
        "redhat.vscode-yaml",
        "ms-vscode.makefile-tools"
    ]
}
```

---

## CI/CD Integration

### GitHub Actions Workflow

**.github/workflows/ci.yml**
```yaml
name: CI/CD Pipeline

on:
  pull_request:
    branches: [ main ]
  push:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run linting
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Run tests
      run: |
        go test -race -coverprofile=coverage.out -covermode=atomic ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Build application
      run: go build -v ./cmd/api
    
    - name: Run integration tests
      run: make test-integration
      env:
        TEST_DATABASE_URL: postgres://user:pass@localhost/testdb
    
  security:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: './...'
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
```

### Branch Protection Rules

```yaml
# Branch protection configuration
protection_rules:
  main:
    required_status_checks:
      strict: true
      contexts:
        - "test"
        - "security"
        - "golangci-lint"
    enforce_admins: true
    required_pull_request_reviews:
      required_approving_review_count: 1
      dismiss_stale_reviews: true
      require_code_owner_reviews: true
    restrictions: null
    allow_force_pushes: false
    allow_deletions: false
```

### Quality Gates

#### Required Checks
- ✅ All tests pass
- ✅ Linting passes (golangci-lint)
- ✅ Security scan passes (gosec)
- ✅ Code coverage ≥ 80%
- ✅ At least one code review approval
- ✅ No merge conflicts
- ✅ Commit message format valid

#### Automated Actions
- **Auto-merge**: Dependabot PRs with passing tests
- **Auto-label**: Label PRs based on changed files
- **Auto-assign**: Assign reviewers based on CODEOWNERS
- **Auto-deploy**: Deploy to staging on main branch updates

---

## Release Management

### Semantic Versioning

#### Version Format
```
MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]

Examples:
- 1.0.0 (stable release)
- 1.1.0-beta.1 (beta prerelease)
- 1.0.1 (patch release)
- 2.0.0-rc.1 (release candidate)
```

#### Version Increment Rules
- **MAJOR**: Breaking changes to API or architecture
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Release Process

#### 1. Release Planning
```bash
# Create release branch
git checkout main
git pull origin main
git checkout -b release/v1.2.0

# Update version in code
./scripts/update-version.sh 1.2.0

# Update CHANGELOG
vim CHANGELOG.md

# Commit version updates
git commit -m "chore(release): prepare v1.2.0"
```

#### 2. Release Testing
```bash
# Run full test suite
make test-all

# Run integration tests
make test-integration

# Run performance tests
make test-performance

# Security scan
make security-scan
```

#### 3. Release Creation
```bash
# Merge to main
git checkout main
git merge release/v1.2.0

# Create and push tag
git tag -a v1.2.0 -m "Release v1.2.0

Features:
- AI-enhanced deployment planning
- Improved policy validation
- WebSocket event streaming

Bug Fixes:
- Fixed deployment timeout issues
- Resolved policy evaluation edge cases

Breaking Changes:
- None"

git push origin v1.2.0
```

#### 4. Release Documentation
```markdown
# Release Notes Template

## [1.2.0] - 2025-06-15

### 🚀 Features
- **AI Integration**: Added AI-enhanced deployment planning with fallback
- **Event Streaming**: Real-time WebSocket event streaming
- **Policy Engine**: Graph-based policy validation

### 🐛 Bug Fixes
- Fixed deployment timeout issues in large applications
- Resolved policy evaluation edge cases with nested conditions
- Corrected API error response formats

### 📚 Documentation
- Updated architecture documentation
- Added API examples for new endpoints
- Improved developer setup guide

### 🔧 Internal
- Refactored AI provider interface
- Improved test coverage to 85%
- Enhanced error handling patterns

### ⬆️ Dependencies
- Updated Go to 1.21.5
- Updated chi router to v5.0.8
- Added testify for better testing

### 🔐 Security
- Enhanced input validation
- Improved error message sanitization
- Updated dependency vulnerabilities
```

### Release Automation

**Makefile Release Targets**
```makefile
# Release management targets
.PHONY: release-prepare release-test release-create

release-prepare:
	@echo "Preparing release $(VERSION)"
	@./scripts/update-version.sh $(VERSION)
	@./scripts/update-changelog.sh $(VERSION)

release-test:
	@echo "Running release tests"
	@make test-all
	@make test-integration
	@make security-scan

release-create:
	@echo "Creating release $(VERSION)"
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@./scripts/create-github-release.sh $(VERSION)
```

---

## Hotfix Procedures

### Emergency Hotfix Process

#### 1. Issue Identification
```bash
# When critical production issue is identified
# 1. Create hotfix branch immediately
git checkout main
git pull origin main
git checkout -b hotfix/critical-deployment-fix

# 2. Document the issue
echo "Critical deployment timeout in production apps" > HOTFIX.md
echo "Affects: All deployments > 10 minutes" >> HOTFIX.md
echo "Impact: Failed deployments, service unavailability" >> HOTFIX.md
```

#### 2. Fix Implementation
```bash
# Implement minimal fix
# Focus on the specific issue only
# Avoid additional changes or refactoring

# Example: Fix timeout configuration
git add internal/deployments/config.go
git commit -m "fix(deployment): increase timeout for large applications

Resolves critical production issue where deployments > 10 minutes
fail due to timeout. Increases default timeout from 5m to 30m.

Fixes: #PROD-123"
```

#### 3. Testing
```bash
# Run targeted tests for the fix
go test ./internal/deployments -run TestDeploymentTimeout

# Run related integration tests
make test-deployment-integration

# Manual verification if possible
./scripts/verify-fix.sh
```

#### 4. Fast-Track Review
```bash
# Create PR with hotfix label
gh pr create \
    --title "hotfix: critical deployment timeout fix" \
    --body "Emergency fix for production deployment timeouts" \
    --label "hotfix,critical,needs-immediate-review"

# Request immediate review from on-call team
gh pr review --request @oncall-team
```

#### 5. Emergency Deployment
```bash
# After approval, merge immediately
gh pr merge --squash

# Create emergency release
git checkout main
git pull origin main
git tag -a v1.1.1 -m "Hotfix v1.1.1: Critical deployment timeout fix"
git push origin v1.1.1

# Deploy to production immediately
make deploy-production VERSION=v1.1.1
```

### Hotfix Quality Gates

#### Minimal Requirements
- ✅ Fix addresses specific production issue
- ✅ Tests pass for affected functionality
- ✅ At least one senior developer review
- ✅ Change is minimal and focused
- ✅ No additional features or refactoring

#### Documentation Requirements
```markdown
# Hotfix Documentation Template

## Hotfix: v1.1.1

### Issue
Brief description of the production issue.

### Root Cause
Technical explanation of what caused the issue.

### Fix
Description of the minimal change implemented.

### Testing
- Manual testing performed
- Automated tests run
- Verification steps

### Rollback Plan
Steps to rollback if the hotfix causes additional issues.

### Follow-up Actions
- [ ] Create proper fix for next minor release
- [ ] Update monitoring to prevent recurrence
- [ ] Review root cause in retrospective
```

---

## Best Practices

### Daily Development Workflow

#### Morning Routine
```bash
# Start of day workflow
git checkout main
git pull origin main

# Check for any urgent issues
gh issue list --label="urgent,bug"

# Plan daily work
gh project list
gh milestone list
```

#### Feature Development
```bash
# Create feature branch
git checkout -b feature/new-ai-provider

# Development cycle
while developing:
    # Make changes
    # Run tests frequently
    go test ./...
    
    # Commit regularly
    git add .
    git commit -m "feat(ai): add anthropic provider interface"
    
    # Push regularly
    git push origin feature/new-ai-provider
done

# Final checks before PR
make lint
make test-all
make build
```

#### End of Day
```bash
# Push any work in progress
git add .
git commit -m "wip: partial implementation of anthropic provider"
git push origin feature/new-ai-provider

# Clean up local branches
git branch --merged | grep -v main | xargs git branch -d
```

### Collaboration Practices

#### Pair Programming
```bash
# Set up shared session
git checkout -b pair/ai-integration-session

# Regular commits during pairing
git commit -m "pair: implement AI provider validation"
git commit -m "pair: add error handling for AI timeouts"
git commit -m "pair: refactor AI response parsing"

# Final cleanup
git rebase -i HEAD~3  # Squash related commits
```

#### Code Reviews
- **Review within 24 hours** of assignment
- **Provide constructive feedback** with examples
- **Approve with confidence** or request changes clearly
- **Follow up** on requested changes

#### Knowledge Sharing
- **Document decisions** in architecture docs
- **Share learnings** in team meetings
- **Update examples** when patterns evolve
- **Mentor junior developers** through reviews

### Tool Configuration

#### Git Aliases
```bash
# Add helpful Git aliases
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.unstage 'reset HEAD --'
git config --global alias.last 'log -1 HEAD'
git config --global alias.visual '!gitk'

# ZTDP-specific aliases
git config --global alias.feature 'checkout -b feature/'
git config --global alias.hotfix 'checkout -b hotfix/'
git config --global alias.sync '!git checkout main && git pull origin main'
```

#### Shell Aliases
```bash
# Add to ~/.bashrc or ~/.zshrc
alias gt="go test ./..."
alias gl="golangci-lint run"
alias gb="go build ./cmd/api"
alias gm="make"
alias gmt="make test-all"

# Git shortcuts
alias gs="git status"
alias ga="git add"
alias gc="git commit"
alias gp="git push"
alias gpl="git pull"
```

This comprehensive Git and code review practice guide ensures consistent, high-quality development practices that support ZTDP's AI-native platform architecture while maintaining clean code standards and efficient collaboration workflows.
