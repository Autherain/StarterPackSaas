package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the users table in the database.
// Note: The unique index on LOWER(email) is created by migrations, not GORM.
type User struct {
	ID               int       `gorm:"primaryKey;autoIncrement"`
	Created          time.Time `gorm:"not null;type:timestamptz"`
	Email            string    `gorm:"not null;type:text"`
	HashedPassword   string    `gorm:"not null;type:text;column:hashed_password"`
	SubscriptionTier string    `gorm:"not null;type:text;column:subscription_tier;index:idx_users_subscription_tier"`
}

// TableName specifies the table name for GORM.
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that sets defaults before creating a user.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Created.IsZero() {
		u.Created = time.Now()
	}
	if u.SubscriptionTier == "" {
		u.SubscriptionTier = "free"
	}
	return nil
}

