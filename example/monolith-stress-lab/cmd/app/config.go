package app

import (
	"os"
)

type Config struct {
	DSN       string
	JWTSecret []byte
	Port      string
}

func LoadConfig() *Config {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "host=127.0.0.1 user=stressuser password=stresspassword dbname=stresslab port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}

	return &Config{
		DSN:       dsn,
		JWTSecret: []byte("super-secret-key-for-stress-lab"),
		Port:      "8080",
	}
}
