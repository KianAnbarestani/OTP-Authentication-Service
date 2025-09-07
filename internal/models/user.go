package models

import "time"

type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Phone        string    `gorm:"uniqueIndex;size:32" json:"phone"`
	RegisteredAt time.Time `gorm:"autoCreateTime" json:"registered_at"`
}
