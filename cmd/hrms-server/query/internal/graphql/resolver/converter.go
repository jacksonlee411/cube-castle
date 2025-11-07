package resolver

import (
	"encoding/json"

	"cube-castle/internal/organization/dto"
)

type stringScalar interface {
	~string
}

func convertToModel[T any](input interface{}) (*T, error) {
	if input == nil {
		return nil, nil
	}
	var out T
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func convertSlice[T any](input interface{}) ([]T, error) {
	if input == nil {
		return nil, nil
	}
	var out []T
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func convertSingleResult[T any](data interface{}, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	return convertToModel[T](data)
}

func convertSliceResult[T any](data interface{}, err error) ([]T, error) {
	if err != nil {
		return nil, err
	}
	return convertSlice[T](data)
}

func convertInput[S any, D any](input *S) (*D, error) {
	if input == nil {
		return nil, nil
	}
	var out D
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func convertInputSlice[S any, D any](input []S) ([]D, error) {
	if input == nil {
		return nil, nil
	}
	var out []D
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func convertInputSlicePointer[S any, D any](input []S) (*[]D, error) {
	if len(input) == 0 {
		return nil, nil
	}
	converted, err := convertInputSlice[S, D](input)
	if err != nil {
		return nil, err
	}
	return &converted, nil
}

func scalarToString[T stringScalar](value T) string {
	return string(value)
}

func scalarPtrToStringPtr[T stringScalar](value *T) *string {
	if value == nil {
		return nil
	}
	str := string(*value)
	return &str
}

func dateToStringPtr(value *dto.Date) *string {
	return scalarPtrToStringPtr(value)
}

func positionCodeToStringPtr(value *dto.PositionCode) *string {
	return scalarPtrToStringPtr(value)
}
