package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/johnayoung/finlib/pkg/storage"
)

// MemoryStore provides an in-memory implementation of the storage interfaces
type MemoryStore struct {
	sync.RWMutex
	data    map[string]map[string]interface{}
	audit   map[string][]storage.AuditEntry
	version map[string]int64
}

// NewMemoryStore creates a new memory store instance
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data:    make(map[string]map[string]interface{}),
		audit:   make(map[string][]storage.AuditEntry),
		version: make(map[string]int64),
	}
}

// Create implements Repository.Create
func (s *MemoryStore) Create(ctx context.Context, entity interface{}) error {
	s.Lock()
	defer s.Unlock()

	entityType := getEntityType(entity)
	id := getEntityID(entity)
	
	if id == "" {
		return fmt.Errorf("entity ID cannot be empty")
	}

	if s.data[entityType] == nil {
		s.data[entityType] = make(map[string]interface{})
	}

	if _, exists := s.data[entityType][id]; exists {
		return fmt.Errorf("entity already exists: %s", id)
	}

	s.data[entityType][id] = entity
	s.version[id] = 1
	s.recordAudit(entityType, id, "CREATE", nil, entity)

	return nil
}

// Read implements Repository.Read
func (s *MemoryStore) Read(ctx context.Context, id string, entity interface{}) error {
	s.RLock()
	defer s.RUnlock()

	entityType := getEntityType(entity)
	
	if s.data[entityType] == nil {
		return fmt.Errorf("entity type not found: %s", entityType)
	}

	stored, exists := s.data[entityType][id]
	if !exists {
		return fmt.Errorf("entity not found: %s", id)
	}

	copyEntity(stored, entity)
	return nil
}

// Update implements Repository.Update
func (s *MemoryStore) Update(ctx context.Context, entity interface{}) error {
	s.Lock()
	defer s.Unlock()

	entityType := getEntityType(entity)
	id := getEntityID(entity)

	if s.data[entityType] == nil {
		return fmt.Errorf("entity type not found: %s", entityType)
	}

	old, exists := s.data[entityType][id]
	if !exists {
		return fmt.Errorf("entity not found: %s", id)
	}

	// Handle optimistic locking
	if versioned, ok := entity.(interface{ GetVersion() int64 }); ok {
		currentVersion := s.version[id]
		if versioned.GetVersion() != currentVersion {
			return &storage.OptimisticLockError{
				EntityType:       entityType,
				EntityID:        id,
				CurrentVersion:  currentVersion,
				ExpectedVersion: versioned.GetVersion(),
			}
		}
	}

	// Update version after successful validation
	s.version[id]++
	s.data[entityType][id] = entity
	s.recordAudit(entityType, id, "UPDATE", old, entity)

	return nil
}

// Delete implements Repository.Delete
func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.Lock()
	defer s.Unlock()

	for entityType, entities := range s.data {
		if stored, exists := entities[id]; exists {
			delete(entities, id)
			s.recordAudit(entityType, id, "DELETE", stored, nil)
			return nil
		}
	}

	return fmt.Errorf("entity not found: %s", id)
}

// Query implements Repository.Query
func (s *MemoryStore) Query(ctx context.Context, query storage.Query, results interface{}) error {
	s.RLock()
	defer s.RUnlock()

	// Implementation would filter and sort based on query parameters
	// For simplicity, this is a basic implementation
	return fmt.Errorf("not implemented")
}

// Count implements Repository.Count
func (s *MemoryStore) Count(ctx context.Context, query storage.Query) (int64, error) {
	s.RLock()
	defer s.RUnlock()

	// Implementation would count based on query parameters
	// For simplicity, this is a basic implementation
	return 0, fmt.Errorf("not implemented")
}

// GetAuditTrail implements AuditableRepository.GetAuditTrail
func (s *MemoryStore) GetAuditTrail(ctx context.Context, entityID string) ([]storage.AuditEntry, error) {
	s.RLock()
	defer s.RUnlock()

	if trail, exists := s.audit[entityID]; exists {
		return trail, nil
	}

	return nil, nil
}

func (s *MemoryStore) recordAudit(entityType, entityID, operation string, oldState, newState interface{}) {
	entry := storage.AuditEntry{
		ID:            fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		EntityType:    entityType,
		EntityID:      entityID,
		Operation:     operation,
		Timestamp:     time.Now(),
		PreviousState: oldState,
		NewState:      newState,
	}

	if s.audit[entityID] == nil {
		s.audit[entityID] = make([]storage.AuditEntry, 0)
	}
	s.audit[entityID] = append(s.audit[entityID], entry)
}

// Helper functions

func getEntityType(entity interface{}) string {
	// In a real implementation, this would use reflection or type assertions
	// to get the entity type name
	return fmt.Sprintf("%T", entity)
}

func getEntityID(entity interface{}) string {
	// In a real implementation, this would use reflection or type assertions
	// to get the entity ID
	if e, ok := entity.(interface{ GetID() string }); ok {
		return e.GetID()
	}
	return ""
}

func copyEntity(src, dst interface{}) {
	// In a real implementation, this would use reflection or type assertions
	// to copy the entity data
	if copier, ok := dst.(interface{ CopyFrom(interface{}) error }); ok {
		_ = copier.CopyFrom(src)
	}
}
