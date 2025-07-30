package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"

	"github.com/gaogu/cube-castle/go-app/ent"
)

func main() {
	client, err := ent.Open("postgres", "postgresql://user:password@localhost:5432/cubecastle?sslmode=disable")
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Create the schema DDL
	var w io.Writer = os.Stdout
	if len(os.Args) > 1 && os.Args[1] == "--file" {
		file, err := os.Create("ent/migrate/schema.sql")
		if err != nil {
			log.Fatalf("failed creating schema file: %v", err)
		}
		defer file.Close()
		w = file
	}

	// Generate and output the schema
	err = client.Schema.WriteTo(ctx, w,
		schema.WithDropIndex(true),
		schema.WithDropColumn(true),
		schema.WithForeignKeys(true),
	)
	if err != nil {
		log.Fatalf("failed generating schema: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "--file" {
		fmt.Println("Database schema written to ent/migrate/schema.sql")
	}
}
