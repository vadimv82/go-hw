package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"errors"
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAll() ([]model.User, error) {
	return s.repo.GetAll()
}

func (s *userService) GetByUsername(username string) (*model.User, error) {
	user, err := s.repo.GetByUsername(username)
	return s.ensureUserExists(user, err)
}

func (s *userService) GetByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	return s.ensureUserExists(user, err)
}

func (s *userService) ensureUserExists(user *model.User, err error) (*model.User, error) {
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user is not found")
	}
	return user, nil
}
