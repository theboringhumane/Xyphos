package store

import (
	"context"
	"encoding/json"
	"fmt"
	"xyphos/internal/models"

	"github.com/dgraph-io/badger/v4"
)

// 🔑 Keyring operations
func (s *BadgerStore) CreateKeyRing(ctx context.Context, keyring *KeyRing) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(keyring)
		if err != nil {
			return fmt.Errorf("failed to marshal keyring: %w", err)
		}

		// Store by ID
		if err := txn.Set([]byte(fmt.Sprintf("keyring:%s", keyring.ID)), data); err != nil {
			return fmt.Errorf("failed to store keyring: %w", err)
		}

		// Store in tenant's keyring list
		return txn.Set([]byte(fmt.Sprintf("tenant:%s:keyrings:%s", keyring.Tenant, keyring.ID)), []byte{1})
	})
}

func (s *BadgerStore) ListKeyRings(ctx context.Context, tenant string) ([]*KeyRing, error) {
	var keyrings []*KeyRing

	err := s.db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("tenant:%s:keyrings:", tenant))
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			keyringID := string(it.Item().Key())[len(prefix):]

			// Get the keyring data
			item, err := txn.Get([]byte(fmt.Sprintf("keyring:%s", keyringID)))
			if err != nil {
				continue // Skip invalid keyrings
			}

			var keyring KeyRing
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &keyring)
			})
			if err != nil {
				continue // Skip invalid keyrings
			}

			keyrings = append(keyrings, &keyring)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list keyrings: %w", err)
	}
	return keyrings, nil
}

func (s *BadgerStore) GetKeyRing(ctx context.Context, id string) (*KeyRing, error) {
	var keyring *KeyRing

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("keyring:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get keyring: %w", err)
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &keyring)
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get keyring by ID: %w", err)
	}
	return keyring, nil
}

func (s *BadgerStore) DeleteKeyRing(ctx context.Context, id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		// Get keyring first to get tenant
		item, err := txn.Get([]byte(fmt.Sprintf("keyring:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get keyring: %w", err)
		}

		var keyring KeyRing
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &keyring)
		})
		if err != nil {
			return fmt.Errorf("failed to unmarshal keyring: %w", err)
		}

		// Delete from tenant's keyring list
		if err := txn.Delete([]byte(fmt.Sprintf("tenant:%s:keyrings:%s", keyring.Tenant, id))); err != nil {
			return fmt.Errorf("failed to delete keyring from tenant list: %w", err)
		}

		// Delete keyring
		return txn.Delete([]byte(fmt.Sprintf("keyring:%s", id)))
	})
}

// 🔐 Key operations
func (s *BadgerStore) CreateKey(ctx context.Context, key *Key) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(key)
		if err != nil {
			return fmt.Errorf("failed to marshal key: %w", err)
		}

		// Store by ID
		if err := txn.Set([]byte(fmt.Sprintf("key:%s", key.ID)), data); err != nil {
			return fmt.Errorf("failed to store key: %w", err)
		}

		// Store in keyring's key list
		return txn.Set([]byte(fmt.Sprintf("keyring:%s:keys:%s", key.KeyRing, key.ID)), []byte{1})
	})
}

func (s *BadgerStore) ListKeys(ctx context.Context, keyringID, tenant string) ([]*Key, error) {
	var keys []*Key

	err := s.db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("keyring:%s:keys:", keyringID))
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			keyID := string(it.Item().Key())[len(prefix):]

			// Get the key data
			item, err := txn.Get([]byte(fmt.Sprintf("key:%s", keyID)))
			if err != nil {
				continue // Skip invalid keys
			}

			var key Key
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &key)
			})
			if err != nil {
				continue // Skip invalid keys
			}

			// Only return keys for the specified tenant
			if key.Tenant == tenant {
				keys = append(keys, &key)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	return keys, nil
}

func (s *BadgerStore) GetKey(ctx context.Context, id string) (*Key, error) {
	var key *Key

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("key:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get key: %w", err)
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &key)
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get key by ID: %w", err)
	}
	return key, nil
}

func (s *BadgerStore) UpdateKey(ctx context.Context, key *Key) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(key)
		if err != nil {
			return fmt.Errorf("failed to marshal key: %w", err)
		}

		return txn.Set([]byte(fmt.Sprintf("key:%s", key.ID)), data)
	})
}

func (s *BadgerStore) DeleteKey(ctx context.Context, id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		// Get key first to get keyring ID
		item, err := txn.Get([]byte(fmt.Sprintf("key:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get key: %w", err)
		}

		var key Key
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &key)
		})
		if err != nil {
			return fmt.Errorf("failed to unmarshal key: %w", err)
		}

		// Delete from keyring's key list
		if err := txn.Delete([]byte(fmt.Sprintf("keyring:%s:keys:%s", key.KeyRing, id))); err != nil {
			return fmt.Errorf("failed to delete key from keyring list: %w", err)
		}

		// Delete key
		return txn.Delete([]byte(fmt.Sprintf("key:%s", id)))
	})
}

// 👤 User operations
func (s *BadgerStore) CreateUser(ctx context.Context, user *models.User) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %w", err)
		}

		// Store by ID
		if err := txn.Set([]byte(fmt.Sprintf("user:%s", user.ID)), data); err != nil {
			return fmt.Errorf("failed to store user: %w", err)
		}

		// Store by GitHub ID for lookup
		return txn.Set([]byte(fmt.Sprintf("user:github:%d", user.GithubID)), []byte(user.ID))
	})
}

