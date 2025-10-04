package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type GitServerService interface {
	CreateGitServer(ctx context.Context, userID, orgID uuid.UUID, req domain.CreateGitServerRequest) (*domain.GitServerResponse, error)
	GetGitServersByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.GitServerResponse, error)
	GetGitServer(ctx context.Context, userID, gitServerID uuid.UUID) (*domain.GitServerResponse, error)
	DeleteGitServer(ctx context.Context, userID, gitServerID uuid.UUID) error
}

type RunnerService interface {
	CreateRunner(ctx context.Context, userID, orgID uuid.UUID, req domain.CreateRunnerRequest) (*domain.RunnerResponse, error)
	GetRunnersByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.RunnerResponse, error)
	GetRunner(ctx context.Context, userID, runnerID uuid.UUID) (*domain.RunnerResponse, error)
	DeleteRunner(ctx context.Context, userID, runnerID uuid.UUID) error
}

type JobService interface {
	CreateJob(ctx context.Context, orgID uuid.UUID, jobType domain.JobType, payload domain.JobPayload) (*domain.JobResponse, error)
	GetJobsByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.JobResponse, error)
	GetPendingJobs(ctx context.Context) ([]domain.JobResponse, error)
	StartJob(ctx context.Context, jobID uuid.UUID) (*domain.JobResponse, error)
	CompleteJob(ctx context.Context, jobID uuid.UUID) (*domain.JobResponse, error)
	FailJob(ctx context.Context, jobID uuid.UUID, errorMessage string) (*domain.JobResponse, error)
}

type gitServerService struct {
	gitServerRepo repo.GitServerRepository
	jobRepo       repo.JobRepository
	orgRepo       repo.OrganizationRepository
	crypto        *crypto.Crypto
	logger        *zap.Logger
}

type runnerService struct {
	runnerRepo repo.RunnerRepository
	jobRepo    repo.JobRepository
	orgRepo    repo.OrganizationRepository
	crypto     *crypto.Crypto
	logger     *zap.Logger
}

type jobService struct {
	jobRepo repo.JobRepository
	orgRepo repo.OrganizationRepository
	logger  *zap.Logger
}

func NewGitServerService(
	gitServerRepo repo.GitServerRepository,
	jobRepo repo.JobRepository,
	orgRepo repo.OrganizationRepository,
	crypto *crypto.Crypto,
	logger *zap.Logger,
) GitServerService {
	return &gitServerService{
		gitServerRepo: gitServerRepo,
		jobRepo:       jobRepo,
		orgRepo:       orgRepo,
		crypto:        crypto,
		logger:        logger,
	}
}

func NewRunnerService(
	runnerRepo repo.RunnerRepository,
	jobRepo repo.JobRepository,
	orgRepo repo.OrganizationRepository,
	crypto *crypto.Crypto,
	logger *zap.Logger,
) RunnerService {
	return &runnerService{
		runnerRepo: runnerRepo,
		jobRepo:    jobRepo,
		orgRepo:    orgRepo,
		crypto:     crypto,
		logger:     logger,
	}
}

func NewJobService(
	jobRepo repo.JobRepository,
	orgRepo repo.OrganizationRepository,
	logger *zap.Logger,
) JobService {
	return &jobService{
		jobRepo: jobRepo,
		orgRepo: orgRepo,
		logger:  logger,
	}
}

