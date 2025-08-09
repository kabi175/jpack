package jpack

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryOptionService(t *testing.T) {
	t.Run("Create with options", func(t *testing.T) {
		options := []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "inactive", DisplayName: "Inactive"},
		}

		service := NewInMemoryOptionService(options)
		assert.NotNil(t, service)
		assert.Equal(t, 2, service.Count())
	})

	t.Run("Create with empty options", func(t *testing.T) {
		service := NewInMemoryOptionService(nil)
		assert.NotNil(t, service)
		assert.Equal(t, 0, service.Count())
	})
}

func TestInMemoryOptionService_GetOptions(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
		{UniqueName: "pending", DisplayName: "Pending"},
	}

	service := NewInMemoryOptionService(options)
	ctx := context.Background()

	t.Run("Get all options", func(t *testing.T) {
		result, err := service.GetOptions(ctx)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, options, result)
	})

	t.Run("Get options with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := service.GetOptions(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("Get options with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Should not timeout for in-memory service
		result, err := service.GetOptions(ctx)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("Return copy of options", func(t *testing.T) {
		result, err := service.GetOptions(ctx)
		assert.NoError(t, err)
		assert.Len(t, result, 3)

		// Modify the returned slice should not affect the original
		result[0].DisplayName = "Modified"

		// Get options again should return original values
		result2, err := service.GetOptions(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "Active", result2[0].DisplayName)
	})
}

func TestInMemoryOptionService_AddOption(t *testing.T) {
	service := NewInMemoryOptionService(nil)

	t.Run("Add new option", func(t *testing.T) {
		option := Option{UniqueName: "active", DisplayName: "Active"}
		service.AddOption(option)

		assert.Equal(t, 1, service.Count())
		assert.True(t, service.HasOption("active"))
	})

	t.Run("Add duplicate option", func(t *testing.T) {
		option := Option{UniqueName: "active", DisplayName: "Active Modified"}
		service.AddOption(option)

		// Should not add duplicate, count should remain 1
		assert.Equal(t, 1, service.Count())

		// Should still have original display name
		result, err := service.GetOptions(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "Active", result[0].DisplayName)
	})

	t.Run("Add multiple options", func(t *testing.T) {
		service.Clear()

		options := []Option{
			{UniqueName: "active", DisplayName: "Active"},
			{UniqueName: "inactive", DisplayName: "Inactive"},
			{UniqueName: "pending", DisplayName: "Pending"},
		}

		for _, option := range options {
			service.AddOption(option)
		}

		assert.Equal(t, 3, service.Count())
	})
}

func TestInMemoryOptionService_RemoveOption(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
		{UniqueName: "pending", DisplayName: "Pending"},
	}

	service := NewInMemoryOptionService(options)

	t.Run("Remove existing option", func(t *testing.T) {
		removed := service.RemoveOption("inactive")
		assert.True(t, removed)
		assert.Equal(t, 2, service.Count())
		assert.False(t, service.HasOption("inactive"))
		assert.True(t, service.HasOption("active"))
		assert.True(t, service.HasOption("pending"))
	})

	t.Run("Remove non-existing option", func(t *testing.T) {
		removed := service.RemoveOption("non-existing")
		assert.False(t, removed)
		assert.Equal(t, 2, service.Count())
	})

	t.Run("Remove last option", func(t *testing.T) {
		service.Clear()
		service.AddOption(Option{UniqueName: "single", DisplayName: "Single"})

		removed := service.RemoveOption("single")
		assert.True(t, removed)
		assert.Equal(t, 0, service.Count())
	})
}

func TestInMemoryOptionService_UpdateOption(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
	}

	service := NewInMemoryOptionService(options)

	t.Run("Update existing option", func(t *testing.T) {
		updated := Option{UniqueName: "active", DisplayName: "Active Updated"}
		success := service.UpdateOption(updated)

		assert.True(t, success)
		assert.Equal(t, 2, service.Count())

		result, err := service.GetOptions(context.Background())
		assert.NoError(t, err)

		// Find the updated option
		for _, option := range result {
			if option.UniqueName == "active" {
				assert.Equal(t, "Active Updated", option.DisplayName)
				break
			}
		}
	})

	t.Run("Update non-existing option", func(t *testing.T) {
		updated := Option{UniqueName: "non-existing", DisplayName: "Non Existing"}
		success := service.UpdateOption(updated)

		assert.False(t, success)
		assert.Equal(t, 2, service.Count())
	})
}

func TestInMemoryOptionService_GetOptionByUniqueName(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
	}

	service := NewInMemoryOptionService(options)

	t.Run("Get existing option by unique name", func(t *testing.T) {
		option, found := service.GetOptionByUniqueName("active")
		assert.True(t, found)
		assert.Equal(t, "active", option.UniqueName)
		assert.Equal(t, "Active", option.DisplayName)
	})

	t.Run("Get non-existing option by unique name", func(t *testing.T) {
		option, found := service.GetOptionByUniqueName("non-existing")
		assert.False(t, found)
		assert.Equal(t, Option{}, option)
	})
}

