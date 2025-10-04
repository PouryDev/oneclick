package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

type ClusterRepository interface {
	CreateCluster(ctx context.Context, cluster *domain.Cluster) (*domain.Cluster, error)
	GetClusterByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error)
	GetClustersByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.ClusterSummary, error)
	UpdateClusterStatus(ctx context.Context, id uuid.UUID, status string) (*domain.Cluster, error)
	UpdateClusterKubeconfig(ctx context.Context, id uuid.UUID, kubeconfigEncrypted []byte, status string) (*domain.Cluster, error)
	UpdateClusterHealth(ctx context.Context, id uuid.UUID, kubeVersion string) (*domain.Cluster, error)
	UpdateClusterNodeCount(ctx context.Context, id uuid.UUID, nodeCount int) (*domain.Cluster, error)
	DeleteCluster(ctx context.Context, id uuid.UUID) error
}

type clusterRepository struct {
	db *sql.DB
}

func NewClusterRepository(db *sql.DB) ClusterRepository {
	return &clusterRepository{db: db}
}

func (r *clusterRepository) CreateCluster(ctx context.Context, cluster *domain.Cluster) (*domain.Cluster, error) {
	query := `
		INSERT INTO clusters (org_id, name, provider, region, kubeconfig_encrypted, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, org_id, name, provider, region, node_count, status, kube_version, last_health_check, created_at, updated_at
	`

	var createdCluster domain.Cluster
	err := r.db.QueryRowContext(ctx, query,
		cluster.OrgID,
		cluster.Name,
		cluster.Provider,
		cluster.Region,
		cluster.KubeconfigEncrypted,
		cluster.Status,
	).Scan(
		&createdCluster.ID,
		&createdCluster.OrgID,
		&createdCluster.Name,
		&createdCluster.Provider,
		&createdCluster.Region,
		&createdCluster.NodeCount,
		&createdCluster.Status,
		&createdCluster.KubeVersion,
		&createdCluster.LastHealthCheck,
		&createdCluster.CreatedAt,
		&createdCluster.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdCluster, nil
}

func (r *clusterRepository) GetClusterByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	query := `
		SELECT id, org_id, name, provider, region, kubeconfig_encrypted, node_count, status, kube_version, last_health_check, created_at, updated_at
		FROM clusters
		WHERE id = $1
	`

	var cluster domain.Cluster
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cluster.ID,
		&cluster.OrgID,
		&cluster.Name,
		&cluster.Provider,
		&cluster.Region,
		&cluster.KubeconfigEncrypted,
		&cluster.NodeCount,
		&cluster.Status,
		&cluster.KubeVersion,
		&cluster.LastHealthCheck,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &cluster, nil
}

func (r *clusterRepository) GetClustersByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.ClusterSummary, error) {
	query := `
		SELECT id, org_id, name, provider, region, node_count, status, kube_version, last_health_check, created_at, updated_at
		FROM clusters
		WHERE org_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []domain.ClusterSummary
	for rows.Next() {
		var cluster domain.ClusterSummary
		var kubeVersion sql.NullString
		var lastHealthCheck sql.NullTime

		err := rows.Scan(
			&cluster.ID,
			&cluster.Name,
			&cluster.Provider,
			&cluster.Region,
			&cluster.NodeCount,
			&cluster.Status,
			&kubeVersion,
			&lastHealthCheck,
			&cluster.CreatedAt,
			&cluster.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (r *clusterRepository) UpdateClusterStatus(ctx context.Context, id uuid.UUID, status string) (*domain.Cluster, error) {
	query := `
		UPDATE clusters
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, name, provider, region, node_count, status, kube_version, last_health_check, created_at, updated_at
	`

	var cluster domain.Cluster
	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&cluster.ID,
		&cluster.OrgID,
		&cluster.Name,
		&cluster.Provider,
		&cluster.Region,
		&cluster.NodeCount,
		&cluster.Status,
		&cluster.KubeVersion,
		&cluster.LastHealthCheck,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &cluster, nil
}

func (r *clusterRepository) UpdateClusterKubeconfig(ctx context.Context, id uuid.UUID, kubeconfigEncrypted []byte, status string) (*domain.Cluster, error) {
	query := `
		UPDATE clusters
		SET kubeconfig_encrypted = $2, status = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, name, provider, region, node_count, status, kube_version, last_health_check, created_at, updated_at
	`

	var cluster domain.Cluster
	err := r.db.QueryRowContext(ctx, query, id, kubeconfigEncrypted, status).Scan(
		&cluster.ID,
		&cluster.OrgID,
		&cluster.Name,
		&cluster.Provider,
		&cluster.Region,
		&cluster.NodeCount,
		&cluster.Status,
		&cluster.KubeVersion,
		&cluster.LastHealthCheck,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &cluster, nil
}

func (r *clusterRepository) UpdateClusterHealth(ctx context.Context, id uuid.UUID, kubeVersion string) (*domain.Cluster, error) {
	query := `
		UPDATE clusters
		SET kube_version = $2, last_health_check = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, name, provider, region, node_count, status, kube_version, last_health_check, created_at, updated_at
	`

	var cluster domain.Cluster
	err := r.db.QueryRowContext(ctx, query, id, kubeVersion).Scan(
		&cluster.ID,
		&cluster.OrgID,
		&cluster.Name,
		&cluster.Provider,
		&cluster.Region,
		&cluster.NodeCount,
		&cluster.Status,
		&cluster.KubeVersion,
		&cluster.LastHealthCheck,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &cluster, nil
}

func (r *clusterRepository) UpdateClusterNodeCount(ctx context.Context, id uuid.UUID, nodeCount int) (*domain.Cluster, error) {
	query := `
		UPDATE clusters
		SET node_count = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, name, provider, region, node_count, status, kube_version, last_health_check, created_at, updated_at
	`

	var cluster domain.Cluster
	err := r.db.QueryRowContext(ctx, query, id, nodeCount).Scan(
		&cluster.ID,
		&cluster.OrgID,
		&cluster.Name,
		&cluster.Provider,
		&cluster.Region,
		&cluster.NodeCount,
		&cluster.Status,
		&cluster.KubeVersion,
		&cluster.LastHealthCheck,
		&cluster.CreatedAt,
		&cluster.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &cluster, nil
}

func (r *clusterRepository) DeleteCluster(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM clusters WHERE id = $1`

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
