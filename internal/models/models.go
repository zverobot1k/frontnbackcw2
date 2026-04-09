package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email     string `gorm:"uniqueIndex;not null" json:"email"`
	Password  string `gorm:"not null" json:"-"`
	Role      string `gorm:"not null;default:user" json:"role"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	IsBlocked bool   `gorm:"not null;default:false" json:"is_blocked"`
}

type Product struct {
	gorm.Model
	Name        string  `gorm:"not null" json:"name"`
	Description string  `gorm:"not null" json:"description"`
	Price       float64 `gorm:"not null" json:"price"`
	Stock       int     `gorm:"not null;default:0" json:"stock"`
	OwnerID     uint    `gorm:"index;not null" json:"owner_id"`
	Owner       User    `json:"-"`
}