func TestInMemoryOptionService_GetOptionByDisplayName(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
	}

	service := NewInMemoryOptionService(options)

	t.Run("Get existing option by display name", func(t *testing.T) {
		option, found := service.GetOptionByDisplayName("Active")
		assert.True(t, found)
		assert.Equal(t, "active", option.UniqueName)
		assert.Equal(t, "Active", option.DisplayName)
	})

	t.Run("Get non-existing option by display name", func(t *testing.T) {
		option, found := service.GetOptionByDisplayName("Non Existing")
		assert.False(t, found)
		assert.Equal(t, Option{}, option)
	})
}

func TestInMemoryOptionService_Clear(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
	}

	service := NewInMemoryOptionService(options)

	t.Run("Clear all options", func(t *testing.T) {
		assert.Equal(t, 2, service.Count())

		service.Clear()

		assert.Equal(t, 0, service.Count())
		assert.False(t, service.HasOption("active"))
		assert.False(t, service.HasOption("inactive"))
	})
}

func TestInMemoryOptionService_Count(t *testing.T) {
	service := NewInMemoryOptionService(nil)

	t.Run("Count empty service", func(t *testing.T) {
		assert.Equal(t, 0, service.Count())
	})

	t.Run("Count with options", func(t *testing.T) {
		service.AddOption(Option{UniqueName: "active", DisplayName: "Active"})
		assert.Equal(t, 1, service.Count())

		service.AddOption(Option{UniqueName: "inactive", DisplayName: "Inactive"})
		assert.Equal(t, 2, service.Count())
	})
}

func TestInMemoryOptionService_HasOption(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
	}

	service := NewInMemoryOptionService(options)

	t.Run("Has existing option", func(t *testing.T) {
		assert.True(t, service.HasOption("active"))
		assert.True(t, service.HasOption("inactive"))
	})

	t.Run("Has non-existing option", func(t *testing.T) {
		assert.False(t, service.HasOption("non-existing"))
	})
}

func TestInMemoryOptionService_HasDisplayName(t *testing.T) {
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
	}

	service := NewInMemoryOptionService(options)

	t.Run("Has existing display name", func(t *testing.T) {
		assert.True(t, service.HasDisplayName("Active"))
		assert.True(t, service.HasDisplayName("Inactive"))
	})

	t.Run("Has non-existing display name", func(t *testing.T) {
		assert.False(t, service.HasDisplayName("Non Existing"))
	})
}

func TestInMemoryOptionService_Concurrency(t *testing.T) {
	service := NewInMemoryOptionService(nil)

	// Test concurrent reads
	t.Run("Concurrent reads", func(t *testing.T) {
		// Add some initial data
		for i := 0; i < 10; i++ {
			service.AddOption(Option{
				UniqueName:  fmt.Sprintf("option%d", i),
				DisplayName: fmt.Sprintf("Option %d", i),
			})
		}

		// Start multiple goroutines reading
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					_, err := service.GetOptions(context.Background())
					assert.NoError(t, err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	// Test concurrent writes
	t.Run("Concurrent writes", func(t *testing.T) {
		service.Clear()

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				for j := 0; j < 10; j++ {
					service.AddOption(Option{
						UniqueName:  fmt.Sprintf("option_%d_%d", id, j),
						DisplayName: fmt.Sprintf("Option %d-%d", id, j),
					})
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should have 100 options (10 goroutines * 10 options each)
		assert.Equal(t, 100, service.Count())
	})
}

func TestInMemoryOptionService_IntegrationWithOptionsFieldType(t *testing.T) {
	// Test that InMemoryOptionService works correctly with the Options FieldType
	options := []Option{
		{UniqueName: "active", DisplayName: "Active"},
		{UniqueName: "inactive", DisplayName: "Inactive"},
		{UniqueName: "pending", DisplayName: "Pending"},
	}

	service := NewInMemoryOptionService(options)
	optionsField := NewOptions(service)
	ctx := context.Background()

	t.Run("Validate with in-memory service", func(t *testing.T) {
		err := optionsField.Validate("active")
		assert.NoError(t, err)

		err = optionsField.Validate("invalid")
		assert.Error(t, err)
	})

	t.Run("Get display name with in-memory service", func(t *testing.T) {
		displayName, err := optionsField.GetDisplayName(ctx, "active")
		assert.NoError(t, err)
		assert.Equal(t, "Active", displayName)
	})

	t.Run("Get unique name with in-memory service", func(t *testing.T) {
		uniqueName, err := optionsField.GetUniqueName(ctx, "Active")
		assert.NoError(t, err)
		assert.Equal(t, "active", uniqueName)
	})

	t.Run("Get all options with in-memory service", func(t *testing.T) {
		allOptions, err := optionsField.GetAllOptions(ctx)
		assert.NoError(t, err)
		assert.Len(t, allOptions, 3)
		assert.Equal(t, options, allOptions)
	})

	t.Run("Dynamic updates work with field type", func(t *testing.T) {
		// Add a new option
		service.AddOption(Option{UniqueName: "completed", DisplayName: "Completed"})

		// Should now validate successfully
		err := optionsField.Validate("completed")
		assert.NoError(t, err)

		// Should be able to get display name
		displayName, err := optionsField.GetDisplayName(ctx, "completed")
		assert.NoError(t, err)
		assert.Equal(t, "Completed", displayName)
	})
}
