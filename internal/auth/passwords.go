package auth

import "github.com/alexedwards/argon2id"

func HashPassword(password string) (string, error) {
	hashed_password, err := argon2id.CreateHash(password, 
		&argon2id.Params{
			Memory: uint32(32),
			Iterations: uint32(3),
			SaltLength: uint32(16),
			Parallelism: uint8(4),
			KeyLength: uint32(32),
		})
	if err != nil {
		return "", err
	}
	return hashed_password, nil
}

func CheckPassword(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}
