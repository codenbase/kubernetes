package app

// User represents a system user mapped to the users table.
type User struct {
	ID           int    `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"uniqueIndex;size:50;not null"`
	PasswordHash string `gorm:"size:255;not null"`
}
