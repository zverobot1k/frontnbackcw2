package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email     string `gorm:"uniqueIndex;not null" json:"email"`
	Password  string `gorm:"not null" json:"-"`
	Role      string `gorm:"not null;default:customer" json:"role"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	IsBlocked bool   `gorm:"not null;default:false" json:"is_blocked"`
}

type Product struct {
	gorm.Model
	Name        string  `gorm:"not null" json:"name"`
	Category    string  `gorm:"index;not null;default:general" json:"category"`
	Description string  `gorm:"not null" json:"description"`
	Price       float64 `gorm:"not null" json:"price"`
	Stock       int     `gorm:"not null;default:0" json:"stock"`
	OwnerID     uint    `gorm:"index;not null" json:"owner_id"`
	Owner       User    `json:"-"`
}

type CartItem struct {
	gorm.Model
	UserID    uint    `gorm:"index;not null" json:"user_id"`
	ProductID uint    `gorm:"index;not null" json:"product_id"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"not null" json:"price"`
	Product   Product `json:"product"`
}

type Order struct {
	gorm.Model
	UserID          uint        `gorm:"index;not null" json:"user_id"`
	Status          string      `gorm:"index;not null;default:pending" json:"status"`
	TotalAmount     float64     `gorm:"not null" json:"total_amount"`
	PaymentProvider string      `gorm:"not null;default:stripe" json:"payment_provider"`
	PaymentRef      string      `gorm:"index" json:"payment_ref"`
	Items           []OrderItem `json:"items"`
}

type OrderItem struct {
	gorm.Model
	OrderID     uint    `gorm:"index;not null" json:"order_id"`
	ProductID   uint    `gorm:"index;not null" json:"product_id"`
	ProductName string  `gorm:"not null" json:"product_name"`
	UnitPrice   float64 `gorm:"not null" json:"unit_price"`
	Quantity    int     `gorm:"not null" json:"quantity"`
	LineTotal   float64 `gorm:"not null" json:"line_total"`
	Product     Product `json:"product"`
}
