package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/infra"
	"github.com/PouryDev/oneclick/internal/app/provisioner"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type InfrastructureService interface {
	ProvisionServices(ctx context.Context, userID, appID uuid.UUID, infraConfigYAML string) (*domain.ProvisionServiceResponse, error)
	GetServicesByApp(ctx context.Context, userID, appID uuid.UUID) ([]domain.ServiceDetail, error)
	GetServiceConfig(ctx context.Context, userID, configID uuid.UUID) (*domain.ServiceConfigRevealResponse, error)
	UnprovisionService(ctx context.Context, userID, serviceID uuid.UUID) error
}

type infrastructureService struct {
	appRepo           repo.ApplicationRepository
	serviceRepo       repo.ServiceRepository
	serviceConfigRepo repo.ServiceConfigRepository
	orgRepo           repo.OrganizationRepository
	parser            *infra.Parser
	provisioner       *provisioner.ServiceProvisioner
	logger            *zap.Logger
}

func NewInfrastructureService(
	appRepo repo.ApplicationRepository,
	serviceRepo repo.ServiceRepository,
	serviceConfigRepo repo.ServiceConfigRepository,
	orgRepo repo.OrganizationRepository,
	parser *infra.Parser,
	provisioner *provisioner.ServiceProvisioner,
	logger *zap.Logger,
) InfrastructureService {
	return &infrastructureService{
		appRepo:           appRepo,
		serviceRepo:       serviceRepo,
		serviceConfigRepo: serviceConfigRepo,
		orgRepo:           orgRepo,
		parser:            parser,
		provisioner:       provisioner,
		logger:            logger,
	}
}

func (s *infrastructureService) ProvisionServices(ctx context.Context, userID, appID uuid.UUID, infraConfigYAML string) (*domain.ProvisionServiceResponse, error) {
	// 1. Get application to verify organization membership
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID for service provisioning", zap.Error(err), zap.String("appID", appID.String()))
		return nil, errors.New("failed to retrieve application")
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// 2. Verify user is a member of the organization (Admin or Owner can provision services)
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization during service provisioning", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return nil, errors.New("insufficient permissions to provision services")
	}

	// 3. Parse infrastructure configuration
	config, err := s.parser.ParseConfig(infraConfigYAML)
	if err != nil {
		s.logger.Error("Failed to parse infrastructure configuration", zap.Error(err))
		return nil, fmt.Errorf("failed to parse infrastructure configuration: %w", err)
	}

	// 4. Validate configuration
	err = s.parser.ValidateConfig(config)
	if err != nil {
		s.logger.Error("Invalid infrastructure configuration", zap.Error(err))
		return nil, fmt.Errorf("invalid infrastructure configuration: %w", err)
	}

	// 5. Process secrets (for future use)
	_, err = s.parser.ProcessSecrets(config)
	if err != nil {
		s.logger.Error("Failed to process secrets", zap.Error(err))
		return nil, fmt.Errorf("failed to process secrets: %w", err)
	}

	// 6. Generate service configurations
	serviceConfigs, err := s.parser.GenerateServiceConfigs(config, app.Name)
	if err != nil {
		s.logger.Error("Failed to generate service configurations", zap.Error(err))
		return nil, fmt.Errorf("failed to generate service configurations: %w", err)
	}

	// 7. Create services and configurations in database
	var createdServices []domain.ServiceSummary
	for _, serviceConfig := range serviceConfigs {
		// Check if service already exists
		existingService, err := s.serviceRepo.GetServiceByNameInApp(ctx, appID, serviceConfig.ServiceName)
		if err != nil && err.Error() != "sql: no rows in result set" {
			s.logger.Error("Failed to check for existing service", zap.Error(err), zap.String("serviceName", serviceConfig.ServiceName))
			return nil, errors.New("failed to check for existing service")
		}
		if existingService != nil {
			s.logger.Warn("Service already exists, skipping", zap.String("serviceName", serviceConfig.ServiceName))
			continue
		}

		// Create service
		service := &domain.Service{
			AppID:     appID,
			Name:      serviceConfig.ServiceName,
			Chart:     serviceConfig.Chart,
			Status:    domain.ServiceStatusPending,
			Namespace: serviceConfig.Namespace,
		}

		createdService, err := s.serviceRepo.CreateService(ctx, service)
		if err != nil {
			s.logger.Error("Failed to create service in database", zap.Error(err), zap.String("serviceName", serviceConfig.ServiceName))
			return nil, errors.New("failed to create service")
		}

		// Create service configurations
		for key, configValue := range serviceConfig.Configs {
			serviceConfig := &domain.ServiceConfig{
				ServiceID: createdService.ID,
				Key:       key,
				Value:     configValue.Value,
				IsSecret:  configValue.IsSecret,
			}

			_, err := s.serviceConfigRepo.CreateServiceConfig(ctx, serviceConfig)
			if err != nil {
				s.logger.Error("Failed to create service configuration", zap.Error(err), zap.String("serviceID", createdService.ID.String()), zap.String("key", key))
				return nil, errors.New("failed to create service configuration")
			}
		}

		createdServices = append(createdServices, createdService.ToSummary())

		// Update service status to provisioning
		_, err = s.serviceRepo.UpdateServiceStatus(ctx, createdService.ID, domain.ServiceStatusProvisioning)
		if err != nil {
			s.logger.Error("Failed to update service status to provisioning", zap.Error(err), zap.String("serviceID", createdService.ID.String()))
			// Continue with other services even if status update fails
		}

		// Start provisioning in background (for MVP, we'll simulate this)
		go s.provisionServiceInBackground(ctx, createdService, serviceConfig.Configs)
	}

	response := &domain.ProvisionServiceResponse{
		Services: createdServices,
		Message:  fmt.Sprintf("Provisioning initiated for %d services", len(createdServices)),
	}

	return response, nil
}

