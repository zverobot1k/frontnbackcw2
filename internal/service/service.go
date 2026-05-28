package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"somewebproject/internal/auth"
	"somewebproject/internal/cache"
	"somewebproject/internal/models"
	"somewebproject/internal/repository"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/paymentintent"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
)

var allowedRoles = map[string]struct{}{
	RoleCustomer: {},
	RoleAdmin:    {},
}

type ProductListFilter struct {
	Query       string
	Category    string
	MinPrice    *float64
	MaxPrice    *float64
	OnlyInStock bool
}

type CartLineInput struct {
	ProductID uint
	Quantity  int
}

type CheckoutResult struct {
	Order        *models.Order
	ClientSecret string
}

type AuthService interface {
	Register(ctx context.Context, email, password, gender string, age int) (*models.User, error)
	Login(ctx context.Context, email, password string) (*auth.TokenPair, *models.User, error)
	Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, *models.User, error)
	Me(ctx context.Context, id uint) (*models.User, error)
}

type UserService interface {
	List(ctx context.Context) ([]models.User, error)
	GetByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, id uint, updates map[string]any) (*models.User, error)
	Block(ctx context.Context, id uint) error
}

type ProductService interface {
	Create(ctx context.Context, ownerID uint, name, category, description string, price float64, stock int) (*models.Product, error)
	List(ctx context.Context, filter ProductListFilter) ([]models.Product, error)
	GetByID(ctx context.Context, id uint) (*models.Product, error)
	Update(ctx context.Context, id uint, updates map[string]any) (*models.Product, error)
	Delete(ctx context.Context, id uint) error
}

type CartService interface {
	GetByUserID(ctx context.Context, userID uint) ([]models.CartItem, error)
	Sync(ctx context.Context, userID uint, lines []CartLineInput) ([]models.CartItem, error)
}

type OrderService interface {
	Checkout(ctx context.Context, userID uint) (*CheckoutResult, error)
	ListMyOrders(ctx context.Context, userID uint) ([]models.Order, error)
	ListAllOrders(ctx context.Context) ([]models.Order, error)
	MarkPaidByPaymentRef(ctx context.Context, paymentRef string) error
}

