package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type DomainService interface {
	CreateDomain(ctx context.Context, userID, appID uuid.UUID, req domain.CreateDomainRequest) (*domain.DomainResponse, error)
	GetDomainsByApp(ctx context.Context, userID, appID uuid.UUID) ([]domain.DomainResponse, error)
	GetDomain(ctx context.Context, userID, domainID uuid.UUID) (*domain.DomainResponse, error)
	RequestCertificate(ctx context.Context, userID, domainID uuid.UUID) (*domain.CertificateRequestResponse, error)
	GetCertificateStatus(ctx context.Context, userID, domainID uuid.UUID) (*domain.CertificateInfo, error)
	DeleteDomain(ctx context.Context, userID, domainID uuid.UUID) error
}

type domainService struct {
	domainRepo repo.DomainRepository
	appRepo    repo.ApplicationRepository
	jobRepo    repo.JobRepository
	orgRepo    repo.OrganizationRepository
	crypto     *crypto.Crypto
	logger     *zap.Logger
}

func NewDomainService(
	domainRepo repo.DomainRepository,
	appRepo repo.ApplicationRepository,
	jobRepo repo.JobRepository,
	orgRepo repo.OrganizationRepository,
	crypto *crypto.Crypto,
	logger *zap.Logger,
) DomainService {
	return &domainService{
		domainRepo: domainRepo,
		appRepo:    appRepo,
		jobRepo:    jobRepo,
		orgRepo:    orgRepo,
		crypto:     crypto,
		logger:     logger,
	}
}

func (s *domainService) CreateDomain(ctx context.Context, userID, appID uuid.UUID, req domain.CreateDomainRequest) (*domain.DomainResponse, error) {
	// Verify user has access to the application
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", appID.String()))
		return nil, errors.New("application not found or access denied")
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user is member of the organization and has admin/owner role
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return nil, errors.New("unauthorized: only organization owners or admins can manage domains")
	}

	// Check if domain already exists for this app
	existingDomain, err := s.domainRepo.GetDomainByDomainInApp(ctx, appID, req.Domain)
	if err != nil {
		s.logger.Error("Failed to check for existing domain", zap.Error(err), zap.String("appID", appID.String()), zap.String("domain", req.Domain))
		return nil, errors.New("failed to check for existing domain")
	}
	if existingDomain != nil {
		return nil, errors.New("domain already exists for this application")
	}

	// Encrypt provider configuration if provided
	providerConfig := req.ProviderConfig
	if providerConfig.CloudflareToken != "" {
		encryptedToken, err := s.crypto.EncryptString(providerConfig.CloudflareToken)
		if err != nil {
			s.logger.Error("Failed to encrypt Cloudflare token", zap.Error(err))
			return nil, errors.New("failed to encrypt provider credentials")
		}
		providerConfig.CloudflareToken = string(encryptedToken)
	}

	if providerConfig.Route53AccessKey != "" {
		encryptedAccessKey, err := s.crypto.EncryptString(providerConfig.Route53AccessKey)
		if err != nil {
			s.logger.Error("Failed to encrypt Route53 access key", zap.Error(err))
			return nil, errors.New("failed to encrypt provider credentials")
		}
		providerConfig.Route53AccessKey = string(encryptedAccessKey)
	}

	if providerConfig.Route53SecretKey != "" {
		encryptedSecretKey, err := s.crypto.EncryptString(providerConfig.Route53SecretKey)
		if err != nil {
			s.logger.Error("Failed to encrypt Route53 secret key", zap.Error(err))
			return nil, errors.New("failed to encrypt provider credentials")
		}
		providerConfig.Route53SecretKey = string(encryptedSecretKey)
	}

	// Create domain record
	domainRecord := &domain.Domain{
		AppID:          appID,
		Domain:         req.Domain,
		Provider:       req.Provider,
		ProviderConfig: providerConfig,
		CertStatus:     domain.CertificateStatusPending,
		ChallengeType:  req.ChallengeType,
	}

	createdDomain, err := s.domainRepo.CreateDomain(ctx, domainRecord)
	if err != nil {
		s.logger.Error("Failed to create domain", zap.Error(err), zap.String("appID", appID.String()))
		return nil, errors.New("failed to create domain")
	}

	// Create background job for domain provisioning
	jobPayload := domain.JobPayload{
		Config: map[string]interface{}{
			"domain_id":      createdDomain.ID.String(),
			"domain":         req.Domain,
			"provider":       req.Provider,
			"challenge_type": req.ChallengeType,
			"app_id":         appID.String(),
		},
	}

	job := &domain.Job{
		OrgID:   app.OrgID,
		Type:    domain.JobTypeDomainProvision,
		Status:  domain.JobStatusPending,
		Payload: jobPayload,
	}

	_, err = s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create domain provisioning job", zap.Error(err), zap.String("domainID", createdDomain.ID.String()))
		// Don't fail the domain creation if job creation fails
		s.logger.Warn("Domain created but job creation failed", zap.String("domainID", createdDomain.ID.String()))
	}

	s.logger.Info("Domain created successfully", zap.String("domainID", createdDomain.ID.String()), zap.String("domain", req.Domain))

	response := createdDomain.ToResponse(nil)
	return &response, nil
}

