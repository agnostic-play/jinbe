package config

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repository for testing
type mockRepository struct {
	data      map[string]ConfitEnt
	shouldErr bool
}

func (m *mockRepository) GetConfigEntity(ctx context.Context, configID string) (ConfitEnt, error) {
	if m.shouldErr {
		return ConfitEnt{}, errors.New("mock error")
	}

	entity, exists := m.data[configID]
	if !exists {
		return ConfitEnt{}, errors.New("config not found")
	}

	return entity, nil
}

func (m *mockRepository) CreateOrUpdate(ctx context.Context, configID string, configVal ConfitEnt) (ConfitEnt, error) {
	if m.shouldErr {
		return ConfitEnt{}, errors.New("mock error")
	}
	m.data[configID] = configVal
	return configVal, nil
}

func (m *mockRepository) Delete(ctx context.Context, configID string) error {
	if m.shouldErr {
		return errors.New("mock error")
	}
	delete(m.data, configID)
	return nil
}

// Test config struct
type TestConfig struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (c TestConfig) GetConfigID() string {
	return "test-config"
}

func (c TestConfig) GetConfigValue() TestConfig {
	return c
}

// TestNewClientConfig tests config client creation
func TestNewClientConfig(t *testing.T) {
	config := TestConfig{Name: "test", Value: 123}
	repo := &mockRepository{data: make(map[string]ConfitEnt)}

	client := NewClientConfig(config, repo)

	assert.NotNil(t, client)
}

// TestClientConfig_Get tests config retrieval
func TestClientConfig_Get(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		expectedConfig := TestConfig{Name: "production", Value: 999}
		rawValue, _ := json.Marshal(expectedConfig)

		repo := &mockRepository{
			data: map[string]ConfitEnt{
				"test-config": {
					ConfigID: "test-config",
					RawValue: string(rawValue),
				},
			},
		}

		baseConfig := TestConfig{}
		client := NewClientConfig(baseConfig, repo)

		result, err := client.Get(context.Background())

		require.NoError(t, err)
		assert.Equal(t, expectedConfig.Name, result.Name)
		assert.Equal(t, expectedConfig.Value, result.Value)
	})

	t.Run("config not found", func(t *testing.T) {
		repo := &mockRepository{
			data: make(map[string]ConfitEnt),
		}

		baseConfig := TestConfig{}
		client := NewClientConfig(baseConfig, repo)

		_, err := client.Get(context.Background())

		assert.Error(t, err)
	})

	t.Run("empty raw value", func(t *testing.T) {
		repo := &mockRepository{
			data: map[string]ConfitEnt{
				"test-config": {
					ConfigID: "test-config",
					RawValue: "",
				},
			},
		}

		baseConfig := TestConfig{}
		client := NewClientConfig(baseConfig, repo)

		_, err := client.Get(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})
}
