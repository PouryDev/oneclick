package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

// DomainRepository defines the interface for managing domains
type DomainRepository interface {
	CreateDomain(ctx context.Context, domain *domain.Domain) (*domain.Domain, error)
	GetDomainByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error)
	GetDomainsByAppID(ctx context.Context, appID uuid.UUID) ([]domain.Domain, error)
	GetDomainByDomainInApp(ctx context.Context, appID uuid.UUID, domainName string) (*domain.Domain, error)
	UpdateDomainCertStatus(ctx context.Context, id uuid.UUID, status domain.CertificateStatus) (*domain.Domain, error)
	UpdateDomainCertSecret(ctx context.Context, id uuid.UUID, secretName string) (*domain.Domain, error)
	UpdateDomainProviderConfig(ctx context.Context, id uuid.UUID, config domain.ProviderConfig) (*domain.Domain, error)
	DeleteDomain(ctx context.Context, id uuid.UUID) error
}

type domainRepo struct {
	db *sql.DB
}

func NewDomainRepository(db *sql.DB) DomainRepository {
	return &domainRepo{db: db}
}

// Domain repository implementation
func (r *domainRepo) CreateDomain(ctx context.Context, d *domain.Domain) (*domain.Domain, error) {
	configBytes, err := json.Marshal(d.ProviderConfig)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO domains (app_id, domain, provider, provider_config, cert_status, cert_secret_name, challenge_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	var id uuid.UUID
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query,
		d.AppID,
		d.Domain,
		d.Provider,
		configBytes,
		d.CertStatus,
		d.CertSecretName,
		d.ChallengeType,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	d.ID = id
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time
	}

	return d, nil
}

func (r *domainRepo) GetDomainByID(ctx context.Context, id uuid.UUID) (*domain.Domain, error) {
	query := `
		SELECT id, app_id, domain, provider, provider_config, cert_status, cert_secret_name, challenge_type, created_at, updated_at
		FROM domains
		WHERE id = $1`

	var d domain.Domain
	var configBytes []byte
	var certSecretName sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID,
		&d.AppID,
		&d.Domain,
		&d.Provider,
		&configBytes,
		&d.CertStatus,
		&certSecretName,
		&d.ChallengeType,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &d.ProviderConfig); err != nil {
		return nil, err
	}

	if certSecretName.Valid {
		d.CertSecretName = certSecretName.String
	}
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time
	}

	return &d, nil
}

func (r *domainRepo) GetDomainsByAppID(ctx context.Context, appID uuid.UUID) ([]domain.Domain, error) {
	query := `
		SELECT id, app_id, domain, provider, provider_config, cert_status, cert_secret_name, challenge_type, created_at, updated_at
		FROM domains
		WHERE app_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []domain.Domain
	for rows.Next() {
		var d domain.Domain
		var configBytes []byte
		var certSecretName sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&d.ID,
			&d.AppID,
			&d.Domain,
			&d.Provider,
			&configBytes,
			&d.CertStatus,
			&certSecretName,
			&d.ChallengeType,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configBytes, &d.ProviderConfig); err != nil {
			return nil, err
		}

		if certSecretName.Valid {
			d.CertSecretName = certSecretName.String
		}
		if createdAt.Valid {
			d.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			d.UpdatedAt = updatedAt.Time
		}

		domains = append(domains, d)
	}

	return domains, nil
}

func (r *domainRepo) GetDomainByDomainInApp(ctx context.Context, appID uuid.UUID, domainName string) (*domain.Domain, error) {
	query := `
		SELECT id, app_id, domain, provider, provider_config, cert_status, cert_secret_name, challenge_type, created_at, updated_at
		FROM domains
		WHERE app_id = $1 AND domain = $2`

	var d domain.Domain
	var configBytes []byte
	var certSecretName sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, appID, domainName).Scan(
		&d.ID,
		&d.AppID,
		&d.Domain,
		&d.Provider,
		&configBytes,
		&d.CertStatus,
		&certSecretName,
		&d.ChallengeType,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &d.ProviderConfig); err != nil {
		return nil, err
	}

	if certSecretName.Valid {
		d.CertSecretName = certSecretName.String
	}
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time
	}

	return &d, nil
}

func (r *domainRepo) UpdateDomainCertStatus(ctx context.Context, id uuid.UUID, status domain.CertificateStatus) (*domain.Domain, error) {
	query := `
		UPDATE domains
		SET cert_status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, domain, provider, provider_config, cert_status, cert_secret_name, challenge_type, created_at, updated_at`

	var d domain.Domain
	var configBytes []byte
	var certSecretName sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&d.ID,
		&d.AppID,
		&d.Domain,
		&d.Provider,
		&configBytes,
		&d.CertStatus,
		&certSecretName,
		&d.ChallengeType,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &d.ProviderConfig); err != nil {
		return nil, err
	}

	if certSecretName.Valid {
		d.CertSecretName = certSecretName.String
	}
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time
	}

	return &d, nil
}

func (r *domainRepo) UpdateDomainCertSecret(ctx context.Context, id uuid.UUID, secretName string) (*domain.Domain, error) {
	query := `
		UPDATE domains
		SET cert_secret_name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, domain, provider, provider_config, cert_status, cert_secret_name, challenge_type, created_at, updated_at`

	var d domain.Domain
	var configBytes []byte
	var certSecretName sql.NullString
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, secretName).Scan(
		&d.ID,
		&d.AppID,
		&d.Domain,
		&d.Provider,
		&configBytes,
		&d.CertStatus,
		&certSecretName,
		&d.ChallengeType,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &d.ProviderConfig); err != nil {
		return nil, err
	}

	if certSecretName.Valid {
		d.CertSecretName = certSecretName.String
	}
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time
	}

	return &d, nil
}

func (r *domainRepo) UpdateDomainProviderConfig(ctx context.Context, id uuid.UUID, config domain.ProviderConfig) (*domain.Domain, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE domains
		SET provider_config = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, domain, provider, provider_config, cert_status, cert_secret_name, challenge_type, created_at, updated_at`

	var d domain.Domain
	var certSecretName sql.NullString
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, id, configBytes).Scan(
		&d.ID,
		&d.AppID,
		&d.Domain,
		&d.Provider,
		&configBytes,
		&d.CertStatus,
		&certSecretName,
		&d.ChallengeType,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &d.ProviderConfig); err != nil {
		return nil, err
	}

	if certSecretName.Valid {
		d.CertSecretName = certSecretName.String
	}
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time
	}

	return &d, nil
}

func (r *domainRepo) DeleteDomain(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM domains WHERE id = $1`
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
