package main

import (
	"context"
	"fmt"
	"log"

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

	// Create schema diff and apply migration
	err = client.Schema.Create(ctx,
		schema.WithDropIndex(true),
		schema.WithDropColumn(true),
		schema.WithForeignKeys(true),
	)
	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	fmt.Println("Database schema migration completed successfully!")
	fmt.Println("Created tables: organization_units, positions, position_attribute_histories, position_occupancy_histories")
}
