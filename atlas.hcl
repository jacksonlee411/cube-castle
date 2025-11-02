env "dev" {
  src = "file://database/schema.sql"
  dev = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"

  migration {
    dir    = "file://database/migrations"
    format = goose
  }
}
