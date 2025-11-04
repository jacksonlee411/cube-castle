package repository

import (
	"database/sql"
	"errors"

	pkglogger "cube-castle/pkg/logger"
)

type OrganizationRepository struct {
	db     *sql.DB
	logger pkglogger.Logger
}

func NewOrganizationRepository(db *sql.DB, baseLogger pkglogger.Logger) *OrganizationRepository {
	return &OrganizationRepository{
		db:     db,
		logger: scopedLogger(baseLogger, "organization", "OrganizationRepository", nil),
	}
}

var (
	ErrOrganizationHasChildren  = errors.New("organization has non-deleted child units")
	ErrOrganizationPrecondition = errors.New("organization precondition failed")
)

type OrganizationHasChildrenError struct {
	Count int
}

func (e *OrganizationHasChildrenError) Error() string {
	return ErrOrganizationHasChildren.Error()
}

func (e *OrganizationHasChildrenError) Is(target error) bool {
	return target == ErrOrganizationHasChildren
}

func NewOrganizationHasChildrenError(count int) error {
	return &OrganizationHasChildrenError{Count: count}
}
