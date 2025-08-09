package jpack

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_Validate(t *testing.T) {
	dt := &DateTime{}

	t.Run("Valid time.Time", func(t *testing.T) {
		now := time.Now()
		err := dt.Validate(now)
		assert.NoError(t, err)
	})

	t.Run("Valid RFC3339 string", func(t *testing.T) {
		err := dt.Validate("2024-12-25T10:00:00Z")
		assert.NoError(t, err)
	})

	t.Run("Valid RFC3339 string with timezone", func(t *testing.T) {
		err := dt.Validate("2024-12-25T10:00:00+05:30")
		assert.NoError(t, err)
	})

	t.Run("Nil value", func(t *testing.T) {
		err := dt.Validate(nil)
		assert.NoError(t, err)
	})

	t.Run("Invalid string format", func(t *testing.T) {
		err := dt.Validate("2024-12-25 10:00:00")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RFC3339")
	})

	t.Run("Invalid type", func(t *testing.T) {
		err := dt.Validate(123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "datetime type")
	})

	t.Run("Pointer to time.Time", func(t *testing.T) {
		now := time.Now()
		err := dt.Validate(&now)
		assert.NoError(t, err)
	})

	t.Run("Nil pointer", func(t *testing.T) {
		var ptr *time.Time
		err := dt.Validate(ptr)
		assert.NoError(t, err)
	})
}

func TestDateTime_Scan(t *testing.T) {
	dt := &DateTime{}
	ctx := context.Background()
	field := &mockField{name: "created_at", fieldType: &DateTime{}}

	t.Run("Valid time.Time from database", func(t *testing.T) {
		now := time.Now()
		row := map[string]any{"created_at": now}

		value, err := dt.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.NotNil(t, value)

		result, ok := value.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, now.UTC(), result)
	})

	t.Run("Valid RFC3339 string from database", func(t *testing.T) {
		row := map[string]any{"created_at": "2024-12-25T10:00:00Z"}

		value, err := dt.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.NotNil(t, value)

		result, ok := value.(time.Time)
		assert.True(t, ok)
		expected, _ := time.Parse(time.RFC3339, "2024-12-25T10:00:00Z")
		assert.Equal(t, expected, result)
	})

	t.Run("String with timezone converted to GMT", func(t *testing.T) {
		row := map[string]any{"created_at": "2024-12-25T10:00:00+05:30"}

		value, err := dt.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.NotNil(t, value)

		result, ok := value.(time.Time)
		assert.True(t, ok)
		// Should be converted to GMT (UTC)
		assert.Equal(t, "UTC", result.Location().String())
	})

	t.Run("Nil value", func(t *testing.T) {
		row := map[string]any{"created_at": nil}

		value, err := dt.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("Missing field", func(t *testing.T) {
		row := map[string]any{}

		value, err := dt.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("Invalid string format", func(t *testing.T) {
		row := map[string]any{"created_at": "invalid-date"}

		value, err := dt.Scan(ctx, field, row)
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "RFC3339")
	})

	t.Run("Invalid type", func(t *testing.T) {
		row := map[string]any{"created_at": 123}

		value, err := dt.Scan(ctx, field, row)
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "datetime type")
	})
}

func TestDateTime_SetValue(t *testing.T) {
	dt := &DateTime{}
	ctx := context.Background()
	field := &mockField{name: "created_at", fieldType: &DateTime{}}

	t.Run("Set time.Time value", func(t *testing.T) {
		now := time.Now()
		row := make(map[string]any)

		err := dt.SetValue(ctx, field, now, row)
		assert.NoError(t, err)

		value, exists := row["created_at"]
		assert.True(t, exists)
		assert.NotNil(t, value)

		result, ok := value.(time.Time)
		assert.True(t, ok)
		assert.Equal(t, now.UTC(), result)
	})

	t.Run("Set RFC3339 string value", func(t *testing.T) {
		row := make(map[string]any)

		err := dt.SetValue(ctx, field, "2024-12-25T10:00:00Z", row)
		assert.NoError(t, err)

		value, exists := row["created_at"]
		assert.True(t, exists)
		assert.NotNil(t, value)

		result, ok := value.(time.Time)
		assert.True(t, ok)
		expected, _ := time.Parse(time.RFC3339, "2024-12-25T10:00:00Z")
		assert.Equal(t, expected, result)
	})

	t.Run("Set string with timezone converted to GMT", func(t *testing.T) {
		row := make(map[string]any)

		err := dt.SetValue(ctx, field, "2024-12-25T10:00:00+05:30", row)
		assert.NoError(t, err)

		value, exists := row["created_at"]
		assert.True(t, exists)
		assert.NotNil(t, value)

		result, ok := value.(time.Time)
		assert.True(t, ok)
		// Should be converted to GMT (UTC)
		assert.Equal(t, "UTC", result.Location().String())
	})

	t.Run("Set nil value", func(t *testing.T) {
		row := make(map[string]any)

		err := dt.SetValue(ctx, field, nil, row)
		assert.NoError(t, err)

		value, exists := row["created_at"]
		assert.True(t, exists)
		assert.Nil(t, value)
	})

	t.Run("Set invalid string format", func(t *testing.T) {
		row := make(map[string]any)

		err := dt.SetValue(ctx, field, "invalid-date", row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RFC3339")
	})

	t.Run("Set invalid type", func(t *testing.T) {
		row := make(map[string]any)

		err := dt.SetValue(ctx, field, 123, row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "datetime type")
	})
}

func TestDateTime_GMTZoneHandling(t *testing.T) {
	dt := &DateTime{}
	ctx := context.Background()
	field := &mockField{name: "created_at", fieldType: &DateTime{}}

	t.Run("Local time converted to GMT", func(t *testing.T) {
		// Create a time in a specific timezone
		loc, _ := time.LoadLocation("America/New_York")
		localTime := time.Date(2024, 12, 25, 10, 0, 0, 0, loc)

		row := make(map[string]any)
		err := dt.SetValue(ctx, field, localTime, row)
		assert.NoError(t, err)

		value := row["created_at"].(time.Time)
		assert.Equal(t, "UTC", value.Location().String())

		// The time should be converted to GMT
		expectedGMT := localTime.UTC()
		assert.Equal(t, expectedGMT, value)
	})

	t.Run("String with timezone offset converted to GMT", func(t *testing.T) {
		row := make(map[string]any)
		err := dt.SetValue(ctx, field, "2024-12-25T10:00:00+05:30", row)
		assert.NoError(t, err)

		value := row["created_at"].(time.Time)
		assert.Equal(t, "UTC", value.Location().String())

		// Verify the time is correctly converted
		expected, _ := time.Parse(time.RFC3339, "2024-12-25T04:30:00Z")
		assert.Equal(t, expected, value)
	})

	t.Run("Scan preserves GMT conversion", func(t *testing.T) {
		// Store a local time (it gets converted to GMT)
		loc, _ := time.LoadLocation("America/New_York")
		localTime := time.Date(2024, 12, 25, 10, 0, 0, 0, loc)

		row := map[string]any{"created_at": localTime.UTC()}

		value, err := dt.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.NotNil(t, value)

		result := value.(time.Time)
		assert.Equal(t, "UTC", result.Location().String())
		assert.Equal(t, localTime.UTC(), result)
	})
}

// Use existing mockField from field_types_test.go
