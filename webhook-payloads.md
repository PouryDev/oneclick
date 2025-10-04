# OneClick Webhook Payload Examples

This directory contains sample webhook payloads for different Git providers that OneClick supports.

## GitHub Webhook Payload

### Push Event (Standard)

```json
{
  "ref": "refs/heads/main",
  "repository": {
    "full_name": "user/example-repo",
    "clone_url": "https://github.com/user/example-repo.git",
    "ssh_url": "git@github.com:user/example-repo.git"
  },
  "commits": [
    {
      "id": "abc123def456",
      "message": "Add new feature",
      "author": {
        "name": "John Doe",
        "email": "john@example.com"
      }
    }
  ],
  "head_commit": {
    "id": "abc123def456",
    "message": "Add new feature",
    "author": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
}
```

### Push Event (Multiple Commits)

```json
{
  "ref": "refs/heads/feature-branch",
  "repository": {
    "full_name": "company/project",
    "clone_url": "https://github.com/company/project.git",
    "ssh_url": "git@github.com:company/project.git"
  },
  "commits": [
    {
      "id": "commit1",
      "message": "Initial commit",
      "author": {
        "name": "Alice Smith",
        "email": "alice@company.com"
      }
    },
    {
      "id": "commit2",
      "message": "Add tests",
      "author": {
        "name": "Bob Johnson",
        "email": "bob@company.com"
      }
    }
  ],
  "head_commit": {
    "id": "commit2",
    "message": "Add tests",
    "author": {
      "name": "Bob Johnson",
      "email": "bob@company.com"
    }
  }
}
```

## GitLab Webhook Payload

### Push Event

```json
{
  "ref": "refs/heads/main",
  "repository": {
    "name": "example-repo",
    "url": "https://gitlab.com/user/example-repo.git",
    "homepage": "https://gitlab.com/user/example-repo"
  },
  "commits": [
    {
      "id": "def456ghi789",
      "message": "Update documentation",
      "author": {
        "name": "Jane Doe",
        "email": "jane@example.com"
      }
    }
  ]
}
```

### Merge Request Event

```json
{
  "object_kind": "merge_request",
  "object_attributes": {
    "id": 123,
    "title": "Add new feature",
    "description": "This MR adds a new feature",
    "state": "opened",
    "source_branch": "feature-branch",
    "target_branch": "main",
    "source": {
      "name": "example-repo",
      "url": "https://gitlab.com/user/example-repo.git"
    }
  }
}
```

## Gitea Webhook Payload

### Push Event

```json
{
  "ref": "refs/heads/main",
  "repository": {
    "full_name": "user/example-repo",
    "clone_url": "https://gitea.example.com/user/example-repo.git",
    "ssh_url": "git@gitea.example.com:user/example-repo.git"
  },
  "commits": [
    {
      "id": "ghi789jkl012",
      "message": "Fix bug in authentication",
      "author": {
        "name": "Mike Wilson",
        "email": "mike@example.com"
      }
    }
  ]
}
```

### Pull Request Event

```json
{
  "action": "opened",
  "number": 1,
  "pull_request": {
    "id": 1,
    "title": "Add new feature",
    "body": "This PR adds a new feature",
    "state": "open",
    "head": {
      "ref": "feature-branch",
      "sha": "abc123"
    },
    "base": {
      "ref": "main",
      "sha": "def456"
    }
  },
  "repository": {
    "full_name": "user/example-repo",
    "clone_url": "https://gitea.example.com/user/example-repo.git"
  }
}
```

## Webhook Headers

### GitHub Headers

```
X-Hub-Signature-256: sha256=abc123def456...
X-Hub-Signature: sha1=def456ghi789... (legacy)
X-GitHub-Event: push
X-GitHub-Delivery: 12345-67890-abcdef
```

### GitLab Headers

```
X-Gitlab-Signature: sha256=abc123def456...
X-Gitlab-Event: Push Hook
X-Gitlab-Token: your-secret-token
```

### Gitea Headers

```
X-Gitea-Signature: sha256=abc123def456...
X-Gitea-Event: push
X-Gitea-Delivery: 12345-67890-abcdef
```

## Testing Webhooks

### Using curl to test GitHub webhook

```bash
curl -X POST http://localhost:8080/hooks/git?provider=github \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=your-signature" \
  -d @github-push-payload.json
```

### Using curl to test GitLab webhook

```bash
curl -X POST http://localhost:8080/hooks/git?provider=gitlab \
  -H "Content-Type: application/json" \
  -H "X-Gitlab-Signature: sha256=your-signature" \
  -d @gitlab-push-payload.json
```

### Using curl to test Gitea webhook

```bash
curl -X POST http://localhost:8080/hooks/git?provider=gitea \
  -H "Content-Type: application/json" \
  -H "X-Gitea-Signature: sha256=your-signature" \
  -d @gitea-push-payload.json
```

## Signature Verification

OneClick supports signature verification for all three providers:

- **GitHub**: Uses HMAC-SHA256 (preferred) or HMAC-SHA1 (legacy)
- **GitLab**: Uses HMAC-SHA256
- **Gitea**: Uses HMAC-SHA256

The webhook secret should be provided as a query parameter:

```
POST /hooks/git?provider=github&secret=your-webhook-secret
```

## Pipeline Integration

When a webhook is received, OneClick will:

1. Verify the signature (if secret is provided)
2. Parse the payload to extract repository information
3. Find the corresponding repository in the database
4. Create a pipeline record (when applications are implemented)
5. Return HTTP 202 Accepted

The pipeline record will contain:

- Repository ID
- Branch name
- Commit SHA
- Commit message
- Author information
- Full webhook payload
