package hooks

import (
	"context"
	"strconv"
	"time"

	"github.com/kabi175/jpack/logger"
	"github.com/kabi175/jpack/schema"
)

// Built-in hook implementations

// TimestampHook adds timestamps to records
type TimestampHook struct {
	CreatedAtField string
	UpdatedAtField string
}

// NewTimestampHook creates a new timestamp hook
func NewTimestampHook(createdAtField, updatedAtField string) *TimestampHook {
	return &TimestampHook{
		CreatedAtField: createdAtField,
		UpdatedAtField: updatedAtField,
	}
}

// BeforeSave adds timestamps before saving
func (h *TimestampHook) BeforeSave(ctx context.Context, rec schema.JRecord) error {
	now := time.Now()

	// Set updated_at
	if h.UpdatedAtField != "" {
		rec.Set(h.UpdatedAtField, now)
	}

	// Set created_at only if it doesn't exist
	if h.CreatedAtField != "" && !rec.Has(h.CreatedAtField) {
		rec.Set(h.CreatedAtField, now)
	}

	return nil
}

// AfterSave is called after saving (no-op for timestamps)
func (h *TimestampHook) AfterSave(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// BeforeUpdate adds updated timestamp before updating
func (h *TimestampHook) BeforeUpdate(ctx context.Context, rec schema.JRecord) error {
	if h.UpdatedAtField != "" {
		rec.Set(h.UpdatedAtField, time.Now())
	}
	return nil
}

// AfterUpdate is called after updating (no-op for timestamps)
func (h *TimestampHook) AfterUpdate(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// ValidationHook validates records before operations
type ValidationHook struct {
	Schema schema.JSchema
}

// NewValidationHook creates a new validation hook
func NewValidationHook(schema schema.JSchema) *ValidationHook {
	return &ValidationHook{
		Schema: schema,
	}
}

// BeforeSave validates record before saving
func (h *ValidationHook) BeforeSave(ctx context.Context, rec schema.JRecord) error {
	return h.Schema.Validate(ctx, rec)
}

// BeforeUpdate validates record before updating
func (h *ValidationHook) BeforeUpdate(ctx context.Context, rec schema.JRecord) error {
	return h.Schema.Validate(ctx, rec)
}

// AfterSave is called after saving (no-op for validation)
func (h *ValidationHook) AfterSave(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// AfterUpdate is called after updating (no-op for validation)
func (h *ValidationHook) AfterUpdate(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// LoggingHook logs operations
type LoggingHook struct {
	Logger interface{} // Would be a proper logger interface in real implementation
}

// NewLoggingHook creates a new logging hook
func NewLoggingHook(logger interface{}) *LoggingHook {
	return &LoggingHook{
		Logger: logger,
	}
}

// BeforeSave logs before saving
func (h *LoggingHook) BeforeSave(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("BeforeSave hook executed")
	return nil
}

// AfterSave logs after saving
func (h *LoggingHook) AfterSave(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("AfterSave hook executed")
	return nil
}

// BeforeDelete logs before deleting
func (h *LoggingHook) BeforeDelete(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("BeforeDelete hook executed")
	return nil
}

// AfterDelete logs after deleting
func (h *LoggingHook) AfterDelete(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("AfterDelete hook executed")
	return nil
}

// BeforeFind logs before finding
func (h *LoggingHook) BeforeFind(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("BeforeFind hook executed")
	return nil
}

// AfterFind logs after finding
func (h *LoggingHook) AfterFind(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("AfterFind hook executed")
	return nil
}

// BeforeUpdate logs before updating
func (h *LoggingHook) BeforeUpdate(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("BeforeUpdate hook executed")
	return nil
}

// AfterUpdate logs after updating
func (h *LoggingHook) AfterUpdate(ctx context.Context, rec schema.JRecord) error {
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("AfterUpdate hook executed")
	return nil
}

// AuditHook tracks changes for audit purposes
type AuditHook struct {
	AuditField string
}

// NewAuditHook creates a new audit hook
func NewAuditHook(auditField string) *AuditHook {
	return &AuditHook{
		AuditField: auditField,
	}
}

// BeforeSave adds audit information before saving
func (h *AuditHook) BeforeSave(ctx context.Context, rec schema.JRecord) error {
	if h.AuditField != "" {
		auditInfo := map[string]any{
			"created_at": time.Now(),
			"created_by": "system", // In real implementation, get from context
		}
		rec.Set(h.AuditField, auditInfo)
	}
	return nil
}

// BeforeUpdate adds audit information before updating
func (h *AuditHook) BeforeUpdate(ctx context.Context, rec schema.JRecord) error {
	if h.AuditField != "" {
		auditInfo := map[string]any{
			"updated_at": time.Now(),
			"updated_by": "system", // In real implementation, get from context
		}
		rec.Set(h.AuditField, auditInfo)
	}
	return nil
}

// AfterSave is called after saving (no-op for audit)
func (h *AuditHook) AfterSave(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// AfterUpdate is called after updating (no-op for audit)
func (h *AuditHook) AfterUpdate(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// SoftDeleteHook implements soft delete functionality
type SoftDeleteHook struct {
	DeletedAtField string
	DeletedByField string
}

// NewSoftDeleteHook creates a new soft delete hook
func NewSoftDeleteHook(deletedAtField, deletedByField string) *SoftDeleteHook {
	return &SoftDeleteHook{
		DeletedAtField: deletedAtField,
		DeletedByField: deletedByField,
	}
}

// BeforeDelete implements soft delete before actual deletion
func (h *SoftDeleteHook) BeforeDelete(ctx context.Context, rec schema.JRecord) error {
	if h.DeletedAtField != "" {
		rec.Set(h.DeletedAtField, time.Now())
	}
	if h.DeletedByField != "" {
		rec.Set(h.DeletedByField, "system") // In real implementation, get from context
	}
	return nil
}

// AfterDelete is called after soft delete (no-op)
func (h *SoftDeleteHook) AfterDelete(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// CacheHook manages cache operations
type CacheHook struct {
	Cache interface{} // Would be a proper cache interface in real implementation
}

// NewCacheHook creates a new cache hook
func NewCacheHook(cache interface{}) *CacheHook {
	return &CacheHook{
		Cache: cache,
	}
}

// AfterSave invalidates cache after saving
func (h *CacheHook) AfterSave(ctx context.Context, rec schema.JRecord) error {
	// In a real implementation, you would invalidate relevant cache entries
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("Cache invalidated after save")
	return nil
}

// AfterUpdate invalidates cache after updating
func (h *CacheHook) AfterUpdate(ctx context.Context, rec schema.JRecord) error {
	// In a real implementation, you would invalidate relevant cache entries
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("Cache invalidated after update")
	return nil
}

// AfterDelete invalidates cache after deleting
func (h *CacheHook) AfterDelete(ctx context.Context, rec schema.JRecord) error {
	// In a real implementation, you would invalidate relevant cache entries
	logger.Hooks.Info().
		Interface("record", rec.ToMap()).
		Msg("Cache invalidated after delete")
	return nil
}

// BeforeSave is called before saving (no-op for cache)
func (h *CacheHook) BeforeSave(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// BeforeUpdate is called before updating (no-op for cache)
func (h *CacheHook) BeforeUpdate(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// BeforeDelete is called before deleting (no-op for cache)
func (h *CacheHook) BeforeDelete(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// BeforeFind is called before finding (no-op for cache)
func (h *CacheHook) BeforeFind(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// AfterFind is called after finding (no-op for cache)
func (h *CacheHook) AfterFind(ctx context.Context, rec schema.JRecord) error {
	return nil
}

// HookManager manages multiple hooks for a schema
type HookManager struct {
	schema schema.JSchema
	hooks  []interface{}
}

// NewHookManager creates a new hook manager
func NewHookManager(schema schema.JSchema) *HookManager {
	return &HookManager{
		schema: schema,
		hooks:  make([]interface{}, 0),
	}
}

// AddTimestampHook adds timestamp hook
func (m *HookManager) AddTimestampHook(createdAtField, updatedAtField string) *HookManager {
	hook := NewTimestampHook(createdAtField, updatedAtField)
	m.hooks = append(m.hooks, hook)
	return m
}

// AddValidationHook adds validation hook
func (m *HookManager) AddValidationHook() *HookManager {
	hook := NewValidationHook(m.schema)
	m.hooks = append(m.hooks, hook)
	return m
}

// AddLoggingHook adds logging hook
func (m *HookManager) AddLoggingHook(logger interface{}) *HookManager {
	hook := NewLoggingHook(logger)
	m.hooks = append(m.hooks, hook)
	return m
}

// AddAuditHook adds audit hook
func (m *HookManager) AddAuditHook(auditField string) *HookManager {
	hook := NewAuditHook(auditField)
	m.hooks = append(m.hooks, hook)
	return m
}

// AddSoftDeleteHook adds soft delete hook
func (m *HookManager) AddSoftDeleteHook(deletedAtField, deletedByField string) *HookManager {
	hook := NewSoftDeleteHook(deletedAtField, deletedByField)
	m.hooks = append(m.hooks, hook)
	return m
}

// AddCacheHook adds cache hook
func (m *HookManager) AddCacheHook(cache interface{}) *HookManager {
	hook := NewCacheHook(cache)
	m.hooks = append(m.hooks, hook)
	return m
}

// RegisterAll registers all hooks in the manager
func (m *HookManager) RegisterAll() error {
	for i, hook := range m.hooks {
		hookID := m.schema.Name() + "_hook_" + strconv.Itoa(i)

		switch h := hook.(type) {
		case *TimestampHook:
			if err := RegisterBeforeSaveHook(hookID+"_before_save", h.BeforeSave, 100, "Timestamp before save"); err != nil {
				return err
			}
			if err := RegisterAfterSaveHook(hookID+"_after_save", h.AfterSave, 100, "Timestamp after save"); err != nil {
				return err
			}
			if err := RegisterBeforeUpdateHook(hookID+"_before_update", h.BeforeUpdate, 100, "Timestamp before update"); err != nil {
				return err
			}
			if err := RegisterAfterUpdateHook(hookID+"_after_update", h.AfterUpdate, 100, "Timestamp after update"); err != nil {
				return err
			}
		case *ValidationHook:
			if err := RegisterBeforeSaveHook(hookID+"_before_save", h.BeforeSave, 200, "Validation before save"); err != nil {
				return err
			}
			if err := RegisterBeforeUpdateHook(hookID+"_before_update", h.BeforeUpdate, 200, "Validation before update"); err != nil {
				return err
			}
		case *LoggingHook:
			if err := RegisterBeforeSaveHook(hookID+"_before_save", h.BeforeSave, 50, "Logging before save"); err != nil {
				return err
			}
			if err := RegisterAfterSaveHook(hookID+"_after_save", h.AfterSave, 50, "Logging after save"); err != nil {
				return err
			}
			if err := RegisterBeforeDeleteHook(hookID+"_before_delete", h.BeforeDelete, 50, "Logging before delete"); err != nil {
				return err
			}
			if err := RegisterAfterDeleteHook(hookID+"_after_delete", h.AfterDelete, 50, "Logging after delete"); err != nil {
				return err
			}
			if err := RegisterBeforeFindHook(hookID+"_before_find", h.BeforeFind, 50, "Logging before find"); err != nil {
				return err
			}
			if err := RegisterAfterFindHook(hookID+"_after_find", h.AfterFind, 50, "Logging after find"); err != nil {
				return err
			}
			if err := RegisterBeforeUpdateHook(hookID+"_before_update", h.BeforeUpdate, 50, "Logging before update"); err != nil {
				return err
			}
			if err := RegisterAfterUpdateHook(hookID+"_after_update", h.AfterUpdate, 50, "Logging after update"); err != nil {
				return err
			}
		case *AuditHook:
			if err := RegisterBeforeSaveHook(hookID+"_before_save", h.BeforeSave, 150, "Audit before save"); err != nil {
				return err
			}
			if err := RegisterBeforeUpdateHook(hookID+"_before_update", h.BeforeUpdate, 150, "Audit before update"); err != nil {
				return err
			}
		case *SoftDeleteHook:
			if err := RegisterBeforeDeleteHook(hookID+"_before_delete", h.BeforeDelete, 300, "Soft delete before delete"); err != nil {
				return err
			}
		case *CacheHook:
			if err := RegisterAfterSaveHook(hookID+"_after_save", h.AfterSave, 10, "Cache after save"); err != nil {
				return err
			}
			if err := RegisterAfterUpdateHook(hookID+"_after_update", h.AfterUpdate, 10, "Cache after update"); err != nil {
				return err
			}
			if err := RegisterAfterDeleteHook(hookID+"_after_delete", h.AfterDelete, 10, "Cache after delete"); err != nil {
				return err
			}
		}
	}

	return nil
}
