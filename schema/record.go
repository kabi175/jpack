package schema

import (
	"encoding/json"
)

// jRecord implements JRecord interface
type jRecord struct {
	data map[string]any
}

// NewJRecord creates a new record
func NewJRecord() JRecord {
	return &jRecord{
		data: make(map[string]any),
	}
}

// NewJRecordFromMap creates a new record from a map
func NewJRecordFromMap(data map[string]any) JRecord {
	return &jRecord{
		data: data,
	}
}

func (r *jRecord) Get(key string) any {
	return r.data[key]
}

func (r *jRecord) Set(key string, value any) JRecord {
	r.data[key] = value
	return r
}

func (r *jRecord) Has(key string) bool {
	_, exists := r.data[key]
	return exists
}

func (r *jRecord) Delete(key string) JRecord {
	delete(r.data, key)
	return r
}

func (r *jRecord) Keys() []string {
	keys := make([]string, 0, len(r.data))
	for k := range r.data {
		keys = append(keys, k)
	}
	return keys
}

func (r *jRecord) ToMap() map[string]any {
	// Return a copy to prevent external modifications
	result := make(map[string]any)
	for k, v := range r.data {
		result[k] = v
	}
	return result
}

func (r *jRecord) FromMap(data map[string]any) JRecord {
	r.data = make(map[string]any)
	for k, v := range data {
		r.data[k] = v
	}
	return r
}

func (r *jRecord) Clone() JRecord {
	return NewJRecordFromMap(r.ToMap())
}

// MarshalJSON implements json.Marshaler
func (r *jRecord) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.data)
}

// UnmarshalJSON implements json.Unmarshaler
func (r *jRecord) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.data)
}
