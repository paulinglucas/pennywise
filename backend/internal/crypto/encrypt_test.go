package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveKey(t *testing.T) {
	key := DeriveKey("test-passphrase")
	assert.Len(t, key, 32)

	key2 := DeriveKey("test-passphrase")
	assert.Equal(t, key, key2)

	key3 := DeriveKey("different-passphrase")
	assert.NotEqual(t, key, key3)
}

func TestEncryptDecrypt(t *testing.T) {
	key := DeriveKey("my-secret-key")
	plaintext := "https://user:pass@beta-bridge.simplefin.org/simplefin" //nolint:gosec // test data, not real credentials

	ciphertext, err := Encrypt(key, plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	assert.NotEmpty(t, ciphertext)

	decrypted, err := Decrypt(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	key := DeriveKey("my-secret-key")
	plaintext := "same-input"

	ct1, err := Encrypt(key, plaintext)
	require.NoError(t, err)

	ct2, err := Encrypt(key, plaintext)
	require.NoError(t, err)

	assert.NotEqual(t, ct1, ct2)
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := DeriveKey("correct-key")
	key2 := DeriveKey("wrong-key")

	ciphertext, err := Encrypt(key1, "secret-data")
	require.NoError(t, err)

	_, err = Decrypt(key2, ciphertext)
	assert.Error(t, err)
}

func TestDecryptInvalidData(t *testing.T) {
	key := DeriveKey("key")

	_, err := Decrypt(key, "not-valid-base64!@#$")
	assert.Error(t, err)

	_, err = Decrypt(key, "dG9vc2hvcnQ=")
	assert.Error(t, err)
}

func TestEncryptEmptyString(t *testing.T) {
	key := DeriveKey("key")

	ciphertext, err := Encrypt(key, "")
	require.NoError(t, err)

	decrypted, err := Decrypt(key, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, "", decrypted)
}
