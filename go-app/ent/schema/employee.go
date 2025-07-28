package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "time"
)

// Employee holds the schema definition for the Employee entity.
type Employee struct {
    ent.Schema
}

// Fields of the Employee.
func (Employee) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique(),
        field.String("name"),
        field.String("email"),
        field.String("position"),
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}