func (s *domainService) GetDomainsByApp(ctx context.Context, userID, appID uuid.UUID) ([]domain.DomainResponse, error) {
	// Verify user has access to the application
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", appID.String()))
		return nil, errors.New("application not found or access denied")
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	domains, err := s.domainRepo.GetDomainsByAppID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get domains by app ID", zap.Error(err), zap.String("appID", appID.String()))
		return nil, errors.New("failed to retrieve domains")
	}

	var responses []domain.DomainResponse
	for _, d := range domains {
		// For now, we don't have certificate info, so pass nil
		// In a real implementation, you would fetch certificate info from Kubernetes
		responses = append(responses, d.ToResponse(nil))
	}

	return responses, nil
}

func (s *domainService) GetDomain(ctx context.Context, userID, domainID uuid.UUID) (*domain.DomainResponse, error) {
	// Get domain
	domainRecord, err := s.domainRepo.GetDomainByID(ctx, domainID)
	if err != nil {
		s.logger.Error("Failed to get domain by ID", zap.Error(err), zap.String("domainID", domainID.String()))
		return nil, errors.New("failed to retrieve domain")
	}
	if domainRecord == nil {
		return nil, errors.New("domain not found")
	}

	// Verify user has access to the application
	app, err := s.appRepo.GetApplicationByID(ctx, domainRecord.AppID)
	if err != nil {
		s.logger.Error("Failed to get application for domain", zap.Error(err), zap.String("appID", domainRecord.AppID.String()))
		return nil, errors.New("application not found for domain")
	}
	if app == nil {
		return nil, errors.New("application not found for domain")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	response := domainRecord.ToResponse(nil)
	return &response, nil
}

func (s *domainService) RequestCertificate(ctx context.Context, userID, domainID uuid.UUID) (*domain.CertificateRequestResponse, error) {
	// Get domain
	domainRecord, err := s.domainRepo.GetDomainByID(ctx, domainID)
	if err != nil {
		s.logger.Error("Failed to get domain by ID for certificate request", zap.Error(err), zap.String("domainID", domainID.String()))
		return nil, errors.New("failed to retrieve domain")
	}
	if domainRecord == nil {
		return nil, errors.New("domain not found")
	}

	// Verify user has access to the application
	app, err := s.appRepo.GetApplicationByID(ctx, domainRecord.AppID)
	if err != nil {
		s.logger.Error("Failed to get application for domain", zap.Error(err), zap.String("appID", domainRecord.AppID.String()))
		return nil, errors.New("application not found for domain")
	}
	if app == nil {
		return nil, errors.New("application not found for domain")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Create background job for certificate request
	jobPayload := domain.JobPayload{
		Config: map[string]interface{}{
			"domain_id":      domainID.String(),
			"domain":         domainRecord.Domain,
			"provider":       domainRecord.Provider,
			"challenge_type": domainRecord.ChallengeType,
			"app_id":         domainRecord.AppID.String(),
		},
	}

	job := &domain.Job{
		OrgID:   app.OrgID,
		Type:    domain.JobTypeCertificateRequest,
		Status:  domain.JobStatusPending,
		Payload: jobPayload,
	}

	_, err = s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create certificate request job", zap.Error(err), zap.String("domainID", domainID.String()))
		return nil, errors.New("failed to request certificate")
	}

	// Update domain status to pending
	_, err = s.domainRepo.UpdateDomainCertStatus(ctx, domainID, domain.CertificateStatusPending)
	if err != nil {
		s.logger.Error("Failed to update domain certificate status", zap.Error(err), zap.String("domainID", domainID.String()))
		// Don't fail the request if status update fails
	}

	s.logger.Info("Certificate request initiated", zap.String("domainID", domainID.String()))

	// Generate DNS challenge instructions if manual provider
	var dnsChallenge *domain.DNSChallenge
	if domainRecord.Provider == domain.DomainProviderManual && domainRecord.ChallengeType == domain.ChallengeTypeDNS01 {
		dnsChallenge = &domain.DNSChallenge{
			RecordName:   fmt.Sprintf("_acme-challenge.%s", domainRecord.Domain),
			RecordType:   "TXT",
			RecordValue:  "placeholder-challenge-value", // This would be generated by cert-manager
			Instructions: domainRecord.GetDNSInstructions(),
		}
	}

	return &domain.CertificateRequestResponse{
		DomainID:     domainID,
		Status:       domain.CertificateStatusPending,
		Message:      "Certificate request initiated",
		DNSChallenge: dnsChallenge,
	}, nil
}

func (s *domainService) GetCertificateStatus(ctx context.Context, userID, domainID uuid.UUID) (*domain.CertificateInfo, error) {
	// Get domain
	domainRecord, err := s.domainRepo.GetDomainByID(ctx, domainID)
	if err != nil {
		s.logger.Error("Failed to get domain by ID for certificate status", zap.Error(err), zap.String("domainID", domainID.String()))
		return nil, errors.New("failed to retrieve domain")
	}
	if domainRecord == nil {
		return nil, errors.New("domain not found")
	}

	// Verify user has access to the application
	app, err := s.appRepo.GetApplicationByID(ctx, domainRecord.AppID)
	if err != nil {
		s.logger.Error("Failed to get application for domain", zap.Error(err), zap.String("appID", domainRecord.AppID.String()))
		return nil, errors.New("application not found for domain")
	}
	if app == nil {
		return nil, errors.New("application not found for domain")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// For MVP, return basic certificate info from database
	// In a real implementation, you would fetch this from Kubernetes cert-manager
	certInfo := &domain.CertificateInfo{
		Status:     domainRecord.CertStatus,
		SecretName: domainRecord.CertSecretName,
	}

	// Add DNS challenge info if manual provider and pending status
	if domainRecord.Provider == domain.DomainProviderManual &&
		domainRecord.ChallengeType == domain.ChallengeTypeDNS01 &&
		domainRecord.CertStatus == domain.CertificateStatusPending {
		certInfo.DNSChallenge = &domain.DNSChallenge{
			RecordName:   fmt.Sprintf("_acme-challenge.%s", domainRecord.Domain),
			RecordType:   "TXT",
			RecordValue:  "placeholder-challenge-value", // This would be generated by cert-manager
			Instructions: domainRecord.GetDNSInstructions(),
		}
	}

	return certInfo, nil
}

func (s *domainService) DeleteDomain(ctx context.Context, userID, domainID uuid.UUID) error {
	// Get domain
	domainRecord, err := s.domainRepo.GetDomainByID(ctx, domainID)
	if err != nil {
		s.logger.Error("Failed to get domain by ID for deletion", zap.Error(err), zap.String("domainID", domainID.String()))
		return errors.New("failed to retrieve domain")
	}
	if domainRecord == nil {
		return errors.New("domain not found")
	}

	// Verify user has access to the application
	app, err := s.appRepo.GetApplicationByID(ctx, domainRecord.AppID)
	if err != nil {
		s.logger.Error("Failed to get application for domain", zap.Error(err), zap.String("appID", domainRecord.AppID.String()))
		return errors.New("application not found for domain")
	}
	if app == nil {
		return errors.New("application not found for domain")
	}

	// Check if user is member of the organization and has admin/owner role
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return fmt.Errorf("failed to check organization role: %w", err)
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return errors.New("unauthorized: only organization owners or admins can delete domains")
	}

	// Delete domain record
	err = s.domainRepo.DeleteDomain(ctx, domainID)
	if err != nil {
		s.logger.Error("Failed to delete domain", zap.Error(err), zap.String("domainID", domainID.String()))
		return errors.New("failed to delete domain")
	}

	s.logger.Info("Domain deleted successfully", zap.String("domainID", domainID.String()))
	return nil
}
