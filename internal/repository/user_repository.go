package repository

import (
	"github.com/mrbayss/golang-simple-ecommerce/internal/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	Repository[entity.User]
	FindByEmail(db *gorm.DB, email string) (*entity.User, error)
}

type userRepository struct {
	RepositoryImpl[entity.User]
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (ur *userRepository) FindByEmail(db *gorm.DB, email string) (*entity.User, error) {

	user, err := ur.RepositoryImpl.GetByColumn(db, "email", email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
