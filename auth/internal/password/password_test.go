package password

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

const testPassword = "a sufficiently long password"

func TestHashAndVerify(t *testing.T) {
	hasher, err := New(19456, 2, 1, 1)
	if err != nil {
		t.Fatalf("create hasher: %v", err)
	}

	encoded, err := hasher.Hash(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	match, err := hasher.Verify(testPassword, encoded)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !match {
		t.Fatal("correct password did not match")
	}

	match, err = hasher.Verify("a different password", encoded)
	if err != nil {
		t.Fatalf("verify incorrect password: %v", err)
	}
	if match {
		t.Fatal("incorrect password matched")
	}
}

func TestHashFreshSalt(t *testing.T) {
	hasher, err := New(19456, 2, 1, 1)
	if err != nil {
		t.Fatalf("create hasher: %v", err)
	}

	first, err := hasher.Hash(testPassword)
	if err != nil {
		t.Fatalf("hash first password: %v", err)
	}

	second, err := hasher.Hash(testPassword)
	if err != nil {
		t.Fatalf("hash second password: %v", err)
	}

	firstParsed, err := parseHash(first)
	if err != nil {
		t.Fatalf("parse first hash: %v", err)
	}

	secondParsed, err := parseHash(second)
	if err != nil {
		t.Fatalf("parse second hash: %v", err)
	}

	if bytes.Equal(firstParsed.salt, secondParsed.salt) {
		t.Fatal("repeated hashing used the same salt")
	}

	for _, encoded := range []string{first, second} {
		match, err := hasher.Verify(testPassword, encoded)
		if err != nil {
			t.Fatalf("verify password: %v", err)
		}
		if !match {
			t.Fatal("password did not match its salted hash")
		}
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	hasher, err := New(19456, 2, 1, 1)
	if err != nil {
		t.Fatalf("create hasher: %v", err)
	}

	encoded, err := hasher.Hash(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	parts := strings.Split(encoded, "$")
	prefix := strings.Join(parts[:4], "$")

	withSalt := func(salt string) string {
		return prefix + "$" + salt + "$" + parts[5]
	}

	withHash := func(hash string) string {
		return prefix + "$" + parts[4] + "$" + hash
	}

	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"invalid format", "invalid"},
		{"too long", strings.Repeat("x", 257)},
		{"unsupported algorithm", strings.Replace(encoded, "argon2id", "argon2i", 1)},
		{"unsupported version", strings.Replace(encoded, "v=19", "v=18", 1)},
		{"missing parameter", strings.Replace(encoded, ",p=1", "", 1)},
		{"invalid cost", strings.Replace(encoded, "t=2", "t=invalid", 1)},
		{"low memory", strings.Replace(encoded, "m=19456", "m=19455", 1)},
		{"high memory", strings.Replace(encoded, "m=19456", "m=262145", 1)},
		{"low iterations", strings.Replace(encoded, "t=2", "t=1", 1)},
		{"high iterations", strings.Replace(encoded, "t=2", "t=11", 1)},
		{"zero parallelism", strings.Replace(encoded, "p=1", "p=0", 1)},
		{"high parallelism", strings.Replace(encoded, "p=1", "p=5", 1)},
		{"invalid salt encoding", withSalt("!")},
		{"empty salt", withSalt("")},
		{"short salt", withSalt("AA")},
		{"long salt", withSalt(base64.RawStdEncoding.EncodeToString(make([]byte, 17)))},
		{"invalid hash encoding", withHash("!")},
		{"empty hash", withHash("")},
		{"short hash", withHash("AA")},
		{"long hash", withHash(base64.RawStdEncoding.EncodeToString(make([]byte, 33)))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match, err := hasher.Verify(testPassword, tc.hash)
			if !errors.Is(err, ErrInvalidHash) {
				t.Fatalf("got error %v, want %v", err, ErrInvalidHash)
			}
			if match {
				t.Fatal("invalid hash matched")
			}
		})
	}
}

func TestHasherBusy(t *testing.T) {
	hasher, err := New(19456, 2, 1, 1)
	if err != nil {
		t.Fatalf("create hasher: %v", err)
	}

	encoded, err := hasher.Hash(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := hasher.acquire(); err != nil {
		t.Fatalf("occupy hashing slot: %v", err)
	}
	t.Cleanup(hasher.release)

	if _, err := hasher.Hash(testPassword); !errors.Is(err, ErrBusy) {
		t.Fatalf("hash: got error %v, want %v", err, ErrBusy)
	}

	if _, err := hasher.Verify(testPassword, encoded); !errors.Is(err, ErrBusy) {
		t.Fatalf("verify: got error %v, want %v", err, ErrBusy)
	}
}

func TestVerifyReleasesSlot(t *testing.T) {
	hasher, err := New(19456, 2, 1, 1)
	if err != nil {
		t.Fatalf("create hasher: %v", err)
	}

	encoded, err := hasher.Hash(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		match    bool
		err      error
	}{
		{"match", testPassword, encoded, true, nil},
		{"mismatch", "a different password", encoded, false, nil},
		{"invalid hash", testPassword, "invalid", false, ErrInvalidHash},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			match, err := hasher.Verify(tc.password, tc.hash)
			if !errors.Is(err, tc.err) {
				t.Fatalf("got error %v, want %v", err, tc.err)
			}
			if match != tc.match {
				t.Fatalf("got match %t, want %t", match, tc.match)
			}

			if err := hasher.acquire(); err != nil {
				t.Fatalf("hashing slot unavailable after verification: %v", err)
			}
			hasher.release()
		})
	}
}

func TestNewInvalidSettings(t *testing.T) {
	tests := []struct {
		name        string
		memoryKiB   uint32
		iterations  uint32
		parallelism uint8
		concurrency int
	}{
		{"low memory", 19455, 2, 1, 1},
		{"high memory", 262145, 2, 1, 1},
		{"low iterations", 19456, 1, 1, 1},
		{"high iterations", 19456, 11, 1, 1},
		{"zero parallelism", 19456, 2, 0, 1},
		{"high parallelism", 19456, 2, 5, 1},
		{"zero concurrency", 19456, 2, 1, 0},
		{"negative concurrency", 19456, 2, 1, -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasher, err := New(
				tc.memoryKiB, tc.iterations,
				tc.parallelism, tc.concurrency,
			)
			if err == nil {
				t.Fatal("invalid settings were accepted")
			}
			if hasher != nil {
				t.Fatal("invalid settings returned a hasher")
			}
		})
	}
}

func TestVerifyStoredCosts(t *testing.T) {
	original, err := New(19456, 2, 1, 1)
	if err != nil {
		t.Fatalf("create original hasher: %v", err)
	}

	encoded, err := original.Hash(testPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	current, err := New(32768, 3, 2, 1)
	if err != nil {
		t.Fatalf("create current hasher: %v", err)
	}

	match, err := current.Verify(testPassword, encoded)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !match {
		t.Fatal("password did not match after changing hashing settings")
	}
}
