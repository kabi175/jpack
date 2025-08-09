package jpack

import (
	"context"
	"sync"
)

// InMemoryOptionService provides an in-memory implementation of OptionService
type InMemoryOptionService struct {
	options []Option
	mu      sync.RWMutex
}

// NewInMemoryOptionService creates a new in-memory option service with the given options
func NewInMemoryOptionService(options []Option) *InMemoryOptionService {
	return &InMemoryOptionService{
		options: options,
	}
}

// GetOptions implements OptionService interface
func (i *InMemoryOptionService) GetOptions(ctx context.Context) ([]Option, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Return a copy of the options to prevent external modification
	result := make([]Option, len(i.options))
	copy(result, i.options)
	return result, nil
}

// AddOption adds a new option to the service
func (i *InMemoryOptionService) AddOption(option Option) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Check if option with same uniqueName already exists
	for _, existing := range i.options {
		if existing.UniqueName == option.UniqueName {
			return // Option already exists, don't add duplicate
		}
	}

	i.options = append(i.options, option)
}

// RemoveOption removes an option by uniqueName
func (i *InMemoryOptionService) RemoveOption(uniqueName string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	for j, option := range i.options {
		if option.UniqueName == uniqueName {
			// Remove the option by slicing
			i.options = append(i.options[:j], i.options[j+1:]...)
			return true
		}
	}
	return false
}

// UpdateOption updates an existing option
func (i *InMemoryOptionService) UpdateOption(option Option) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	for j, existing := range i.options {
		if existing.UniqueName == option.UniqueName {
			i.options[j] = option
			return true
		}
	}
	return false
}

// GetOptionByUniqueName returns an option by its uniqueName
func (i *InMemoryOptionService) GetOptionByUniqueName(uniqueName string) (Option, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, option := range i.options {
		if option.UniqueName == uniqueName {
			return option, true
		}
	}
	return Option{}, false
}

// GetOptionByDisplayName returns an option by its displayName
func (i *InMemoryOptionService) GetOptionByDisplayName(displayName string) (Option, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, option := range i.options {
		if option.DisplayName == displayName {
			return option, true
		}
	}
	return Option{}, false
}

// Clear removes all options
func (i *InMemoryOptionService) Clear() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.options = nil
}

// Count returns the number of options
func (i *InMemoryOptionService) Count() int {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return len(i.options)
}

// HasOption checks if an option with the given uniqueName exists
func (i *InMemoryOptionService) HasOption(uniqueName string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, option := range i.options {
		if option.UniqueName == uniqueName {
			return true
		}
	}
	return false
}

// HasDisplayName checks if an option with the given displayName exists
func (i *InMemoryOptionService) HasDisplayName(displayName string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, option := range i.options {
		if option.DisplayName == displayName {
			return true
		}
	}
	return false
}
