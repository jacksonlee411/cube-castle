package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type scalarString string

func (scalarString) ImplementsGraphQLType(name string) bool { return true }

func (s *scalarString) unmarshal(name string, input interface{}) error {
	if input == nil {
		return nil
	}
	str, ok := input.(string)
	if !ok {
		return fmt.Errorf("%s: 期望字符串，实际得到 %T", name, input)
	}
	trimmed := strings.TrimSpace(str)
	value := scalarString(trimmed)
	*s = value
	return nil
}

func (s scalarString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

type PositionCode string

func (PositionCode) ImplementsGraphQLType(name string) bool { return name == "PositionCode" }

func (p *PositionCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("PositionCode", input); err != nil {
		return err
	}
	*p = PositionCode(base)
	return nil
}

type JobFamilyGroupCode string

func (JobFamilyGroupCode) ImplementsGraphQLType(name string) bool {
	return name == "JobFamilyGroupCode"
}

func (c *JobFamilyGroupCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobFamilyGroupCode", input); err != nil {
		return err
	}
	*c = JobFamilyGroupCode(base)
	return nil
}

type JobFamilyCode string

func (JobFamilyCode) ImplementsGraphQLType(name string) bool { return name == "JobFamilyCode" }

func (c *JobFamilyCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobFamilyCode", input); err != nil {
		return err
	}
	*c = JobFamilyCode(base)
	return nil
}

type JobRoleCode string

func (JobRoleCode) ImplementsGraphQLType(name string) bool { return name == "JobRoleCode" }

func (c *JobRoleCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobRoleCode", input); err != nil {
		return err
	}
	*c = JobRoleCode(base)
	return nil
}

type JobLevelCode string

func (JobLevelCode) ImplementsGraphQLType(name string) bool { return name == "JobLevelCode" }

func (c *JobLevelCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobLevelCode", input); err != nil {
		return err
	}
	*c = JobLevelCode(base)
	return nil
}

type UUID string

func (UUID) ImplementsGraphQLType(name string) bool { return name == "UUID" }

func (u *UUID) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("UUID", input); err != nil {
		return err
	}
	*u = UUID(base)
	return nil
}

type Date string

func (Date) ImplementsGraphQLType(name string) bool { return name == "Date" }

func (d *Date) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("Date", input); err != nil {
		return err
	}
	*d = Date(base)
	return nil
}

type DateTime string

func (DateTime) ImplementsGraphQLType(name string) bool { return name == "DateTime" }

func (d *DateTime) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("DateTime", input); err != nil {
		return err
	}
	*d = DateTime(base)
	return nil
}

type JSON map[string]interface{}

func (JSON) ImplementsGraphQLType(name string) bool { return name == "JSON" }

func (j *JSON) UnmarshalGraphQL(input interface{}) error {
	if input == nil {
		*j = nil
		return nil
	}
	switch value := input.(type) {
	case map[string]interface{}:
		*j = JSON(value)
		return nil
	default:
		return fmt.Errorf("JSON: 期望对象类型，实际得到 %T", input)
	}
}
