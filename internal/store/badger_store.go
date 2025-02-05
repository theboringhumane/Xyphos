package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
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

		// Store in owner's keyring list
		return txn.Set([]byte(fmt.Sprintf("owner:%s:keyrings:%s", keyring.Owner, keyring.ID)), []byte{1})
	})
}

func (s *BadgerStore) ListKeyRings(ctx context.Context, owner string) ([]*KeyRing, error) {
	var keyrings []*KeyRing

	err := s.db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("owner:%s:keyrings:", owner))
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

		// Delete from owner's keyring list
		if err := txn.Delete([]byte(fmt.Sprintf("owner:%s:keyrings:%s", keyring.Owner, id))); err != nil {
			return fmt.Errorf("failed to delete keyring from owner list: %w", err)
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
		return txn.Set([]byte(fmt.Sprintf("keyring:%s:keys:%s", key.ID, key.ID)), []byte{1})
	})
}

func (s *BadgerStore) ListKeys(ctx context.Context, keyringName, tenant string) ([]*Key, error) {
	var keys []*Key

	err := s.db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("key:%s", keyringName))
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var key Key
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &key)
			})
			if err != nil {
				continue // Skip invalid keys
			}

			// Only return keys for the specified keyring and tenant
			if key.KeyRing == keyringName && key.Tenant == tenant {
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

func (s *BadgerStore) GetKeyByName(ctx context.Context, name string) (*Key, error) {
	var key *Key

	err := s.db.View(func(txn *badger.Txn) error {
		prefix := []byte("key:")
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var k Key
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &k)
			})
			if err != nil {
				continue
			}

			if k.Name == name {
				key = &k
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get key by name: %w", err)
	}

	if key == nil {
		return nil, fmt.Errorf("key not found with name: %s", name)
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
		if err := txn.Delete([]byte(fmt.Sprintf("keyring:%s:keys:%s", key.ID, id))); err != nil {
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

// 🏢 Tenant operations
func (s *BadgerStore) CreateTenant(ctx context.Context, tenant *Tenant) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(tenant)
		if err != nil {
			return fmt.Errorf("failed to marshal tenant: %w", err)
		}

		// Store by ID
		if err := txn.Set([]byte(fmt.Sprintf("tenant:%s", tenant.ID)), data); err != nil {
			return fmt.Errorf("failed to store tenant: %w", err)
		}

		// Store in keyring's tenant list
		return txn.Set([]byte(fmt.Sprintf("keyring:%s:tenants:%s", tenant.KeyRing, tenant.ID)), []byte{1})
	})
}

func (s *BadgerStore) GetTenant(ctx context.Context, name string) (*Tenant, error) {
	var tenant *Tenant

	err := s.db.View(func(txn *badger.Txn) error {
		// Iterate through all tenants to find by name
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("tenant:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var t Tenant
			err := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &t)
			})
			if err != nil {
				continue // Skip invalid entries
			}

			if t.Name == name {
				tenant = &t
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by name: %w", err)
	}

	if tenant == nil {
		return nil, fmt.Errorf("tenant with name %s not found", name)
	}

	return tenant, nil
}

func (s *BadgerStore) ListTenants(ctx context.Context, keyRingID string) ([]*Tenant, error) {
	var tenants []*Tenant

	err := s.db.View(func(txn *badger.Txn) error {
		prefix := []byte(fmt.Sprintf("keyring:%s:tenants:", keyRingID))
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			tenantID := string(it.Item().Key())[len(prefix):]

			// Get the tenant data
			item, err := txn.Get([]byte(fmt.Sprintf("tenant:%s", tenantID)))
			if err != nil {
				continue // Skip invalid tenants
			}

			var tenant Tenant
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &tenant)
			})
			if err != nil {
				continue // Skip invalid tenants
			}

			tenants = append(tenants, &tenant)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

func (s *BadgerStore) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(tenant)
		if err != nil {
			return fmt.Errorf("failed to marshal tenant: %w", err)
		}

		tenant.UpdatedAt = time.Now()
		return txn.Set([]byte(fmt.Sprintf("tenant:%s", tenant.ID)), data)
	})
}

func (s *BadgerStore) DeleteTenant(ctx context.Context, id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		// Get tenant first to get keyring ID
		item, err := txn.Get([]byte(fmt.Sprintf("tenant:%s", id)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("failed to get tenant: %w", err)
		}

		var tenant Tenant
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &tenant)
		})
		if err != nil {
			return fmt.Errorf("failed to unmarshal tenant: %w", err)
		}

		// Delete from keyring's tenant list
		if err := txn.Delete([]byte(fmt.Sprintf("keyring:%s:tenants:%s", tenant.KeyRing, id))); err != nil {
			return fmt.Errorf("failed to delete tenant from keyring list: %w", err)
		}

		// Delete tenant
		return txn.Delete([]byte(fmt.Sprintf("tenant:%s", id)))
	})
}
