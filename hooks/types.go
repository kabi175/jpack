// Package hooks provides lifecycle hooks for JPack.
// It includes before/after hooks for all CRUD operations and record lifecycle events.
//
// The hooks system allows you to intercept and modify operations at various points
// in the data lifecycle, enabling features like validation, logging, auditing, and more.
//
// Example:
//
//	// Create hook registry
//	registry := hooks.NewHookRegistry()
//
//	// Register a before save hook
//	hook := hooks.Hook{
//		ID:          "timestamp",
//		Type:        hooks.HookBeforeSave,
//		Function: func(ctx context.Context, rec schema.JRecord) error {
//			rec.Set("updated_at", time.Now())
//			return nil
//		},
//		Priority:    100,
//		Description: "Add timestamp",
//	}
//	err := registry.Register(hook)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register an after save hook
//	err = hooks.RegisterAfterSaveHook("logging", func(ctx context.Context, rec schema.JRecord) error {
//		log.Printf("Saved record: %+v", rec.ToMap())
//		return nil
//	}, 50, "Log save operation")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register enhanced hook with more context
//	enhancedHook := hooks.EnhancedHook{
//		ID:          "validation",
//		Type:        hooks.HookBeforeSave,
//		Function: func(ctx context.Context, hookCtx hooks.HookContext) error {
//			// Access schema, record, and repository
//			schema := hookCtx.Schema
//			record := hookCtx.Record
//			repository := hookCtx.Repository
//
//			// Perform validation
//			return schema.Validate(ctx, record)
//		},
//		Priority:    200,
//		Description: "Validate data",
//	}
//	enhancedRegistry := hooks.NewEnhancedHookRegistry()
//	err = enhancedRegistry.Register(enhancedHook)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get hooks for execution
//	beforeSaveHooks := registry.GetHooks(hooks.HookBeforeSave)
//	for _, hook := range beforeSaveHooks {
//		fmt.Printf("Hook: %s - %s\n", hook.ID, hook.Description)
//	}
//
//	// Execute hooks using executor
//	executor := hooks.NewHookExecutor(registry)
//	err = executor.ExecuteBeforeSave(ctx, userRecord)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use hooks with repository operations
//	userRepo, err := client.GetRepository("User")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	user := schema.NewJRecord().Set("id", "user_001").Set("name", "John")
//	savedUser, err := userRepo.Save(context.Background(), user)
//	// Hooks will be executed automatically
//
//	// Register hooks for different operations
//	err = hooks.RegisterBeforeDeleteHook("cleanup", func(ctx context.Context, rec schema.JRecord) error {
//		// Clean up related resources
//		return nil
//	}, 100, "Cleanup before delete")
//
//	err = hooks.RegisterAfterFindHook("enrich", func(ctx context.Context, rec schema.JRecord) error {
//		// Enrich found records with additional data
//		rec.Set("enriched", true)
//		return nil
//	}, 75, "Enrich found records")
//
//	err = hooks.RegisterBeforeUpdateHook("audit", func(ctx context.Context, rec schema.JRecord) error {
//		// Log update operation
//		log.Printf("Updating record: %s", rec.Get("id"))
//		return nil
//	}, 50, "Audit update operations")
//
//	err = hooks.RegisterAfterUpdateHook("notification", func(ctx context.Context, rec schema.JRecord) error {
//		// Send notification about update
//		return nil
//	}, 25, "Send update notification")
package hooks