func (s *infrastructureService) GetServicesByApp(ctx context.Context, userID, appID uuid.UUID) ([]domain.ServiceDetail, error) {
	// 1. Get application to verify organization membership
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application by ID", zap.Error(err), zap.String("appID", appID.String()))
		return nil, errors.New("failed to retrieve application")
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// 2. Verify user is a member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role == "" {
		return nil, errors.New("user does not have access to this application's organization")
	}

	// 3. Get services
	serviceSummaries, err := s.serviceRepo.GetServicesByAppID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get services by application ID", zap.Error(err), zap.String("appID", appID.String()))
		return nil, errors.New("failed to retrieve services")
	}

	var serviceDetails []domain.ServiceDetail
	for _, summary := range serviceSummaries {
		// Get service configurations
		configs, err := s.serviceConfigRepo.GetServiceConfigsByServiceID(ctx, summary.ID)
		if err != nil {
			s.logger.Warn("Failed to get service configurations", zap.Error(err), zap.String("serviceID", summary.ID.String()))
			// Continue without configs if not found or error
		}

		// Convert to ServiceDetail
		service := &domain.Service{
			ID:        summary.ID,
			AppID:     appID,
			Name:      summary.Name,
			Chart:     summary.Chart,
			Status:    summary.Status,
			Namespace: summary.Namespace,
			CreatedAt: summary.CreatedAt,
			UpdatedAt: summary.UpdatedAt,
		}

		var domainConfigs []domain.ServiceConfig
		for _, config := range configs {
			domainConfigs = append(domainConfigs, domain.ServiceConfig{
				ID:        config.ID,
				ServiceID: summary.ID,
				Key:       config.Key,
				Value:     config.Value,
				IsSecret:  config.IsSecret,
			})
		}

		detail := service.ToDetail(domainConfigs)
		serviceDetails = append(serviceDetails, detail)
	}

	return serviceDetails, nil
}

func (s *infrastructureService) GetServiceConfig(ctx context.Context, userID, configID uuid.UUID) (*domain.ServiceConfigRevealResponse, error) {
	// 1. Get service configuration
	config, err := s.serviceConfigRepo.GetServiceConfigByID(ctx, configID)
	if err != nil {
		s.logger.Error("Failed to get service configuration by ID", zap.Error(err), zap.String("configID", configID.String()))
		return nil, errors.New("failed to retrieve service configuration")
	}
	if config == nil {
		return nil, errors.New("service configuration not found")
	}

	// 2. Get service to verify organization membership
	service, err := s.serviceRepo.GetServiceByID(ctx, config.ServiceID)
	if err != nil {
		s.logger.Error("Failed to get service by ID", zap.Error(err), zap.String("serviceID", config.ServiceID.String()))
		return nil, errors.New("failed to retrieve service")
	}
	if service == nil {
		return nil, errors.New("service not found")
	}

	// 3. Get application to verify organization membership
	app, err := s.appRepo.GetApplicationByID(ctx, service.AppID)
	if err != nil {
		s.logger.Error("Failed to get application by ID", zap.Error(err), zap.String("appID", service.AppID.String()))
		return nil, errors.New("failed to retrieve application")
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// 4. Verify user is a member of the organization (Admin or Owner can reveal secrets)
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, errors.New("failed to verify organization membership")
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return nil, errors.New("insufficient permissions to reveal service configuration")
	}

	response := &domain.ServiceConfigRevealResponse{
		Config: config.ToResponse(),
	}

	return response, nil
}

