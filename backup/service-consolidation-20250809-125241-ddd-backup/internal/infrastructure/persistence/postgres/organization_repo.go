package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"github.com/cube-castle/cmd/organization-command-server/internal/domain/entities"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/repositories"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/valueobjects"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/logging"
)

// PostgresOrganizationRepository implements OrganizationRepository using PostgreSQL
type PostgresOrganizationRepository struct {
	pool   *pgxpool.Pool
	logger logging.Logger
}

// NewPostgresOrganizationRepository creates a new PostgreSQL repository
func NewPostgresOrganizationRepository(pool *pgxpool.Pool, logger logging.Logger) *PostgresOrganizationRepository {
	return &PostgresOrganizationRepository{
		pool:   pool,
		logger: logger,
	}
}

// Create creates a new organization in the database
func (r *PostgresOrganizationRepository) Create(ctx context.Context, org *entities.Organization) error {
	query := `
		INSERT INTO organization_units (
			code, parent_code, tenant_id, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	var parentCodePtr *string
	if org.ParentCode() != nil && !org.ParentCode().IsEmpty() {
		s := org.ParentCode().String()
		parentCodePtr = &s
	}

	_, err := r.pool.Exec(ctx, query,
		org.Code().String(),
		parentCodePtr,
		org.TenantID(),
		org.Name(),
		org.UnitType().String(),
		org.Status().String(),
		org.Level(),
		org.Path(),
		org.SortOrder(),
		org.Description(),
		org.CreatedAt(),
		org.UpdatedAt(),
	)

	if err != nil {
		r.logger.Error("failed to create organization",
			"code", org.Code().String(),
			"error", err,
		)
		return fmt.Errorf("failed to create organization: %w", err)
	}

	r.logger.Debug("organization created successfully",
		"code", org.Code().String(),
		"name", org.Name(),
	)

	return nil
}

// Update updates an existing organization in the database
func (r *PostgresOrganizationRepository) Update(ctx context.Context, org *entities.Organization) error {
	query := `
		UPDATE organization_units 
		SET name = $1, status = $2, sort_order = $3, description = $4, updated_at = $5
		WHERE code = $6 AND tenant_id = $7`

	result, err := r.pool.Exec(ctx, query,
		org.Name(),
		org.Status().String(),
		org.SortOrder(),
		org.Description(),
		org.UpdatedAt(),
		org.Code().String(),
		org.TenantID(),
	)

	if err != nil {
		r.logger.Error("failed to update organization",
			"code", org.Code().String(),
			"error", err,
		)
		return fmt.Errorf("failed to update organization: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("organization not found: %s", org.Code().String())
	}

	r.logger.Debug("organization updated successfully",
		"code", org.Code().String(),
		"rows_affected", result.RowsAffected(),
	)

	return nil
}

// Delete soft deletes an organization (implementation uses Update with status change)
func (r *PostgresOrganizationRepository) Delete(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) error {
	query := `
		UPDATE organization_units 
		SET status = 'INACTIVE', updated_at = $1 
		WHERE code = $2 AND tenant_id = $3`

	result, err := r.pool.Exec(ctx, query, time.Now(), code.String(), tenantID)
	if err != nil {
		r.logger.Error("failed to delete organization",
			"code", code.String(),
			"error", err,
		)
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("organization not found: %s", code.String())
	}

	r.logger.Debug("organization deleted successfully",
		"code", code.String(),
	)

	return nil
}

// FindByCode finds an organization by its code
func (r *PostgresOrganizationRepository) FindByCode(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) (*entities.Organization, error) {
	query := `
		SELECT code, parent_code, tenant_id, name, unit_type, status, 
		       level, path, sort_order, description, created_at, updated_at
		FROM organization_units 
		WHERE code = $1 AND tenant_id = $2`

	var (
		codeStr, name, unitTypeStr, statusStr, path string
		parentCodePtr, descriptionPtr               *string
		level, sortOrder                            int
		orgTenantID                                 uuid.UUID
		createdAt, updatedAt                        time.Time
	)

	err := r.pool.QueryRow(ctx, query, code.String(), tenantID).Scan(
		&codeStr, &parentCodePtr, &orgTenantID, &name, &unitTypeStr, &statusStr,
		&level, &path, &sortOrder, &descriptionPtr, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization not found: %s", code.String())
		}
		r.logger.Error("failed to find organization",
			"code", code.String(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	// Parse domain objects
	orgCode, _ := valueobjects.NewOrganizationCode(codeStr)
	unitType, _ := entities.ParseUnitType(unitTypeStr)
	status, _ := entities.ParseStatus(statusStr)

	var parentCode *valueobjects.OrganizationCode
	if parentCodePtr != nil {
		pc, _ := valueobjects.NewOrganizationCode(*parentCodePtr)
		parentCode = &pc
	}

	// Create organization entity (simplified constructor for loading from DB)
	org, err := entities.NewOrganization(
		orgCode, name, unitType, orgTenantID, parentCode, level, path, sortOrder, descriptionPtr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization entity: %w", err)
	}

	// Clear events since this is loaded from DB, not a new creation
	org.ClearEvents()

	return org, nil
}

// FindChildren finds child organizations
func (r *PostgresOrganizationRepository) FindChildren(ctx context.Context, parentCode valueobjects.OrganizationCode, tenantID uuid.UUID) ([]*entities.Organization, error) {
	query := `
		SELECT code, parent_code, tenant_id, name, unit_type, status, 
		       level, path, sort_order, description, created_at, updated_at
		FROM organization_units 
		WHERE parent_code = $1 AND tenant_id = $2
		ORDER BY sort_order, name`

	rows, err := r.pool.Query(ctx, query, parentCode.String(), tenantID)
	if err != nil {
		r.logger.Error("failed to find children organizations",
			"parent_code", parentCode.String(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to find children organizations: %w", err)
	}
	defer rows.Close()

	var organizations []*entities.Organization

	for rows.Next() {
		var (
			codeStr, name, unitTypeStr, statusStr, path string
			parentCodePtr, descriptionPtr               *string
			level, sortOrder                            int
			orgTenantID                                 uuid.UUID
			createdAt, updatedAt                        time.Time
		)

		err := rows.Scan(
			&codeStr, &parentCodePtr, &orgTenantID, &name, &unitTypeStr, &statusStr,
			&level, &path, &sortOrder, &descriptionPtr, &createdAt, &updatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan organization row", "error", err)
			continue
		}

		// Parse domain objects
		orgCode, _ := valueobjects.NewOrganizationCode(codeStr)
		unitType, _ := entities.ParseUnitType(unitTypeStr)

		var parentCodeObj *valueobjects.OrganizationCode
		if parentCodePtr != nil {
			pc, _ := valueobjects.NewOrganizationCode(*parentCodePtr)
			parentCodeObj = &pc
		}

		org, err := entities.NewOrganization(
			orgCode, name, unitType, orgTenantID, parentCodeObj, level, path, sortOrder, descriptionPtr,
		)
		if err != nil {
			r.logger.Error("failed to create organization entity", "error", err)
			continue
		}

		org.ClearEvents()
		organizations = append(organizations, org)
	}

	return organizations, nil
}

// Exists checks if an organization exists
func (r *PostgresOrganizationRepository) Exists(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM organization_units WHERE code = $1 AND tenant_id = $2`

	var count int
	err := r.pool.QueryRow(ctx, query, code.String(), tenantID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to check organization existence",
			"code", code.String(),
			"error", err,
		)
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}

	return count > 0, nil
}