type authService struct {
	users      repository.UserRepository
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type userService struct {
	repo  repository.UserRepository
	cache cache.Cache
}

type productService struct {
	repo      repository.ProductRepository
	cartRepo  repository.CartRepository
	orderRepo repository.OrderRepository
	cache     cache.Cache
	stripeKey string
}

type cartService struct {
	productRepo repository.ProductRepository
	cartRepo    repository.CartRepository
	cache       cache.Cache
}

type orderService struct {
	productRepo repository.ProductRepository
	cartRepo    repository.CartRepository
	orderRepo   repository.OrderRepository
	cache       cache.Cache
	stripeKey   string
}

func NewAuthService(users repository.UserRepository, secret string) AuthService {
	return &authService{
		users:      users,
		secret:     secret,
		accessTTL:  15 * time.Minute,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

func NewUserService(repo repository.UserRepository, c cache.Cache) UserService {
	return &userService{repo: repo, cache: c}
}

func NewProductService(repo repository.ProductRepository, cartRepo repository.CartRepository, orderRepo repository.OrderRepository, c cache.Cache, stripeKey string) ProductService {
	return &productService{repo: repo, cartRepo: cartRepo, orderRepo: orderRepo, cache: c, stripeKey: stripeKey}
}

func NewCartService(productRepo repository.ProductRepository, cartRepo repository.CartRepository, c cache.Cache) CartService {
	return &cartService{productRepo: productRepo, cartRepo: cartRepo, cache: c}
}

func NewOrderService(productRepo repository.ProductRepository, cartRepo repository.CartRepository, orderRepo repository.OrderRepository, c cache.Cache, stripeKey string) OrderService {
	return &orderService{productRepo: productRepo, cartRepo: cartRepo, orderRepo: orderRepo, cache: c, stripeKey: stripeKey}
}

func (s *authService) Register(ctx context.Context, email, password, gender string, age int) (*models.User, error) {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("email and password are required")
	}

	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, errors.New("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    email,
		Password: string(hash),
		Role:     RoleCustomer,
		Age:      age,
		Gender:   gender,
	}

	if users, err := s.users.List(ctx); err == nil && len(users) == 0 {
		user.Role = RoleAdmin
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*auth.TokenPair, *models.User, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	if user.IsBlocked {
		return nil, nil, errors.New("user is blocked")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	pair, err := auth.GenerateTokenPair(user, s.secret, s.accessTTL, s.refreshTTL)
	if err != nil {
		return nil, nil, err
	}

	return pair, user, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, *models.User, error) {
	claims, err := auth.ParseToken(refreshToken, s.secret)
	if err != nil {
		return nil, nil, err
	}

	if claims.TokenType != auth.TokenTypeRefresh {
		return nil, nil, errors.New("invalid refresh token")
	}

	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, err
	}

	if user.IsBlocked {
		return nil, nil, errors.New("user is blocked")
	}

	pair, err := auth.GenerateTokenPair(user, s.secret, s.accessTTL, s.refreshTTL)
	if err != nil {
		return nil, nil, err
	}

	return pair, user, nil
}

func (s *authService) Me(ctx context.Context, id uint) (*models.User, error) {
	return s.users.FindByID(ctx, id)
}

// UserService methods with caching
func (s *userService) List(ctx context.Context) ([]models.User, error) {
	const cacheKey = "users:list"
	const cacheTTL = 1 * time.Minute

	var users []models.User
	if err := s.cache.Get(ctx, cacheKey, &users); err == nil {
		return users, nil
	}

	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, users, cacheTTL)
	return users, nil
}

func (s *userService) GetByID(ctx context.Context, id uint) (*models.User, error) {
	cacheKey := fmt.Sprintf("users:%d", id)
	const cacheTTL = 1 * time.Minute

	var cachedUser models.User
	if err := s.cache.Get(ctx, cacheKey, &cachedUser); err == nil {
		return &cachedUser, nil
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, user, cacheTTL)
	return user, nil
}

func (s *userService) Update(ctx context.Context, id uint, updates map[string]any) (*models.User, error) {
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	if email, ok := updates["email"].(string); ok && strings.TrimSpace(email) == "" {
		return nil, errors.New("email cannot be empty")
	}

	if role, ok := updates["role"].(string); ok {
		if _, found := allowedRoles[strings.TrimSpace(strings.ToLower(role))]; !found {
			return nil, errors.New("role must be customer or admin")
		}
		updates["role"] = strings.TrimSpace(strings.ToLower(role))
	}

	user, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("users:%d", id)
	_ = s.cache.Delete(ctx, cacheKey)
	_ = s.cache.Delete(ctx, "users:list")

	return user, nil
}

func (s *userService) Block(ctx context.Context, id uint) error {
	err := s.repo.Block(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("users:%d", id)
	_ = s.cache.Delete(ctx, cacheKey)
	_ = s.cache.Delete(ctx, "users:list")

	return nil
}

// ProductService methods with caching
func (s *productService) Create(ctx context.Context, ownerID uint, name, category, description string, price float64, stock int) (*models.Product, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(description) == "" || strings.TrimSpace(category) == "" {
		return nil, errors.New("name, category and description are required")
	}
	if price < 0 {
		return nil, errors.New("price must be positive")
	}
	if stock < 0 {
		return nil, errors.New("stock must be positive")
	}

	product := &models.Product{
		Name:        name,
		Category:    strings.TrimSpace(strings.ToLower(category)),
		Description: description,
		Price:       price,
		Stock:       stock,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	_ = s.cache.Clear(ctx, "products:list:*")
	return product, nil
}

func (s *productService) List(ctx context.Context, filter ProductListFilter) ([]models.Product, error) {
	cacheKey := fmt.Sprintf(
		"products:list:%s:%s:%v:%v:%t",
		strings.TrimSpace(strings.ToLower(filter.Query)),
		strings.TrimSpace(strings.ToLower(filter.Category)),
		filter.MinPrice,
		filter.MaxPrice,
		filter.OnlyInStock,
	)
	const cacheTTL = 10 * time.Minute

	var products []models.Product
	if err := s.cache.Get(ctx, cacheKey, &products); err == nil {
		return products, nil
	}

	products, err := s.repo.List(ctx, repository.ProductFilter{
		Query:       filter.Query,
		Category:    filter.Category,
		MinPrice:    filter.MinPrice,
		MaxPrice:    filter.MaxPrice,
		OnlyInStock: filter.OnlyInStock,
	})
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, products, cacheTTL)
	return products, nil
}

func (s *productService) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	cacheKey := fmt.Sprintf("products:%d", id)
	const cacheTTL = 10 * time.Minute

	var cachedProduct models.Product
	if err := s.cache.Get(ctx, cacheKey, &cachedProduct); err == nil {
		return &cachedProduct, nil
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, product, cacheTTL)
	return product, nil
}

func (s *productService) Update(ctx context.Context, id uint, updates map[string]any) (*models.Product, error) {
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	if name, ok := updates["name"].(string); ok && strings.TrimSpace(name) == "" {
		return nil, errors.New("name cannot be empty")
	}
	if description, ok := updates["description"].(string); ok && strings.TrimSpace(description) == "" {
		return nil, errors.New("description cannot be empty")
	}
	if category, ok := updates["category"].(string); ok {
		trimmed := strings.TrimSpace(strings.ToLower(category))
		if trimmed == "" {
			return nil, errors.New("category cannot be empty")
		}
		updates["category"] = trimmed
	}
	if price, ok := updates["price"].(float64); ok && price < 0 {
		return nil, errors.New("price must be positive")
	}
	if stock, ok := updates["stock"].(int); ok && stock < 0 {
		return nil, errors.New("stock must be positive")
	}

	product, err := s.repo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("products:%d", id)
	_ = s.cache.Delete(ctx, cacheKey)
	_ = s.cache.Clear(ctx, "products:list:*")

	return product, nil
}

func (s *productService) Delete(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("products:%d", id)
	_ = s.cache.Delete(ctx, cacheKey)
	_ = s.cache.Clear(ctx, "products:list:*")

	return nil
}

func (s *cartService) GetByUserID(ctx context.Context, userID uint) ([]models.CartItem, error) {
	cacheKey := fmt.Sprintf("cart:%d", userID)
	const cacheTTL = 2 * time.Minute

	var items []models.CartItem
	if err := s.cache.Get(ctx, cacheKey, &items); err == nil {
		return items, nil
	}

	items, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, items, cacheTTL)
	return items, nil
}

func (s *cartService) Sync(ctx context.Context, userID uint, lines []CartLineInput) ([]models.CartItem, error) {
	normalized := make([]models.CartItem, 0, len(lines))

	for i := range lines {
		line := lines[i]
		if line.ProductID == 0 || line.Quantity <= 0 {
			continue
		}

		product, err := s.productRepo.FindByID(ctx, line.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %d not found", line.ProductID)
		}

		if line.Quantity > product.Stock {
			return nil, fmt.Errorf("insufficient stock for product %d", line.ProductID)
		}

		normalized = append(normalized, models.CartItem{
			UserID:    userID,
			ProductID: line.ProductID,
			Quantity:  line.Quantity,
			Price:     product.Price,
		})
	}

	items, err := s.cartRepo.ReplaceByUserID(ctx, userID, normalized)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("cart:%d", userID))
	return items, nil
}

