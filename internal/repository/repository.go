package repository

import (
	"context"
	"strings"

	"somewebproject/internal/models"

	"gorm.io/gorm"
)

type ProductFilter struct {
	Query       string
	Category    string
	MinPrice    *float64
	MaxPrice    *float64
	OnlyInStock bool
}

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	List(ctx context.Context) ([]models.User, error)
	Update(ctx context.Context, id uint, updates map[string]any) (*models.User, error)
	Block(ctx context.Context, id uint) error
}

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	FindByID(ctx context.Context, id uint) (*models.Product, error)
	List(ctx context.Context, filter ProductFilter) ([]models.Product, error)
	Update(ctx context.Context, id uint, updates map[string]any) (*models.Product, error)
	DecreaseStock(ctx context.Context, id uint, quantity int) error
	Delete(ctx context.Context, id uint) error
}

type CartRepository interface {
	GetByUserID(ctx context.Context, userID uint) ([]models.CartItem, error)
	ReplaceByUserID(ctx context.Context, userID uint, items []models.CartItem) ([]models.CartItem, error)
	ClearByUserID(ctx context.Context, userID uint) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	ListByUserID(ctx context.Context, userID uint) ([]models.Order, error)
	ListAll(ctx context.Context) ([]models.Order, error)
	MarkPaidByPaymentRef(ctx context.Context, paymentRef string) (bool, error)
}

type userRepo struct {
	db *gorm.DB
}

type productRepo struct {
	db *gorm.DB
}

type cartRepo struct {
	db *gorm.DB
}

type orderRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func NewProductRepo(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func NewCartRepo(db *gorm.DB) CartRepository {
	return &cartRepo{db: db}
}

func NewOrderRepo(db *gorm.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) List(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepo) Update(ctx context.Context, id uint, updates map[string]any) (*models.User, error) {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *userRepo) Block(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("is_blocked", true).Error
}

func (r *productRepo) Create(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepo) FindByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepo) List(ctx context.Context, filter ProductFilter) ([]models.Product, error) {
	var products []models.Product
	query := r.db.WithContext(ctx).Model(&models.Product{})

	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("lower(name) LIKE ? OR lower(description) LIKE ?", like, like)
	}

	if category := strings.TrimSpace(filter.Category); category != "" {
		query = query.Where("lower(category) = ?", strings.ToLower(category))
	}

	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}

	if filter.OnlyInStock {
		query = query.Where("stock > 0")
	}

	if err := query.Order("created_at desc").Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepo) Update(ctx context.Context, id uint, updates map[string]any) (*models.Product, error) {
	if err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *productRepo) DecreaseStock(ctx context.Context, id uint, quantity int) error {
	result := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ? AND stock >= ?", id, quantity).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *productRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, id).Error
}

func (r *cartRepo) GetByUserID(ctx context.Context, userID uint) ([]models.CartItem, error) {
	var items []models.CartItem
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at asc").
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *cartRepo) ReplaceByUserID(ctx context.Context, userID uint, items []models.CartItem) ([]models.CartItem, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		for i := range items {
			items[i].ID = 0
			items[i].UserID = userID
		}

		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByUserID(ctx, userID)
}

func (r *cartRepo) ClearByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.CartItem{}).Error
}

func (r *orderRepo) Create(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepo) ListByUserID(ctx context.Context, userID uint) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepo) ListAll(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Order("created_at desc").
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepo) MarkPaidByPaymentRef(ctx context.Context, paymentRef string) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("payment_provider = ? AND payment_ref = ?", "stripe", paymentRef).
		Where("status <> ?", "paid").
		Update("status", "paid")

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