import (
	"context"
	"fmt"
	"sync"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// HookType represents the type of hook.
// It defines when the hook should be executed in the operation lifecycle.
type HookType string

const (
	// HookBeforeSave executes before saving a record
	HookBeforeSave HookType = "before_save"
	// HookAfterSave executes after saving a record
	HookAfterSave HookType = "after_save"
	// HookBeforeDelete executes before deleting a record
	HookBeforeDelete HookType = "before_delete"
	// HookAfterDelete executes after deleting a record
	HookAfterDelete HookType = "after_delete"
	// HookBeforeFind executes before finding records
	HookBeforeFind HookType = "before_find"
	// HookAfterFind executes after finding records
	HookAfterFind HookType = "after_find"
	// HookBeforeUpdate executes before updating records
	HookBeforeUpdate HookType = "before_update"
	// HookAfterUpdate executes after updating records
	HookAfterUpdate HookType = "after_update"
)

// HookFunc represents a hook function.
// It defines the signature for hook functions that can be registered.
//
// Example:
//
//	hookFunc := func(ctx context.Context, rec schema.JRecord) error {
//		// Modify record or perform side effects
//		rec.Set("updated_at", time.Now())
//		return nil
//	}
type HookFunc func(ctx context.Context, rec schema.JRecord) error

// Hook represents a registered hook.
// It contains all the information needed to execute a hook.
//
// Example:
//
//	hook := hooks.Hook{
//		ID:          "timestamp",
//		Type:        hooks.HookBeforeSave,
//		Function:    timestampHook,
//		Priority:    100,
//		Description: "Add timestamp to record",
//	}
//
//	registry.Register(hook)
type Hook struct {
	// ID is the unique identifier for the hook
	ID string
	// Type is the hook type
	Type HookType
	// Function is the hook function to execute
	Function HookFunc
	// Priority determines execution order (higher priority first)
	Priority int
	// Description describes what the hook does
	Description string
}

// HookRegistry manages hooks.
// It provides registration, retrieval, and management of hooks.
//
// Example:
//
//	registry := hooks.NewHookRegistry()
//
//	// Register a hook
//	hook := hooks.Hook{
//		ID:          "timestamp",
//		Type:        hooks.HookBeforeSave,
//		Function:    timestampHook,
//		Priority:    100,
//		Description: "Add timestamp",
//	}
//	err := registry.Register(hook)
//
//	// Get hooks for a specific type
//	hooks := registry.GetHooks(hooks.HookBeforeSave)
//
//	// Unregister a hook
//	err = registry.Unregister("timestamp")
//
//	// Clear all hooks
//	err = registry.ClearAllHooks()
type HookRegistry interface {
	// Register registers a hook
	Register(hook Hook) error
	// Unregister removes a hook by ID
	Unregister(id string) error
	// GetHooks returns all hooks of a specific type
	GetHooks(hookType HookType) []Hook
	// ClearHooks removes all hooks of a specific type
	ClearHooks(hookType HookType) error
	// ClearAllHooks removes all registered hooks
	ClearAllHooks() error
}

// jHookRegistry implements HookRegistry
type jHookRegistry struct {
	hooks map[HookType][]Hook
	mutex sync.RWMutex
}

// NewHookRegistry creates a new hook registry.
// It initializes a registry for managing hooks.
//
// Example:
//
//	registry := hooks.NewHookRegistry()
//
//	// Register hooks
//	hook := hooks.Hook{
//		ID:          "timestamp",
//		Type:        hooks.HookBeforeSave,
//		Function:    timestampHook,
//		Priority:    100,
//		Description: "Add timestamp",
//	}
//	registry.Register(hook)
func NewHookRegistry() HookRegistry {
	return &jHookRegistry{
		hooks: make(map[HookType][]Hook),
	}
}

// Global hook registry
var globalHookRegistry HookRegistry = NewHookRegistry()

// GetGlobalHookRegistry returns the global hook registry.
// This provides access to the default hook registry.
//
// Example:
//
//	registry := hooks.GetGlobalHookRegistry()
//
//	// Register hooks
//	hook := hooks.Hook{
//		ID:          "timestamp",
//		Type:        hooks.HookBeforeSave,
//		Function:    timestampHook,
//		Priority:    100,
//		Description: "Add timestamp",
//	}
//	registry.Register(hook)
func GetGlobalHookRegistry() HookRegistry {
	return globalHookRegistry
}

func (r *jHookRegistry) Register(hook Hook) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if hook.ID == "" {
		logger.Hooks.Error().Msg("hook ID cannot be empty")
		return fmt.Errorf("hook ID cannot be empty")
	}

	if hook.Function == nil {
		logger.Hooks.Error().Msg("hook function cannot be nil")
		return fmt.Errorf("hook function cannot be nil")
	}

	// Check if hook with same ID already exists
	for _, existingHooks := range r.hooks {
		for _, existingHook := range existingHooks {
			if existingHook.ID == hook.ID {
				logger.Hooks.Error().
					Str("hook_id", hook.ID).
					Msg("hook with ID already exists")
				return fmt.Errorf("hook with ID '%s' already exists", hook.ID)
			}
		}
	}

	// Add hook to the appropriate type
	if r.hooks[hook.Type] == nil {
		r.hooks[hook.Type] = make([]Hook, 0)
	}

	r.hooks[hook.Type] = append(r.hooks[hook.Type], hook)

	// Sort hooks by priority (higher priority first)
	r.sortHooksByPriority(hook.Type)

	return nil
}

