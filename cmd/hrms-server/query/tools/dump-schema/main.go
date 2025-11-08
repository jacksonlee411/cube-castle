package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	graphqlruntime "cube-castle/cmd/hrms-server/query/internal/graphql"
)

func main() {
	out := flag.String("out", "", "Path to write the schema snapshot (defaults to stdout)")
	flag.Parse()

	schema := graphqlruntime.SnapshotSDL()
	if *out == "" {
		fmt.Print(schema)
		return
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*out, []byte(schema), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write schema snapshot: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "✅ GraphQL runtime schema snapshot saved to %s\n", *out)
}
