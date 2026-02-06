---
id: task-001.02
title: Create error handling infrastructure with oops
status: To Do
assignee: []
created_date: '2026-02-06 01:57'
labels:
  - errors
  - phase-1
dependencies: []
parent_task_id: task-1
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Set up structured error handling using samber/oops library.

## Files to Create
- `internal/errors/domains.go` - Error domain constants
- `internal/errors/codes.go` - Error code constants
- `internal/errors/errors.go` - Error builders

## Implementation Steps

### 1. Create internal/errors/domains.go
```go
package errors

const (
    DomainRegistry  = "registry"
    DomainConfig    = "config"
    DomainHTTP      = "http"
    DomainGitHub    = "github"
    DomainGenerator = "generator"
)
```

### 2. Create internal/errors/codes.go
```go
package errors

const (
    // Registry errors
    CodeNotFound     = "not_found"
    CodeInvalidEntry = "invalid_entry"

    // HTTP errors
    CodeNetworkFailure = "network_failure"
    CodeRateLimited    = "rate_limited"
    CodeTimeout        = "timeout"

    // Config errors
    CodeConfigLoad  = "config_load"
    CodeConfigParse = "config_parse"
    CodeConfigWrite = "config_write"

    // GitHub errors
    CodeRepoNotFound = "repo_not_found"
    CodePathNotFound = "path_not_found"
    CodeAPIError     = "api_error"
)
```

### 3. Create internal/errors/errors.go
```go
package errors

import "github.com/samber/oops"

var (
    RegistryErr  = oops.In(DomainRegistry).Tags("registry")
    HTTPErr      = oops.In(DomainHTTP).Tags("http", "network")
    ConfigErr    = oops.In(DomainConfig).Tags("config")
    GitHubErr    = oops.In(DomainGitHub).Tags("github", "api")
    GeneratorErr = oops.In(DomainGenerator).Tags("generator")
)
```

## Quality Checks
```bash
task lint
task build
```

## Commit
```bash
git add internal/errors/
git commit -m "feat: add error handling infrastructure with oops"
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 internal/errors/domains.go exists with 5 domain constants
- [ ] #2 internal/errors/codes.go exists with all error codes
- [ ] #3 internal/errors/errors.go exists with domain-specific error builders
- [ ] #4 task lint passes with no errors
- [ ] #5 task build passes
<!-- AC:END -->
