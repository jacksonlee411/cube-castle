package repository

import (
	"database/sql"
	"errors"
	"log"
)

type OrganizationRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewOrganizationRepository(db *sql.DB, logger *log.Logger) *OrganizationRepository {
	return &OrganizationRepository{db: db, logger: logger}
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