func (r *jHookRegistry) Unregister(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for hookType, hooks := range r.hooks {
		for i, hook := range hooks {
			if hook.ID == id {
				// Remove hook from slice
				r.hooks[hookType] = append(hooks[:i], hooks[i+1:]...)
				return nil
			}
		}
	}

	logger.Hooks.Error().
		Str("hook_id", id).
		Msg("hook with ID not found")
	return fmt.Errorf("hook with ID '%s' not found", id)
}

func (r *jHookRegistry) GetHooks(hookType HookType) []Hook {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	hooks := r.hooks[hookType]
	if hooks == nil {
		return make([]Hook, 0)
	}

	// Return a copy to prevent external modifications
	result := make([]Hook, len(hooks))
	copy(result, hooks)
	return result
}

func (r *jHookRegistry) ClearHooks(hookType HookType) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.hooks[hookType] = make([]Hook, 0)
	return nil
}

func (r *jHookRegistry) ClearAllHooks() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.hooks = make(map[HookType][]Hook)
	return nil
}

func (r *jHookRegistry) sortHooksByPriority(hookType HookType) {
	hooks := r.hooks[hookType]
	if len(hooks) <= 1 {
		return
	}

	// Simple bubble sort by priority (higher priority first)
	for i := 0; i < len(hooks)-1; i++ {
		for j := 0; j < len(hooks)-i-1; j++ {
			if hooks[j].Priority < hooks[j+1].Priority {
				hooks[j], hooks[j+1] = hooks[j+1], hooks[j]
			}
		}
	}
}

// HookExecutor executes hooks.
// It provides methods to execute hooks at different points in the operation lifecycle.
//
// Example:
//
//	executor := hooks.NewHookExecutor(registry)
//
//	// Execute before save hooks
//	err := executor.ExecuteBeforeSave(ctx, record)
//	if err != nil {
//		return err
//	}
//
//	// Perform save operation
//	savedRecord, err := repository.Save(ctx, record)
//	if err != nil {
//		return err
//	}
//
//	// Execute after save hooks
//	err = executor.ExecuteAfterSave(ctx, savedRecord)
//	if err != nil {
//		return err
//	}
type HookExecutor interface {
	// ExecuteBeforeSave executes all before save hooks
	ExecuteBeforeSave(ctx context.Context, rec schema.JRecord) error
	// ExecuteAfterSave executes all after save hooks
	ExecuteAfterSave(ctx context.Context, rec schema.JRecord) error
	// ExecuteBeforeDelete executes all before delete hooks
	ExecuteBeforeDelete(ctx context.Context, rec schema.JRecord) error
	// ExecuteAfterDelete executes all after delete hooks
	ExecuteAfterDelete(ctx context.Context, rec schema.JRecord) error
	// ExecuteBeforeFind executes all before find hooks
	ExecuteBeforeFind(ctx context.Context, rec schema.JRecord) error
	// ExecuteAfterFind executes all after find hooks
	ExecuteAfterFind(ctx context.Context, rec schema.JRecord) error
	// ExecuteBeforeUpdate executes all before update hooks
	ExecuteBeforeUpdate(ctx context.Context, rec schema.JRecord) error
	// ExecuteAfterUpdate executes all after update hooks
	ExecuteAfterUpdate(ctx context.Context, rec schema.JRecord) error
}

// jHookExecutor implements HookExecutor
type jHookExecutor struct {
	registry HookRegistry
}

// NewHookExecutor creates a new hook executor.
// It initializes an executor with the provided hook registry.
//
// Example:
//
//	registry := hooks.NewHookRegistry()
//	executor := hooks.NewHookExecutor(registry)
//
//	// Execute hooks
//	err := executor.ExecuteBeforeSave(ctx, record)
func NewHookExecutor(registry HookRegistry) HookExecutor {
	return &jHookExecutor{
		registry: registry,
	}
}

// GetGlobalHookExecutor returns the global hook executor.
// This provides access to the default hook executor.
//
// Example:
//
//	executor := hooks.GetGlobalHookExecutor()
//
//	// Execute hooks
//	err := executor.ExecuteBeforeSave(ctx, record)
func GetGlobalHookExecutor() HookExecutor {
	return NewHookExecutor(globalHookRegistry)
}

func (e *jHookExecutor) ExecuteBeforeSave(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookBeforeSave, rec)
}

