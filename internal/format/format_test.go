package format

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	input := Envelope{
		Nonce:      []byte("123456789012"),
		WrappedKey: []byte("wrapped-key"),
		Ciphertext: []byte("ciphertext"),
	}

	data, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if data[4] != Version {
		t.Fatalf("expected version byte %x, got %x", Version, data[4])
	}

	output, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if !bytes.Equal(output.Nonce, input.Nonce) {
		t.Fatal("nonce mismatch")
	}
	if !bytes.Equal(output.WrappedKey, input.WrappedKey) {
		t.Fatal("wrapped key mismatch")
	}
	if !bytes.Equal(output.Ciphertext, input.Ciphertext) {
		t.Fatal("ciphertext mismatch")
	}
}

func TestUnmarshalRejectsInvalidMagic(t *testing.T) {
	t.Parallel()

	_, err := Unmarshal([]byte("BADC"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing ENVC header") {
		t.Fatalf("expected invalid header error, got %v", err)
	}
}

func TestUnmarshalRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	data := []byte(Magic + "\x02\x0c\x00\x00")
	_, err := Unmarshal(data)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported file format version") {
		t.Fatalf("expected unsupported version error, got %v", err)
	}
}

func TestUnmarshalRejectsTruncatedPayload(t *testing.T) {
	t.Parallel()

	data, err := Marshal(Envelope{
		Nonce:      []byte("123456789012"),
		WrappedKey: []byte("wrapped-key"),
		Ciphertext: []byte("ciphertext"),
	})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	_, err = Unmarshal(data[:headerSize+len("123456789012")+len("wrapped-key")-3])
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unexpected end of data") {
		t.Fatalf("expected truncated payload error, got %v", err)
	}
}
