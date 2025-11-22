package graphqlruntime

import (
	"fmt"

	"cube-castle/internal/organization/dto"
	"github.com/99designs/gqlgen/graphql"
)

// MarshalDate 将领域 Date 序列化为 GraphQL 字符串。
func MarshalDate(value dto.Date) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalDate 解析 GraphQL 字符串为领域 Date。
func UnmarshalDate(v interface{}) (dto.Date, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.Date(str), err
}

// MarshalDateTime 将领域 DateTime 序列化为 GraphQL 字符串。
func MarshalDateTime(value dto.DateTime) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalDateTime 解析 GraphQL 字符串为领域 DateTime。
func UnmarshalDateTime(v interface{}) (dto.DateTime, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.DateTime(str), err
}

// MarshalJobFamilyGroupCode 序列化职族组编码。
func MarshalJobFamilyGroupCode(value dto.JobFamilyGroupCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobFamilyGroupCode 解析职族组编码。
func UnmarshalJobFamilyGroupCode(v interface{}) (dto.JobFamilyGroupCode, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.JobFamilyGroupCode(str), err
}

// MarshalJobFamilyCode 序列化职族编码。
func MarshalJobFamilyCode(value dto.JobFamilyCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobFamilyCode 解析职族编码。
func UnmarshalJobFamilyCode(v interface{}) (dto.JobFamilyCode, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.JobFamilyCode(str), err
}

// MarshalJobRoleCode 序列化职位角色编码。
func MarshalJobRoleCode(value dto.JobRoleCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobRoleCode 解析职位角色编码。
func UnmarshalJobRoleCode(v interface{}) (dto.JobRoleCode, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.JobRoleCode(str), err
}

// MarshalJobLevelCode 序列化职位级别编码。
func MarshalJobLevelCode(value dto.JobLevelCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalJobLevelCode 解析职位级别编码。
func UnmarshalJobLevelCode(v interface{}) (dto.JobLevelCode, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.JobLevelCode(str), err
}

// MarshalPositionCode 序列化职位编码。
func MarshalPositionCode(value dto.PositionCode) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalPositionCode 解析职位编码。
func UnmarshalPositionCode(v interface{}) (dto.PositionCode, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.PositionCode(str), err
}

// MarshalUUID 序列化通用 UUID。
func MarshalUUID(value dto.UUID) graphql.Marshaler {
	return graphql.MarshalString(string(value))
}

// UnmarshalUUID 解析 GraphQL UUID。
func UnmarshalUUID(v interface{}) (dto.UUID, error) {
	str, err := graphql.UnmarshalString(v)
	return dto.UUID(str), err
}

// MarshalJSON 将任意 JSON 对象序列化为 GraphQL 值。
func MarshalJSON(value dto.JSON) graphql.Marshaler {
	if value == nil {
		return graphql.Null
	}
	return graphql.MarshalAny(map[string]interface{}(value))
}

// UnmarshalJSON 解析任意 JSON 对象。
func UnmarshalJSON(v interface{}) (dto.JSON, error) {
	if v == nil {
		return nil, nil
	}
	switch data := v.(type) {
	case map[string]interface{}:
		return dto.JSON(data), nil
	default:
		return nil, fmt.Errorf("JSON: 期望对象类型，实际得到 %T", v)
	}
}
