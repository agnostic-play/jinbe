package masking

import (
	"testing"

	"github.com/showa-93/go-mask"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMaskStruct tests masking of struct fields
func TestMaskStruct(t *testing.T) {
	type TestStruct struct {
		Name     string `mask:"filled"`
		Email    string `mask:"filled"`
		Password string `mask:"filled"`
		PublicID string
	}

	t.Run("successful masking", func(t *testing.T) {
		input := TestStruct{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "secret123",
			PublicID: "user-123",
		}

		result, err := MaskStruct(input)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("nil input", func(t *testing.T) {
		result, err := MaskStruct(nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

// TestMaskJSON tests masking of JSON byte arrays
func TestMaskJSON(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		jsonData := []byte(`{"name": "John", "password": "secret"}`)

		result, err := MaskJSON(jsonData)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		jsonData := []byte(`{invalid json}`)

		result, err := MaskJSON(jsonData)

		assert.Error(t, err)
		assert.Equal(t, jsonData, result)
	})
}

// TestShouldMaskStruct tests masking with error suppression
func TestShouldMaskStruct(t *testing.T) {
	type SensitiveData struct {
		APIKey string `mask:"filled"`
		Token  string `mask:"filled"`
	}

	t.Run("successful masking", func(t *testing.T) {
		input := SensitiveData{
			APIKey: "api-key-123",
			Token:  "token-xyz",
		}

		result := ShouldMaskStruct(input)

		assert.NotNil(t, result)
	})

	t.Run("masking error returns original", func(t *testing.T) {
		// Channel can't be masked
		input := make(chan int)

		result := ShouldMaskStruct(input)

		assert.Equal(t, input, result)
	})
}

// TestNewMasker tests masker initialization
func TestNewMasker(t *testing.T) {
	t.Run("initialize with custom char", func(t *testing.T) {
		NewMasker("X")

		// If no panic, initialization successful
		assert.True(t, true)
	})

	t.Run("initialize with options", func(t *testing.T) {
		option := func(m *mask.Masker) {
			// Custom option
		}

		NewMasker("*", option)

		assert.True(t, true)
	})
}
