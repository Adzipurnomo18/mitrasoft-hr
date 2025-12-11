package crypto

import "golang.org/x/crypto/bcrypt"

// HashPassword membuat hash bcrypt dari plain password.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword membandingkan hash yang tersimpan dengan plain password.
// URUTAN ARGUMEN HARUS: (hashed, plain).
func ComparePassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
