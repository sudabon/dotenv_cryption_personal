package crypto

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{1}, DataKeySize)
	nonce, ciphertext, err := Encrypt([]byte("HELLO=world"), key)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	plaintext, err := Decrypt(nonce, ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}

	if string(plaintext) != "HELLO=world" {
		t.Fatalf("expected original plaintext, got %q", plaintext)
	}
}

func TestEncryptRejectsInvalidKeyLength(t *testing.T) {
	t.Parallel()

	_, _, err := Encrypt([]byte("HELLO=world"), []byte("short"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid key length") {
		t.Fatalf("expected invalid key length error, got %v", err)
	}
}

func TestDecryptDetectsTampering(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{1}, DataKeySize)
	nonce, ciphertext, err := Encrypt([]byte("HELLO=world"), key)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = Decrypt(nonce, ciphertext, key)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "gcm authentication failed") {
		t.Fatalf("expected gcm authentication error, got %v", err)
	}
}

func TestWrapAndUnwrapKey(t *testing.T) {
	t.Parallel()

	dataKey := bytes.Repeat([]byte{2}, DataKeySize)
	masterKey := bytes.Repeat([]byte{3}, DataKeySize)

	wrappedKey, err := WrapKey(dataKey, masterKey)
	if err != nil {
		t.Fatalf("WrapKey returned error: %v", err)
	}

	unwrappedKey, err := UnwrapKey(wrappedKey, masterKey)
	if err != nil {
		t.Fatalf("UnwrapKey returned error: %v", err)
	}

	if !bytes.Equal(unwrappedKey, dataKey) {
		t.Fatal("expected unwrapped key to match original")
	}
}

func TestUnwrapRejectsWrongMasterKey(t *testing.T) {
	t.Parallel()

	dataKey := bytes.Repeat([]byte{2}, DataKeySize)
	masterKey := bytes.Repeat([]byte{3}, DataKeySize)

	wrappedKey, err := WrapKey(dataKey, masterKey)
	if err != nil {
		t.Fatalf("WrapKey returned error: %v", err)
	}

	_, err = UnwrapKey(wrappedKey, bytes.Repeat([]byte{4}, DataKeySize))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "gcm authentication failed") {
		t.Fatalf("expected gcm authentication error, got %v", err)
	}
}

func TestGenerateMasterKey(t *testing.T) {
	t.Parallel()

	key, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey returned error: %v", err)
	}
	if len(key) != DataKeySize {
		t.Fatalf("expected %d byte key, got %d", DataKeySize, len(key))
	}
}
