package connections

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
)

type ConnectionStore struct {
	mu    sync.RWMutex
	conns map[string]map[string]*sql.DB
}

func NewConnectionStore() *ConnectionStore {
	return &ConnectionStore{
		conns: make(map[string]map[string]*sql.DB),
	}
}

func (s *ConnectionStore) Set(projectID string, shardID string, db *sql.DB) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.conns[projectID]
	if !ok {
		s.conns[projectID] = make(map[string]*sql.DB)
	}

	oldDB, ok := s.conns[projectID][shardID]
	if ok {
		_ = oldDB.Close()
	}

	s.conns[projectID][shardID] = db
}

func (s *ConnectionStore) Get(projectID, shardID string) (*sql.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projectConns, ok := s.conns[projectID]
	if !ok {
		return nil, errors.New("No shards for project found")
	}

	db, ok := projectConns[shardID]
	if !ok {
		return nil, fmt.Errorf("no connection found for shard")
	}

	return db, nil
}