func (e *jHookExecutor) ExecuteAfterSave(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookAfterSave, rec)
}

func (e *jHookExecutor) ExecuteBeforeDelete(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookBeforeDelete, rec)
}

func (e *jHookExecutor) ExecuteAfterDelete(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookAfterDelete, rec)
}

func (e *jHookExecutor) ExecuteBeforeFind(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookBeforeFind, rec)
}

func (e *jHookExecutor) ExecuteAfterFind(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookAfterFind, rec)
}

func (e *jHookExecutor) ExecuteBeforeUpdate(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookBeforeUpdate, rec)
}

func (e *jHookExecutor) ExecuteAfterUpdate(ctx context.Context, rec schema.JRecord) error {
	return e.executeHooks(ctx, HookAfterUpdate, rec)
}

func (e *jHookExecutor) executeHooks(ctx context.Context, hookType HookType, rec schema.JRecord) error {
	hooks := e.registry.GetHooks(hookType)

	for _, hook := range hooks {
		if err := hook.Function(ctx, rec); err != nil {
			logger.Hooks.Error().
				Str("hook_id", hook.ID).
				Err(err).
				Msg("hook execution failed")
			return fmt.Errorf("hook '%s' failed: %w", hook.ID, err)
		}
	}

	return nil
}

// Convenience functions for registering hooks

// RegisterBeforeSaveHook registers a before save hook.
// This is a convenience function for registering before save hooks.
//
// Example:
//
//	err := hooks.RegisterBeforeSaveHook("timestamp", func(ctx context.Context, rec schema.JRecord) error {
//		rec.Set("updated_at", time.Now())
//		return nil
//	}, 100, "Add timestamp")
func RegisterBeforeSaveHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeSave,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterSaveHook registers an after save hook.
// This is a convenience function for registering after save hooks.
//
// Example:
//
//	err := hooks.RegisterAfterSaveHook("logging", func(ctx context.Context, rec schema.JRecord) error {
//		log.Printf("Saved record: %+v", rec.ToMap())
//		return nil
//	}, 50, "Log save operation")
func RegisterAfterSaveHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookAfterSave,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterBeforeDeleteHook registers a before delete hook.
// This is a convenience function for registering before delete hooks.
//
// Example:
//
//	err := hooks.RegisterBeforeDeleteHook("audit", func(ctx context.Context, rec schema.JRecord) error {
//		log.Printf("Deleting record: %+v", rec.ToMap())
//		return nil
//	}, 100, "Audit delete operation")
func RegisterBeforeDeleteHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeDelete,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterDeleteHook registers an after delete hook.
// This is a convenience function for registering after delete hooks.
//
// Example:
//
//	err := hooks.RegisterAfterDeleteHook("cleanup", func(ctx context.Context, rec schema.JRecord) error {
//		// Perform cleanup operations
//		return nil
//	}, 50, "Cleanup after delete")
func RegisterAfterDeleteHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookAfterDelete,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterBeforeFindHook registers a before find hook.
// This is a convenience function for registering before find hooks.
//
// Example:
//
//	err := hooks.RegisterBeforeFindHook("filter", func(ctx context.Context, rec schema.JRecord) error {
//		// Apply filters or modify query
//		return nil
//	}, 100, "Apply query filters")
func RegisterBeforeFindHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeFind,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterFindHook registers an after find hook.
// This is a convenience function for registering after find hooks.
//
// Example:
//
//	err := hooks.RegisterAfterFindHook("transform", func(ctx context.Context, rec schema.JRecord) error {
//		// Transform or enrich found records
//		return nil
//	}, 50, "Transform found records")
func RegisterAfterFindHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookAfterFind,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterBeforeUpdateHook registers a before update hook.
// This is a convenience function for registering before update hooks.
//
// Example:
//
//	err := hooks.RegisterBeforeUpdateHook("validation", func(ctx context.Context, rec schema.JRecord) error {
//		// Validate update data
//		return nil
//	}, 100, "Validate update data")
func RegisterBeforeUpdateHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeUpdate,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterUpdateHook registers an after update hook.
// This is a convenience function for registering after update hooks.
//
// Example:
//
//	err := hooks.RegisterAfterUpdateHook("audit", func(ctx context.Context, rec schema.JRecord) error {
//		// Log update operation
//		return nil
//	}, 50, "Log update operation")
func RegisterAfterUpdateHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookAfterUpdate,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// HookContext provides additional context for hooks
type HookContext struct {
	Schema     schema.JSchema
	Repository interface{}
	Operation  string
	Metadata   map[string]any
}

