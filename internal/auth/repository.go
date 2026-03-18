package auth

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	Delete(id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *repository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *repository) FindByID(id string) (*User, error) {
	var user User
	err := r.db.First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *repository) Delete(id int64) error {

	var user User
	err := r.db.First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return r.db.Delete(&user).Error
}
