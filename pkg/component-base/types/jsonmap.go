package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type JSONMap map[string]string

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONMap) // 如果为空，设置为一个空的 map
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	// 如果数据为空字符串，设置为空 map
	if len(bytes) == 0 {
		*m = make(JSONMap)
		return nil
	}

	// 尝试解析 JSON 数据
	return json.Unmarshal(bytes, m)
}

func (m *JSONMap) GormDataType() string {
	return "text"
}

func (m *JSONMap) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	// returns different database type based on driver name
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "text"
	}
	return ""
}
