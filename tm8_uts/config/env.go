package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadEnv membaca file .env dan memasukkannya ke environment.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: file .env tidak ditemukan")
	}
}

// GetEnv mengambil nilai teks, kalau kosong dipakai nilai cadangan.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// GetEnvInt sama seperti GetEnv, tapi hasilnya angka.
func GetEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}