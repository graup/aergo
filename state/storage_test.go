package state

import (
	"testing"

	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/assert"
)

// func TestCacheBasic(t *testing.T) {
// 	t.Log("StorageCache")
// }

func TestStorageBasic(t *testing.T) {
	storage := newBufferedStorage(nil, nil)
	v1 := types.GetHashID([]byte("v1"))
	v2 := types.GetHashID([]byte("v2"))

	storage.checkpoint(0)
	storage.put(newValueEntry(v1, []byte{1})) // rev 1

	storage.checkpoint(1)
	storage.put(newValueEntry(v2, []byte{2})) // rev 3

	storage.checkpoint(2)
	storage.put(newValueEntry(v1, []byte{3})) // rev 5
	storage.put(newValueEntry(v2, []byte{4})) // rev 6
	storage.put(newValueEntry(v2, []byte{5})) // rev 7

	storage.checkpoint(3)
	storage.put(newValueEntry(v1, []byte{6})) // rev 9
	storage.put(newValueEntry(v2, []byte{7})) // rev 10

	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2).Value())
	assert.Equal(t, []byte{6}, storage.get(v1).Value())
	assert.Equal(t, []byte{7}, storage.get(v2).Value())

	storage.rollback(3)
	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2).Value())
	assert.Equal(t, []byte{3}, storage.get(v1).Value())
	assert.Equal(t, []byte{5}, storage.get(v2).Value())

	storage.rollback(1)
	t.Log("v1", storage.get(v1).Value(), "v2", storage.get(v2))
	assert.Equal(t, []byte{1}, storage.get(v1).Value())
	assert.Nil(t, storage.get(v2))

	storage.rollback(0)
	t.Log("v1", storage.get(v1), "v2", storage.get(v2))
	assert.Nil(t, storage.get(v1))
	assert.Nil(t, storage.get(v2))
}