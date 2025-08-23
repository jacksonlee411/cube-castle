package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Date 自定义日期类型，用于处理PostgreSQL的date类型
type Date struct {
	time.Time
}

// NewDate 创建新的日期
func NewDate(year int, month time.Month, day int) *Date {
	return &Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// ParseDate 解析日期字符串 (YYYY-MM-DD)
func ParseDate(s string) (*Date, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &Date{t}, nil
}

// MarshalJSON 实现JSON序列化
func (d *Date) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format("2006-01-02"))
}

// UnmarshalJSON 实现JSON反序列化
func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" || s == "null" {
		return nil
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = *parsed
	return nil
}

// Scan 实现sql.Scanner接口
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*d = Date{v}
		return nil
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = *parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Date", value)
	}
}

// Value 实现driver.Valuer接口
func (d Date) Value() (driver.Value, error) {
	return d.Time, nil
}

// String 返回日期字符串
func (d *Date) String() string {
	if d == nil {
		return ""
	}
	return d.Format("2006-01-02")
}