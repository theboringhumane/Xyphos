package store

import (
	"context"
	"encoding/json"
	"fmt"
	"xyphos/internal/models"

	"github.com/dgraph-io/badger/v4"
)

// 🔐 KMSStore defines operations for managing KMS resources
type KMSStore interface {
	// Project operations
	ListAllProjects(ctx context.Context) ([]*models.Project, error)
	ListProjects(ctx context.Context, ownerID string) ([]*models.Project, error)
	CreateProject(ctx context.Context, project *models.Project) error
	GetProject(ctx context.Context, id string) (*models.Project, error)

	// Location operations
	ListLocations(ctx context.Context) ([]*models.Location, error)
	GetLocation(ctx context.Context, id string) (*models.Location, error)

	// Keyring operations
	CreateKeyRing(ctx context.Context, keyring *KeyRing) error
	ListKeyRings(ctx context.Context, owner string) ([]*KeyRing, error)
	GetKeyRing(ctx context.Context, id string) (*KeyRing, error)
	DeleteKeyRing(ctx context.Context, id string) error

	// Tenant operations
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenant(ctx context.Context, name string) (*Tenant, error)
	ListTenants(ctx context.Context, keyRingID string) ([]*Tenant, error)
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, id string) error

	// Key operations
	CreateKey(ctx context.Context, key *Key) error
	ListKeys(ctx context.Context, keyringName, tenant string) ([]*Key, error)
	GetKey(ctx context.Context, id string) (*Key, error)
	GetKeyByName(ctx context.Context, name string) (*Key, error)
	UpdateKey(ctx context.Context, key *Key) error
	DeleteKey(ctx context.Context, id string) error

	// Close closes the store
	Close() error
}

// BadgerKMSStore implements KMSStore using BadgerDB
type BadgerKMSStore struct {
	db *badger.DB
}

// NewBadgerKMSStore creates a new BadgerKMSStore
func NewBadgerKMSStore(db *badger.DB) *BadgerKMSStore {
	return &BadgerKMSStore{db: db}
}

// Helper functions for key generation
func projectKey(id string) []byte {
	return []byte(fmt.Sprintf("project:%s", id))
}

func clientConfigKey(id string) []byte {
	return []byte(fmt.Sprintf("client-config:%s", id))
}

// Project operations implementation
func (s *BadgerKMSStore) CreateProject(ctx context.Context, project *models.Project) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(project)
		if err != nil {
			return err
		}
		return txn.Set(projectKey(project.ID), data)
	})
}

func (s *BadgerKMSStore) GetProject(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(projectKey(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &project)
		})
	})
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *BadgerKMSStore) ListProjects(ctx context.Context, ownerID string) ([]*models.Project, error) {
	var projects []*models.Project
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("project:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var project models.Project
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &project)
			})
			if err != nil {
				return err
			}
			if project.OwnerID == ownerID {
				projects = append(projects, &project)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// Location operations implementation
func (s *BadgerKMSStore) GetLocation(ctx context.Context, id string) (*models.Location, error) {
	for _, loc := range models.PredefinedLocations {
		if loc.ID == id {
			return &loc, nil
		}
	}
	return nil, fmt.Errorf("location not found")
}

func (s *BadgerKMSStore) ListLocations(ctx context.Context) ([]*models.Location, error) {
	locations := make([]*models.Location, len(models.PredefinedLocations))
	for i := range models.PredefinedLocations {
		locations[i] = &models.PredefinedLocations[i]
	}
	return locations, nil
}

// Client configuration operations implementation
func (s *BadgerKMSStore) CreateClientConfig(ctx context.Context, config *models.ClientConfig) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return txn.Set(clientConfigKey(config.ID), data)
	})
}

func (s *BadgerKMSStore) GetClientConfig(ctx context.Context, id string) (*models.ClientConfig, error) {
	var config models.ClientConfig
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(clientConfigKey(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (s *BadgerKMSStore) ListClientConfigs(ctx context.Context, userID string) ([]*models.ClientConfig, error) {
	var configs []*models.ClientConfig
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("client-config:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var config models.ClientConfig
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &config)
			})
			if err != nil {
				return err
			}
			if config.UserID == userID {
				configs = append(configs, &config)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (s *BadgerKMSStore) UpdateClientConfig(ctx context.Context, config *models.ClientConfig) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return txn.Set(clientConfigKey(config.ID), data)
	})
}

// ListAllProjects lists all projects in the system
func (s *BadgerKMSStore) ListAllProjects(ctx context.Context) ([]*models.Project, error) {
	var projects []*models.Project
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("project:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var project models.Project
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &project)
			})
			if err != nil {
				return err
			}
			projects = append(projects, &project)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}
