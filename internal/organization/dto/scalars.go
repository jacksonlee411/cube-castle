package dto

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
)

type scalarString string

func (scalarString) ImplementsGraphQLType(_ string) bool { return true }

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

// PositionCode 表示 GraphQL PositionCode 标量。
type PositionCode string

// ImplementsGraphQLType 声明 PositionCode 满足 gqlgen 标量接口。
func (PositionCode) ImplementsGraphQLType(name string) bool { return name == "PositionCode" }

// UnmarshalGraphQL 解析 PositionCode。
func (p *PositionCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("PositionCode", input); err != nil {
		return err
	}
	*p = PositionCode(base)
	return nil
}

// JobFamilyGroupCode 表示职族组编码标量。
type JobFamilyGroupCode string

// ImplementsGraphQLType 声明 JobFamilyGroupCode 满足 gqlgen 接口。
func (JobFamilyGroupCode) ImplementsGraphQLType(name string) bool {
	return name == "JobFamilyGroupCode"
}

// UnmarshalGraphQL 解析 JobFamilyGroupCode。
func (c *JobFamilyGroupCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobFamilyGroupCode", input); err != nil {
		return err
	}
	*c = JobFamilyGroupCode(base)
	return nil
}

// JobFamilyCode 表示职族编码标量。
type JobFamilyCode string

// ImplementsGraphQLType 声明 JobFamilyCode 满足 gqlgen 接口。
func (JobFamilyCode) ImplementsGraphQLType(name string) bool { return name == "JobFamilyCode" }

// UnmarshalGraphQL 解析 JobFamilyCode。
func (c *JobFamilyCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobFamilyCode", input); err != nil {
		return err
	}
	*c = JobFamilyCode(base)
	return nil
}

// JobRoleCode 表示职位角色编码标量。
type JobRoleCode string

// ImplementsGraphQLType 声明 JobRoleCode 满足 gqlgen 接口。
func (JobRoleCode) ImplementsGraphQLType(name string) bool { return name == "JobRoleCode" }

// UnmarshalGraphQL 解析 JobRoleCode。
func (c *JobRoleCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobRoleCode", input); err != nil {
		return err
	}
	*c = JobRoleCode(base)
	return nil
}

// JobLevelCode 表示职位级别编码标量。
type JobLevelCode string

// ImplementsGraphQLType 声明 JobLevelCode 满足 gqlgen 接口。
func (JobLevelCode) ImplementsGraphQLType(name string) bool { return name == "JobLevelCode" }

// UnmarshalGraphQL 解析 JobLevelCode。
func (c *JobLevelCode) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("JobLevelCode", input); err != nil {
		return err
	}
	*c = JobLevelCode(base)
	return nil
}

// UUID 表示 GraphQL UUID 标量。
type UUID string

// ImplementsGraphQLType 声明 UUID 满足 gqlgen 接口。
func (UUID) ImplementsGraphQLType(name string) bool { return name == "UUID" }

// UnmarshalGraphQL 解析 UUID。
func (u *UUID) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("UUID", input); err != nil {
		return err
	}
	*u = UUID(base)
	return nil
}

// Date 表示 GraphQL Date 标量。
type Date string

// ImplementsGraphQLType 声明 Date 满足 gqlgen 接口。
func (Date) ImplementsGraphQLType(name string) bool { return name == "Date" }

// UnmarshalGraphQL 解析 Date。
func (d *Date) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("Date", input); err != nil {
		return err
	}
	*d = Date(base)
	return nil
}

// DateTime 表示 GraphQL DateTime 标量。
type DateTime string

// ImplementsGraphQLType 声明 DateTime 满足 gqlgen 接口。
func (DateTime) ImplementsGraphQLType(name string) bool { return name == "DateTime" }

// UnmarshalGraphQL 解析 DateTime。
func (d *DateTime) UnmarshalGraphQL(input interface{}) error {
	var base scalarString
	if err := base.unmarshal("DateTime", input); err != nil {
		return err
	}
	*d = DateTime(base)
	return nil
}

// JSON 表示 GraphQL JSON 标量。
type JSON map[string]interface{}

// ImplementsGraphQLType 声明 JSON 满足 gqlgen 接口。
func (JSON) ImplementsGraphQLType(name string) bool { return name == "JSON" }

// UnmarshalGraphQL 解析 JSON。
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

// MarshalDate 将 Date 序列化为 GraphQL 字符串。
func MarshalDate(value Date) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalDate 解析 GraphQL Date。
func UnmarshalDate(v interface{}) (Date, error) {
	str, err := graphql.UnmarshalString(v)
	return Date(str), err
}

// MarshalDateTime 将 DateTime 序列化为 GraphQL 字符串。
func MarshalDateTime(value DateTime) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalDateTime 解析 GraphQL DateTime。
func UnmarshalDateTime(v interface{}) (DateTime, error) {
	str, err := graphql.UnmarshalString(v)
	return DateTime(str), err
}

// MarshalJobFamilyGroupCode 序列化职族组编码。
func MarshalJobFamilyGroupCode(value JobFamilyGroupCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobFamilyGroupCode 解析职族组编码。
func UnmarshalJobFamilyGroupCode(v interface{}) (JobFamilyGroupCode, error) {
	str, err := graphql.UnmarshalString(v)
	return JobFamilyGroupCode(str), err
}

// MarshalJobFamilyCode 序列化职族编码。
func MarshalJobFamilyCode(value JobFamilyCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobFamilyCode 解析职族编码。
func UnmarshalJobFamilyCode(v interface{}) (JobFamilyCode, error) {
	str, err := graphql.UnmarshalString(v)
	return JobFamilyCode(str), err
}

// MarshalJobRoleCode 序列化职位角色编码。
func MarshalJobRoleCode(value JobRoleCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobRoleCode 解析职位角色编码。
func UnmarshalJobRoleCode(v interface{}) (JobRoleCode, error) {
	str, err := graphql.UnmarshalString(v)
	return JobRoleCode(str), err
}

// MarshalJobLevelCode 序列化职位级别编码。
func MarshalJobLevelCode(value JobLevelCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobLevelCode 解析职位级别编码。
func UnmarshalJobLevelCode(v interface{}) (JobLevelCode, error) {
	str, err := graphql.UnmarshalString(v)
	return JobLevelCode(str), err
}

// MarshalPositionCode 序列化 PositionCode。
func MarshalPositionCode(value PositionCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalPositionCode 解析 PositionCode。
func UnmarshalPositionCode(v interface{}) (PositionCode, error) {
	str, err := graphql.UnmarshalString(v)
	return PositionCode(str), err
}

// MarshalUUID 序列化 UUID。
func MarshalUUID(value UUID) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalUUID 解析 UUID。
func UnmarshalUUID(v interface{}) (UUID, error) {
	str, err := graphql.UnmarshalString(v)
	return UUID(str), err
}

// MarshalJSON 序列化 JSON 标量。
func MarshalJSON(value JSON) graphql.Marshaler {
	if value == nil {
		return graphql.Null
	}
	return graphql.MarshalAny(map[string]interface{}(value))
}

// UnmarshalJSON 解析 JSON 标量。
func UnmarshalJSON(v interface{}) (JSON, error) {
	if v == nil {
		return nil, nil
	}
	switch data := v.(type) {
	case map[string]interface{}:
		return JSON(data), nil
	default:
		return nil, fmt.Errorf("JSON: 期望对象类型，实际得到 %T", v)
	}
}
