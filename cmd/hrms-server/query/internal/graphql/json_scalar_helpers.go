package graphqlruntime

import (
    "context"

    "cube-castle/internal/organization/dto"
    "github.com/99designs/gqlgen/graphql"
    "github.com/vektah/gqlparser/v2/ast"
)

func (ec *executionContext) unmarshalInputJSON(ctx context.Context, v interface{}) (dto.JSON, error) {
    return UnmarshalJSON(v)
}

func (ec *executionContext) _JSON(ctx context.Context, sel ast.SelectionSet, v dto.JSON) graphql.Marshaler {
    return MarshalJSON(v)
}
