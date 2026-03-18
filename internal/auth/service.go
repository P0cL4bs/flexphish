package auth

import (
	"errors"
	"fmt"
)

type Service interface {
	Register(email, password string, role Role) (*User, error)
	Authenticate(email, password string) (*User, error)
	DeleteByEmail(email string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) DeleteByEmail(email string) error {

	if email == "" {
		return fmt.Errorf("email is required")
	}

	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	return s.repo.Delete(user.ID)
}

func (s *service) Register(email, password string, role Role) (*User, error) {

	existing, _ := s.repo.FindByEmail(email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Authenticate(email, password string) (*User, error) {

	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user disabled")
	}

	return user, nil
}
