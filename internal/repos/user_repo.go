package repos

import (
	"context"

	"OTP-Authenticate-Service/internal/models"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// CreateIfNotExist registers a user if not existing
func (r *UserRepo) CreateIfNotExist(ctx context.Context, phone string) (*models.User, error) {
	user := &models.User{Phone: phone}
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).FirstOrCreate(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// GetByID retrieves user by uint ID
func (r *UserRepo) GetByID(ctx context.Context, id uint) (*models.User, error) {
	user := &models.User{}
	if err := r.db.WithContext(ctx).First(user, id).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// List retrieves users with pagination and optional phone search
func (r *UserRepo) List(ctx context.Context, page, limit int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	query := r.db.WithContext(ctx).Model(&models.User{})

	if search != "" {
		query = query.Where("phone LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
