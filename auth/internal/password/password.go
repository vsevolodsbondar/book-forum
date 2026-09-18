package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrBusy        = errors.New("password hashing capacity exhausted")
	ErrInvalidHash = errors.New("invalid password hash")
)

// Hasher handles password hashing and verification with shared concurrency limits.
type Hasher struct {
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
	slots       chan struct{}
}

type parsedHash struct {
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	hash        []byte
}

// New creates a Hasher with validated costs and a shared concurrency limit.
func New(memoryKiB, iterations uint32, parallelism uint8, maxConcurrency int) (*Hasher, error) {
	if memoryKiB < 19456 || memoryKiB > 262144 ||
		iterations < 2 || iterations > 10 ||
		parallelism < 1 || parallelism > 4 {
		return nil, errors.New("password hashing costs outside accepted bounds")
	}

	if maxConcurrency <= 0 {
		return nil, errors.New("password hashing concurrency must be positive")
	}

	return &Hasher{
		memoryKiB:   memoryKiB,
		iterations:  iterations,
		parallelism: parallelism,
		slots:       make(chan struct{}, maxConcurrency),
	}, nil
}

// Hash generates a salted, self-describing Argon2id password hash.
func (h *Hasher) Hash(password string) (string, error) {
	if err := h.acquire(); err != nil {
		return "", err
	}
	defer h.release()

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password), salt,
		h.iterations, h.memoryKiB, h.parallelism, 32,
	)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memoryKiB, h.iterations, h.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// Verify reports whether the password matches the encoded hash.
func (h *Hasher) Verify(password, encoded string) (bool, error) {
	parsed, err := parseHash(encoded)
	if err != nil {
		return false, err
	}

	if err := h.acquire(); err != nil {
		return false, err
	}
	defer h.release()

	hash := argon2.IDKey(
		[]byte(password), parsed.salt,
		parsed.iterations, parsed.memoryKiB, parsed.parallelism, 32,
	)

	return subtle.ConstantTimeCompare(hash, parsed.hash) == 1, nil
}

func parseHash(encoded string) (parsedHash, error) {
	if len(encoded) > 256 {
		return parsedHash{}, ErrInvalidHash
	}

	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" ||
		parts[1] != "argon2id" ||
		parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return parsedHash{}, ErrInvalidHash
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return parsedHash{}, ErrInvalidHash
	}

	var costs [3]uint64
	for i, prefix := range []string{"m=", "t=", "p="} {
		value, ok := strings.CutPrefix(params[i], prefix)
		if !ok {
			return parsedHash{}, ErrInvalidHash
		}

		cost, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return parsedHash{}, ErrInvalidHash
		}
		costs[i] = cost
	}

	if costs[0] < 19456 || costs[0] > 262144 ||
		costs[1] < 2 || costs[1] > 10 ||
		costs[2] < 1 || costs[2] > 4 {
		return parsedHash{}, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return parsedHash{}, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(hash) != 32 {
		return parsedHash{}, ErrInvalidHash
	}

	return parsedHash{
		memoryKiB:   uint32(costs[0]),
		iterations:  uint32(costs[1]),
		parallelism: uint8(costs[2]),
		salt:        salt,
		hash:        hash,
	}, nil
}

func (h *Hasher) acquire() error {
	select {
	case h.slots <- struct{}{}:
		return nil
	default:
		return ErrBusy
	}
}

func (h *Hasher) release() {
	<-h.slots
}
