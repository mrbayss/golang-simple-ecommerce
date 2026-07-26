package repository

import "github.com/mrbayss/golang-simple-ecommerce/internal/entity"

type UserRepository interface {
	Repository[entity.User]
}

type userRepository struct {
	RepositoryImpl[entity.User]
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}
