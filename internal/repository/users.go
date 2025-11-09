package repository

import (
	"context"
	"cruder/internal/model"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

var ErrUniqueConstraint = errors.New("username or email already exists")

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	GetByUUID(uuid string) (*model.User, error)
	Create(user *model.User) (*model.User, error)
	Update(uuid string, user *model.User) (*model.User, error)
	Delete(uuid string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.QueryContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(id int64) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByUUID(uuid string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE uuid = $1`, uuid).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(user *model.User) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(
		context.Background(),
		`INSERT INTO users (username, email, full_name) VALUES ($1, $2, $3) RETURNING id, uuid, username, email, full_name`,
		user.Username, user.Email, user.FullName,
	).Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrUniqueConstraint
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(uuid string, user *model.User) (*model.User, error) {
	var u model.User
	err := r.db.QueryRowContext(
		context.Background(),
		`UPDATE users SET username = $1, email = $2, full_name = $3 WHERE uuid = $4 RETURNING id, uuid, username, email, full_name`,
		user.Username, user.Email, user.FullName, uuid,
	).Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if isUniqueConstraintError(err) {
			return nil, ErrUniqueConstraint
		}
		return nil, err
	}
	return &u, nil
}

func isUniqueConstraintError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

func (r *userRepository) Delete(uuid string) error {
	result, err := r.db.ExecContext(context.Background(), `DELETE FROM users WHERE uuid = $1`, uuid)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
