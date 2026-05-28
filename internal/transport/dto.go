package transport

import "time"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	IsBlocked bool      `json:"is_blocked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UpdateUserRequest struct {
	Email  *string `json:"email"`
	Age    *int    `json:"age"`
	Gender *string `json:"gender"`
	Role   *string `json:"role"`
}

type ProductCreateRequest struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type ProductUpdateRequest struct {
	Name        *string  `json:"name"`
	Category    *string  `json:"category"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
}

type ProductResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	OwnerID     uint      `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CartItemSyncRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type SyncCartRequest struct {
	Items []CartItemSyncRequest `json:"items"`
}

type CartItemResponse struct {
	ID       uint            `json:"id"`
	Product  ProductResponse `json:"product"`
	Price    float64         `json:"price"`
	Quantity int             `json:"quantity"`
}

type CartResponse struct {
	Items []CartItemResponse `json:"items"`
	Total float64            `json:"total"`
}

type CheckoutResponse struct {
	OrderID            uint    `json:"order_id"`
	Status             string  `json:"status"`
	TotalAmount        float64 `json:"total_amount"`
	PaymentReference   string  `json:"payment_reference"`
	StripeClientSecret string  `json:"stripe_client_secret,omitempty"`
}

type OrderItemResponse struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	UnitPrice   float64 `json:"unit_price"`
	Quantity    int     `json:"quantity"`
	LineTotal   float64 `json:"line_total"`
}

type OrderResponse struct {
	ID              uint                `json:"id"`
	Status          string              `json:"status"`
	TotalAmount     float64             `json:"total_amount"`
	PaymentProvider string              `json:"payment_provider"`
	PaymentRef      string              `json:"payment_ref"`
	Items           []OrderItemResponse `json:"items"`
	CreatedAt       time.Time           `json:"created_at"`
}
