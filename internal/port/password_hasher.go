package port

/*
PasswordHasher is the contract for hashing and verifying passwords.
The usecase layer depends only on this interface, so the underlying
algorithm (bcrypt, Argon2, etc.) can be swapped without touching
business logic.
*/
type PasswordHasher interface {
	/*
		Hash produces a salted hash of the given plaintext password,
		safe to persist.
	*/
	Hash(password string) (string, error)

	/*
		Verify reports whether the given plaintext password matches the
		previously hashed value.
	*/
	Verify(password, hashed string) (bool, error)
}
