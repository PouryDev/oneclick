package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Repository represents a Git repository
type Repository struct {
	ID            uuid.UUID       `json:"id"`
	OrgID         uuid.UUID       `json:"org_id"`
	Type          string          `json:"type"`
	URL           string          `json:"url"`
	DefaultBranch string          `json:"default_branch"`
	Config        json.RawMessage `json:"config"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// RepositorySummary represents a repository in list views
type RepositorySummary struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	URL           string    `json:"url"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RepositoryConfig represents repository configuration
type RepositoryConfig struct {
	Token     string `json:"token,omitempty"`      // Encrypted access token
	WebhookID string `json:"webhook_id,omitempty"` // Webhook ID for deletion
	Secret    string `json:"secret,omitempty"`     // Webhook secret for signature verification
}

// Request/Response DTOs
type CreateRepositoryRequest struct {
	Type          string `json:"type" validate:"required,oneof=github gitlab gitea"`
	URL           string `json:"url" validate:"required,url"`
	DefaultBranch string `json:"default_branch" validate:"required"`
	Token         string `json:"token,omitempty"` // Optional access token
}

type RepositoryResponse struct {
	ID            uuid.UUID       `json:"id"`
	Type          string          `json:"type"`
	URL           string          `json:"url"`
	DefaultBranch string          `json:"default_branch"`
	Config        json.RawMessage `json:"config"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// Webhook payload structures
type WebhookPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		SSHURL   string `json:"ssh_url"`
	} `json:"repository"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"head_commit"`
}

// GitLab webhook payload structure
type GitLabWebhookPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Homepage string `json:"homepage"`
	} `json:"repository"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

// Gitea webhook payload structure
type GiteaWebhookPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		SSHURL   string `json:"ssh_url"`
	} `json:"repository"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
}

// Valid repository types
const (
	RepoTypeGitHub = "github"
	RepoTypeGitLab = "gitlab"
	RepoTypeGitea  = "gitea"
)

// ValidStatuses returns a list of valid repository types
func ValidRepositoryTypes() []string {
	return []string{RepoTypeGitHub, RepoTypeGitLab, RepoTypeGitea}
}

// IsValidRepositoryType checks if a type is valid
func IsValidRepositoryType(repoType string) bool {
	for _, validType := range ValidRepositoryTypes() {
		if repoType == validType {
			return true
		}
	}
	return false
}

// ToResponse converts a Repository to RepositoryResponse
func (r *Repository) ToResponse() RepositoryResponse {
	return RepositoryResponse{
		ID:            r.ID,
		Type:          r.Type,
		URL:           r.URL,
		DefaultBranch: r.DefaultBranch,
		Config:        r.Config,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

// ToSummary converts a Repository to RepositorySummary
func (r *Repository) ToSummary() RepositorySummary {
	return RepositorySummary{
		ID:            r.ID,
		Type:          r.Type,
		URL:           r.URL,
		DefaultBranch: r.DefaultBranch,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

// GetConfig parses the repository config JSON
func (r *Repository) GetConfig() (*RepositoryConfig, error) {
	var config RepositoryConfig
	if len(r.Config) > 0 {
		err := json.Unmarshal(r.Config, &config)
		if err != nil {
			return nil, err
		}
	}
	return &config, nil
}

// SetConfig sets the repository config JSON
func (r *Repository) SetConfig(config *RepositoryConfig) error {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	r.Config = configBytes
	return nil
}
