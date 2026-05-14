package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"somewebproject/internal/auth"
	"somewebproject/internal/cache"
	"somewebproject/internal/models"
	"somewebproject/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleUser   = "user"
	RoleSeller = "seller"
	RoleAdmin  = "admin"
)

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
	Create(ctx context.Context, ownerID uint, name, description string, price float64, stock int) (*models.Product, error)
	List(ctx context.Context) ([]models.Product, error)
	GetByID(ctx context.Context, id uint) (*models.Product, error)
	Update(ctx context.Context, id uint, updates map[string]any) (*models.Product, error)
	Delete(ctx context.Context, id uint) error
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
	repo  repository.ProductRepository
	cache cache.Cache
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

func NewProductService(repo repository.ProductRepository, c cache.Cache) ProductService {
	return &productService{repo: repo, cache: c}
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
		Role:     RoleUser,
		Age:      age,
		Gender:   gender,
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
func (s *productService) Create(ctx context.Context, ownerID uint, name, description string, price float64, stock int) (*models.Product, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(description) == "" {
		return nil, errors.New("name and description are required")
	}
	if price < 0 {
		return nil, errors.New("price must be positive")
	}
	if stock < 0 {
		return nil, errors.New("stock must be positive")
	}

	product := &models.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(ctx, "products:list")
	return product, nil
}

func (s *productService) List(ctx context.Context) ([]models.Product, error) {
	const cacheKey = "products:list"
	const cacheTTL = 10 * time.Minute

	var products []models.Product
	if err := s.cache.Get(ctx, cacheKey, &products); err == nil {
		return products, nil
	}

	products, err := s.repo.List(ctx)
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
	_ = s.cache.Delete(ctx, "products:list")

	return product, nil
}

func (s *productService) Delete(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("products:%d", id)
	_ = s.cache.Delete(ctx, cacheKey)
	_ = s.cache.Delete(ctx, "products:list")

	return nil
}
