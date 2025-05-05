/*
Copyright © 2024, 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package partitionmap

import (
	"fmt"
	"hash/crc32"
	"slices"
	"strings"
	"sync"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// The number of partitions to use for the key/value pairs.
	numberOfPartitionsInPartitionMap = 64 // within byte range
)

type (
	// `tKeyMap` contains a partition's key/value pairs
	// Type definition provided for better readability and clarity.
	tKeyMap[V any] map[string]V

	// `tPartition` implements a single partition in a `tPartitionList`
	tPartition[V any] struct {
		mtx *sync.RWMutex
		kv  tKeyMap[V] // the final key/value store
	}

	// `tPartitionList` is a slice of `tPartition` instances.
	// Type definition provided for better readability and clarity.
	tPartitionList[V any] []*tPartition[V]

	// `TPartitionMap` is a slice of partitions holding the
	// key/value pairs.
	TPartitionMap[V any] tPartitionList[V]
)

// ---------------------------------------------------------------------------
// `tPartition` constructor:

// `newPartition()` is a constructor function that creates and
// initialises a new partition.
// Each partition holds a set of key-value pairs.
//
// The function takes no parameters and returns a pointer to a new
// `tPartition` instance.
//
// The returned partition is initialised with a read-write mutex and
// an empty map (tKeyMap[V]).
//
// Example usage:
//
//	partition := newPartition[int]()
//	partition.put("key1", 10)
//	partition.put("key2", 20)
//	value, ok := partition.get("key1")
//	fmt.Println(value, ok) // Output: 10 true
func newPartition[V any]() *tPartition[V] {
	p := &tPartition[V]{
		mtx: new(sync.RWMutex),
		kv:  make(tKeyMap[V]),
	}

	return p
} // newPartition()

// ---------------------------------------------------------------------------
// `tPartition` methods:

// `del()` removes a key/value pair from the partition.
//
// This method is used to delete a key/value pair from the partition.
// If the key is found in the partition, it is removed.
//
// Parameters:
//   - `aKey`: The key of the key-value pair to be deleted.
//
// Returns:
//   - `*tPartition[V]`: The partition itself, allowing method chaining.
func (p *tPartition[V]) del(aKey string) *tPartition[V] {
	if nil == p {
		return p
	}
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if _, ok := p.kv[aKey]; ok {
		delete((*p).kv, aKey)
	}

	return p
} // del()

// `get()` retrieves a key/value pair from the partition.
//
// This method is used to fetch the value associated with a given
// key from the partition.
// If the key is found in the partition, the function returns the
// associated value and a boolean value indicating whether the key
// was found.
// If the key is not found, the method returns the zero value of
// type `V` and a boolean value of `false`.
//
// Parameters:
//   - `aKey`: The key of the key/value pair to be retrieved.
//
// Returns:
//   - `V`: The value associated with the key (if found).
//   - `bool`: Indicating whether the key was found.
func (p tPartition[V]) get(aKey string) (rVal V, rOk bool) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	rVal, rOk = p.kv[aKey]

	return
} // get()

// `keys()` returns a slice of all keys in the partition.
//
// The partition holds a set of key/value pairs. This method retrieves
// the keys from the partition and returns them in a sorted slice.
//
// Returns:
//   - `[]string`: A slice of the keys in the current partition.
func (p tPartition[V]) keys() (rKeys []string) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	for k := range p.kv {
		rKeys = append(rKeys, k)
	}
	slices.Sort(rKeys)

	return
} // keys()

// `put()` stores a key-value pair in the partition.
// If the key already exists, it will be updated.
//
// Parameters:
//   - `aKey`: The key to be store in the partition.
//   - `aValue`: The value associated with the key.
//
// Returns:
//   - `*tPartition[V]`: The partition itself, allowing method chaining.
func (p *tPartition[V]) put(aKey string, aVal V) *tPartition[V] {
	if nil == p {
		return p
	}
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.kv[aKey] = aVal

	return p
} // put()

// `String()` returns a string representation of the partition.
//
// The method iterates over all key/value pairs in the partition
// and concatenates their string representations.
// The keys in the returned string are sorted in ascending order.
//
// Returns:
//   - `string`: A string representation of the partition.
func (p tPartition[V]) String() string {
	p.mtx.RLock()
	// Get keys while holding the lock
	keys := make([]string, 0, len(p.kv))
	for k := range p.kv {
		keys = append(keys, k)
	}

	// Get values while still holding the lock
	values := make(map[string]V, len(keys))
	for _, k := range keys {
		values[k] = p.kv[k]
	}
	p.mtx.RUnlock()

	// Process outside the lock
	slices.Sort(keys)

	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(fmt.Sprintf("%q: '%v'\n", k, values[k]))
	}

	return builder.String()
} // String()

// ---------------------------------------------------------------------------
// `TPartitionMap` constructor:

// `NewPartitionMap()` creates and initialises a new partition map.
//
// This function is a constructor that returns a pointer to a new
// `TPartitionMap` instance with the specified value type.
//
// The returned partition map is initialised with the predefined number
// of partitions (64), but the actual partition instances are created
// lazily when needed.
//
// Example usage:
//
//	pm := NewPartitionMap[string]()
//	pm.Put("key1", "value1")
//	value, exists := pm.Get("key1")
//
// Returns:
//   - `*TPartitionMap[V]`: A pointer to the newly created partition map.
func NewPartitionMap[V any]() *TPartitionMap[V] {
	// Unfortunately, Go doesn't support the use of sparse arrays
	// (i.e. slices). That forces us to initialise the whole list
	// at once. With 64 possible values/indices that takes 512 bytes.
	result := make(TPartitionMap[V], numberOfPartitionsInPartitionMap)

	// Leave the partitions to lazy/late initialisation;
	// see `partition()`.

	return &result
} // NewPartitionMap()

var (
	// `crc32Table`: To avoid (re-)allocation with every call
	// to `partitionIndex()` we create it here once.
	gCrc32Table = crc32.MakeTable(crc32.Castagnoli)
)

// ---------------------------------------------------------------------------
// `TPartitionMap` methods:

// `partitionIndex()` computes the partition index for a given key.
// It uses the CRC32 algorithm to generate a hash value for the key,
// then takes the modulus of the hash value with the number of partitions
// to obtain the partition index.
//
// Parameters:
//   - `aKey`: The key for which the partition index is to be computed.
//
// Returns:
//   - `uint32`: The partition index to use for the given key.
func (pm TPartitionMap[V]) partitionIndex(aKey string) (rIdx uint32) {
	// SHA-512 has a very low chance of collisions. For small data sizes
	// up to 255 bytes, SHA-256 is typically faster than SHA-512 on 32-bit
	// systems. However, on 64-bit systems, SHA-512 can be faster than
	// SHA-256, even for small data sizes.
	// Here, we don't care for possible collisions but only for speed.
	// So we choose CRC32 instead of the cryptographically more secure
	// SHA-256 and SHA-512 algorithms.
	// If there would, in fact, happen to be a collision of two different
	// keys, the only consequence would be that both keys end up in the
	// same partition: So what?

	cs32 := crc32.Checksum([]byte(aKey), gCrc32Table)
	rIdx = cs32 % numberOfPartitionsInPartitionMap

	return
} // partitionIndex()

// `partition()` retrieves a partition from the partition map based
// on the provided key.
//
// If the partition already exists, it is returned along with a boolean
// value indicating its existence.
//
// If the partition does not exist and the create parameter is set to
// `true`, a new partition is created and returned.
//
// If the partition does not exist and the create parameter is set to
// `false`, the method returns `nil` and a boolean value of `false`.
//
// Parameters:
//   - `aKey`: The key used to identify the partition.
//   - `aCreate`: A boolean value indicating whether a new partition for the given key should be created if it does not exist.
//
// Returns:
//   - `*tPartition[V]`: The partition associated with the provided key, or `nil` if the partition does not exist and `aCreate` is `false`.
//   - `bool`: A boolean value indicating whether the partition was successfully retrieved.
func (pm *TPartitionMap[V]) partition(aKey string, aCreate bool) (*tPartition[V], bool) {
	if nil == pm {
		return nil, false
	}
	idx := pm.partitionIndex(aKey)

	p := (*pm)[idx]
	if nil != p {
		return p, true
	}

	if !aCreate {
		return nil, false
	}

	// Here we do the lazy initialisation of the required `tPartition`:
	p = newPartition[V]()
	(*pm)[idx] = p

	return p, true
} // partition()

//
// CRUD interface
//
// `C`: create == Put()
// `R`: read == Get()
// `U`: update == Put()
// `D`: delete == Delete()
//

// `Delete()` removes a key/value pair from the partition map.
//
// Parameters:
//   - `aKey`: The key of the key-value pair to be deleted.
//
// Returns:
//   - `*TPartitionMap[V]`: The partition map itself, allowing method chaining.
func (pm *TPartitionMap[V]) Delete(aKey string) *TPartitionMap[V] {
	if nil == pm {
		return pm
	}
	if p, ok := pm.partition(aKey, false); ok {
		p.del(aKey)
	}

	return pm
} // Delete()

// `Get()` retrieves a key-value/pair from the partition map.
//
// If the partition map contains a key-value pair with the specified key,
// the function returns the associated value and a boolean value indicating
// whether the key was found. If the key is not found, the function returns
// the zero value of type V and a boolean value of false.
//
// Parameters:
//   - `aKey`: The key of the key/value pair to be retrieved.
//
// Returns:
//   - `V`: The value associated with the key ()if found).
//   - `bool`: Indicating for whether the key was found.
func (pm TPartitionMap[V]) Get(aKey string) (V, bool) {
	var zeroVal V
	if nil == pm {
		return zeroVal, false
	}

	if p, ok := pm.partition(aKey, false); ok {
		return p.get(aKey)
	}

	return zeroVal, false
} // Get()

// `Keys()` returns a slice of all keys in the partition map.
//
// The partition map is divided into multiple partitions, each holding
// a subset of the keys. This method retrieves the keys from all
// partitions and returns them in a sorted slice.
//
// The returned slice is a copy of the keys from all partitions, sorted
// in ascending order.
//
// Return:
//   - `[]string`: A slice of all the keys in the current partition map.
func (pm TPartitionMap[V]) Keys() []string {
	// Pre-allocate to avoid multiple reallocations
	totalKeys := 0
	for _, p := range pm {
		if nil != p {
			p.mtx.RLock()
			totalKeys += len(p.kv)
			p.mtx.RUnlock()
		}
	}
	result := make([]string, 0, totalKeys)

	// Collect all keys
	for _, p := range pm {
		if nil != p {
			result = append(result, p.keys()...)
		}
	}

	slices.Sort(result)

	return result
} // Keys()

// `Len()` returns the total number of key/value pairs in the partition map.
//
// Returns:
//   - `rLen int`: The number of all key/value pairs in the partition map.
func (pm TPartitionMap[V]) Len() (rLen int) {
	for _, p := range pm {
		if nil != p {
			p.mtx.RLock()
			rLen += len(p.kv)
			p.mtx.RUnlock()
		}
	}

	return
} // Len()

// `Put()` stores a key-value pair into the partition map.
// If the key already exists, it will be updated.
//
// Parameters:
//   - `aKey`: The key to be put into the partition map.
//   - `aValue`: The value associated with the key.
//
// Returns:
//   - `*TPartitionMap[V]`: The partition map itself, allowing method chaining.
func (pm *TPartitionMap[V]) Put(aKey string, aValue V) *TPartitionMap[V] {
	if nil == pm {
		return pm
	}
	if p, ok := pm.partition(aKey, true); ok {
		// Store the key/value pair in the partition
		p.put(aKey, aValue)
	}

	return pm
} // Put()

// `String()` returns a string representation of the `TPartitionMap`.
// It iterates over all existing partitions and concatenates their
// string representations.
//
// The keys in returned string are sorted in ascending order.
//
// Return:
//   - `string`: A string representation of the partition map.
func (pm TPartitionMap[V]) String() string {
	if nil == pm {
		return ""
	}

	var builder strings.Builder
	for _, p := range pm {
		if nil != p {
			builder.WriteString(p.String())
		}
	}

	return builder.String()
} // String()

/* _EoF_ */