func (s *infrastructureService) UnprovisionService(ctx context.Context, userID, serviceID uuid.UUID) error {
	// 1. Get service
	service, err := s.serviceRepo.GetServiceByID(ctx, serviceID)
	if err != nil {
		s.logger.Error("Failed to get service by ID for unprovisioning", zap.Error(err), zap.String("serviceID", serviceID.String()))
		return errors.New("failed to retrieve service")
	}
	if service == nil {
		return errors.New("service not found")
	}

	// 2. Get application to verify organization membership
	app, err := s.appRepo.GetApplicationByID(ctx, service.AppID)
	if err != nil {
		s.logger.Error("Failed to get application by ID", zap.Error(err), zap.String("appID", service.AppID.String()))
		return errors.New("failed to retrieve application")
	}
	if app == nil {
		return errors.New("application not found")
	}

	// 3. Verify user is a member of the organization (Admin or Owner can unprovision services)
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role for organization during service unprovisioning", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return errors.New("failed to verify organization membership")
	}
	if role != domain.RoleAdmin && role != domain.RoleOwner {
		return errors.New("insufficient permissions to unprovision service")
	}

	// 4. Update service status to stopped
	_, err = s.serviceRepo.UpdateServiceStatus(ctx, serviceID, domain.ServiceStatusStopped)
	if err != nil {
		s.logger.Error("Failed to update service status to stopped", zap.Error(err), zap.String("serviceID", serviceID.String()))
		return errors.New("failed to update service status")
	}

	// 5. Unprovision service in background
	go s.unprovisionServiceInBackground(ctx, service)

	return nil
}

// provisionServiceInBackground provisions a service in the background
func (s *infrastructureService) provisionServiceInBackground(ctx context.Context, service *domain.Service, configs map[string]infra.ConfigValue) {
	s.logger.Info("Starting background service provisioning",
		zap.String("serviceID", service.ID.String()),
		zap.String("serviceName", service.Name),
		zap.String("chart", service.Chart),
	)

	// Convert configs to domain.ServiceConfig
	var domainConfigs []domain.ServiceConfig
	for key, configValue := range configs {
		domainConfigs = append(domainConfigs, domain.ServiceConfig{
			ServiceID: service.ID,
			Key:       key,
			Value:     configValue.Value,
			IsSecret:  configValue.IsSecret,
		})
	}

	// Provision service
	err := s.provisioner.ProvisionService(ctx, service, domainConfigs)
	if err != nil {
		s.logger.Error("Failed to provision service in background",
			zap.String("serviceID", service.ID.String()),
			zap.String("serviceName", service.Name),
			zap.Error(err),
		)

		// Update service status to failed
		_, updateErr := s.serviceRepo.UpdateServiceStatus(ctx, service.ID, domain.ServiceStatusFailed)
		if updateErr != nil {
			s.logger.Error("Failed to update service status to failed",
				zap.String("serviceID", service.ID.String()),
				zap.Error(updateErr),
			)
		}
		return
	}

	// Update service status to running
	_, err = s.serviceRepo.UpdateServiceStatus(ctx, service.ID, domain.ServiceStatusRunning)
	if err != nil {
		s.logger.Error("Failed to update service status to running",
			zap.String("serviceID", service.ID.String()),
			zap.Error(err),
		)
	}

	s.logger.Info("Service provisioned successfully in background",
		zap.String("serviceID", service.ID.String()),
		zap.String("serviceName", service.Name),
	)
}

// unprovisionServiceInBackground unprovisions a service in the background
func (s *infrastructureService) unprovisionServiceInBackground(ctx context.Context, service *domain.Service) {
	s.logger.Info("Starting background service unprovisioning",
		zap.String("serviceID", service.ID.String()),
		zap.String("serviceName", service.Name),
	)

	// Unprovision service
	err := s.provisioner.UnprovisionService(ctx, service)
	if err != nil {
		s.logger.Error("Failed to unprovision service in background",
			zap.String("serviceID", service.ID.String()),
			zap.String("serviceName", service.Name),
			zap.Error(err),
		)

		// Update service status to failed
		_, updateErr := s.serviceRepo.UpdateServiceStatus(ctx, service.ID, domain.ServiceStatusFailed)
		if updateErr != nil {
			s.logger.Error("Failed to update service status to failed",
				zap.String("serviceID", service.ID.String()),
				zap.Error(updateErr),
			)
		}
		return
	}

	// Delete service from database
	err = s.serviceRepo.DeleteService(ctx, service.ID)
	if err != nil {
		s.logger.Error("Failed to delete service from database",
			zap.String("serviceID", service.ID.String()),
			zap.Error(err),
		)
	}

	s.logger.Info("Service unprovisioned successfully in background",
		zap.String("serviceID", service.ID.String()),
		zap.String("serviceName", service.Name),
	)
}
