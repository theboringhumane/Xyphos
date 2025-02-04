// Package store provides data storage implementations
package store

import (
	"github.com/dgraph-io/badger/v4"
)

// 🔒 DB returns the underlying BadgerDB instance
func (s *BadgerStore) DB() *badger.DB {
	return s.db
}

// 🔐 Close closes the BadgerDB instance
func (s *BadgerStore) Close() error {
	return s.db.Close()
}
