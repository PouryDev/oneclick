package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

type ServiceRepository interface {
	CreateService(ctx context.Context, service *domain.Service) (*domain.Service, error)
	GetServiceByID(ctx context.Context, id uuid.UUID) (*domain.Service, error)
	GetServicesByAppID(ctx context.Context, appID uuid.UUID) ([]domain.ServiceSummary, error)
	GetServiceByNameInApp(ctx context.Context, appID uuid.UUID, name string) (*domain.Service, error)
	UpdateServiceStatus(ctx context.Context, id uuid.UUID, status domain.ServiceStatus) (*domain.Service, error)
	DeleteService(ctx context.Context, id uuid.UUID) error
}

type ServiceConfigRepository interface {
	CreateServiceConfig(ctx context.Context, config *domain.ServiceConfig) (*domain.ServiceConfig, error)
	GetServiceConfigByID(ctx context.Context, id uuid.UUID) (*domain.ServiceConfig, error)
	GetServiceConfigsByServiceID(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceConfigSummary, error)
	GetServiceConfigByKey(ctx context.Context, serviceID uuid.UUID, key string) (*domain.ServiceConfig, error)
	UpdateServiceConfigValue(ctx context.Context, id uuid.UUID, value string) (*domain.ServiceConfig, error)
	DeleteServiceConfig(ctx context.Context, id uuid.UUID) error
}

type serviceRepository struct {
	db *sql.DB
}

type serviceConfigRepository struct {
	db *sql.DB
}

func NewServiceRepository(db *sql.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func NewServiceConfigRepository(db *sql.DB) ServiceConfigRepository {
	return &serviceConfigRepository{db: db}
}

// Service Repository Implementation

func (r *serviceRepository) CreateService(ctx context.Context, service *domain.Service) (*domain.Service, error) {
	query := `
		INSERT INTO services (app_id, name, chart, status, namespace)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, app_id, name, chart, status, namespace, created_at, updated_at
	`

	var createdService domain.Service
	err := r.db.QueryRowContext(ctx, query,
		service.AppID,
		service.Name,
		service.Chart,
		service.Status,
		service.Namespace,
	).Scan(
		&createdService.ID,
		&createdService.AppID,
		&createdService.Name,
		&createdService.Chart,
		&createdService.Status,
		&createdService.Namespace,
		&createdService.CreatedAt,
		&createdService.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdService, nil
}

func (r *serviceRepository) GetServiceByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	query := `
		SELECT id, app_id, name, chart, status, namespace, created_at, updated_at
		FROM services
		WHERE id = $1
	`

	var service domain.Service
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&service.ID,
		&service.AppID,
		&service.Name,
		&service.Chart,
		&service.Status,
		&service.Namespace,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &service, nil
}

func (r *serviceRepository) GetServicesByAppID(ctx context.Context, appID uuid.UUID) ([]domain.ServiceSummary, error) {
	query := `
		SELECT id, app_id, name, chart, status, namespace, created_at, updated_at
		FROM services
		WHERE app_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []domain.ServiceSummary
	for rows.Next() {
		var service domain.ServiceSummary
		var serviceID, appID uuid.UUID

		err := rows.Scan(
			&serviceID,
			&appID,
			&service.Name,
			&service.Chart,
			&service.Status,
			&service.Namespace,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		service.ID = serviceID
		services = append(services, service)
	}

	return services, nil
}

func (r *serviceRepository) GetServiceByNameInApp(ctx context.Context, appID uuid.UUID, name string) (*domain.Service, error) {
	query := `
		SELECT id, app_id, name, chart, status, namespace, created_at, updated_at
		FROM services
		WHERE app_id = $1 AND name = $2
	`

	var service domain.Service
	err := r.db.QueryRowContext(ctx, query, appID, name).Scan(
		&service.ID,
		&service.AppID,
		&service.Name,
		&service.Chart,
		&service.Status,
		&service.Namespace,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &service, nil
}

func (r *serviceRepository) UpdateServiceStatus(ctx context.Context, id uuid.UUID, status domain.ServiceStatus) (*domain.Service, error) {
	query := `
		UPDATE services
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, name, chart, status, namespace, created_at, updated_at
	`

	var service domain.Service
	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&service.ID,
		&service.AppID,
		&service.Name,
		&service.Chart,
		&service.Status,
		&service.Namespace,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &service, nil
}

func (r *serviceRepository) DeleteService(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM services WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Service Config Repository Implementation

func (r *serviceConfigRepository) CreateServiceConfig(ctx context.Context, config *domain.ServiceConfig) (*domain.ServiceConfig, error) {
	query := `
		INSERT INTO service_configs (service_id, key, value, is_secret)
		VALUES ($1, $2, $3, $4)
		RETURNING id, service_id, key, value, is_secret, created_at, updated_at
	`

	var createdConfig domain.ServiceConfig
	err := r.db.QueryRowContext(ctx, query,
		config.ServiceID,
		config.Key,
		config.Value,
		config.IsSecret,
	).Scan(
		&createdConfig.ID,
		&createdConfig.ServiceID,
		&createdConfig.Key,
		&createdConfig.Value,
		&createdConfig.IsSecret,
		&createdConfig.CreatedAt,
		&createdConfig.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdConfig, nil
}

func (r *serviceConfigRepository) GetServiceConfigByID(ctx context.Context, id uuid.UUID) (*domain.ServiceConfig, error) {
	query := `
		SELECT id, service_id, key, value, is_secret, created_at, updated_at
		FROM service_configs
		WHERE id = $1
	`

	var config domain.ServiceConfig
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&config.ServiceID,
		&config.Key,
		&config.Value,
		&config.IsSecret,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &config, nil
}

func (r *serviceConfigRepository) GetServiceConfigsByServiceID(ctx context.Context, serviceID uuid.UUID) ([]domain.ServiceConfigSummary, error) {
	query := `
		SELECT id, service_id, key, value, is_secret, created_at, updated_at
		FROM service_configs
		WHERE service_id = $1
		ORDER BY key
	`

	rows, err := r.db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.ServiceConfigSummary
	for rows.Next() {
		var config domain.ServiceConfigSummary
		var configID, serviceID uuid.UUID
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&configID,
			&serviceID,
			&config.Key,
			&config.Value,
			&config.IsSecret,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		config.ID = configID
		configs = append(configs, config)
	}

	return configs, nil
}

func (r *serviceConfigRepository) GetServiceConfigByKey(ctx context.Context, serviceID uuid.UUID, key string) (*domain.ServiceConfig, error) {
	query := `
		SELECT id, service_id, key, value, is_secret, created_at, updated_at
		FROM service_configs
		WHERE service_id = $1 AND key = $2
	`

	var config domain.ServiceConfig
	err := r.db.QueryRowContext(ctx, query, serviceID, key).Scan(
		&config.ID,
		&config.ServiceID,
		&config.Key,
		&config.Value,
		&config.IsSecret,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &config, nil
}

func (r *serviceConfigRepository) UpdateServiceConfigValue(ctx context.Context, id uuid.UUID, value string) (*domain.ServiceConfig, error) {
	query := `
		UPDATE service_configs
		SET value = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, service_id, key, value, is_secret, created_at, updated_at
	`

	var config domain.ServiceConfig
	err := r.db.QueryRowContext(ctx, query, id, value).Scan(
		&config.ID,
		&config.ServiceID,
		&config.Key,
		&config.Value,
		&config.IsSecret,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &config, nil
}

func (r *serviceConfigRepository) DeleteServiceConfig(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM service_configs WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