// GenerateNextCode generates the next available organization code
func (r *PostgresOrganizationRepository) GenerateNextCode(ctx context.Context, tenantID uuid.UUID) (valueobjects.OrganizationCode, error) {
	query := `
		SELECT COALESCE(MAX(CAST(code AS INTEGER)), 1000000) + 1
		FROM organization_units 
		WHERE tenant_id = $1 AND code ~ '^[0-9]+$'`

	var nextCodeInt int
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(&nextCodeInt)
	if err != nil {
		r.logger.Error("failed to generate next organization code",
			"tenant_id", tenantID,
			"error", err,
		)
		return valueobjects.OrganizationCode{}, fmt.Errorf("failed to generate next organization code: %w", err)
	}

	codeStr := strconv.Itoa(nextCodeInt)
	code, err := valueobjects.NewOrganizationCode(codeStr)
	if err != nil {
		return valueobjects.OrganizationCode{}, fmt.Errorf("failed to create organization code: %w", err)
	}

	r.logger.Debug("generated next organization code",
		"code", codeStr,
		"tenant_id", tenantID,
	)

	return code, nil
}

// HasChildren checks if an organization has children
func (r *PostgresOrganizationRepository) HasChildren(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM organization_units WHERE parent_code = $1 AND tenant_id = $2`

	var count int
	err := r.pool.QueryRow(ctx, query, code.String(), tenantID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to check for child organizations",
			"code", code.String(),
			"error", err,
		)
		return false, fmt.Errorf("failed to check for child organizations: %w", err)
	}

	return count > 0, nil
}

// GetParentInfo gets parent organization information
func (r *PostgresOrganizationRepository) GetParentInfo(ctx context.Context, parentCode valueobjects.OrganizationCode, tenantID uuid.UUID) (*repositories.ParentInfo, error) {
	query := `SELECT level, path FROM organization_units WHERE code = $1 AND tenant_id = $2`

	var level int
	var path string

	err := r.pool.QueryRow(ctx, query, parentCode.String(), tenantID).Scan(&level, &path)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("parent organization not found: %s", parentCode.String())
		}
		r.logger.Error("failed to get parent organization info",
			"parent_code", parentCode.String(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to get parent organization info: %w", err)
	}

	return &repositories.ParentInfo{
		Level: level,
		Path:  path,
	}, nil
}