func (s *orderService) Checkout(ctx context.Context, userID uint) (*CheckoutResult, error) {
	items, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, errors.New("cart is empty")
	}

	orderItems := make([]models.OrderItem, 0, len(items))
	total := 0.0

	for i := range items {
		line := items[i]
		product, err := s.productRepo.FindByID(ctx, line.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %d not found", line.ProductID)
		}

		if line.Quantity > product.Stock {
			return nil, fmt.Errorf("insufficient stock for product %d", line.ProductID)
		}

		lineTotal := product.Price * float64(line.Quantity)
		total += lineTotal

		orderItems = append(orderItems, models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			UnitPrice:   product.Price,
			Quantity:    line.Quantity,
			LineTotal:   lineTotal,
		})
	}

	paymentRef := ""
	clientSecret := ""
	if strings.TrimSpace(s.stripeKey) != "" {
		stripe.Key = s.stripeKey
		params := &stripe.PaymentIntentParams{
			Amount:   stripe.Int64(int64(math.Round(total * 100))),
			Currency: stripe.String(string(stripe.CurrencyUSD)),
		}
		params.Metadata = map[string]string{
			"user_id": fmt.Sprintf("%d", userID),
		}

		pi, err := paymentintent.New(params)
		if err != nil {
			return nil, fmt.Errorf("stripe checkout failed: %w", err)
		}

		paymentRef = pi.ID
		clientSecret = pi.ClientSecret
	}

	for i := range orderItems {
		if err := s.productRepo.DecreaseStock(ctx, orderItems[i].ProductID, orderItems[i].Quantity); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("insufficient stock for product %d", orderItems[i].ProductID)
			}
			return nil, err
		}
	}

	order := &models.Order{
		UserID:          userID,
		Status:          "created",
		TotalAmount:     total,
		PaymentProvider: "stripe",
		PaymentRef:      paymentRef,
		Items:           orderItems,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	if err := s.cartRepo.ClearByUserID(ctx, userID); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, fmt.Sprintf("cart:%d", userID))
	_ = s.cache.Clear(ctx, "products:list:*")

	return &CheckoutResult{Order: order, ClientSecret: clientSecret}, nil
}

func (s *orderService) ListMyOrders(ctx context.Context, userID uint) ([]models.Order, error) {
	return s.orderRepo.ListByUserID(ctx, userID)
}

func (s *orderService) ListAllOrders(ctx context.Context) ([]models.Order, error) {
	return s.orderRepo.ListAll(ctx)
}

func (s *orderService) MarkPaidByPaymentRef(ctx context.Context, paymentRef string) error {
	paymentRef = strings.TrimSpace(paymentRef)
	if paymentRef == "" {
		return errors.New("empty payment reference")
	}

	updated, err := s.orderRepo.MarkPaidByPaymentRef(ctx, paymentRef)
	if err != nil {
		return err
	}

	if !updated {
		return nil
	}

	_ = s.cache.Clear(ctx, "orders:*")
	return nil
}