// GitServer service implementation
func (s *gitServerService) CreateGitServer(ctx context.Context, userID, orgID uuid.UUID, req domain.CreateGitServerRequest) (*domain.GitServerResponse, error) {
	// Verify user is a member of the organization (Admin or Owner can create git servers)
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return nil, errors.New("insufficient permissions to create git server")
	}

	// Check if git server with same domain already exists in organization
	existingGitServer, err := s.gitServerRepo.GetGitServerByDomainInOrg(ctx, orgID, req.Domain)
	if err != nil {
		s.logger.Error("Failed to check for existing git server", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("domain", req.Domain))
		return nil, errors.New("failed to check for existing git server")
	}
	if existingGitServer != nil {
		return nil, errors.New("git server with this domain already exists in organization")
	}

	// Create git server record
	gitServer := &domain.GitServer{
		OrgID:   orgID,
		Type:    req.Type,
		Domain:  req.Domain,
		Storage: req.Storage,
		Status:  domain.GitServerStatusPending,
		Config:  domain.GitServerConfig{},
	}

	createdGitServer, err := s.gitServerRepo.CreateGitServer(ctx, gitServer)
	if err != nil {
		s.logger.Error("Failed to create git server", zap.Error(err), zap.String("orgID", orgID.String()))
		return nil, errors.New("failed to create git server")
	}

	// Create background job for git server installation
	jobPayload := domain.JobPayload{
		GitServerID: &createdGitServer.ID,
		Config: map[string]interface{}{
			"type":    req.Type,
			"domain":  req.Domain,
			"storage": req.Storage,
		},
	}

	job := &domain.Job{
		OrgID:   orgID,
		Type:    domain.JobTypeGitServerInstall,
		Status:  domain.JobStatusPending,
		Payload: jobPayload,
	}

	_, err = s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create git server installation job", zap.Error(err), zap.String("gitServerID", createdGitServer.ID.String()))
		// Don't fail the git server creation if job creation fails
		s.logger.Warn("Git server created but job creation failed", zap.String("gitServerID", createdGitServer.ID.String()))
	}

	// Update git server status to provisioning
	_, err = s.gitServerRepo.UpdateGitServerStatus(ctx, createdGitServer.ID, domain.GitServerStatusProvisioning)
	if err != nil {
		s.logger.Error("Failed to update git server status to provisioning", zap.Error(err), zap.String("gitServerID", createdGitServer.ID.String()))
		// Don't fail the operation if status update fails
	}

	s.logger.Info("Git server created successfully", zap.String("gitServerID", createdGitServer.ID.String()), zap.String("domain", req.Domain))

	response := createdGitServer.ToResponse()
	return &response, nil
}

func (s *gitServerService) GetGitServersByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.GitServerResponse, error) {
	// Verify user is a member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	gitServers, err := s.gitServerRepo.GetGitServersByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get git servers by organization ID", zap.Error(err), zap.String("orgID", orgID.String()))
		return nil, errors.New("failed to retrieve git servers")
	}

	var responses []domain.GitServerResponse
	for _, gitServer := range gitServers {
		responses = append(responses, gitServer.ToResponse())
	}

	return responses, nil
}

func (s *gitServerService) GetGitServer(ctx context.Context, userID, gitServerID uuid.UUID) (*domain.GitServerResponse, error) {
	// Get git server
	gitServer, err := s.gitServerRepo.GetGitServerByID(ctx, gitServerID)
	if err != nil {
		s.logger.Error("Failed to get git server by ID", zap.Error(err), zap.String("gitServerID", gitServerID.String()))
		return nil, errors.New("failed to retrieve git server")
	}
	if gitServer == nil {
		return nil, errors.New("git server not found")
	}

	// Verify user is a member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, gitServer.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", gitServer.OrgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role == "" {
		return nil, errors.New("user does not have access to this git server")
	}

	response := gitServer.ToResponse()
	return &response, nil
}

func (s *gitServerService) DeleteGitServer(ctx context.Context, userID, gitServerID uuid.UUID) error {
	// Get git server
	gitServer, err := s.gitServerRepo.GetGitServerByID(ctx, gitServerID)
	if err != nil {
		s.logger.Error("Failed to get git server by ID for deletion", zap.Error(err), zap.String("gitServerID", gitServerID.String()))
		return errors.New("failed to retrieve git server")
	}
	if gitServer == nil {
		return errors.New("git server not found")
	}

	// Verify user is a member of the organization (Admin or Owner can delete git servers)
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, gitServer.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", gitServer.OrgID.String()), zap.String("userID", userID.String()))
		return errors.New("failed to verify organization membership")
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return errors.New("insufficient permissions to delete git server")
	}

	// Create background job for git server removal
	jobPayload := domain.JobPayload{
		GitServerID: &gitServer.ID,
		Config: map[string]interface{}{
			"type":   gitServer.Type,
			"domain": gitServer.Domain,
		},
	}

	job := &domain.Job{
		OrgID:   gitServer.OrgID,
		Type:    domain.JobTypeGitServerStop,
		Status:  domain.JobStatusPending,
		Payload: jobPayload,
	}

	_, err = s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create git server removal job", zap.Error(err), zap.String("gitServerID", gitServer.ID.String()))
		// Continue with deletion even if job creation fails
	}

	// Update git server status to stopped
	_, err = s.gitServerRepo.UpdateGitServerStatus(ctx, gitServerID, domain.GitServerStatusStopped)
	if err != nil {
		s.logger.Error("Failed to update git server status to stopped", zap.Error(err), zap.String("gitServerID", gitServerID.String()))
		return errors.New("failed to update git server status")
	}

	// Delete git server record
	err = s.gitServerRepo.DeleteGitServer(ctx, gitServerID)
	if err != nil {
		s.logger.Error("Failed to delete git server", zap.Error(err), zap.String("gitServerID", gitServerID.String()))
		return errors.New("failed to delete git server")
	}

	s.logger.Info("Git server deleted successfully", zap.String("gitServerID", gitServerID.String()))
	return nil
}

