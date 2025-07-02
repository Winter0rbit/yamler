# GitHub Repository Setup

This document provides instructions for setting up GitHub Actions and branch protection for the Yamler project.

## GitHub Actions Setup

The repository includes automated testing workflow in `.github/workflows/test.yml` that:

- Tests on Go versions 1.19, 1.20, and 1.21
- Runs linting with golangci-lint
- Reports test coverage to Codecov
- Runs on pushes and pull requests to main/develop branches

## Branch Protection Setup

To ensure code quality and prevent direct pushes to main branch:

### Manual Setup (via GitHub Web UI):

1. Go to your repository on GitHub
2. Click **Settings** tab
3. Click **Branches** in the left sidebar
4. Click **Add rule** or **Add branch protection rule**
5. Configure the rule:
   - **Branch name pattern**: `main`
   - ✅ **Require status checks to pass before merging**
   - ✅ **Require branches to be up to date before merging**
   - Select required status checks:
     - `test (1.19)`
     - `test (1.20)` 
     - `test (1.21)`
     - `lint`
   - ✅ **Require pull request reviews before merging**
   - ✅ **Dismiss stale PR approvals when new commits are pushed**
   - ✅ **Require review from code owners** (if you have CODEOWNERS file)
   - ✅ **Restrict pushes that create files larger than 100 MB**
   - ✅ **Include administrators** (applies rules to admins too)

6. Click **Create** to save the rule

### Automated Setup (via GitHub CLI):

```bash
# Install GitHub CLI if not already installed
# macOS: brew install gh
# Linux: see https://github.com/cli/cli/blob/trunk/docs/install_linux.md

# Authenticate with GitHub
gh auth login

# Create branch protection rule
gh api repos/:owner/:repo/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["test (1.19)","test (1.20)","test (1.21)","lint"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
  --field restrictions=null
```

## Benefits

With this setup:
- No code can be merged to main without passing all tests
- All PRs require review before merging
- Automated testing ensures code quality
- Consistent formatting is enforced via linting
- Coverage reports help track test completeness 