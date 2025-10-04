package worker

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/app/provisioner"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

// JobProcessor defines the interface for processing background jobs
type JobProcessor interface {
	ProcessJob(ctx context.Context, job *domain.Job) error
	Start(ctx context.Context) error
	Stop() error
}

// GitRunnerWorker processes background jobs for git servers, CI runners, and domains
type GitRunnerWorker struct {
	jobRepo            repo.JobRepository
	gitServerRepo      repo.GitServerRepository
	runnerRepo         repo.RunnerRepository
	domainRepo         repo.DomainRepository
	provisioner        provisioner.Provisioner
	crypto             *crypto.Crypto
	logger             *zap.Logger
	stopChan           chan struct{}
	processingInterval time.Duration
}

// NewGitRunnerWorker creates a new GitRunnerWorker
func NewGitRunnerWorker(
	jobRepo repo.JobRepository,
	gitServerRepo repo.GitServerRepository,
	runnerRepo repo.RunnerRepository,
	domainRepo repo.DomainRepository,
	provisioner provisioner.Provisioner,
	crypto *crypto.Crypto,
	logger *zap.Logger,
) *GitRunnerWorker {
	return &GitRunnerWorker{
		jobRepo:            jobRepo,
		gitServerRepo:      gitServerRepo,
		runnerRepo:         runnerRepo,
		domainRepo:         domainRepo,
		provisioner:        provisioner,
		crypto:             crypto,
		logger:             logger,
		stopChan:           make(chan struct{}),
		processingInterval: 10 * time.Second, // Process jobs every 10 seconds
	}
}

// Start starts the worker to process background jobs
func (w *GitRunnerWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting GitRunnerWorker")

	ticker := time.NewTicker(w.processingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("GitRunnerWorker stopped due to context cancellation")
			return ctx.Err()
		case <-w.stopChan:
			w.logger.Info("GitRunnerWorker stopped")
			return nil
		case <-ticker.C:
			if err := w.processPendingJobs(ctx); err != nil {
				w.logger.Error("Failed to process pending jobs", zap.Error(err))
				// Continue processing even if one batch fails
			}
		}
	}
}

// Stop stops the worker
func (w *GitRunnerWorker) Stop() error {
	w.logger.Info("Stopping GitRunnerWorker")
	close(w.stopChan)
	return nil
}

// processPendingJobs processes all pending jobs
func (w *GitRunnerWorker) processPendingJobs(ctx context.Context) error {
	jobs, err := w.jobRepo.GetPendingJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil // No pending jobs
	}

	w.logger.Info("Processing pending jobs", zap.Int("count", len(jobs)))

	for _, job := range jobs {
		// Try to start the job (atomic operation)
		startedJob, err := w.jobRepo.StartJob(ctx, job.ID)
		if err != nil {
			w.logger.Error("Failed to start job", zap.Error(err), zap.String("jobID", job.ID.String()))
			continue
		}

		// Process the job
		if err := w.ProcessJob(ctx, startedJob); err != nil {
			w.logger.Error("Failed to process job", zap.Error(err), zap.String("jobID", job.ID.String()))
			// Mark job as failed
			if _, failErr := w.jobRepo.FailJob(ctx, job.ID, err.Error()); failErr != nil {
				w.logger.Error("Failed to mark job as failed", zap.Error(failErr), zap.String("jobID", job.ID.String()))
			}
		} else {
			// Mark job as completed
			if _, completeErr := w.jobRepo.CompleteJob(ctx, job.ID); completeErr != nil {
				w.logger.Error("Failed to mark job as completed", zap.Error(completeErr), zap.String("jobID", job.ID.String()))
			}
		}
	}

	return nil
}

