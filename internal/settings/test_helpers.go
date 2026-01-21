package settings

import (
	"context"
	"errors"
)

// mockRepository implements RepositoryInterface for testing
type mockRepository struct {
	apiKeys map[string]*APIKeyModel
	err     error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		apiKeys: make(map[string]*APIKeyModel),
	}
}

func (m *mockRepository) GetAPIKey(ctx context.Context, serviceName string) (*APIKeyModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	key, ok := m.apiKeys[serviceName]
	if !ok {
		return nil, errors.New("not found")
	}
	return key, nil
}

func (m *mockRepository) GetAllAPIKeys(ctx context.Context) ([]APIKeyModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	var keys []APIKeyModel
	for _, key := range m.apiKeys {
		keys = append(keys, *key)
	}
	return keys, nil
}

func (m *mockRepository) UpsertAPIKey(ctx context.Context, apiKey *APIKeyModel) error {
	if m.err != nil {
		return m.err
	}
	m.apiKeys[apiKey.ServiceName] = apiKey
	return nil
}

func (m *mockRepository) DeleteAPIKey(ctx context.Context, serviceName string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.apiKeys, serviceName)
	return nil
}

// mockRepositoryWithOnce extends mockRepository to support one-time error
type mockRepositoryWithOnce struct {
	*mockRepository
	firstGetAllKeysError error
	getCallCount         int
}

func (m *mockRepositoryWithOnce) GetAllAPIKeys(ctx context.Context) ([]APIKeyModel, error) {
	m.getCallCount++
	if m.getCallCount == 1 && m.firstGetAllKeysError != nil {
		return nil, m.firstGetAllKeysError
	}
	return m.mockRepository.GetAllAPIKeys(ctx)
}
