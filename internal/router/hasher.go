package router

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
)

// computes hashes for shard-key values.
type Hasher struct{}

// creates a new hasher instance.
func NewHasher() *Hasher {
	return &Hasher{}
}

// computes a deterministic hash for a shard-key value.
func (h *Hasher) Hash(value any) HashValue {
	hasher := fnv.New64a()

	switch v := value.(type) {

	case string:
		_, _ = hasher.Write([]byte(v))

	case int:
		binary.Write(hasher, binary.LittleEndian, int64(v))

	case int32:
		binary.Write(hasher, binary.LittleEndian, int64(v))

	case int64:
		binary.Write(hasher, binary.LittleEndian, v)

	case uint:
		binary.Write(hasher, binary.LittleEndian, uint64(v))

	case uint32:
		binary.Write(hasher, binary.LittleEndian, uint64(v))

	case uint64:
		binary.Write(hasher, binary.LittleEndian, v)

	default:
		_, _ = hasher.Write([]byte(fmt.Sprintf("%v", v)))
	}

	return HashValue(hasher.Sum64())
}