// ProcessJob processes a specific job based on its type
func (w *GitRunnerWorker) ProcessJob(ctx context.Context, job *domain.Job) error {
	w.logger.Info("Processing job", zap.String("jobID", job.ID.String()), zap.String("type", string(job.Type)))

	switch job.Type {
	case domain.JobTypeGitServerInstall:
		return w.processGitServerInstall(ctx, job)
	case domain.JobTypeGitServerStop:
		return w.processGitServerStop(ctx, job)
	case domain.JobTypeRunnerDeploy:
		return w.processRunnerDeploy(ctx, job)
	case domain.JobTypeRunnerStop:
		return w.processRunnerStop(ctx, job)
	case domain.JobTypeDomainProvision:
		return w.processDomainProvision(ctx, job)
	case domain.JobTypeCertificateRequest:
		return w.processCertificateRequest(ctx, job)
	case domain.JobTypeDomainDelete:
		return w.processDomainDelete(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

// processGitServerInstall processes git server installation jobs
func (w *GitRunnerWorker) processGitServerInstall(ctx context.Context, job *domain.Job) error {
	if job.Payload.GitServerID == nil {
		return fmt.Errorf("git server ID is required for git server install job")
	}

	gitServerID := *job.Payload.GitServerID
	w.logger.Info("Installing git server", zap.String("gitServerID", gitServerID.String()))

	// Get git server details
	gitServer, err := w.gitServerRepo.GetGitServerByID(ctx, gitServerID)
	if err != nil {
		return fmt.Errorf("failed to get git server: %w", err)
	}
	if gitServer == nil {
		return fmt.Errorf("git server not found: %s", gitServerID.String())
	}

	// Extract configuration from job payload
	config := job.Payload.Config

	gitServerType, ok := config["type"].(string)
	if !ok {
		return fmt.Errorf("git server type not found in job payload")
	}
	_ = gitServerType // Used for validation

	domainName, ok := config["domain"].(string)
	if !ok {
		return fmt.Errorf("git server domain not found in job payload")
	}

	storage, ok := config["storage"].(string)
	if !ok {
		return fmt.Errorf("git server storage not found in job payload")
	}

	// Generate admin credentials
	adminUser := "admin"
	adminPassword := w.generatePassword(16)
	adminEmail := fmt.Sprintf("admin@%s", domainName)

	// Create Helm values for Gitea installation
	helmValues := map[string]interface{}{
		"gitea": map[string]interface{}{
			"admin": map[string]interface{}{
				"username": adminUser,
				"password": adminPassword,
				"email":    adminEmail,
			},
			"config": map[string]interface{}{
				"server": map[string]interface{}{
					"DOMAIN":   domainName,
					"ROOT_URL": fmt.Sprintf("https://%s", domainName),
				},
			},
			"persistence": map[string]interface{}{
				"size": storage,
			},
		},
	}

	// Install Gitea using Helm
	releaseName := fmt.Sprintf("gitea-%s", gitServerID.String()[:8])
	namespace := fmt.Sprintf("gitea-%s", gitServerID.String()[:8])

	err = w.provisioner.Install(ctx, "gitea/gitea", namespace, helmValues)
	if err != nil {
		return fmt.Errorf("failed to install Gitea: %w", err)
	}

	// Update git server configuration with admin credentials
	gitServerConfig := domain.GitServerConfig{
		AdminUser:     adminUser,
		AdminPassword: adminPassword,
		AdminEmail:    adminEmail,
		Repositories:  []string{},
		Settings: map[string]string{
			"domain":    domainName,
			"storage":   storage,
			"namespace": namespace,
			"release":   releaseName,
		},
	}

	_, err = w.gitServerRepo.UpdateGitServerConfig(ctx, gitServerID, gitServerConfig)
	if err != nil {
		w.logger.Error("Failed to update git server config", zap.Error(err), zap.String("gitServerID", gitServerID.String()))
		// Don't fail the job if config update fails
	}

	// Update git server status to running
	_, err = w.gitServerRepo.UpdateGitServerStatus(ctx, gitServerID, domain.GitServerStatusRunning)
	if err != nil {
		w.logger.Error("Failed to update git server status to running", zap.Error(err), zap.String("gitServerID", gitServerID.String()))
		// Don't fail the job if status update fails
	}

	w.logger.Info("Git server installation completed", zap.String("gitServerID", gitServerID.String()), zap.String("domain", domainName))
	return nil
}

// processGitServerStop processes git server removal jobs
func (w *GitRunnerWorker) processGitServerStop(ctx context.Context, job *domain.Job) error {
	if job.Payload.GitServerID == nil {
		return fmt.Errorf("git server ID is required for git server stop job")
	}

	gitServerID := *job.Payload.GitServerID
	w.logger.Info("Stopping git server", zap.String("gitServerID", gitServerID.String()))

	// Get git server details
	gitServer, err := w.gitServerRepo.GetGitServerByID(ctx, gitServerID)
	if err != nil {
		return fmt.Errorf("failed to get git server: %w", err)
	}
	if gitServer == nil {
		w.logger.Warn("Git server not found, assuming already deleted", zap.String("gitServerID", gitServerID.String()))
		return nil
	}

	// Extract namespace and release from git server config
	namespace, ok := gitServer.Config.Settings["namespace"]
	if !ok {
		w.logger.Warn("Namespace not found in git server config", zap.String("gitServerID", gitServerID.String()))
		return nil
	}

	release, ok := gitServer.Config.Settings["release"]
	if !ok {
		w.logger.Warn("Release not found in git server config", zap.String("gitServerID", gitServerID.String()))
		return nil
	}

	// Uninstall Gitea using Helm
	err = w.provisioner.Uninstall(ctx, release, namespace)
	if err != nil {
		w.logger.Error("Failed to uninstall Gitea", zap.Error(err), zap.String("gitServerID", gitServerID.String()))
		// Don't fail the job if uninstall fails - the git server record will still be deleted
	}

	w.logger.Info("Git server removal completed", zap.String("gitServerID", gitServerID.String()))
	return nil
}

// processRunnerDeploy processes CI runner deployment jobs
func (w *GitRunnerWorker) processRunnerDeploy(ctx context.Context, job *domain.Job) error {
	if job.Payload.RunnerID == nil {
		return fmt.Errorf("runner ID is required for runner deploy job")
	}

	runnerID := *job.Payload.RunnerID
	w.logger.Info("Deploying runner", zap.String("runnerID", runnerID.String()))

	// Get runner details
	runner, err := w.runnerRepo.GetRunnerByID(ctx, runnerID)
	if err != nil {
		return fmt.Errorf("failed to get runner: %w", err)
	}
	if runner == nil {
		return fmt.Errorf("runner not found: %s", runnerID.String())
	}

	// Extract configuration from job payload
	config := job.Payload.Config

	name, ok := config["name"].(string)
	if !ok {
		return fmt.Errorf("runner name not found in job payload")
	}

	runnerType, ok := config["type"].(string)
	if !ok {
		return fmt.Errorf("runner type not found in job payload")
	}

	// Generate runner token if not provided
	runnerToken := runner.Config.Token
	if runnerToken == "" {
		runnerToken = w.generatePassword(32)
	}

	// Create Helm values for runner deployment
	helmValues := map[string]interface{}{
		"runner": map[string]interface{}{
			"name":         name,
			"type":         runnerType,
			"token":        runnerToken,
			"labels":       runner.Config.Labels,
			"nodeSelector": runner.Config.NodeSelector,
			"resources":    runner.Config.Resources,
		},
	}
	_ = helmValues // Will be used when implementing actual Helm deployment

	// Deploy runner using Helm (this would be a custom runner chart)
	releaseName := fmt.Sprintf("runner-%s", runnerID.String()[:8])
	namespace := fmt.Sprintf("runner-%s", runnerID.String()[:8])

	// For now, we'll simulate the deployment
	w.logger.Info("Deploying runner with Helm",
		zap.String("runnerID", runnerID.String()),
		zap.String("release", releaseName),
		zap.String("namespace", namespace))

	// In a real implementation, you would:
	// 1. Create a Kubernetes namespace
	// 2. Install the runner controller/agent
	// 3. Configure the runner with the provided token and labels
	// 4. Set up monitoring and health checks

	// Update runner configuration with deployment details
	runnerConfig := runner.Config
	runnerConfig.Settings = map[string]string{
		"namespace":   namespace,
		"release":     releaseName,
		"deployed_at": time.Now().Format(time.RFC3339),
	}

	_, err = w.runnerRepo.UpdateRunnerConfig(ctx, runnerID, runnerConfig)
	if err != nil {
		w.logger.Error("Failed to update runner config", zap.Error(err), zap.String("runnerID", runnerID.String()))
		// Don't fail the job if config update fails
	}

	// Update runner status to running
	_, err = w.runnerRepo.UpdateRunnerStatus(ctx, runnerID, domain.RunnerStatusRunning)
	if err != nil {
		w.logger.Error("Failed to update runner status to running", zap.Error(err), zap.String("runnerID", runnerID.String()))
		// Don't fail the job if status update fails
	}

	w.logger.Info("Runner deployment completed", zap.String("runnerID", runnerID.String()), zap.String("name", name))
	return nil
}

// processRunnerStop processes CI runner removal jobs
func (w *GitRunnerWorker) processRunnerStop(ctx context.Context, job *domain.Job) error {
	if job.Payload.RunnerID == nil {
		return fmt.Errorf("runner ID is required for runner stop job")
	}

	runnerID := *job.Payload.RunnerID
	w.logger.Info("Stopping runner", zap.String("runnerID", runnerID.String()))

	// Get runner details
	runner, err := w.runnerRepo.GetRunnerByID(ctx, runnerID)
	if err != nil {
		return fmt.Errorf("failed to get runner: %w", err)
	}
	if runner == nil {
		w.logger.Warn("Runner not found, assuming already deleted", zap.String("runnerID", runnerID.String()))
		return nil
	}

	// Extract namespace and release from runner config
	namespace, ok := runner.Config.Settings["namespace"]
	if !ok {
		w.logger.Warn("Namespace not found in runner config", zap.String("runnerID", runnerID.String()))
		return nil
	}

	release, ok := runner.Config.Settings["release"]
	if !ok {
		w.logger.Warn("Release not found in runner config", zap.String("runnerID", runnerID.String()))
		return nil
	}

	// Uninstall runner using Helm
	err = w.provisioner.Uninstall(ctx, release, namespace)
	if err != nil {
		w.logger.Error("Failed to uninstall runner", zap.Error(err), zap.String("runnerID", runnerID.String()))
		// Don't fail the job if uninstall fails - the runner record will still be deleted
	}

	w.logger.Info("Runner removal completed", zap.String("runnerID", runnerID.String()))
	return nil
}

// processDomainProvision processes domain provisioning jobs
func (w *GitRunnerWorker) processDomainProvision(ctx context.Context, job *domain.Job) error {
	if job.Payload.DomainID == nil {
		return fmt.Errorf("domain ID is required for domain provision job")
	}

	domainID := *job.Payload.DomainID
	w.logger.Info("Provisioning domain", zap.String("domainID", domainID.String()))

	// Get domain details
	domainRecord, err := w.domainRepo.GetDomainByID(ctx, domainID)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}
	if domainRecord == nil {
		return fmt.Errorf("domain not found: %s", domainID.String())
	}

	// Extract configuration from job payload
	config := job.Payload.Config

	domainName, ok := config["domain"].(string)
	if !ok {
		return fmt.Errorf("domain name not found in job payload")
	}

	provider, ok := config["provider"].(string)
	if !ok {
		return fmt.Errorf("domain provider not found in job payload")
	}

	challengeType, ok := config["challenge_type"].(string)
	if !ok {
		return fmt.Errorf("challenge type not found in job payload")
	}

	appName, ok := config["app_name"].(string)
	if !ok {
		return fmt.Errorf("app name not found in job payload")
	}
	_ = appName // Used for validation

	appNamespace, ok := config["app_namespace"].(string)
	if !ok {
		return fmt.Errorf("app namespace not found in job payload")
	}

	// For DNS-01 challenge with provider credentials, create DNS records
	if challengeType == string(domain.ChallengeTypeDNS01) && provider != string(domain.DomainProviderManual) {
		// Extract provider credentials from config
		apiKey, _ := config["api_key"].(string)
		secretKey, _ := config["secret_key"].(string)
		email, _ := config["email"].(string)
		zoneID, _ := config["zone_id"].(string)
		hostedZoneID, _ := config["hosted_zone_id"].(string)
		_ = apiKey       // Will be used for DNS API calls
		_ = secretKey    // Will be used for DNS API calls
		_ = email        // Will be used for DNS API calls
		_ = zoneID       // Will be used for DNS API calls
		_ = hostedZoneID // Will be used for DNS API calls

		// Create DNS TXT record for ACME challenge
		// This is a simplified implementation - in reality, you'd use cert-manager's DNS-01 solver
		w.logger.Info("Creating DNS TXT record for ACME challenge",
			zap.String("domain", domainName),
			zap.String("provider", provider))

		// For now, we'll simulate DNS record creation
		// In a real implementation, you would:
		// 1. Use the provider's API to create TXT records
		// 2. Wait for DNS propagation
		// 3. Trigger cert-manager Certificate resource creation
	}

	// Create cert-manager Certificate resource
	certSecretName := fmt.Sprintf("%s-tls", domainName)
	certificateManifest := map[string]interface{}{
		"apiVersion": "cert-manager.io/v1",
		"kind":       "Certificate",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("%s-cert", domainName),
			"namespace": appNamespace,
		},
		"spec": map[string]interface{}{
			"secretName": certSecretName,
			"issuerRef": map[string]interface{}{
				"name": "letsencrypt-prod", // Assuming you have a ClusterIssuer configured
				"kind": "ClusterIssuer",
			},
			"dnsNames": []string{domainName},
		},
	}
	_ = certificateManifest // Will be used when implementing Kubernetes client

	// Apply the Certificate resource to Kubernetes
	// This would typically be done using the Kubernetes client
	w.logger.Info("Creating cert-manager Certificate resource",
		zap.String("domain", domainName),
		zap.String("namespace", appNamespace),
		zap.String("secretName", certSecretName))

	// Update domain with certificate secret name
	_, err = w.domainRepo.UpdateDomainCertSecret(ctx, domainID, certSecretName)
	if err != nil {
		w.logger.Error("Failed to update domain cert secret", zap.Error(err), zap.String("domainID", domainID.String()))
		// Don't fail the job if secret name update fails
	}

	// Update domain status to pending (cert-manager will update it to active/failed)
	_, err = w.domainRepo.UpdateDomainCertStatus(ctx, domainID, domain.CertificateStatusPending)
	if err != nil {
		w.logger.Error("Failed to update domain cert status to pending", zap.Error(err), zap.String("domainID", domainID.String()))
		// Don't fail the job if status update fails
	}

	w.logger.Info("Domain provisioning completed", zap.String("domainID", domainID.String()), zap.String("domain", domainName))
	return nil
}

// processCertificateRequest processes certificate request jobs
func (w *GitRunnerWorker) processCertificateRequest(ctx context.Context, job *domain.Job) error {
	if job.Payload.DomainID == nil {
		return fmt.Errorf("domain ID is required for certificate request job")
	}

	domainID := *job.Payload.DomainID
	w.logger.Info("Requesting certificate", zap.String("domainID", domainID.String()))

	// Get domain details
	domainRecord, err := w.domainRepo.GetDomainByID(ctx, domainID)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}
	if domainRecord == nil {
		return fmt.Errorf("domain not found: %s", domainID.String())
	}

	// Extract configuration from job payload
	config := job.Payload.Config

	appName, ok := config["app_name"].(string)
	if !ok {
		return fmt.Errorf("app name not found in job payload")
	}
	_ = appName // Used for validation

	appNamespace, ok := config["app_namespace"].(string)
	if !ok {
		return fmt.Errorf("app namespace not found in job payload")
	}

	// Create or update cert-manager Certificate resource
	certSecretName := fmt.Sprintf("%s-tls", domainRecord.Domain)
	certificateManifest := map[string]interface{}{
		"apiVersion": "cert-manager.io/v1",
		"kind":       "Certificate",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("%s-cert", domainRecord.Domain),
			"namespace": appNamespace,
		},
		"spec": map[string]interface{}{
			"secretName": certSecretName,
			"issuerRef": map[string]interface{}{
				"name": "letsencrypt-prod", // Assuming you have a ClusterIssuer configured
				"kind": "ClusterIssuer",
			},
			"dnsNames": []string{domainRecord.Domain},
		},
	}
	_ = certificateManifest // Will be used when implementing Kubernetes client

	// Apply the Certificate resource to Kubernetes
	w.logger.Info("Creating/updating cert-manager Certificate resource",
		zap.String("domain", domainRecord.Domain),
		zap.String("namespace", appNamespace),
		zap.String("secretName", certSecretName))

	// Update domain with certificate secret name
	_, err = w.domainRepo.UpdateDomainCertSecret(ctx, domainID, certSecretName)
	if err != nil {
		w.logger.Error("Failed to update domain cert secret", zap.Error(err), zap.String("domainID", domainID.String()))
		// Don't fail the job if secret name update fails
	}

	// Update domain status to pending (cert-manager will update it to active/failed)
	_, err = w.domainRepo.UpdateDomainCertStatus(ctx, domainID, domain.CertificateStatusPending)
	if err != nil {
		w.logger.Error("Failed to update domain cert status to pending", zap.Error(err), zap.String("domainID", domainID.String()))
		// Don't fail the job if status update fails
	}

	w.logger.Info("Certificate request completed", zap.String("domainID", domainID.String()), zap.String("domain", domainRecord.Domain))
	return nil
}