// Runner service implementation
func (s *runnerService) CreateRunner(ctx context.Context, userID, orgID uuid.UUID, req domain.CreateRunnerRequest) (*domain.RunnerResponse, error) {
	// Verify user is a member of the organization (Admin or Owner can create runners)
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return nil, errors.New("insufficient permissions to create runner")
	}

	// Check if runner with same name already exists in organization
	existingRunner, err := s.runnerRepo.GetRunnerByNameInOrg(ctx, orgID, req.Name)
	if err != nil {
		s.logger.Error("Failed to check for existing runner", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("name", req.Name))
		return nil, errors.New("failed to check for existing runner")
	}
	if existingRunner != nil {
		return nil, errors.New("runner with this name already exists in organization")
	}

	// Encrypt token if provided
	encryptedToken := ""
	if req.Token != "" {
		encryptedTokenBytes, err := s.crypto.EncryptString(req.Token)
		if err != nil {
			s.logger.Error("Failed to encrypt runner token", zap.Error(err))
			return nil, errors.New("failed to encrypt runner token")
		}
		encryptedToken = string(encryptedTokenBytes)
	}

	// Create runner configuration
	config := domain.RunnerConfig{
		Labels:       req.Labels,
		NodeSelector: req.NodeSelector,
		Resources:    req.Resources,
		Token:        encryptedToken,
		URL:          req.URL,
		Settings:     make(map[string]string),
	}

	// Create runner record
	runner := &domain.Runner{
		OrgID:  orgID,
		Name:   req.Name,
		Type:   req.Type,
		Config: config,
		Status: domain.RunnerStatusPending,
	}

	createdRunner, err := s.runnerRepo.CreateRunner(ctx, runner)
	if err != nil {
		s.logger.Error("Failed to create runner", zap.Error(err), zap.String("orgID", orgID.String()))
		return nil, errors.New("failed to create runner")
	}

	// Create background job for runner deployment
	jobPayload := domain.JobPayload{
		RunnerID: &createdRunner.ID,
		Config: map[string]interface{}{
			"name":          req.Name,
			"type":          req.Type,
			"labels":        req.Labels,
			"node_selector": req.NodeSelector,
			"resources":     req.Resources,
			"url":           req.URL,
		},
	}

	job := &domain.Job{
		OrgID:   orgID,
		Type:    domain.JobTypeRunnerDeploy,
		Status:  domain.JobStatusPending,
		Payload: jobPayload,
	}

	_, err = s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create runner deployment job", zap.Error(err), zap.String("runnerID", createdRunner.ID.String()))
		// Don't fail the runner creation if job creation fails
		s.logger.Warn("Runner created but job creation failed", zap.String("runnerID", createdRunner.ID.String()))
	}

	// Update runner status to provisioning
	_, err = s.runnerRepo.UpdateRunnerStatus(ctx, createdRunner.ID, domain.RunnerStatusProvisioning)
	if err != nil {
		s.logger.Error("Failed to update runner status to provisioning", zap.Error(err), zap.String("runnerID", createdRunner.ID.String()))
		// Don't fail the operation if status update fails
	}

	s.logger.Info("Runner created successfully", zap.String("runnerID", createdRunner.ID.String()), zap.String("name", req.Name))

	response := createdRunner.ToResponse()
	return &response, nil
}

func (s *runnerService) GetRunnersByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.RunnerResponse, error) {
	// Verify user is a member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	runners, err := s.runnerRepo.GetRunnersByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get runners by organization ID", zap.Error(err), zap.String("orgID", orgID.String()))
		return nil, errors.New("failed to retrieve runners")
	}

	var responses []domain.RunnerResponse
	for _, runner := range runners {
		responses = append(responses, runner.ToResponse())
	}

	return responses, nil
}

func (s *runnerService) GetRunner(ctx context.Context, userID, runnerID uuid.UUID) (*domain.RunnerResponse, error) {
	// Get runner
	runner, err := s.runnerRepo.GetRunnerByID(ctx, runnerID)
	if err != nil {
		s.logger.Error("Failed to get runner by ID", zap.Error(err), zap.String("runnerID", runnerID.String()))
		return nil, errors.New("failed to retrieve runner")
	}
	if runner == nil {
		return nil, errors.New("runner not found")
	}

	// Verify user is a member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, runner.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", runner.OrgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role == "" {
		return nil, errors.New("user does not have access to this runner")
	}

	response := runner.ToResponse()
	return &response, nil
}

