package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword — хеш пароля bcrypt. В таблице users хранится только хеш:
// даже при утечке базы пароли не узнать (в хеш входит случайная соль).
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPassword сравнивает введённый пароль с хешем из базы.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
