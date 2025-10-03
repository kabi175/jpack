package hooks

import (
	"context"
	"fmt"
	"sync"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// HookType represents the type of hook
type HookType string

const (
	HookBeforeSave   HookType = "before_save"
	HookAfterSave    HookType = "after_save"
	HookBeforeDelete HookType = "before_delete"
	HookAfterDelete  HookType = "after_delete"
	HookBeforeFind   HookType = "before_find"
	HookAfterFind    HookType = "after_find"
	HookBeforeUpdate HookType = "before_update"
	HookAfterUpdate  HookType = "after_update"
)

// HookFunc represents a hook function
type HookFunc func(ctx context.Context, rec schema.JRecord) error

// Hook represents a registered hook
type Hook struct {
	ID          string
	Type        HookType
	Function    HookFunc
	Priority    int
	Description string
}

// HookRegistry manages hooks
type HookRegistry interface {
	Register(hook Hook) error
	Unregister(id string) error
	GetHooks(hookType HookType) []Hook
	ClearHooks(hookType HookType) error
	ClearAllHooks() error
}

// jHookRegistry implements HookRegistry
type jHookRegistry struct {
	hooks map[HookType][]Hook
	mutex sync.RWMutex
}

// NewHookRegistry creates a new hook registry
func NewHookRegistry() HookRegistry {
	return &jHookRegistry{
		hooks: make(map[HookType][]Hook),
	}
}

// Global hook registry
var globalHookRegistry HookRegistry = NewHookRegistry()

// GetGlobalHookRegistry returns the global hook registry
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

// HookExecutor executes hooks
type HookExecutor interface {
	ExecuteBeforeSave(ctx context.Context, rec schema.JRecord) error
	ExecuteAfterSave(ctx context.Context, rec schema.JRecord) error
	ExecuteBeforeDelete(ctx context.Context, rec schema.JRecord) error
	ExecuteAfterDelete(ctx context.Context, rec schema.JRecord) error
	ExecuteBeforeFind(ctx context.Context, rec schema.JRecord) error
	ExecuteAfterFind(ctx context.Context, rec schema.JRecord) error
	ExecuteBeforeUpdate(ctx context.Context, rec schema.JRecord) error
	ExecuteAfterUpdate(ctx context.Context, rec schema.JRecord) error
}

// jHookExecutor implements HookExecutor
type jHookExecutor struct {
	registry HookRegistry
}

// NewHookExecutor creates a new hook executor
func NewHookExecutor(registry HookRegistry) HookExecutor {
	return &jHookExecutor{
		registry: registry,
	}
}

// GetGlobalHookExecutor returns the global hook executor
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

// RegisterBeforeSaveHook registers a before save hook
func RegisterBeforeSaveHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeSave,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterSaveHook registers an after save hook
func RegisterAfterSaveHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookAfterSave,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterBeforeDeleteHook registers a before delete hook
func RegisterBeforeDeleteHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeDelete,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterDeleteHook registers an after delete hook
func RegisterAfterDeleteHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookAfterDelete,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterBeforeFindHook registers a before find hook
func RegisterBeforeFindHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeFind,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterFindHook registers an after find hook
func RegisterAfterFindHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookAfterFind,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterBeforeUpdateHook registers a before update hook
func RegisterBeforeUpdateHook(id string, fn HookFunc, priority int, description string) error {
	return globalHookRegistry.Register(Hook{
		ID:          id,
		Type:        HookBeforeUpdate,
		Function:    fn,
		Priority:    priority,
		Description: description,
	})
}

// RegisterAfterUpdateHook registers an after update hook
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
