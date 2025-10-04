package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainProvider defines the DNS provider type
type DomainProvider string

const (
	DomainProviderCloudflare DomainProvider = "cloudflare"
	DomainProviderRoute53    DomainProvider = "route53"
	DomainProviderManual     DomainProvider = "manual"
)

// CertificateStatus defines the status of a certificate
type CertificateStatus string

const (
	CertificateStatusPending CertificateStatus = "pending"
	CertificateStatusActive  CertificateStatus = "active"
	CertificateStatusFailed  CertificateStatus = "failed"
	CertificateStatusExpired CertificateStatus = "expired"
)

// ChallengeType defines the ACME challenge type
type ChallengeType string

const (
	ChallengeTypeHTTP01 ChallengeType = "http-01"
	ChallengeTypeDNS01  ChallengeType = "dns-01"
)

// JobType for domain provisioning
const (
	JobTypeDomainProvision    JobType = "domain_provision"
	JobTypeCertificateRequest JobType = "certificate_request"
	JobTypeDomainDelete       JobType = "domain_delete"
)

// Domain represents a domain managed by OneClick
type Domain struct {
	ID             uuid.UUID         `json:"id"`
	AppID          uuid.UUID         `json:"app_id"`
	Domain         string            `json:"domain"`
	Provider       DomainProvider    `json:"provider"`
	ProviderConfig ProviderConfig    `json:"provider_config"`
	CertStatus     CertificateStatus `json:"cert_status"`
	CertSecretName string            `json:"cert_secret_name,omitempty"`
	ChallengeType  ChallengeType     `json:"challenge_type"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ProviderConfig contains configuration for DNS providers
type ProviderConfig struct {
	// Cloudflare configuration
	CloudflareToken  string `json:"cloudflare_token,omitempty"`
	CloudflareZoneID string `json:"cloudflare_zone_id,omitempty"`

	// Route53 configuration
	Route53AccessKey string `json:"route53_access_key,omitempty"`
	Route53SecretKey string `json:"route53_secret_key,omitempty"`
	Route53Region    string `json:"route53_region,omitempty"`

	// Manual configuration
	DNSInstructions string `json:"dns_instructions,omitempty"`

	// Common settings
	Settings map[string]string `json:"settings,omitempty"`
}

// CertificateInfo represents certificate information
type CertificateInfo struct {
	Status       CertificateStatus `json:"status"`
	SecretName   string            `json:"secret_name,omitempty"`
	IssuedAt     *time.Time        `json:"issued_at,omitempty"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	Error        string            `json:"error,omitempty"`
	DNSChallenge *DNSChallenge     `json:"dns_challenge,omitempty"`
}

// DNSChallenge represents DNS challenge information
type DNSChallenge struct {
	RecordName   string `json:"record_name"`
	RecordType   string `json:"record_type"`
	RecordValue  string `json:"record_value"`
	Instructions string `json:"instructions,omitempty"`
}

// Request/Response DTOs

// CreateDomainRequest is the request body for creating a domain
type CreateDomainRequest struct {
	Domain         string         `json:"domain" validate:"required,min=3,max=255"`
	Provider       DomainProvider `json:"provider" validate:"required,oneof=cloudflare route53 manual"`
	ProviderConfig ProviderConfig `json:"provider_config,omitempty"`
	ChallengeType  ChallengeType  `json:"challenge_type" validate:"required,oneof=http-01 dns-01"`
}

// DomainResponse is the response body for domain details
type DomainResponse struct {
	ID              uuid.UUID         `json:"id"`
	AppID           uuid.UUID         `json:"app_id"`
	Domain          string            `json:"domain"`
	Provider        DomainProvider    `json:"provider"`
	ProviderConfig  ProviderConfig    `json:"provider_config"`
	CertStatus      CertificateStatus `json:"cert_status"`
	CertSecretName  string            `json:"cert_secret_name,omitempty"`
	ChallengeType   ChallengeType     `json:"challenge_type"`
	CertificateInfo *CertificateInfo  `json:"certificate_info,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// CertificateRequestResponse is the response for certificate requests
type CertificateRequestResponse struct {
	DomainID     uuid.UUID         `json:"domain_id"`
	Status       CertificateStatus `json:"status"`
	Message      string            `json:"message"`
	DNSChallenge *DNSChallenge     `json:"dns_challenge,omitempty"`
}

// Conversion methods

// ToResponse converts Domain to DomainResponse
func (d *Domain) ToResponse(certInfo *CertificateInfo) DomainResponse {
	// Mask sensitive provider configuration
	config := d.ProviderConfig
	if config.CloudflareToken != "" {
		config.CloudflareToken = "***MASKED***"
	}
	if config.Route53AccessKey != "" {
		config.Route53AccessKey = "***MASKED***"
	}
	if config.Route53SecretKey != "" {
		config.Route53SecretKey = "***MASKED***"
	}

	return DomainResponse{
		ID:              d.ID,
		AppID:           d.AppID,
		Domain:          d.Domain,
		Provider:        d.Provider,
		ProviderConfig:  config,
		CertStatus:      d.CertStatus,
		CertSecretName:  d.CertSecretName,
		ChallengeType:   d.ChallengeType,
		CertificateInfo: certInfo,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

// Validation helpers

// IsValidDomainProvider checks if the domain provider is valid
func IsValidDomainProvider(p string) bool {
	switch p {
	case string(DomainProviderCloudflare), string(DomainProviderRoute53), string(DomainProviderManual):
		return true
	default:
		return false
	}
}

// IsValidCertificateStatus checks if the certificate status is valid
func IsValidCertificateStatus(s string) bool {
	switch s {
	case string(CertificateStatusPending), string(CertificateStatusActive),
		string(CertificateStatusFailed), string(CertificateStatusExpired):
		return true
	default:
		return false
	}
}

// IsValidChallengeType checks if the challenge type is valid
func IsValidChallengeType(t string) bool {
	switch t {
	case string(ChallengeTypeHTTP01), string(ChallengeTypeDNS01):
		return true
	default:
		return false
	}
}

// Helper methods

// RequiresDNSProvider checks if the domain requires DNS provider configuration
func (d *Domain) RequiresDNSProvider() bool {
	return d.Provider != DomainProviderManual && d.ChallengeType == ChallengeTypeDNS01
}

// IsDNSProviderConfigured checks if DNS provider is properly configured
func (d *Domain) IsDNSProviderConfigured() bool {
	switch d.Provider {
	case DomainProviderCloudflare:
		return d.ProviderConfig.CloudflareToken != "" && d.ProviderConfig.CloudflareZoneID != ""
	case DomainProviderRoute53:
		return d.ProviderConfig.Route53AccessKey != "" && d.ProviderConfig.Route53SecretKey != ""
	case DomainProviderManual:
		return true // Manual doesn't need API credentials
	default:
		return false
	}
}

// GetDNSInstructions returns DNS challenge instructions for manual providers
func (d *Domain) GetDNSInstructions() string {
	if d.Provider == DomainProviderManual {
		return d.ProviderConfig.DNSInstructions
	}
	return ""
}
