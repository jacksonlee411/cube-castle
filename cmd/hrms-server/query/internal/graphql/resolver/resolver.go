package resolver

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import "cube-castle/internal/organization"

type Resolver struct {
	QueryResolver *organization.QueryResolver
}

func New(queryResolver *organization.QueryResolver) *Resolver {
	return &Resolver{QueryResolver: queryResolver}
}
