/*
Copyright © 2024, 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package partitionmap

import (
	"cmp"
	"fmt"
	"hash/crc32"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// The number of partitions to use for the key/value pairs.
	numberOfPartitionsInMap = 128 // well within byte range
)

type (
	// `tKeyMap` contains a partition's key/value pairs.
	tKeyMap[K cmp.Ordered, V any] map[K]V

	// `tPartition` implements a single partition in a `tPartitionList`.
	tPartition[K cmp.Ordered, V any] struct {
		sync.RWMutex               // protect the key/value store
		kv           tKeyMap[K, V] // the key/value store
	}

	// `tPartitionList` is a slice of `tPartition` instances.
	tPartitionList[K cmp.Ordered, V any] []*tPartition[K, V]

	// `TPartitionMap` is a slice of partitions holding the
	// key/value pairs.
	TPartitionMap[K cmp.Ordered, V any] struct {
		sync.RWMutex         // protect the list of partitions
		tPartitionList[K, V] // the list of partitions
	}
)

// ---------------------------------------------------------------------------
// `tPartition` constructor:

// `newPartition()` is a constructor function that creates and
// initialises a new partition.
// Each partition holds a set of key/value pairs.
//
// The function takes no parameters and returns a pointer to a new
// `tPartition` instance.
//
// The returned partition is initialised with a read-write mutex and
// an empty map.
//
// Example usage:
//
//	partition := newPartition[string, int]()
//	partition.put("key1", 10)
//	partition.put("key2", 20)
//	value, ok := partition.get("key1")
//	fmt.Println(value, ok) // Output: 10 true
//
// Returns:
//   - `*tPartition[K, V]`: A pointer to a newly created partition.
func newPartition[K cmp.Ordered, V any]() *tPartition[K, V] {
	p := &tPartition[K, V]{
		kv: make(tKeyMap[K, V]),
	}

	return p
} // newPartition()

// ---------------------------------------------------------------------------
// `tPartition` methods:

// `clear()` removes all key/value pairs from the partition.
//
// Returns:
//   - `*tPartition[K, V]`: The partition itself, allowing method chaining.
func (p *tPartition[K, V]) clear() *tPartition[K, V] {
	if nil == p {
		return nil
	}

	p.Lock()
	// For maps, `clear()` deletes all entries,
	// resulting in an empty map.
	clear(p.kv)
	p.Unlock()

	return p
} // clear()

// `clone()` creates a deep copy of the partition's key/value pairs.
//
// The method returns a copy of the key/value pairs. This is a shallow
// clone: the new keys and values are set using ordinary assignment.
//
// Returns:
//   - `tKeyMap[K, V]`: A deep copy of the partition's key/value pairs.
func (p *tPartition[K, V]) clone() tKeyMap[K, V] {
	p.RLock()
	result := maps.Clone(p.kv)
	p.RUnlock()

	return result
} // clone()

// `del()` removes a key/value pair from the partition.
//
// This method is used to delete a key/value pair from the partition.
// If the key is found in the partition, it is removed.
//
// Parameters:
//   - `aKey`: The key of the key/value pair to be deleted.
//
// Returns:
//   - `*tPartition[K, V]`: The partition itself, allowing method chaining.
func (p *tPartition[K, V]) del(aKey K) *tPartition[K, V] {
	if nil == p {
		return nil
	}

	p.Lock()
	delete(p.kv, aKey)
	p.Unlock()

	return p
} // del()

// `forEach()` executes the provided function for each key/value pair
// in the partition.
//
// Parameters:
//   - `aFunc`: The function to execute for each key/value pair.
//
// Returns:
//   - `*tPartition[K, V]`: The partition itself, allowing method chaining.
func (p *tPartition[K, V]) forEach(aFunc func(aKey K, aValue V)) *tPartition[K, V] {
	if nil == p {
		return nil
	}

	// Create a snapshot of keys/values under lock to avoid
	// holding lock during callback
	kvMap := p.clone()

	// Execute callback outside of lock
	for k, v := range kvMap {
		aFunc(k, v)
	}

	return p
} // forEach()

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
func (p *tPartition[K, V]) get(aKey K) (rVal V, rOk bool) {
	if nil == p {
		return
	}

	p.RLock()
	rVal, rOk = p.kv[aKey]
	p.RUnlock()

	return
} // get()

// `keys()` returns a slice of all keys in the partition.
//
// The partition holds a set of key/value pairs. This method retrieves
// the keys from the partition and returns them in a sorted slice.
//
// Returns:
//   - `rKeys`: A slice of the keys in the current partition.
func (p *tPartition[K, V]) keys() (rKeys []K) {
	if nil == p {
		return
	}

	p.RLock()
	for k := range p.kv {
		rKeys = append(rKeys, k)
	}
	p.RUnlock()

	slices.Sort(rKeys)

	return
} // keys()

// `len()` returns the number of key/value pairs in the partition.
//
// Returns:
//   - `rLen`: The number of key/value pairs in the partition.
func (p *tPartition[K, V]) len() (rLen int) {
	if nil != p {
		p.RLock()
		rLen = len(p.kv)
		p.RUnlock()
	}

	return
} // len()

// `put()` stores a key/value pair in the partition.
// If the key already exists, it will be updated.
//
// Parameters:
//   - `aKey`: The key to be store in the partition.
//   - `aValue`: The value associated with the key.
//
// Returns:
//   - `*tPartition[K, V]`: The partition itself, allowing method chaining.
func (p *tPartition[K, V]) put(aKey K, aVal V) *tPartition[K, V] {
	if nil == p {
		return nil
	}

	p.Lock()
	p.kv[aKey] = aVal
	p.Unlock()

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
func (p *tPartition[K, V]) String() string {
	if nil == p {
		return ""
	}

	// Create a snapshot of keys/values under lock to
	// avoid holding lock during processing
	kvMap := p.clone()
	keys := make([]K, 0, len(kvMap))
	for k := range kvMap {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(fmt.Sprintf("%v: '%v'\n", k, kvMap[k]))
	}

	return builder.String()
} // String()

// ---------------------------------------------------------------------------
// `TPartitionMap` constructor:

// `New()` creates and initialises a new partitioned map instance.
//
// This is a constructor function that returns a pointer to a new
// `TPartitionMap` instance with the specified key and value types.
//
// The returned partitioned map is initialised with the predefined number
// of partitions (128), but the actual partition instances are created
// lazily when needed.
//
// Example usage:
//
//	pm := New[string, string]()
//	pm.Put("key1", "value1")
//	value, exists := pm.Get("key1")
//
// Returns:
//   - `*TPartitionMap[K, V]`: A pointer to a newly created partitioned map.
func New[K cmp.Ordered, V any]() *TPartitionMap[K, V] {
	// Unfortunately, Go doesn't support the use of sparse arrays
	// (i.e. slices). That forces us to initialise the whole list
	// at once. With 128 possible values/indices that takes 1024 bytes.
	//
	// An empty map is allocated with enough space to hold the
	// specified number of elements.
	result := &TPartitionMap[K, V]{
		tPartitionList: make(tPartitionList[K, V], numberOfPartitionsInMap),
	}

	// Leave the partition instances to lazy/late initialisation;
	// see `TPartitionMap.partition()`.

	return result
} // New()

// ---------------------------------------------------------------------------
// `TPartitionMap` methods:

var (
	// `crc32Table`: To avoid (re-)allocation with every call
	// to `partitionIndex()` we create it here once.
	gCrc32Table = crc32.MakeTable(crc32.Castagnoli)
)

// `partitionIndex()` computes the partition index for a given key.
// It uses the CRC32 algorithm to generate a hash value for the key,
// then takes the modulus of the hash value with the number of partitions
// to obtain the partition index.
//
// Parameters:
//   - `aKey`: The key for which the partition index is to be computed.
//
// Returns:
//   - `uint8`: The partition index to use for the given key.
func partitionIndex[K cmp.Ordered](aKey K) uint8 {
	var (
		uintKey uint64
		key     []byte
	)

	switch val := any(aKey).(type) {
	case int: // negative values would turn into a modulo == 0
		uintKey = uint64(val)
	case int8:
		uintKey = uint64(val)
	case int16:
		uintKey = uint64(val)
	case int32:
		uintKey = uint64(val)
	case int64:
		uintKey = uint64(val)
	case uint:
		uintKey = uint64(val)
	case uint8:
		uintKey = uint64(val)
	case uint16:
		uintKey = uint64(val)
	case uint32:
		uintKey = uint64(val)
	case uint64:
		uintKey = val
	case uintptr:
		uintKey = uint64(val)
	case float32:
		key = []byte(strconv.FormatFloat(float64(val), 'f', -1, 32))
	case float64:
		key = []byte(strconv.FormatFloat(val, 'f', -1, 64))
	case string:
		key = []byte(val)
	default:
		key = fmt.Appendf(nil, "%v", aKey)
	} // switch

	if 0 < uintKey {
		return uint8(uintKey % numberOfPartitionsInMap) //#nosec G115
	}

	// We use CRC32 for speed and adequate distribution.
	// While it's not cryptographically secure, it's perfect for our
	// partitioning needs.
	// If two different keys hash to the same partition, they'll
	// simply share a partition.

	cs32 := crc32.Checksum(key, gCrc32Table)
	return uint8(cs32 % numberOfPartitionsInMap)
} // partitionIndex()

// `partition()` retrieves a partition from the partitioned map based
// on the provided key.
//
// If the partition already exists, it is returned along with a boolean
// value indicating its existence.
//
// If the partition doesn't exist yet and the create parameter is set to
// `true`, a new partition is created and returned.
//
// If the partition doesn't exist and the create parameter is set to
// `false`, the method returns `nil` and a boolean value of `false`.
//
// Parameters:
//   - `aKey`: The key used to identify the partition.
//   - `aCreate`: A boolean value indicating whether a new partition for the given key should be created if it doesn't exist yet.
//
// Returns:
//   - `*tPartition[K, V]`: The partition associated with the provided key, or `nil` if the partition does not exist and `aCreate` is `false`.
//   - `bool`: A boolean value indicating whether the partition was successfully retrieved.
func (pm *TPartitionMap[K, V]) partition(aKey K, aCreate bool) (*tPartition[K, V], bool) {
	if nil == pm {
		return nil, false
	}
	idx := partitionIndex(aKey)

	pm.RLock()
	p := (pm.tPartitionList)[idx]
	pm.RUnlock()

	if nil != p {
		return p, true
	}

	if !aCreate {
		return nil, false
	}

	// Here we do the lazy initialisation of the required `tPartition`:
	p = newPartition[K, V]()
	pm.Lock()
	(pm.tPartitionList)[idx] = p
	pm.Unlock()

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

// `Clear()` removes all key/value pairs from the partitioned map.
//
// Returns:
//   - `*TPartitionMap[K, V]`: The partitioned map itself, allowing method chaining.
func (pm *TPartitionMap[K, V]) Clear() *TPartitionMap[K, V] {
	if nil == pm {
		return nil
	}

	pm.Lock()
	for _, p := range pm.tPartitionList {
		p.clear()
	}
	pm.Unlock()

	return pm
} // Clear()

// `Delete()` removes a key/value pair from the partitioned map.
//
// Parameters:
//   - `aKey`: The key of the key/value pair to be deleted.
//
// Returns:
//   - `*TPartitionMap[K, V]`: The partitioned map itself, allowing method chaining.
func (pm *TPartitionMap[K, V]) Delete(aKey K) *TPartitionMap[K, V] {
	if nil == pm {
		return nil
	}

	if p, ok := pm.partition(aKey, false); ok {
		p.del(aKey)
	}

	return pm
} // Delete()

// `ForEach()` executes the provided function for each key/value pair
// in the partitioned map.
//
// Parameters:
//   - `aFunc`: The function to execute for each key/value pair.
//
// Returns:
//   - `*TPartitionMap[K, V]`: The partitioned map itself, allowing method chaining.
func (pm *TPartitionMap[K, V]) ForEach(aFunc func(aKey K, aValue V)) *TPartitionMap[K, V] {
	if nil == pm {
		return nil
	}

	pm.RLock()
	for _, p := range pm.tPartitionList {
		p.forEach(aFunc)
	}
	pm.RUnlock()

	return pm
} // ForEach()

// `Get()` retrieves a key/value pair from the partitioned map.
//
// If the partitioned map contains a key/value pair with the specified key,
// the function returns the associated value and a boolean value indicating
// whether the key was found. If the key is not found, the function returns
// the zero value of type V and a boolean value of false.
//
// Parameters:
//   - `aKey`: The key of the key/value pair to be retrieved.
//
// Returns:
//   - `V`: The value associated with the key (if found).
//   - `bool`: Indicating for whether the key was found.
func (pm *TPartitionMap[K, V]) Get(aKey K) (V, bool) {
	var zeroVal V
	if nil == pm {
		return zeroVal, false
	}

	if p, ok := pm.partition(aKey, false); ok {
		return p.get(aKey)
	}

	return zeroVal, false
} // Get()

// `GetOrDefault()` retrieves a value for the given key, or returns
// the given default value if the key doesn't exist in the partitioned map.
//
// Parameters:
//   - `aKey`: The key to look up.
//   - `aDefault`: The default value to return if the key is not found.
//
// Returns:
//   - `V`: The value associated with `aKey`, or `aDefault` if not found.
func (pm *TPartitionMap[K, V]) GetOrDefault(aKey K, aDefault V) V {
	if nil != pm {
		if result, ok := pm.Get(aKey); ok {
			return result
		}
	}

	return aDefault
} // GetOrDefault()

// `Keys()` returns a slice of all keys in the partitioned map.
//
// The partitioned map is divided into multiple partitions, each holding
// a subset of the keys. This method retrieves the keys from all
// partitions and returns them in a sorted slice.
//
// The returned slice is a copy of the keys from all partitions, sorted
// in ascending order.
//
// Returns:
//   - `[]K`: A slice of all the keys in the current partitioned map.
func (pm *TPartitionMap[K, V]) Keys() []K {
	if nil == pm {
		return nil
	}

	totalKeys := 0
	pm.RLock()
	for _, p := range pm.tPartitionList {
		totalKeys += p.len()
	}
	pm.RUnlock()

	if 0 == totalKeys {
		// No point in wasting time and resources ...
		return []K{}
	}

	result := make([]K, 0, totalKeys)

	// Collect all keys
	pm.RLock()
	for _, p := range pm.tPartitionList {
		if nil != p {
			result = append(result, p.keys()...)
		}
	}
	pm.RUnlock()

	slices.Sort(result)

	return result
} // Keys()

// `Len()` returns the total number of key/value pairs in the partitioned map.
//
// Returns:
//   - `rLen`: The number of all key/value pairs in the partitioned map.
func (pm *TPartitionMap[K, V]) Len() (rLen int) {
	if nil == pm {
		return
	}

	pm.RLock()
	for _, p := range pm.tPartitionList {
		rLen += p.len()
	}
	pm.RUnlock()

	return
} // Len()

type (
	// `TMetrics` provides statistics about the partition usage.
	//
	// `Parts` is the number of partitions that are actually in use.
	// `Keys` is the total number of keys across all partitions.
	// `Avg` is the average number of keys per partition.
	// `PartKeys` is a map where the key is the partition index and the
	// value is the number of keys in that partition.
	TMetrics struct {
		Parts    int
		Keys     int
		Avg      int
		PartKeys map[int]int
	}
)

// `PartitionStats()` returns statistics about the partition usage.
//
// This method returns information about how many partitions are actually
// in use, how many keys are distributed across the partitions, and how
// many key/value pairs are stored in each partition.
// This can be useful for monitoring and optimising the distribution
// of keys across partitions.
//
// Returns:
//   - `*TMetrics`: A pointer to a `TMetrics` instance containing the statistics.
func (pm *TPartitionMap[K, V]) PartitionStats() *TMetrics {
	if nil == pm {
		return nil
	}

	pLen := 0
	result := &TMetrics{
		PartKeys: make(map[int]int),
	}

	pm.RLock()
	for idx, p := range pm.tPartitionList {
		if nil != p {
			result.Parts++
			pLen = p.len()
			result.Keys += pLen
			result.PartKeys[idx] = pLen
		}
	}
	pm.RUnlock()

	if (0 == result.Parts) || (0 == result.Keys) {
		return result
	}
	result.Avg = result.Keys / result.Parts

	return result
} // PartitionStats()

// `Put()` stores a key/value pair into the partitioned map.
// If the key already exists, it will be updated.
//
// Parameters:
//   - `aKey`: The key to be put into the partitioned map.
//   - `aValue`: The value associated with the key.
//
// Returns:
//   - `*TPartitionMap[K, V]`: The partitioned map itself, allowing method chaining.
func (pm *TPartitionMap[K, V]) Put(aKey K, aValue V) *TPartitionMap[K, V] {
	if nil == pm {
		return nil
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
// Returns:
//   - `string`: A string representation of the partitioned map.
func (pm *TPartitionMap[K, V]) String() string {
	if nil == pm {
		return ""
	}

	var builder strings.Builder
	pm.RLock()
	for _, p := range pm.tPartitionList {
		builder.WriteString(p.String())
	}
	pm.RUnlock()

	return builder.String()
} // String()

// `Values()` returns a slice of all values in the partitioned map.
//
// The partitioned map is divided into multiple partitions, each holding
// a subset of the values. This method retrieves the values from all
// partitions and returns them in a slice.
//
// The order of values in the returned slice corresponds to the order
// of keys returned by the `Keys()` method.
//
// Returns:
//   - `[]V`: A slice of all the values in the current partitioned map.
func (pm *TPartitionMap[K, V]) Values() []V {
	if nil == pm {
		return nil
	}
	var (
		ok  bool
		p   *tPartition[K, V]
		val V
	)

	// No locking, as `Keys()` and `partition()` are already thread-safe.

	keys := pm.Keys()
	result := make([]V, 0, len(keys))

	for _, key := range keys {
		if p, ok = pm.partition(key, false); ok {
			if val, ok = p.get(key); ok {
				result = append(result, val)
			}
		}
	} // for key

	return result
} // Values()

/* _EoF_ */
