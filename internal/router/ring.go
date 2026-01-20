package router

// ring maps hash values to shard IDs.
type Ring struct {
	shards []ShardID
}

// constructs a routing ring from active shard IDs.
func NewRing(shards []ShardID) *Ring {
	return &Ring{
		shards: shards,
	}
}

// maps a hash value to a single shard.
func (r *Ring) LocateShard(hash HashValue) ShardID {
	if len(r.shards) == 0 {
		panic("ring has no shards")
	}

	pos := int(hash % HashValue(len(r.shards)))
	return r.shards[pos]
}

// maps multiple hash values to shard IDs.
// ensures uniqueness of returned shards.
func (r *Ring) LocateShards(hashes []HashValue) []ShardID {
	if len(hashes) == 0 {
		return nil
	}

	seen := make(map[ShardID]struct{})
	result := make([]ShardID, 0, len(hashes))

	for _, h := range hashes {
		shard := r.LocateShard(h)
		if _, ok := seen[shard]; ok {
			continue
		}
		seen[shard] = struct{}{}
		result = append(result, shard)
	}

	return result
}

// Size returns number of active shards.
func (r *Ring) Size() int {
	return len(r.shards)
}
