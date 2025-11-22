package graphqlruntime

import (
	"context"

	"cube-castle/internal/organization/dto"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

func (ec *executionContext) unmarshalInputJSON(_ context.Context, v interface{}) (dto.JSON, error) {
	return UnmarshalJSON(v)
}

func (ec *executionContext) _JSON(_ context.Context, _ ast.SelectionSet, v dto.JSON) graphql.Marshaler {
	return MarshalJSON(v)
}