func (s *BadgerStore) GetUserByGithubID(ctx context.Context, githubID int) (*models.User, error) {
	var user *models.User
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("user:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var u models.User
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &u)
			})
			if err != nil {
				continue // Skip invalid users
			}
			if u.GithubID == githubID {
				user = &u
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user by GitHub ID: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *BadgerStore) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user *models.User
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("user:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get user: %w", err)
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

// 👤 UpdateUser updates an existing user
func (s *BadgerStore) UpdateUser(ctx context.Context, user *models.User) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %w", err)
		}

		// Update user data
		if err := txn.Set([]byte(fmt.Sprintf("user:%s", user.ID)), data); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

// 🔑 Client configuration operations
func (s *BadgerStore) CreateClientConfig(ctx context.Context, config *models.ClientConfig) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal client config: %w", err)
		}

		// Store by ID
		if err := txn.Set([]byte(fmt.Sprintf("client-config:%s", config.ID)), data); err != nil {
			return fmt.Errorf("failed to store client config: %w", err)
		}

		// Store by client ID for lookup
		if err := txn.Set([]byte(fmt.Sprintf("client-config-by-client-id:%s", config.ClientID)), []byte(config.ID)); err != nil {
			return fmt.Errorf("failed to store client config by client ID: %w", err)
		}

		return nil
	})
}

func (s *BadgerStore) GetClientConfig(ctx context.Context, id string) (*models.ClientConfig, error) {
	var config *models.ClientConfig
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("client-config:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get client config: %w", err)
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &config)
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get client config by ID: %w", err)
	}
	return config, nil
}

func (s *BadgerStore) GetClientConfigByClientID(ctx context.Context, clientID string) (*models.ClientConfig, error) {
	var configID string
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("client-config-by-client-id:%s", clientID)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get client config ID: %w", err)
		}
		return item.Value(func(val []byte) error {
			configID = string(val)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get client config by client ID: %w", err)
	}
	if configID == "" {
		return nil, nil
	}
	return s.GetClientConfig(ctx, configID)
}

func (s *BadgerStore) ListClientConfigs(ctx context.Context, userID string) ([]*models.ClientConfig, error) {
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
				continue // Skip invalid configs
			}
			if config.UserID == userID {
				configs = append(configs, &config)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list client configs: %w", err)
	}
	return configs, nil
}

func (s *BadgerStore) UpdateClientConfig(ctx context.Context, config *models.ClientConfig) error {
	return s.db.Update(func(txn *badger.Txn) error {
		// Get the old config to check if client ID has changed
		oldItem, err := txn.Get([]byte(fmt.Sprintf("client-config:%s", config.ID)))
		if err != nil {
			return fmt.Errorf("failed to get old client config: %w", err)
		}

		var oldConfig models.ClientConfig
		err = oldItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &oldConfig)
		})
		if err != nil {
			return fmt.Errorf("failed to unmarshal old client config: %w", err)
		}

		// If client ID has changed, update the lookup index
		if oldConfig.ClientID != config.ClientID {
			// Delete old lookup
			if err := txn.Delete([]byte(fmt.Sprintf("client-config-by-client-id:%s", oldConfig.ClientID))); err != nil {
				return fmt.Errorf("failed to delete old client config lookup: %w", err)
			}
			// Add new lookup
			if err := txn.Set([]byte(fmt.Sprintf("client-config-by-client-id:%s", config.ClientID)), []byte(config.ID)); err != nil {
				return fmt.Errorf("failed to store new client config lookup: %w", err)
			}
		}

		// Update the main config
		data, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal client config: %w", err)
		}
		return txn.Set([]byte(fmt.Sprintf("client-config:%s", config.ID)), data)
	})
}

// 📋 ListAllProjects lists all projects in the system
func (s *BadgerStore) ListAllProjects(ctx context.Context) ([]*models.Project, error) {
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
				continue // Skip invalid projects
			}
			projects = append(projects, &project)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all projects: %w", err)
	}
	return projects, nil
}

// 📦 Project operations
func (s *BadgerStore) CreateProject(ctx context.Context, project *models.Project) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(project)
		if err != nil {
			return fmt.Errorf("failed to marshal project: %w", err)
		}
		return txn.Set([]byte(fmt.Sprintf("project:%s", project.ID)), data)
	})
}

func (s *BadgerStore) ListProjects(ctx context.Context, ownerID string) ([]*models.Project, error) {
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
				continue // Skip invalid projects
			}
			if project.OwnerID == ownerID {
				projects = append(projects, &project)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

func (s *BadgerStore) GetProject(ctx context.Context, id string) (*models.Project, error) {
	var project *models.Project
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("project:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get project: %w", err)
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &project)
		})
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project by ID: %w", err)
	}
	return project, nil
}

// 📍 Location operations
func (s *BadgerStore) ListLocations(ctx context.Context) ([]*models.Location, error) {
	locations := make([]*models.Location, len(models.PredefinedLocations))
	for i := range models.PredefinedLocations {
		locations[i] = &models.PredefinedLocations[i]
	}
	return locations, nil
}

func (s *BadgerStore) GetLocation(ctx context.Context, id string) (*models.Location, error) {
	for _, loc := range models.PredefinedLocations {
		if loc.ID == id {
			return &loc, nil
		}
	}
	return nil, fmt.Errorf("location not found")
}
