package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

/*
Argon2Hasher implements repository.PasswordHasher using Argon2id, the
winner of the 2015 Password Hashing Competition. Unlike bcrypt, it is
memory-hard, making GPU/ASIC-based brute-force attacks significantly
more expensive.
*/
type Argon2Hasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

/*
NewArgon2Hasher builds an Argon2Hasher using reasonable default
parameters: 1 iteration, 64MB memory, 4 threads, 32-byte output.
*/
func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		time:    1,
		memory:  64 * 1024,
		threads: 4,
		keyLen:  32,
	}
}

/*
Hash generates a random 16-byte salt, derives an Argon2id hash from
the password, and encodes both as a single "salt$hash" string
(base64, unpadded) suitable for storage in a single database column.
*/
func (a *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, a.time, a.memory, a.threads, a.keyLen)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s$%s", encodedSalt, encodedHash), nil
}

/*
Verify decodes the stored "salt$hash" string, re-derives the hash
from the given password using the same salt, and compares the two in
constant time to avoid leaking information via timing side channels.
*/
func (a *Argon2Hasher) Verify(password, stored string) (bool, error) {
	parts := strings.Split(stored, "$")
	if len(parts) != 2 {
		return false, errors.New("invalid stored hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}
	originalHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	newHash := argon2.IDKey([]byte(password), salt, a.time, a.memory, a.threads, a.keyLen)

	return subtle.ConstantTimeCompare(originalHash, newHash) == 1, nil
}