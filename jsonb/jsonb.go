package chi_jsonb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONBValue marshals v to a JSON string for Postgres jsonb columns.
func JSONBValue(v any) (driver.Value, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// JSONBScan unmarshals a jsonb column into dest. dest must be a pointer.
func JSONBScan(value any, dest any) error {
	if value == nil {
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("JSONBScan: value must be []byte or string")
	}
	return json.Unmarshal(data, dest)
}
