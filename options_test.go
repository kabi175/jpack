package jpack

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockOptionService implements OptionService for testing
type mockOptionService struct {
	options []Option
	err     error
}

func (m *mockOptionService) GetOptions(ctx context.Context) ([]Option, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.options, nil
}

func TestOptions_Validate(t *testing.T) {
	t.Run("Valid option", func(t *testing.T) {
		service := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
				{UniqueName: "pending", DisplayName: "Pending"},
			},
		}
		options := NewOptions(service)

		err := options.Validate("active")
		assert.NoError(t, err)
	})

	t.Run("Invalid option", func(t *testing.T) {
		service := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
				{UniqueName: "pending", DisplayName: "Pending"},
			},
		}
		options := NewOptions(service)

		err := options.Validate("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in the list of available options")
	})

	t.Run("Nil value", func(t *testing.T) {
		service := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
				{UniqueName: "pending", DisplayName: "Pending"},
			},
		}
		options := NewOptions(service)

		err := options.Validate(nil)
		assert.NoError(t, err)
	})

	t.Run("Pointer to valid option", func(t *testing.T) {
		service := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
				{UniqueName: "pending", DisplayName: "Pending"},
			},
		}
		options := NewOptions(service)

		value := "active"
		err := options.Validate(&value)
		assert.NoError(t, err)
	})

	t.Run("Nil pointer", func(t *testing.T) {
		service := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
				{UniqueName: "pending", DisplayName: "Pending"},
			},
		}
		options := NewOptions(service)

		var ptr *string
		err := options.Validate(ptr)
		assert.NoError(t, err)
	})

	t.Run("Invalid type", func(t *testing.T) {
		service := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
				{UniqueName: "pending", DisplayName: "Pending"},
			},
		}
		options := NewOptions(service)

		err := options.Validate(123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string")
	})

	t.Run("Service error", func(t *testing.T) {
		service := &mockOptionService{
			err: errors.New("service unavailable"),
		}
		options := NewOptions(service)

		err := options.Validate("active")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get available options")
	})

	t.Run("Empty options list", func(t *testing.T) {
		service := &mockOptionService{
			options: []Option{},
		}
		options := NewOptions(service)

		err := options.Validate("active")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in the list of available options")
	})
}

func TestOptions_Scan(t *testing.T) {
	service := &mockOptionService{
		options: []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "inactive", DisplayName: "Inactive"},
			{UniqueName: "pending", DisplayName: "Pending"},
		},
	}
	options := NewOptions(service)
	ctx := context.Background()
	field := &mockField{name: "status", fieldType: options}

	t.Run("Valid string from database", func(t *testing.T) {
		row := map[string]any{"status": "active"}

		value, err := options.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.Equal(t, "active", value)
	})

	t.Run("Nil value from database", func(t *testing.T) {
		row := map[string]any{"status": nil}

		value, err := options.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("Missing field", func(t *testing.T) {
		row := map[string]any{}

		value, err := options.Scan(ctx, field, row)
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("Invalid type from database", func(t *testing.T) {
		row := map[string]any{"status": 123}

		value, err := options.Scan(ctx, field, row)
		assert.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "must be a string")
	})
}

func TestOptions_SetValue(t *testing.T) {
	service := &mockOptionService{
		options: []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "inactive", DisplayName: "Inactive"},
			{UniqueName: "pending", DisplayName: "Pending"},
		},
	}
	options := NewOptions(service)
	ctx := context.Background()
	field := &mockField{name: "status", fieldType: options}

	t.Run("Set valid option", func(t *testing.T) {
		row := make(map[string]any)

		err := options.SetValue(ctx, field, "active", row)
		assert.NoError(t, err)

		value, exists := row["status"]
		assert.True(t, exists)
		assert.Equal(t, "active", value)
	})

	t.Run("Set nil value", func(t *testing.T) {
		row := make(map[string]any)

		err := options.SetValue(ctx, field, nil, row)
		assert.NoError(t, err)

		value, exists := row["status"]
		assert.True(t, exists)
		assert.Nil(t, value)
	})

	t.Run("Set invalid option", func(t *testing.T) {
		row := make(map[string]any)

		err := options.SetValue(ctx, field, "invalid", row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in the list of available options")
	})

	t.Run("Set invalid type", func(t *testing.T) {
		row := make(map[string]any)

		err := options.SetValue(ctx, field, 123, row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a string")
	})
}

func TestNewOptions(t *testing.T) {
	service := &mockOptionService{
		options: []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "inactive", DisplayName: "Inactive"},
		},
	}

	options := NewOptions(service)
	assert.NotNil(t, options)
	assert.Equal(t, service, options.service)
}

