package services

import (
	"context"
	"testing"
	"time"
	"xyphos/internal/hsm"
	"xyphos/internal/models"
	"xyphos/internal/store"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 🧪 TestKeyRotationService_Start tests starting the key rotation service
func TestKeyRotationService_Start(t *testing.T) {
	// Setup
	mockStore := &mockKMSStore{}
	mockHSM, err := hsm.NewSoftwareHSM()
	require.NoError(t, err)

	config := RotationConfig{
		Period:          100 * time.Millisecond,
		RotationWindow:  24 * time.Hour,
		DefaultLifetime: 30 * 24 * time.Hour,
	}

	service, err := NewKeyRotationService(mockStore, mockHSM, config)
	require.NoError(t, err)

	// Test
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = service.Start(ctx)
	require.NoError(t, err)

	// Wait for a few rotation cycles
	time.Sleep(350 * time.Millisecond)

	// Stop the service
	service.Stop()
}

// 🧪 TestKeyRotationService_RotateKeyNow tests immediate key rotation
func TestKeyRotationService_RotateKeyNow(t *testing.T) {
	// Setup
	mockStore := &mockKMSStore{}
	mockHSM, err := hsm.NewSoftwareHSM()
	require.NoError(t, err)

	config := RotationConfig{
		Period:          time.Hour,
		RotationWindow:  24 * time.Hour,
		DefaultLifetime: 30 * 24 * time.Hour,
	}

	service, err := NewKeyRotationService(mockStore, mockHSM, config)
	require.NoError(t, err)

	// Setup test key
	testKey := &store.Key{
		ID:             "test-key",
		KeyRing:        "test-keyring",
		Tenant:         "test-tenant",
		Algorithm:      "AES-256-GCM",
		Purpose:        "ENCRYPT_DECRYPT",
		CreatedAt:      time.Now(),
		CurrentVersion: 1,
		Versions: []store.KeyVersion{
			{
				Version:      1,
				State:        "ENABLED",
				CreatedAt:    time.Now().Add(-48 * time.Hour),
				RotationTime: time.Now().Add(-24 * time.Hour),
				EncryptedKey: []byte("test-key-material"),
			},
		},
	}

	mockStore.keys = map[string]*store.Key{
		"test-key": testKey,
	}

	// Test
	ctx := context.Background()
	err = service.RotateKeyNow(ctx, "test-key")
	require.NoError(t, err)

	// Verify
	rotatedKey := mockStore.keys["test-key"]
	assert.Equal(t, 2, len(rotatedKey.Versions))
	assert.Equal(t, 2, rotatedKey.CurrentVersion)
	assert.Equal(t, "DEPRECATED", rotatedKey.Versions[0].State)
	assert.Equal(t, "ENABLED", rotatedKey.Versions[1].State)
}

// 🧪 TestKeyRotationService_CheckAndRotateKeys tests the key rotation check logic
func TestKeyRotationService_CheckAndRotateKeys(t *testing.T) {
	// Setup
	mockStore := &mockKMSStore{}
	mockHSM, err := hsm.NewSoftwareHSM()
	require.NoError(t, err)

	config := RotationConfig{
		Period:          time.Hour,
		RotationWindow:  24 * time.Hour,
		DefaultLifetime: 30 * 24 * time.Hour,
	}

	service, err := NewKeyRotationService(mockStore, mockHSM, config)
	require.NoError(t, err)

	// Create test project and keyring
	project := &models.Project{
		ID:        "test-project",
		Name:      "Test Project",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		OwnerID:   "test-tenant",
	}

	keyring := &store.KeyRing{
		ID:        "test-keyring",
		Name:      "Test KeyRing",
		Owner:     "test-tenant",
		CreatedAt: time.Now(),
	}

	// Create test keys
	needsRotation := &store.Key{
		ID:             "needs-rotation",
		KeyRing:        keyring.ID,
		Tenant:         "test-tenant",
		RotationPeriod: 30 * 24 * time.Hour,
		NextRotation:   time.Now().Add(30 * 24 * time.Hour),
		Algorithm:      "AES-256-GCM",
		Purpose:        "ENCRYPT_DECRYPT",
		CreatedAt:      time.Now(),
		CurrentVersion: 1,
		Versions: []store.KeyVersion{
			{
				Version:      1,
				State:        "ENABLED",
				CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
				RotationTime: time.Now().Add(-time.Hour),
				EncryptedKey: []byte("old-key-material"),
			},
		},
	}

	noRotationNeeded := &store.Key{
		ID:             "no-rotation",
		KeyRing:        keyring.ID,
		Tenant:         "test-tenant",
		Algorithm:      "AES-256-GCM",
		Purpose:        "ENCRYPT_DECRYPT",
		CreatedAt:      time.Now(),
		CurrentVersion: 1,
		RotationPeriod: 30 * 24 * time.Hour,
		NextRotation:   time.Now().Add(30 * 24 * time.Hour),
		Versions: []store.KeyVersion{
			{
				Version:      1,
				State:        "ENABLED",
				CreatedAt:    time.Now(),
				RotationTime: time.Now().Add(30 * 24 * time.Hour),
				EncryptedKey: []byte("new-key-material"),
			},
		},
	}

	mockStore.projects = []*models.Project{project}
	mockStore.keyrings = []*store.KeyRing{keyring}
	mockStore.keys = map[string]*store.Key{
		"needs-rotation": needsRotation,
		"no-rotation":    noRotationNeeded,
	}

	// Test
	ctx := context.Background()
	err = service.RotateKeyNow(ctx, "needs-rotation")
	require.NoError(t, err)

	// Verify
	rotatedKey := mockStore.keys["needs-rotation"]
	assert.Equal(t, 2, len(rotatedKey.Versions))
	assert.Equal(t, 2, rotatedKey.CurrentVersion)
	assert.Equal(t, "DEPRECATED", rotatedKey.Versions[0].State)
	assert.Equal(t, "ENABLED", rotatedKey.Versions[1].State)

	nonRotatedKey := mockStore.keys["no-rotation"]
	assert.Equal(t, 1, len(nonRotatedKey.Versions))
	assert.Equal(t, 1, nonRotatedKey.CurrentVersion)
	assert.Equal(t, "ENABLED", nonRotatedKey.Versions[0].State)
}

// 🔧 Mock KMS store for testing
type mockKMSStore struct {
	tenants  []*models.Tenant
	projects []*models.Project
	keyrings []*store.KeyRing
	keys     map[string]*store.Key
}

func (m *mockKMSStore) ListAllProjects(ctx context.Context) ([]*models.Project, error) {
	return m.projects, nil
}

func (m *mockKMSStore) ListProjects(ctx context.Context, ownerID string) ([]*models.Project, error) {
	var result []*models.Project
	for _, p := range m.projects {
		if p.OwnerID == ownerID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockKMSStore) CreateProject(ctx context.Context, project *models.Project) error {
	m.projects = append(m.projects, project)
	return nil
}

func (m *mockKMSStore) GetProject(ctx context.Context, id string) (*models.Project, error) {
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockKMSStore) ListLocations(ctx context.Context) ([]*models.Location, error) {
	return []*models.Location{{ID: "global", Name: "Global"}}, nil
}

func (m *mockKMSStore) GetLocation(ctx context.Context, id string) (*models.Location, error) {
	if id == "global" {
		return &models.Location{ID: "global", Name: "Global"}, nil
	}
	return nil, nil
}

func (m *mockKMSStore) CreateKeyRing(ctx context.Context, keyring *store.KeyRing) error {
	m.keyrings = append(m.keyrings, keyring)
	return nil
}

func (m *mockKMSStore) ListKeyRings(ctx context.Context, owner string) ([]*store.KeyRing, error) {
	var result []*store.KeyRing
	for _, kr := range m.keyrings {
		if kr.Owner == owner {
			result = append(result, kr)
		}
	}
	return result, nil
}

func (m *mockKMSStore) GetKeyRing(ctx context.Context, id string) (*store.KeyRing, error) {
	for _, kr := range m.keyrings {
		if kr.ID == id {
			return kr, nil
		}
	}
	return nil, nil
}

func (m *mockKMSStore) DeleteKeyRing(ctx context.Context, id string) error {
	for i, kr := range m.keyrings {
		if kr.ID == id {
			m.keyrings = append(m.keyrings[:i], m.keyrings[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockKMSStore) CreateKey(ctx context.Context, key *store.Key) error {
	m.keys[key.ID] = key
	return nil
}

func (m *mockKMSStore) ListKeys(ctx context.Context, keyringID, tenant string) ([]*store.Key, error) {
	var result []*store.Key
	for _, k := range m.keys {
		if k.KeyRing == keyringID && k.Tenant == tenant {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *mockKMSStore) GetKey(ctx context.Context, id string) (*store.Key, error) {
	return m.keys[id], nil
}

func (m *mockKMSStore) UpdateKey(ctx context.Context, key *store.Key) error {
	// Create a deep copy of the key to avoid pointer issues
	updatedKey := &store.Key{
		ID:             key.ID,
		KeyRing:        key.KeyRing,
		Tenant:         key.Tenant,
		Algorithm:      key.Algorithm,
		Purpose:        key.Purpose,
		RotationPeriod: key.RotationPeriod,
		NextRotation:   key.NextRotation,
		CreatedAt:      key.CreatedAt,
		CurrentVersion: key.CurrentVersion,
		Versions:       make([]store.KeyVersion, len(key.Versions)),
	}

	// Deep copy each version
	for i, version := range key.Versions {
		updatedKey.Versions[i] = store.KeyVersion{
			Version:      version.Version,
			State:        version.State,
			CreatedAt:    version.CreatedAt,
			RotationTime: version.RotationTime,
			EncryptedKey: make([]byte, len(version.EncryptedKey)),
		}
		copy(updatedKey.Versions[i].EncryptedKey, version.EncryptedKey)
	}

	// Update the key in the map
	m.keys[key.ID] = updatedKey
	return nil
}

func (m *mockKMSStore) DeleteKey(ctx context.Context, id string) error {
	delete(m.keys, id)
	return nil
}

func (m *mockKMSStore) Close() error {
	return nil
}

func (m *mockKMSStore) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	m.tenants = append(m.tenants, tenant)
	return nil
}

func (m *mockKMSStore) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	for _, t := range m.tenants {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockKMSStore) ListTenants(ctx context.Context, keyRingID string) ([]*models.Tenant, error) {
	return m.tenants, nil
}

func (m *mockKMSStore) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	for i, t := range m.tenants {
		if t.ID == tenant.ID {
			m.tenants[i] = tenant
			return nil
		}
	}
	return nil
}

func (m *mockKMSStore) DeleteTenant(ctx context.Context, id string) error {
	for i, t := range m.tenants {
		if t.ID == id {
			m.tenants = append(m.tenants[:i], m.tenants[i+1:]...)
			return nil
		}
	}
	return nil
}
