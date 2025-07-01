# GitHub Repository Setup

## Automatic Testing Setup

This repository is configured with GitHub Actions for automatic testing. The workflow runs:
- On push to `main` and `develop` branches
- On Pull Request creation targeting `main` and `develop` branches

## Branch Protection Rules

To prevent merging PRs without passing tests, configure Branch Protection Rules:

### Setup Steps:

1. Go to **Settings** → **Branches** in your GitHub repository
2. Click **Add rule** for the `main` branch
3. Configure the following parameters:

#### Basic Settings:
- ✅ **Require pull request reviews before merging**
- ✅ **Require status checks to pass before merging**
- ✅ **Require branches to be up to date before merging**

#### Status Checks:
In the "Status checks" section, add:
- ✅ **test (1.19)** 
- ✅ **test (1.20)**
- ✅ **test (1.21)**
- ✅ **lint**

#### Additional Settings:
- ✅ **Restrict pushes that create files that exceed the path length limit**
- ✅ **Require linear history** (optional)
- ✅ **Include administrators** (so rules apply to everyone)

### Result:
After setup:
- Cannot merge PR until all tests pass
- Cannot push directly to main
- All changes must go through PRs

## Alternative Setup via GitHub CLI:

If you have GitHub CLI installed, you can configure via command:

```bash
gh api repos/Winter0rbit/yamler/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["test (1.19)","test (1.20)","test (1.21)","lint"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1}' \
  --field restrictions=null
```

## Verifying Setup:

1. Create a test PR with changes  
2. Ensure status checks appear
3. Try to merge before tests complete - should be blocked
4. After tests pass - merge should be allowed 