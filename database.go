package main

import (
	"database/sql"
	"sync"

	"github.com/ip812/go-template/database"
)

type DBWrapper interface {
	Queries() *database.Queries
	DB() *sql.DB
	IsReady() bool
}

type SwappableDB struct {
	mu    sync.RWMutex
	db    *sql.DB
	q     *database.Queries
	ready bool
}

func NewSwappableDB() *SwappableDB {
	return &SwappableDB{}
}

func (s *SwappableDB) Swap(db *sql.DB, queries *database.Queries) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db = db
	s.q = queries
	s.ready = true
}

func (s *SwappableDB) Queries() *database.Queries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.ready {
		return nil
	}
	return s.q
}

func (s *SwappableDB) DB() *sql.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.ready {
		return nil
	}
	return s.db
}

func (s *SwappableDB) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}
