package main

import (
	"encoding/hex"
	"testing"
)

func TestChecksum(t *testing.T) {
	got := checksum("ABCD")
	want := (int('A') + int('B') + int('C') + int('D')) % 256

	if got != want {
		t.Fatalf("checksum mismatch: got %d, want %d", got, want)
	}
}

func TestManualDecrypt(t *testing.T) {
	encryptedHex := encryptLikeClient("PING")

	got, err := manualDecrypt(encryptedHex)
	if err != nil {
		t.Fatalf("manualDecrypt returned error: %v", err)
	}

	if got != "PING" {
		t.Fatalf("manualDecrypt mismatch: got %q, want %q", got, "PING")
	}
}

func TestManualDecryptRejectsInvalidHex(t *testing.T) {
	_, err := manualDecrypt("invalid")
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func encryptLikeClient(payload string) string {
	p := []byte(payload)
	encrypted := make([]byte, 4)

	for i := 0; i < 4; i++ {
		encrypted[i] = p[i] ^ MANUAL_KEY[i]
	}

	final := make([]byte, 4)
	final[0] = encrypted[2]
	final[1] = encrypted[3]
	final[2] = encrypted[0]
	final[3] = encrypted[1]

	return hex.EncodeToString(final)
}
