package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/johnayoung/finlib/pkg/storage"
)

// TestEntity is a simple entity for testing
type TestEntity struct {
	id      string
	version int64
	data    string
}

func (e *TestEntity) GetID() string      { return e.id }
func (e *TestEntity) GetVersion() int64  { return e.version }
func (e *TestEntity) SetVersion(v int64) { e.version = v }
func (e *TestEntity) CopyFrom(src interface{}) error {
	if s, ok := src.(*TestEntity); ok {
		*e = *s
	}
	return nil
}

// SimpleEntity is a non-versioned entity for testing
type SimpleEntity struct {
	id   string
	data string
}

func (e *SimpleEntity) GetID() string { return e.id }
func (e *SimpleEntity) CopyFrom(src interface{}) error {
	if s, ok := src.(*SimpleEntity); ok {
		*e = *s
	}
	return nil
}

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	t.Run("Create", func(t *testing.T) {
		entity := &TestEntity{id: "test1", data: "test data"}
		err := store.Create(ctx, entity)
		assert.NoError(t, err)

		// Test duplicate creation
		err = store.Create(ctx, entity)
		assert.Error(t, err)
	})

	t.Run("Read", func(t *testing.T) {
		entity := &TestEntity{}
		err := store.Read(ctx, "test1", entity)
		assert.NoError(t, err)
		assert.Equal(t, "test1", entity.id)
		assert.Equal(t, "test data", entity.data)

		// Test reading non-existent entity
		err = store.Read(ctx, "nonexistent", entity)
		assert.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		entity := &TestEntity{id: "test1", version: 1, data: "updated data"}
		err := store.Update(ctx, entity)
		assert.NoError(t, err)

		// Verify update
		updated := &TestEntity{}
		err = store.Read(ctx, "test1", updated)
		assert.NoError(t, err)
		assert.Equal(t, "updated data", updated.data)

		// Test optimistic locking
		conflictEntity := &TestEntity{id: "test1", version: 1, data: "conflict data"}
		err = store.Update(ctx, conflictEntity)
		assert.Error(t, err)
		_, ok := err.(*storage.OptimisticLockError)
		assert.True(t, ok)
	})

	t.Run("Update Non-Versioned Entity", func(t *testing.T) {
		// Create and update a simple entity without version
		simple := &SimpleEntity{id: "simple1", data: "initial"}
		err := store.Create(ctx, simple)
		assert.NoError(t, err)

		simple.data = "updated"
		err = store.Update(ctx, simple)
		assert.NoError(t, err)

		// Verify update
		updated := &SimpleEntity{}
		err = store.Read(ctx, "simple1", updated)
		assert.NoError(t, err)
		assert.Equal(t, "updated", updated.data)
	})

	t.Run("Delete", func(t *testing.T) {
		err := store.Delete(ctx, "test1")
		assert.NoError(t, err)

		// Verify deletion
		entity := &TestEntity{}
		err = store.Read(ctx, "test1", entity)
		assert.Error(t, err)

		// Test deleting non-existent entity
		err = store.Delete(ctx, "nonexistent")
		assert.Error(t, err)
	})

	t.Run("Audit Trail", func(t *testing.T) {
		// Create a new entity
		entity := &SimpleEntity{id: "test2", data: "audit test"}
		err := store.Create(ctx, entity)
		assert.NoError(t, err)

		// Update the entity
		entity.data = "updated audit test"
		err = store.Update(ctx, entity)
		assert.NoError(t, err)

		// Delete the entity
		err = store.Delete(ctx, entity.id)
		assert.NoError(t, err)

		// Check audit trail
		trail, err := store.GetAuditTrail(ctx, entity.id)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(trail))
		assert.Equal(t, "CREATE", trail[0].Operation)
		assert.Equal(t, "UPDATE", trail[1].Operation)
		assert.Equal(t, "DELETE", trail[2].Operation)
	})
}

func TestConcurrency(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	t.Run("Concurrent Reads", func(t *testing.T) {
		entity := &TestEntity{id: "concurrent", data: "initial"}
		err := store.Create(ctx, entity)
		assert.NoError(t, err)

		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				e := &TestEntity{}
				err := store.Read(ctx, "concurrent", e)
				assert.NoError(t, err)
				assert.Equal(t, "initial", e.data)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("Concurrent Updates", func(t *testing.T) {
		done := make(chan bool)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func(i int) {
				e := &TestEntity{id: "concurrent", version: int64(i), data: "updated"}
				err := store.Update(ctx, e)
				if err != nil {
					errors <- err
				}
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		close(errors)
		errorCount := 0
		for err := range errors {
			_, ok := err.(*storage.OptimisticLockError)
			assert.True(t, ok)
			errorCount++
		}

		// We expect some updates to fail due to optimistic locking
		assert.True(t, errorCount > 0)
	})
}