func (s *runnerService) DeleteRunner(ctx context.Context, userID, runnerID uuid.UUID) error {
	// Get runner
	runner, err := s.runnerRepo.GetRunnerByID(ctx, runnerID)
	if err != nil {
		s.logger.Error("Failed to get runner by ID for deletion", zap.Error(err), zap.String("runnerID", runnerID.String()))
		return errors.New("failed to retrieve runner")
	}
	if runner == nil {
		return errors.New("runner not found")
	}

	// Verify user is a member of the organization (Admin or Owner can delete runners)
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, runner.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", runner.OrgID.String()), zap.String("userID", userID.String()))
		return errors.New("failed to verify organization membership")
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return errors.New("insufficient permissions to delete runner")
	}

	// Create background job for runner removal
	jobPayload := domain.JobPayload{
		RunnerID: &runner.ID,
		Config: map[string]interface{}{
			"name": runner.Name,
			"type": runner.Type,
		},
	}

	job := &domain.Job{
		OrgID:   runner.OrgID,
		Type:    domain.JobTypeRunnerStop,
		Status:  domain.JobStatusPending,
		Payload: jobPayload,
	}

	_, err = s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create runner removal job", zap.Error(err), zap.String("runnerID", runner.ID.String()))
		// Continue with deletion even if job creation fails
	}

	// Update runner status to stopped
	_, err = s.runnerRepo.UpdateRunnerStatus(ctx, runnerID, domain.RunnerStatusStopped)
	if err != nil {
		s.logger.Error("Failed to update runner status to stopped", zap.Error(err), zap.String("runnerID", runnerID.String()))
		return errors.New("failed to update runner status")
	}

	// Delete runner record
	err = s.runnerRepo.DeleteRunner(ctx, runnerID)
	if err != nil {
		s.logger.Error("Failed to delete runner", zap.Error(err), zap.String("runnerID", runnerID.String()))
		return errors.New("failed to delete runner")
	}

	s.logger.Info("Runner deleted successfully", zap.String("runnerID", runnerID.String()))
	return nil
}

// Job service implementation
func (s *jobService) CreateJob(ctx context.Context, orgID uuid.UUID, jobType domain.JobType, payload domain.JobPayload) (*domain.JobResponse, error) {
	job := &domain.Job{
		OrgID:   orgID,
		Type:    jobType,
		Status:  domain.JobStatusPending,
		Payload: payload,
	}

	createdJob, err := s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create job", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("type", string(jobType)))
		return nil, errors.New("failed to create job")
	}

	response := createdJob.ToResponse()
	return &response, nil
}

func (s *jobService) GetJobsByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.JobResponse, error) {
	// Verify user is a member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", orgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	jobs, err := s.jobRepo.GetJobsByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get jobs by organization ID", zap.Error(err), zap.String("orgID", orgID.String()))
		return nil, errors.New("failed to retrieve jobs")
	}

	var responses []domain.JobResponse
	for _, job := range jobs {
		responses = append(responses, job.ToResponse())
	}

	return responses, nil
}

func (s *jobService) GetPendingJobs(ctx context.Context) ([]domain.JobResponse, error) {
	jobs, err := s.jobRepo.GetPendingJobs(ctx)
	if err != nil {
		s.logger.Error("Failed to get pending jobs", zap.Error(err))
		return nil, errors.New("failed to retrieve pending jobs")
	}

	var responses []domain.JobResponse
	for _, job := range jobs {
		responses = append(responses, job.ToResponse())
	}

	return responses, nil
}

func (s *jobService) StartJob(ctx context.Context, jobID uuid.UUID) (*domain.JobResponse, error) {
	job, err := s.jobRepo.StartJob(ctx, jobID)
	if err != nil {
		s.logger.Error("Failed to start job", zap.Error(err), zap.String("jobID", jobID.String()))
		return nil, errors.New("failed to start job")
	}

	response := job.ToResponse()
	return &response, nil
}

func (s *jobService) CompleteJob(ctx context.Context, jobID uuid.UUID) (*domain.JobResponse, error) {
	job, err := s.jobRepo.CompleteJob(ctx, jobID)
	if err != nil {
		s.logger.Error("Failed to complete job", zap.Error(err), zap.String("jobID", jobID.String()))
		return nil, errors.New("failed to complete job")
	}

	response := job.ToResponse()
	return &response, nil
}

func (s *jobService) FailJob(ctx context.Context, jobID uuid.UUID, errorMessage string) (*domain.JobResponse, error) {
	job, err := s.jobRepo.FailJob(ctx, jobID, errorMessage)
	if err != nil {
		s.logger.Error("Failed to fail job", zap.Error(err), zap.String("jobID", jobID.String()))
		return nil, errors.New("failed to fail job")
	}

	response := job.ToResponse()
	return &response, nil
}