// EnhancedHookFunc represents a hook function with additional context
type EnhancedHookFunc func(ctx context.Context, rec schema.JRecord, hookCtx HookContext) error

// EnhancedHook represents a hook with additional context
type EnhancedHook struct {
	ID          string
	Type        HookType
	Function    EnhancedHookFunc
	Priority    int
	Description string
}

// EnhancedHookRegistry manages enhanced hooks
type EnhancedHookRegistry interface {
	Register(hook EnhancedHook) error
	Unregister(id string) error
	GetHooks(hookType HookType) []EnhancedHook
	ClearHooks(hookType HookType) error
	ClearAllHooks() error
}

// jEnhancedHookRegistry implements EnhancedHookRegistry
type jEnhancedHookRegistry struct {
	hooks map[HookType][]EnhancedHook
	mutex sync.RWMutex
}

// NewEnhancedHookRegistry creates a new enhanced hook registry
func NewEnhancedHookRegistry() EnhancedHookRegistry {
	return &jEnhancedHookRegistry{
		hooks: make(map[HookType][]EnhancedHook),
	}
}

// Global enhanced hook registry
var globalEnhancedHookRegistry EnhancedHookRegistry = NewEnhancedHookRegistry()

// GetGlobalEnhancedHookRegistry returns the global enhanced hook registry
func GetGlobalEnhancedHookRegistry() EnhancedHookRegistry {
	return globalEnhancedHookRegistry
}

func (r *jEnhancedHookRegistry) Register(hook EnhancedHook) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if hook.ID == "" {
		logger.Hooks.Error().Msg("enhanced hook ID cannot be empty")
		return fmt.Errorf("hook ID cannot be empty")
	}

	if hook.Function == nil {
		logger.Hooks.Error().Msg("enhanced hook function cannot be nil")
		return fmt.Errorf("hook function cannot be nil")
	}

	// Check if hook with same ID already exists
	for _, existingHooks := range r.hooks {
		for _, existingHook := range existingHooks {
			if existingHook.ID == hook.ID {
				logger.Hooks.Error().
					Str("hook_id", hook.ID).
					Msg("enhanced hook with ID already exists")
				return fmt.Errorf("hook with ID '%s' already exists", hook.ID)
			}
		}
	}

	// Add hook to the appropriate type
	if r.hooks[hook.Type] == nil {
		r.hooks[hook.Type] = make([]EnhancedHook, 0)
	}

	r.hooks[hook.Type] = append(r.hooks[hook.Type], hook)

	// Sort hooks by priority (higher priority first)
	r.sortHooksByPriority(hook.Type)

	return nil
}

func (r *jEnhancedHookRegistry) Unregister(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for hookType, hooks := range r.hooks {
		for i, hook := range hooks {
			if hook.ID == id {
				// Remove hook from slice
				r.hooks[hookType] = append(hooks[:i], hooks[i+1:]...)
				return nil
			}
		}
	}

	logger.Hooks.Error().
		Str("hook_id", id).
		Msg("enhanced hook with ID not found")
	return fmt.Errorf("hook with ID '%s' not found", id)
}

func (r *jEnhancedHookRegistry) GetHooks(hookType HookType) []EnhancedHook {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	hooks := r.hooks[hookType]
	if hooks == nil {
		return make([]EnhancedHook, 0)
	}

	// Return a copy to prevent external modifications
	result := make([]EnhancedHook, len(hooks))
	copy(result, hooks)
	return result
}

func (r *jEnhancedHookRegistry) ClearHooks(hookType HookType) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.hooks[hookType] = make([]EnhancedHook, 0)
	return nil
}

func (r *jEnhancedHookRegistry) ClearAllHooks() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.hooks = make(map[HookType][]EnhancedHook)
	return nil
}

func (r *jEnhancedHookRegistry) sortHooksByPriority(hookType HookType) {
	hooks := r.hooks[hookType]
	if len(hooks) <= 1 {
		return
	}

	// Simple bubble sort by priority (higher priority first)
	for i := 0; i < len(hooks)-1; i++ {
		for j := 0; j < len(hooks)-i-1; j++ {
			if hooks[j].Priority < hooks[j+1].Priority {
				hooks[j], hooks[j+1] = hooks[j+1], hooks[j]
			}
		}
	}
}
