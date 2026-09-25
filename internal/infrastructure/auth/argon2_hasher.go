package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

/*
Argon2Hasher implements port.PasswordHasher using Argon2id, the
winner of the 2015 Password Hashing Competition. Unlike bcrypt, it is
memory-hard, making GPU/ASIC-based brute-force attacks significantly
more expensive.

Every hash allocates `memory` KiB, so slots caps how many run at the
same time. Without it, a burst of login requests could exhaust server
RAM (100 concurrent hashes at 64MB each is already 6.4GB).
*/
type Argon2Hasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	slots   chan struct{}
}

/*
NewArgon2Hasher builds an Argon2Hasher using reasonable default
parameters: 1 iteration, 64MB memory, 4 threads, 32-byte output.
At most runtime.NumCPU() hashes are computed concurrently; extra
callers wait for a free slot.
*/
func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		time:    1,
		memory:  64 * 1024,
		threads: 4,
		keyLen:  32,
		slots:   make(chan struct{}, runtime.NumCPU()),
	}
}

/*
idKey derives an Argon2id key while holding one of the hasher's slots.
*/
func (a *Argon2Hasher) idKey(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte {
	a.slots <- struct{}{}
	defer func() { <-a.slots }()

	return argon2.IDKey(password, salt, time, memory, threads, keyLen)
}

/*
Hash generates a random 16-byte salt, derives an Argon2id hash from
the password, and encodes it in the standard PHC string format:

	$argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>

Storing the parameters alongside the hash means they can be raised
later without breaking verification of passwords hashed earlier.
*/
func (a *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := a.idKey([]byte(password), salt, a.time, a.memory, a.threads, a.keyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, a.memory, a.time, a.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

/*
Verify re-derives the hash from the given password using the salt and
parameters stored in the hash, and compares the two in constant time
to avoid leaking information via timing side channels. It accepts both
the PHC format produced by Hash and the older "salt$hash" format, which
is verified with the hasher's current parameters.
*/
func (a *Argon2Hasher) Verify(password, stored string) (bool, error) {
	var (
		time, memory uint32
		threads      uint8
		encodedSalt  string
		encodedHash  string
	)

	parts := strings.Split(stored, "$")

	switch len(parts) {
	case 6:
		if parts[1] != "argon2id" {
			return false, errors.New("unsupported hash algorithm")
		}

		var version int
		if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
			return false, err
		}
		if version != argon2.Version {
			return false, errors.New("unsupported argon2 version")
		}

		if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
			return false, err
		}

		encodedSalt, encodedHash = parts[4], parts[5]
	case 2:
		time, memory, threads = a.time, a.memory, a.threads
		encodedSalt, encodedHash = parts[0], parts[1]
	default:
		return false, errors.New("invalid stored hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return false, err
	}
	originalHash, err := base64.RawStdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false, err
	}

	newHash := a.idKey([]byte(password), salt, time, memory, threads, uint32(len(originalHash)))

	return subtle.ConstantTimeCompare(originalHash, newHash) == 1, nil
}