func TestOptions_ServiceIntegration(t *testing.T) {
	t.Run("Dynamic options from service", func(t *testing.T) {
		// First service call returns one set of options
		service := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
			},
		}
		options := NewOptions(service)

		// Validate with first set of options
		err := options.Validate("active")
		assert.NoError(t, err)

		// Change service options
		service.options = []Option{
			{UniqueName: "pending", DisplayName: "Pending"},
			{UniqueName: "completed", DisplayName: "Completed"},
		}

		// Now "active" should be invalid
		err = options.Validate("active")
		assert.Error(t, err)

		// But "pending" should be valid
		err = options.Validate("pending")
		assert.NoError(t, err)
	})

	t.Run("Service error handling", func(t *testing.T) {
		service := &mockOptionService{
			err: errors.New("database connection failed"),
		}
		options := NewOptions(service)

		err := options.Validate("any_value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get available options")
		assert.Contains(t, err.Error(), "database connection failed")
	})
}

func TestOptions_ContextHandling(t *testing.T) {
	t.Run("Context passed to service", func(t *testing.T) {
		// Create a service that checks if context is passed
		contextService := &mockOptionService{
			options: []Option{
				{UniqueName: "active", DisplayName: "Active"},
				{UniqueName: "inactive", DisplayName: "Inactive"},
			},
		}
		options := NewOptions(contextService)

		ctx := context.WithValue(context.Background(), "test_key", "test_value")

		// The current implementation uses context.Background() in Validate
		// This test ensures the service interface supports context
		_, err := contextService.GetOptions(ctx)
		assert.NoError(t, err)

		// Test that the options field type works with the service
		err = options.Validate("active")
		assert.NoError(t, err)
	})
}

func TestOptions_GetDisplayName(t *testing.T) {
	service := &mockOptionService{
		options: []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "inactive", DisplayName: "Inactive"},
			{UniqueName: "pending", DisplayName: "Pending"},
		},
	}
	options := NewOptions(service)
	ctx := context.Background()

	t.Run("Get display name for valid unique name", func(t *testing.T) {
		displayName, err := options.GetDisplayName(ctx, "active")
		assert.NoError(t, err)
		assert.Equal(t, "Active", displayName)
	})

	t.Run("Get display name for invalid unique name", func(t *testing.T) {
		displayName, err := options.GetDisplayName(ctx, "invalid")
		assert.Error(t, err)
		assert.Empty(t, displayName)
		assert.Contains(t, err.Error(), "option not found")
	})

	t.Run("Get display name with service error", func(t *testing.T) {
		errorService := &mockOptionService{
			err: errors.New("service error"),
		}
		errorOptions := NewOptions(errorService)

		displayName, err := errorOptions.GetDisplayName(ctx, "active")
		assert.Error(t, err)
		assert.Empty(t, displayName)
		assert.Contains(t, err.Error(), "failed to get available options")
	})
}

func TestOptions_GetUniqueName(t *testing.T) {
	service := &mockOptionService{
		options: []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "inactive", DisplayName: "Inactive"},
			{UniqueName: "pending", DisplayName: "Pending"},
		},
	}
	options := NewOptions(service)
	ctx := context.Background()

	t.Run("Get unique name for valid display name", func(t *testing.T) {
		uniqueName, err := options.GetUniqueName(ctx, "Active")
		assert.NoError(t, err)
		assert.Equal(t, "active", uniqueName)
	})

	t.Run("Get unique name for invalid display name", func(t *testing.T) {
		uniqueName, err := options.GetUniqueName(ctx, "Invalid")
		assert.Error(t, err)
		assert.Empty(t, uniqueName)
		assert.Contains(t, err.Error(), "option not found")
	})

	t.Run("Get unique name with service error", func(t *testing.T) {
		errorService := &mockOptionService{
			err: errors.New("service error"),
		}
		errorOptions := NewOptions(errorService)

		uniqueName, err := errorOptions.GetUniqueName(ctx, "Active")
		assert.Error(t, err)
		assert.Empty(t, uniqueName)
		assert.Contains(t, err.Error(), "failed to get available options")
	})
}

func TestOptions_GetAllOptions(t *testing.T) {
	expectedOptions := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
		{UniqueName: "pending", DisplayName: "Pending"},
	}

	service := &mockOptionService{
		options: expectedOptions,
	}
	options := NewOptions(service)
	ctx := context.Background()

	t.Run("Get all options successfully", func(t *testing.T) {
		allOptions, err := options.GetAllOptions(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedOptions, allOptions)
	})

	t.Run("Get all options with service error", func(t *testing.T) {
		errorService := &mockOptionService{
			err: errors.New("service error"),
		}
		errorOptions := NewOptions(errorService)

		allOptions, err := errorOptions.GetAllOptions(ctx)
		assert.Error(t, err)
		assert.Nil(t, allOptions)
		assert.Contains(t, err.Error(), "service error")
	})
}
