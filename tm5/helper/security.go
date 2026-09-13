package helper

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost menentukan berapa kali proses hashing diulang secara internal.
// Cost 12 dipilih sebagai kompromi: cukup lambat buat mempersulit penyerang
// menebak password satu per satu, tapi masih cukup cepat buat sekali login.
const bcryptCost = 12

// HashPassword mengubah password menjadi hash yang tidak dapat dikembalikan.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// VerifyPassword membandingkan password dengan hash-nya.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}