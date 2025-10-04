package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

type OrganizationRepository interface {
	CreateOrganization(ctx context.Context, name string) (*domain.Organization, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error

	// User-Organization relationships
	AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) (*domain.UserOrganization, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domain.UserOrganizationSummary, error)
	GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error)
	GetUserRoleInOrganization(ctx context.Context, userID, orgID uuid.UUID) (string, error)
	UpdateUserRoleInOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) (*domain.UserOrganization, error)
	RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error

	// Helper methods
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type organizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) CreateOrganization(ctx context.Context, name string) (*domain.Organization, error) {
	query := `
		INSERT INTO organizations (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`

	var org domain.Organization
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&org.ID,
		&org.Name,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &org, nil
}

func (r *organizationRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	var org domain.Organization
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &org, nil
}

func (r *organizationRepository) UpdateOrganization(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error) {
	query := `
		UPDATE organizations
		SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, created_at, updated_at
	`

	var org domain.Organization
	err := r.db.QueryRowContext(ctx, query, id, name).Scan(
		&org.ID,
		&org.Name,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &org, nil
}

func (r *organizationRepository) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1`

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

func (r *organizationRepository) AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) (*domain.UserOrganization, error) {
	query := `
		INSERT INTO user_organizations (user_id, org_id, role)
		VALUES ($1, $2, $3)
		RETURNING user_id, org_id, role, created_at, updated_at
	`

	var uo domain.UserOrganization
	err := r.db.QueryRowContext(ctx, query, userID, orgID, role).Scan(
		&uo.UserID,
		&uo.OrgID,
		&uo.Role,
		&uo.CreatedAt,
		&uo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &uo, nil
}

func (r *organizationRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domain.UserOrganizationSummary, error) {
	query := `
		SELECT o.id, o.name, o.created_at, o.updated_at, uo.role
		FROM organizations o
		JOIN user_organizations uo ON o.id = uo.org_id
		WHERE uo.user_id = $1
		ORDER BY o.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []domain.UserOrganizationSummary
	for rows.Next() {
		var org domain.UserOrganizationSummary
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.Role,
		)
		if err != nil {
			return nil, err
		}

		// Set default counts (will be updated when clusters/apps are implemented)
		org.ClustersCount = 0
		org.AppsCount = 0

		orgs = append(orgs, org)
	}

	return orgs, nil
}

func (r *organizationRepository) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error) {
	query := `
		SELECT u.id, u.name, u.email, uo.role, uo.created_at, uo.updated_at
		FROM users u
		JOIN user_organizations uo ON u.id = uo.user_id
		WHERE uo.org_id = $1
		ORDER BY uo.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.OrganizationMember
	for rows.Next() {
		var member domain.OrganizationMember
		err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.Email,
			&member.Role,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, nil
}

func (r *organizationRepository) GetUserRoleInOrganization(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	query := `
		SELECT role FROM user_organizations
		WHERE user_id = $1 AND org_id = $2
	`

	var role string
	err := r.db.QueryRowContext(ctx, query, userID, orgID).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return role, nil
}

func (r *organizationRepository) UpdateUserRoleInOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) (*domain.UserOrganization, error) {
	query := `
		UPDATE user_organizations
		SET role = $3, updated_at = NOW()
		WHERE user_id = $1 AND org_id = $2
		RETURNING user_id, org_id, role, created_at, updated_at
	`

	var uo domain.UserOrganization
	err := r.db.QueryRowContext(ctx, query, userID, orgID, role).Scan(
		&uo.UserID,
		&uo.OrgID,
		&uo.Role,
		&uo.CreatedAt,
		&uo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &uo, nil
}

func (r *organizationRepository) RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error {
	query := `
		DELETE FROM user_organizations
		WHERE user_id = $1 AND org_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, orgID)
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

func (r *organizationRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
