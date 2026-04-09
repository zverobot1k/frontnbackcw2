package repository

import (
	"context"

	"somewebproject/internal/models"

	"gorm.io/gorm"
)

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
	List(ctx context.Context) ([]models.Product, error)
	Update(ctx context.Context, id uint, updates map[string]any) (*models.Product, error)
	Delete(ctx context.Context, id uint) error
}

type userRepo struct {
	db *gorm.DB
}

type productRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func NewProductRepo(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
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

func (r *productRepo) List(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&products).Error; err != nil {
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

func (r *productRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Product{}, id).Error
}