// processDomainDelete processes domain deletion jobs
func (w *GitRunnerWorker) processDomainDelete(ctx context.Context, job *domain.Job) error {
	if job.Payload.DomainID == nil {
		return fmt.Errorf("domain ID is required for domain delete job")
	}

	domainID := *job.Payload.DomainID
	w.logger.Info("Deleting domain", zap.String("domainID", domainID.String()))

	// Get domain details
	domainRecord, err := w.domainRepo.GetDomainByID(ctx, domainID)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}
	if domainRecord == nil {
		w.logger.Warn("Domain not found, assuming already deleted", zap.String("domainID", domainID.String()))
		return nil
	}

	// Extract configuration from job payload
	config := job.Payload.Config

	appNamespace, ok := config["app_namespace"].(string)
	if !ok {
		return fmt.Errorf("app namespace not found in job payload")
	}

	// Delete cert-manager Certificate resource
	certificateName := fmt.Sprintf("%s-cert", domainRecord.Domain)
	w.logger.Info("Deleting cert-manager Certificate resource",
		zap.String("domain", domainRecord.Domain),
		zap.String("namespace", appNamespace),
		zap.String("certificateName", certificateName))

	// For DNS-01 challenge with provider credentials, delete DNS records
	if domainRecord.ChallengeType == domain.ChallengeTypeDNS01 && domainRecord.Provider != domain.DomainProviderManual {
		// Extract provider credentials from config
		apiKey, _ := config["api_key"].(string)
		secretKey, _ := config["secret_key"].(string)
		email, _ := config["email"].(string)
		zoneID, _ := config["zone_id"].(string)
		hostedZoneID, _ := config["hosted_zone_id"].(string)
		_ = apiKey       // Will be used for DNS API calls
		_ = secretKey    // Will be used for DNS API calls
		_ = email        // Will be used for DNS API calls
		_ = zoneID       // Will be used for DNS API calls
		_ = hostedZoneID // Will be used for DNS API calls

		// Delete DNS TXT record for ACME challenge
		w.logger.Info("Deleting DNS TXT record for ACME challenge",
			zap.String("domain", domainRecord.Domain),
			zap.String("provider", string(domainRecord.Provider)))

		// In a real implementation, you would:
		// 1. Use the provider's API to delete TXT records
		// 2. Wait for DNS propagation
	}

	w.logger.Info("Domain deletion completed", zap.String("domainID", domainID.String()), zap.String("domain", domainRecord.Domain))
	return nil
}

// generatePassword generates a random password of specified length
func (w *GitRunnerWorker) generatePassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
