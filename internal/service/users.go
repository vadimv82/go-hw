package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"database/sql"
	"errors"
)

var ErrUniqueConstraint = repository.ErrUniqueConstraint

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	GetByUUID(uuid string) (*model.User, error)
	Create(user *model.User) (*model.User, error)
	Update(uuid string, user *model.User) (*model.User, error)
	Delete(uuid string) error
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

func (s *userService) GetByUUID(uuid string) (*model.User, error) {
	user, err := s.repo.GetByUUID(uuid)
	return s.ensureUserExists(user, err)
}

func (s *userService) Create(user *model.User) (*model.User, error) {
	createdUser, err := s.repo.Create(user)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueConstraint) {
			return nil, ErrUniqueConstraint
		}
		return nil, err
	}
	return createdUser, nil
}

func (s *userService) Update(uuid string, user *model.User) (*model.User, error) {
	updatedUser, err := s.repo.Update(uuid, user)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueConstraint) {
			return nil, ErrUniqueConstraint
		}
		return nil, err
	}
	if updatedUser == nil {
		return nil, errors.New("user is not found")
	}
	return updatedUser, nil
}

func (s *userService) Delete(uuid string) error {
	err := s.repo.Delete(uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user is not found")
		}
		return err
	}
	return nil
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
