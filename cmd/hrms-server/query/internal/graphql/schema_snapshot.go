package graphqlruntime

import (
	"bytes"

	"github.com/vektah/gqlparser/v2/formatter"
)

// SnapshotSDL returns the GraphQL schema compiled into the gqlgen runtime.
func SnapshotSDL() string {
	var buf bytes.Buffer
	formatter.NewFormatter(&buf).FormatSchema(parsedSchema)
	return buf.String()
}
