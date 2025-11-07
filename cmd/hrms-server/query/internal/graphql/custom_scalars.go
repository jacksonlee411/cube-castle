package graphqlruntime

import (
    "fmt"

    "cube-castle/internal/organization/dto"
    "github.com/99designs/gqlgen/graphql"
)

func MarshalDate(value dto.Date) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalDate(v interface{}) (dto.Date, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.Date(str), err
}

func MarshalDateTime(value dto.DateTime) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalDateTime(v interface{}) (dto.DateTime, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.DateTime(str), err
}

func MarshalJobFamilyGroupCode(value dto.JobFamilyGroupCode) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalJobFamilyGroupCode(v interface{}) (dto.JobFamilyGroupCode, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.JobFamilyGroupCode(str), err
}

func MarshalJobFamilyCode(value dto.JobFamilyCode) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalJobFamilyCode(v interface{}) (dto.JobFamilyCode, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.JobFamilyCode(str), err
}

func MarshalJobRoleCode(value dto.JobRoleCode) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalJobRoleCode(v interface{}) (dto.JobRoleCode, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.JobRoleCode(str), err
}

func MarshalJobLevelCode(value dto.JobLevelCode) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalJobLevelCode(v interface{}) (dto.JobLevelCode, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.JobLevelCode(str), err
}

func MarshalPositionCode(value dto.PositionCode) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalPositionCode(v interface{}) (dto.PositionCode, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.PositionCode(str), err
}

func MarshalUUID(value dto.UUID) graphql.Marshaler {
    return graphql.MarshalString(string(value))
}

func UnmarshalUUID(v interface{}) (dto.UUID, error) {
    str, err := graphql.UnmarshalString(v)
    return dto.UUID(str), err
}

func MarshalJSON(value dto.JSON) graphql.Marshaler {
    if value == nil {
        return graphql.Null
    }
    return graphql.MarshalAny(map[string]interface{}(value))
}

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
